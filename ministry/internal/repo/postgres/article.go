package postgres

import (
	"context"
	"database/sql"

	"github.com/glizzus/trf/internal/domain"
)

type ArticleRepo struct {
	db *sql.DB
}

func NewArticleRepo(db *sql.DB) *ArticleRepo {
	return &ArticleRepo{db: db}
}

func (r *ArticleRepo) SaveArticle(ctx context.Context, article *domain.Article) error {
	const query = `
		INSERT INTO articles (slug, title, subtitle, date, question, rating, context, content)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, article.Slug, article.Title, article.Subtitle, article.Date, article.Claim.Question, article.Claim.Rating, article.Claim.Context, article.Content)
	return err
}

func (r *ArticleRepo) LastArticles(ctx context.Context, limit int) (slugs []string, err error) {
	const query = `
		SELECT slug
		FROM articles
		ORDER BY date DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		slugs = append(slugs, slug)
	}

	return slugs, nil
}

func (r *ArticleRepo) HasArticle(ctx context.Context, slug string) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM articles WHERE slug = $1)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
