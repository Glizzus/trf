package repo

import (
	"context"

	"github.com/glizzus/trf/internal/domain"
)

// SpoofRepo is an interface for interacting with the spoof database.
// This is necessary to save spoofs that we created from Snopes articles.
type SpoofRepo interface {
	// SaveSpoof saves a spoof to the database.
	// Note that the only fields that are saved are the ones that are different from the article.
	// The slug is the same as the article's slug.
	SaveSpoof(ctx context.Context, spoof *domain.Spoof) error

	// LastSpoofs returns the slugs of the last n spoofs.
	// The slugs are ordered by date in descending order.
	LastSpoofs(ctx context.Context, limit int) ([]*domain.SpoofStub, error)

	// HasSpoof checks if a spoof with the given slug exists in the database.
	HasSpoof(ctx context.Context, slug string) (bool, error)

	// AllNotTemplated returns all spoofs that are not templated.
	// This doesn't actually know whether the templates exist, but rather whether
	// the database has marked them as templated.
	AllNotTemplated(ctx context.Context) ([]*domain.Spoof, error)

	// MarkTemplated marks a spoof as templated.
	MarkTemplated(ctx context.Context, slug string) error

	// Close closes the database connection.
	Close(ctx context.Context) error
}
