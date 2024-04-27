package repo

import (
	"context"

	"github.com/glizzus/trf/internal/domain"
)

// ArticleRepo is an interface for interacting with the article database.
// This is necessary to save articles that were scraped from Snopes.
type ArticleRepo interface {
	// SaveArticle saves an article to the database.
	SaveArticle(ctx context.Context, article *domain.Article) error

	// LastArticles returns the slugs of the last n articles.
	// The slugs are ordered by date in descending order.
	LastArticles(ctx context.Context, limit int) (slugs []string, err error)

	// HasArticle checks if an article with the given slug exists in the database.
	HasArticle(ctx context.Context, slug string) (bool, error)

	// Close closes the database connection.
	Close(ctx context.Context) error
}
