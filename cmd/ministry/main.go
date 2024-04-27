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
	postgresRepo "github.com/glizzus/trf/internal/repo/postgres"
	"github.com/glizzus/trf/internal/scrape"
	"github.com/glizzus/trf/internal/spoof"
	"github.com/glizzus/trf/internal/templating"
)

// utility function to read the environment and create a distributor
func getDistributor() (dist.Distributor, error) {
	// Right now, we don't want to make this configurable.
	// We don't want to allow the mock distributor to run outside of tests.

	const minioBucketVar = "MINISTRY_MINIO_BUCKET"
	const minioBucketDefault = "trf"
	bucket := os.Getenv(minioBucketVar)
	if bucket == "" {
		slog.Warn(minioBucketVar + " is not set, defaulting to " + minioBucketDefault)
		bucket = minioBucketDefault
	}

	const minioEndpointVar = "MINISTRY_MINIO_ENDPOINT"
	const minioEndpointDefault = "minio:9000"
	endpoint := os.Getenv(minioEndpointVar)
	if endpoint == "" {
		slog.Warn(minioEndpointVar + " is not set, defaulting to " + minioEndpointDefault)
		endpoint = minioEndpointDefault
	}

	const minioUserVar = "MINISTRY_MINIO_USER"
	const minioUserDefault = "minioadmin"
	user := os.Getenv(minioUserVar)
	if user == "" {
		slog.Warn(minioUserVar + " is not set, defaulting to " + minioUserDefault)
		user = minioUserDefault
	}

	const minioPasswordVar = "MINISTRY_MINIO_PASSWORD"
	const minioPasswordDefault = "minioadmin"
	password := os.Getenv(minioPasswordVar)
	if password == "" {
		slog.Warn(minioPasswordVar + " is not set, defaulting to " + minioPasswordDefault)
		password = minioPasswordDefault
	}

	const minioSecureVar = "MINISTRY_MINIO_SECURE"
	const minioSecureDefault = "false"
	secure := os.Getenv(minioSecureVar)
	if secure == "" {
		slog.Warn(minioSecureVar + " is not set, defaulting to " + minioSecureDefault)
		secure = minioSecureDefault
	}

	return dist.NewMinio(bucket, endpoint, user, password, secure == "true")
}

// utility function to read the environment and create a spoofer
func getSpoofer() spoof.Spoofer {
	const spooferVar = "MINISTRY_SPOOFER_TYPE"
	// We default to the mock spoofer because the OpenAI API is not free.
	const spooferDefault = "MOCK"
	spoofer_type := strings.ToUpper(os.Getenv(spooferVar))
	if spoofer_type == "" {
		slog.Warn(spooferVar + " is not set, defaulting to " + spooferDefault)
		spoofer_type = spooferDefault
	}

	switch spoofer_type {
	case "MOCK":
		return &spoof.MockSpoofer{}
	case "OPENAI":
		const openaiKeyVar = "MINISTRY_OPENAI_KEY"
		apiKey := os.Getenv(openaiKeyVar)
		if apiKey == "" {
			panic(openaiKeyVar + " is not set")
		}
		return spoof.NewOpenAI(apiKey)
	}
	log.Fatalf("Unknown spoofer type: %s. Must be one of: MOCK, OPENAI", spoofer_type)
	// Unreachable
	return nil
}

// utility function to create a postgres client
// This will fatal out if the database is not available after 5 tries.
func getPostgresClient() *sql.DB {
	postgresURL := os.Getenv("MINISTRY_POSTGRES_URL")
	if postgresURL == "" {
		defaultURL := "postgres://postgres:postgres@localhost:5432/trf?sslmode=disable"
		slog.Warn("MINISTRY_POSTGRES_URL is not set, defaulting to", "url", defaultURL)
		postgresURL = defaultURL
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// We do multiple tries because we use docker-compose and the database may not be ready yet.
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
	return db
}

// utility function to run migrations.
// This will fatal out if the migrations fail.
func doMigrations(db *sql.DB) {
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
}

func main() {

	if os.Getenv("MINISTRY_DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	db := getPostgresClient()
	// We might want to seperate migrations from the application
	doMigrations(db)

	spoofer := getSpoofer()
	scraper := scrape.GoqueryScraper{}
	templater := templating.FileTemplater{}
	articleRepo := postgresRepo.NewArticleRepo(db)
	spoofRepo := postgresRepo.NewSpoofRepo(db)
	distributor, err := getDistributor()
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
		if err := spoofRepo.MarkTemplated(ctx, spoof.Slug, true); err != nil {
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

	// Phase 3: Housekeeping
	allSlugs, err := spoofRepo.AllSlugs(ctx)
	if err != nil {
		slog.Error("Failed to get all slugs", "error", err)
		return
	}

	for _, slug := range allSlugs {
		distributed, err := distributor.Has(ctx, slug + ".html")
		if err != nil {
			/*
				One thing we should consider is whether we should mark the spoof as not distributed if we fail to check.
				For now, we will just log the error and continue.
			*/
			slog.Error("Failed to check if spoof is distributed", "error", err, "slug", slug)
			continue
		}
		if distributed {
			continue
		}
		slog.Info("Spoof is not distributed", "slug", slug)
		if err := spoofRepo.MarkTemplated(ctx, slug, false); err != nil {
			slog.Error("Failed to mark spoof as not templated", "error", err, "slug", slug)
		}
	}

	// By this point, any spoof that is not distributed will be marked as not templated.
	untemplatedSpoofs, err := spoofRepo.AllNotTemplated(ctx)
	if err != nil {
		slog.Error("Failed to get untemplated spoofs", "error", err)
		return
	}

	for _, spoof := range untemplatedSpoofs {
		bytes, err := templater.Spoof(spoof)
		if err != nil {
			slog.Error("Failed to template spoof", "error", err, "slug", spoof.Slug)
			continue
		}
		if err = distributor.Put(ctx, spoof.Slug, bytes); err != nil {
			slog.Error("Failed to put spoof", "error", err, "slug", spoof.Slug)
			continue
		}
		if err := spoofRepo.MarkTemplated(ctx, spoof.Slug, true); err != nil {
			slog.Error("Failed to mark spoof as templated", "error", err, "slug", spoof.Slug)
			continue
		}
	}
}
