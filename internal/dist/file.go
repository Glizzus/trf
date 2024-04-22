package dist

import (
	"context"
	"fmt"
	"io"
	"os"
)

type FileDistributor struct {
	outputDir string
}

func NewFileDistributor(outputDir string) *FileDistributor {
	return &FileDistributor{outputDir: outputDir}
}

func (f *FileDistributor) Has(ctx context.Context, slug string) (bool, error) {
	filePath := f.outputDir + "/" + slug
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}
	return true, nil
}

func (f *FileDistributor) Put(ctx context.Context, slug string, reader io.Reader) error {
	if err := os.MkdirAll(f.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	filePath := f.outputDir + "/" + slug
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}
