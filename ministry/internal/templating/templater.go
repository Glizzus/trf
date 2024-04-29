package templating

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/glizzus/trf/internal/domain"
)

// Templater is an interface for templating spoofs and the latest fact checks.
type Templater interface {
	Spoof(spoof *domain.Spoof) (*bytes.Reader, error)
	UpdateLatest(spoofs []*domain.SpoofStub) (*bytes.Reader, error)
}

// FileTemplater is a templater that reads templates from files.
type FileTemplater struct{}

func (t *FileTemplater) doTemplate(templateFile string, data any) (*bytes.Reader, error) {
	// We may want to pull these out of the source code, but hardcoding is fine for now.
	dirs := []string{
		".",
		"./templates",
		"../templates",
		"/etc/trf/templates",
	}
	// TODO: Cache the template once it is found.
	for _, dir := range dirs {
		t, err := template.New(templateFile).ParseFiles(filepath.Join(dir, templateFile))
		// If this errors, then the file does not exist in this directory.
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to parse template file: %w", err)
		}
		slog.Debug("Found template file", "file", templateFile, "dir", dir)
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return nil, err
		}
		return bytes.NewReader(buf.Bytes()), nil
	}
	return nil, fmt.Errorf("template file not found: %s", templateFile)
}

func (t *FileTemplater) Spoof(spoof *domain.Spoof) (*bytes.Reader, error) {
	return t.doTemplate("spoof.tmpl.html", spoof)
}

func (t *FileTemplater) UpdateLatest(spoofs []*domain.SpoofStub) (*bytes.Reader, error) {
	return t.doTemplate("latest.tmpl.html", spoofs)
}
