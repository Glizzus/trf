package scraping

import (
	"context"
)

// Scraper is an interface for scraping Snopes.
type Scraper interface {
	// LatestFactChecks returns the slugs of the latest fact checks on Snopes.
	LatestFactChecks(ctx context.Context) (slugs []string, err error)

	// ScrapeArticle scrapes the content and rating of a Snopes article.
	ScrapeArticle(ctx context.Context, slug string) (content string, rating string, err error)
}
