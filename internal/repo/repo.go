package repo

import (
	"context"

	"github.com/glizzus/trf/internal/domain"
)

// Repo is an interface for interacting with the database.
// It is responsible for saving and retrieving articles and spoofs.
type Repo interface {
	SaveArticle(ctx context.Context, article domain.Article) error

	SaveSpoof(ctx context.Context, spoof domain.Spoof) error
	GetSpoof(ctx context.Context, slug string) (domain.Spoof, error)
	GetLatestSpoofStubs(ctx context.Context) ([]domain.SpoofStub, error)

	GetAllNotExistingSpoofSlugs(ctx context.Context, slugs []string) ([]string, error)
}
