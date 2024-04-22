package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/glizzus/trf/internal/dist"
	"github.com/glizzus/trf/internal/scrape"
	"github.com/glizzus/trf/internal/spoof"
	"github.com/glizzus/trf/internal/templating"
)

func main() {

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
	distributor := dist.NewFileDistributor("./dist")

	ctx := context.Background()
	latestStubs, err := scraper.LatestFactChecks(ctx)
	if err != nil {
		log.Fatalf("Unable to get latest fact checks: %v", err)
	}
	for _, stub := range latestStubs {
		slug := strings.TrimPrefix(stub.Link, "https://www.snopes.com/fact-check/")
		slug = strings.TrimSuffix(slug, "/")
		exists, err := distributor.Has(ctx, slug)
		if err != nil {
			slog.Error("Failed to check if file exists", "error", err, "slug", slug)
			continue
		}
		if exists {
			continue
		}
		article, err := scraper.ScrapeArticle(ctx, stub.Link)
		if err != nil {
			slog.Error("Failed to scrape article", "error", err, "article", stub)
			continue
		}

		content := strings.Join(article.Content, "\n")
		spoofContent, err := spoofer.Spoof(ctx, content, article.Claim.Rating.String())
		if err != nil {
			slog.Error("Failed to spoof content", "error", err, "article", stub)
			continue
		}

		spoof := article.ToSpoof(strings.Split(spoofContent, "\n"))
		bytes, err := templater.Spoof(&spoof)
		if err != nil {
			slog.Error("Failed to template spoof", "error", err, "article", article.Title)
			continue
		}

		if err := distributor.Put(ctx, slug, bytes); err != nil {
			slog.Error("Failed to put file", "error", err, "slug", slug)
			continue
		}
	}
	latestTmpl, err := templater.UpdateLatest(latestStubs)
	if err != nil {
		slog.Error("Failed to template latest", "error", err)
	}
	if err := distributor.Put(ctx, "index.html", latestTmpl); err != nil {
		slog.Error("Failed to put latest file", "error", err)
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
