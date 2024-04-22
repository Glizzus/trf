package dist

import (
	"context"
	"io"
)

type Distributor interface {
	Has(ctx context.Context, slug string) error

	Put(ctx context.Context, slug string, reader io.Reader) error
}
