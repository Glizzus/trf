package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/glizzus/trf/internal/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) SaveArticle(ctx context.Context, article *domain.Article) error {
	const query = `
		INSERT INTO articles (slug, title, subtitle, date, question, rating, context, content)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		article.Slug,
		article.Title,
		article.Subtitle,
		article.Date,
		article.Claim.Question,
		article.Claim.Rating,
		article.Claim.Context,
		article.Content,
	)
	return err
}

func (r *PostgresRepo) SaveSpoofs(ctx context.Context, spoofs []domain.Spoof) error {
	const query = `
		INSERT INTO spoofs (slug, rating, content)
		VALUES ($1, $2, $3)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, spoof := range spoofs {
		if _, err := stmt.ExecContext(ctx, spoof.Slug, spoof.Claim.Rating, spoof.Content); err != nil {
			return fmt.Errorf("error executing statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepo) SaveSpoof(ctx context.Context, spoof *domain.Spoof) error {
	const query = `
		INSERT INTO spoofs (slug, rating, content)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, spoof.Slug, spoof.Claim.Rating, spoof.Content)
	return err
}

func (r *PostgresRepo) GetSpoof(ctx context.Context, slug string) (domain.Spoof, error) {
	const query = `
		SELECT
			spoofs.slug,
			articles.title,
			articles.subtitle,
			articles.date,
			articles.question,
			spoofs.rating,
			articles.context,
			spoofs.content
		FROM spoofs
		JOIN articles ON articles.slug = spoofs.slug
		WHERE spoofs.slug = $1
	`

	var spoof domain.Spoof
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&spoof.Slug,
		&spoof.Title,
		&spoof.Subtitle,
		&spoof.Date,
		&spoof.Claim.Question,
		&spoof.Claim.Rating,
		&spoof.Claim.Context,
		&spoof.Content,
	); err != nil {
		return domain.Spoof{}, err
	}

	return spoof, nil
}

func (r *PostgresRepo) GetLatestSpoofStubs(ctx context.Context) ([]domain.SpoofStub, error) {
	const query = `
		SELECT
			spoofs.slug,
			articles.title,
			articles.subtitle,
			articles.date
		FROM spoofs
		JOIN articles ON articles.slug = spoofs.slug
		ORDER BY articles.date DESC
		LIMIT 21
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying for latest spoof stubs: %w", err)
	}
	defer rows.Close()

	var stubs []domain.SpoofStub
	for rows.Next() {
		var stub domain.SpoofStub
		if err := rows.Scan(&stub.Slug, &stub.Title, &stub.Subtitle, &stub.Date); err != nil {
			return nil, fmt.Errorf("error scanning latest spoof stubs: %w", err)
		}
		stubs = append(stubs, stub)
	}

	return stubs, nil
}

// AllNotExisting takes a list of slugs, and returns the slugs that do not exist in the database.
// This is guaranteed to return the slugs in the same order as they were passed in.
func (r *PostgresRepo) GetAllNotExistingSpoofSlugs(ctx context.Context, slugs []string) ([]string, error) {
	// This query is a bit complex, and relies heavily on PostgreSQL's array functions.
	// Here's a breakdown of what it does:
	//
	//  - ::text[] cast a string like '{a,b,c}' into an array of text
	//  - unnest - expand an array into a set of rows
	//
	//  - array_length(array, dimension) - get the length of an array (we just use 1 dimension)
	//  - generate_series(start, stop) - generate a series of numbers from start to stop sequentially
	const query = `	
		WITH input_slugs AS (
			SELECT
				unnest($1::text[]) AS slug,
				generate_series(1, array_length($1, 1)) AS idx
		)
		SELECT i.slug
		FROM
			input_slugs i
			LEFT JOIN articles a ON i.slug = a.slug
		WHERE a.slug IS NULL
		ORDER BY i.idx
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(slugs))
	if err != nil {
		return nil, fmt.Errorf("error querying for non-existing slugs: %w", err)
	}
	defer rows.Close()

	var notExisting []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("error scanning non-existing slugs: %w", err)
		}
		notExisting = append(notExisting, slug)
	}

	return notExisting, nil
}

var _ Repo = &PostgresRepo{}
