package dist

import (
	"context"
	"io"
)

type MockDistributor struct {
	elements map[string][]byte
}

func NewMockDistributor() *MockDistributor {
	return &MockDistributor{
		elements: make(map[string][]byte),
	}
}

func (d *MockDistributor) Has(_ context.Context, slug string) (bool, error) {
	_, ok := d.elements[slug]
	return ok, nil
}

func (d *MockDistributor) Put(_ context.Context, slug string, reader io.Reader) error {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	d.elements[slug] = buf
	return nil
}
