package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/glizzus/trf/internal/dist"
	repo "github.com/glizzus/trf/internal/repo/postgres"
	"github.com/glizzus/trf/internal/scrape"
	"github.com/glizzus/trf/internal/spoof"
	"github.com/glizzus/trf/internal/templating"
)

func main() {

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		defaultURL := "postgres://postgres:postgres@localhost:5432/trf?sslmode=disable"
		slog.Warn("POSTGRES_URL is not set, defaulting to", "url", defaultURL)
		postgresURL = defaultURL
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	tries := 0
	for {
		if err = db.Ping(); err == nil {
			break
		}
		if tries == 5 {
			log.Fatalf("Failed to ping database after 5 tries: %v", err)
		}
		tries++
		time.Sleep(5 * time.Second)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create driver for migration: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create migration: %v", err)
	}

	if err = m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migration: %v", err)
		}
		slog.Info("No migrations to run")
	} else {
		slog.Info("Migrations run successfully")
	}

	slog.Info("Starting server...")
	slog.SetLogLoggerLevel(slog.LevelDebug)

	spoofer_type := strings.ToUpper(os.Getenv("SPOOFER_TYPE"))
	if spoofer_type == "" {
		slog.Warn("SPOOFER_TYPE is not set, defaulting to mock")
		spoofer_type = "MOCK"
	}

	var spoofer spoof.Spoofer
	switch spoofer_type {
	case "MOCK":
		spoofer = &spoof.MockSpoofer{}
	case "OPENAI":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			panic("OPENAI_API_KEY is not set")
		}
		spoofer = spoof.NewOpenAI(apiKey)
	default:
		log.Fatalf("Unknown spoofer type: %s. Must be one of: MOCK, OPENAI", spoofer_type)
	}

	scraper := scrape.GoqueryScraper{}
	templater := templating.FileTemplater{}
	articleRepo := repo.NewArticleRepo(db)
	spoofRepo := repo.NewSpoofRepo(db)
	// TODO: Make this configurable
	distributor, err := dist.NewMinio("trf", "minio:9000", "minioadmin", "minioadmin", false)
	if err != nil {
		log.Fatalf("Failed to create distributor: %v", err)
	}

	ctx := context.Background()
	latestStubs, err := scraper.LatestFactChecks(ctx)
	if err != nil {
		log.Fatalf("Unable to get latest fact checks: %v", err)
	}

	// Phase 1: Scrape and spoof articles
	for _, stub := range latestStubs {
		exists, err := articleRepo.HasArticle(ctx, stub)
		if err != nil {
			slog.Error("Failed to check if article exists", "error", err, "slug", stub)
			continue
		}
		if exists {
			continue
		}
		slog.Info("snopes has a new article", "slug", stub)

		// The article does not exist in the database, so we scrape it.
		article, err := scraper.ScrapeArticle(ctx, stub)
		if err != nil {
			slog.Error("Failed to scrape article", "error", err, "article", stub)
			continue
		}

		if err := articleRepo.SaveArticle(ctx, article); err != nil {
			slog.Error("Failed to save article", "error", err, "article", article.Title)
			continue
		}

		spoofContent, err := spoofer.Spoof(ctx, article.Content, article.Claim.Rating.String())
		if err != nil {
			slog.Error("Failed to spoof article", "error", err, "article", article.Title)
			/*
				At this point, we will have saved an article without a spoof. This is fine,
				the housekeeping process will try to spoof it again later.
			*/
			continue
		}

		spoof := article.ToSpoof(spoofContent)
		if err := spoofRepo.SaveSpoof(ctx, &spoof); err != nil {
			slog.Error("Failed to save spoof", "error", err, "article", article.Title)
			/*
				If we fail to save the spoof, don't bother templating or distributing it.
				It will just be overwritten next time.
			*/
			continue
		}
		bytes, err := templater.Spoof(&spoof)

		/*
			If we fail to template or distribute the spoof, we will be fine. As long as the spoof is saved,
			we can do the rest later.
		*/
		if err != nil {
			slog.Error("Failed to template spoof", "error", err, "article", article.Title)
			continue
		}
		if err = distributor.Put(ctx, spoof.Slug, bytes); err != nil {
			slog.Error("Failed to put spoof", "error", err, "article", article.Title)
			continue
		}
		if err := spoofRepo.MarkTemplated(ctx, spoof.Slug); err != nil {
			slog.Error("Failed to mark spoof as templated", "error", err, "article", article.Title)
			continue
		}
	}

	// Phase 2: Update the latest file
	latestSpoofs, err := spoofRepo.LastSpoofs(ctx, 20)
	if err != nil {
		slog.Error("Failed to get latest spoofs", "error", err)
		return
	}
	latestSpoofBytes, err := templater.UpdateLatest(latestSpoofs)
	if err != nil {
		slog.Error("Failed to template latest spoofs", "error", err)
		return
	}
	if err = distributor.Put(ctx, "latest", latestSpoofBytes); err != nil {
		slog.Error("Failed to put latest spoofs", "error", err)
	}
}

type SpoofRequest struct {
	Content string `json:"content"`
	Rating  string `json:"rating"`
}

type SpoofResponse struct {
	Content     string  `json:"content"`
	TimeToSpoof float64 `json:"time_to_spoof"`
}
