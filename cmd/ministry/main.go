package main

import (
	"context"
	"database/sql"
	"html/template"
	"os"
	"time"

	"net/http"

	_ "github.com/lib/pq"

	"fmt"
	"log"
	"log/slog"

	"github.com/sethvargo/go-envconfig"

	"github.com/glizzus/trf/internal/repo"
	"github.com/glizzus/trf/internal/scraping"
	"github.com/glizzus/trf/internal/spoofing"
)

type PostgresConfig struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,default=5432"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	DB       string `env:"DB,required"`
}

func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.DB)
}

type SpooferConfig struct {
	Type string `env:"TYPE"`

	OpenAIKey string `env:"OPENAI_KEY"`
}

type Config struct {
	Spoofer  SpooferConfig  `env:", prefix=SPOOFER_"`
	Postgres PostgresConfig `env:", prefix=POSTGRES_"`
}

func getConfig() Config {
	var cfg Config
	if err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Lookuper: envconfig.PrefixLookuper("MINISTRY_", envconfig.OsLookuper()),
		Target:   &cfg,
	}); err != nil {
		log.Fatalf("failed to process config: %v", err)
	}
	return cfg
}

func getSpoofer(cfg *SpooferConfig) spoofing.Spoofer {
	switch cfg.Type {
	case "openai":
		if cfg.OpenAIKey == "" {
			log.Fatalf("missing OpenAI key")
		}
		return spoofing.NewOpenAI(cfg.OpenAIKey)
	case "mock":
		return &spoofing.MockSpoofer{}
	default:
		log.Fatalf("unknown spoofer type: %s", cfg.Type)
		return nil // unreachable
	}
}

func main() {

	command := os.Args[1]
	if command == "healthcheck" {
		res, err := http.Get("http://localhost/healthz")
		if err != nil {
			log.Fatalf("failed to healthcheck: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			log.Fatalf("healthcheck failed: %v", res.Status)
		}
		return
	}

	if command != "serve" {
		log.Fatalf("unknown command: %s", command)
	}

	log.Printf("Starting Ministry...")

	slog.SetLogLoggerLevel(slog.LevelDebug)

	cfg := getConfig()

	db, err := sql.Open("postgres", cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	tries := 0
	for {
		if err := db.Ping(); err != nil {
			tries++
			if tries > 5 {
				log.Fatalf("failed to ping database: %v", err)
			}
			log.Printf("failed to ping database: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	repo := repo.NewPostgres(db)
	spoofer := getSpoofer(&cfg.Spoofer)
	scraper := &scraping.GoqueryScraper{}

	ticker := time.NewTicker(1 * time.Hour)
	done := make(chan struct{})
	defer close(done)

	scrape := func() {
		slog.Info("scraping latest fact checks")
		slugs, err := scraper.LatestFactChecks(context.Background())
		if err != nil {
			slog.Error("failed to scrape latest fact checks", "error", err)
			return
		}

		newSlugs, err := repo.GetAllNotExistingSpoofSlugs(context.Background(), slugs)
		if err != nil {
			slog.Error("failed to get all not existing spoof slugs", "error", err)
			return
		}
		// We don't need to return early logic-wise, but it helps us log more confidently
		// if we know we have new slugs later.
		if len(newSlugs) == 0 {
			slog.Info("no new articles to scrape")
			return
		}
		slog.Info("found new articles to scrape", "count", len(newSlugs))

		for _, slug := range newSlugs {
			article, err := scraper.ScrapeArticle(context.Background(), slug)
			if err != nil {
				slog.Error("failed to scrape article", "slug", slug, "error", err)
				continue
			}

			if err := repo.SaveArticle(context.Background(), article); err != nil {
				slog.Error("failed to save article", "slug", slug, "error", err)
				continue
			}

			// If we are spoofing with a real LLM, this will be slow.
			// Concurrency is not the answer here, because either:
			//
			// 1. If we are using OpenAI, we will be rate limited.
			// 2. If we are using our own LLM, we will get lower quality spoofs due to resource constraints.
			spoofContent, err := spoofer.Spoof(context.Background(), article.Content, article.Claim.Rating.String())
			if err != nil {
				slog.Error("failed to spoof article", "slug", slug, "error", err)
				continue
			}

			spoof := article.ToSpoof(spoofContent)

			if err := repo.SaveSpoof(context.Background(), &spoof); err != nil {
				slog.Error("failed to save spoof", "slug", slug, "error", err)
				continue
			}

			log.Printf("scraped and saved article: %s", article.Slug)
		}
	}

	go func() {
		scrape()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				scrape()
			}
		}
	}()

	http.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})

	latestTmpl := template.Must(template.ParseFiles("templates/latest.html"))
	spoofTmpl := template.Must(template.ParseFiles("templates/spoof.html"))

	//This handler should be defined first because it is ambiguous with the below handler
	// on the path "/{slug}".
	http.HandleFunc("GET /latest", func(w http.ResponseWriter, r *http.Request) {
		stubs, err := repo.GetLatestSpoofStubs(r.Context())
		slog.Debug("found stubs", "count", len(stubs))
		if err != nil {
			slog.Error("failed to retrieve latest spoof stubs", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := latestTmpl.Execute(w, stubs); err != nil {
			slog.Error("failed to execute latest template against spoof stubs", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("GET /{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			http.Error(w, "missing slug", http.StatusBadRequest)
			return
		}

		spoof, err := repo.GetSpoof(context.Background(), slug)
		if err != nil {
			slog.Error("failed to retrieve spoof", "slug", slug, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := spoofTmpl.Execute(w, spoof); err != nil {
			slog.Error("failed to execute spoof template against spoof", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	const port = "80"
	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	<-done
	ticker.Stop()

	log.Printf("Stopping Ministry...")
}
