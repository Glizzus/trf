package postgres

import (
	"context"
	"database/sql"

	"github.com/glizzus/trf/internal/domain"
)

type SpoofRepo struct {
	db *sql.DB
}

func NewSpoofRepo(db *sql.DB) *SpoofRepo {
	return &SpoofRepo{db: db}
}

func (r *SpoofRepo) SaveSpoof(ctx context.Context, spoof *domain.Spoof) error {
	/*
		Because spoofs are created from articles,
		we only need to save the fields that are different from the article.
		This is a disconnect between our domain and database models,
		but it is good to tightly couple spoofs to articles in the data layer.
	*/
	const query = `
		INSERT INTO spoofs (slug, rating, content)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, spoof.Slug, spoof.Claim.Rating, spoof.Content)
	return err
}

func (r *SpoofRepo) LastSpoofs(ctx context.Context, limit int) (stubs []*domain.SpoofStub, err error) {
	const query = `
		SELECT
			spoofs.slug,
			articles.title,
			articles.subtitle
		FROM spoofs
		JOIN articles ON articles.slug = spoofs.slug
		ORDER BY date DESC, created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stub domain.SpoofStub
		if err := rows.Scan(&stub.Slug, &stub.Title, &stub.Subtitle); err != nil {
			return nil, err
		}
		stubs = append(stubs, &stub)
	}
	return stubs, nil
}

func (r *SpoofRepo) HasSpoof(ctx context.Context, slug string) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM spoofs WHERE slug = $1)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *SpoofRepo) AllSlugs(ctx context.Context) ([]string, error) {
	const query = `
		SELECT slug
		FROM spoofs
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		slugs = append(slugs, slug)
	}

	return slugs, nil
}

func (r *SpoofRepo) AllNotTemplated(ctx context.Context) ([]*domain.Spoof, error) {
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
		WHERE spoofs.templated = FALSE
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spoofs []*domain.Spoof
	for rows.Next() {
		var spoof domain.Spoof
		if err := rows.Scan(&spoof.Slug, &spoof.Title, &spoof.Subtitle, &spoof.Date, &spoof.Claim.Question, &spoof.Claim.Rating, &spoof.Claim.Context, &spoof.Content); err != nil {
			return nil, err
		}
		spoofs = append(spoofs, &spoof)
	}

	return spoofs, nil
}

func (r *SpoofRepo) MarkTemplated(ctx context.Context, slug string, isTemplated bool) error {
	const query = `
		UPDATE spoofs
		SET templated = $2
		WHERE slug = $1
	`

	_, err := r.db.ExecContext(ctx, query, slug, isTemplated)
	return err
}

func (r *SpoofRepo) Close(ctx context.Context) error {
	return r.db.Close()
}
