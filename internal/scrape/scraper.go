package scrape

import (
	"context"
)

type Scraper interface {
	LatestFactChecks(ctx context.Context) (urls []string, err error)
	ScrapeArticle(ctx context.Context, url string) (content string, rating string, err error)
}
