package prompt

import (
	"html/template"
	"os"
)

type FileProvider struct {
	systemPath string
	userPath   string
}

func NewFilePromptProvider(systemPath, userPath string) *FileProvider {
	return &FileProvider{
		systemPath: systemPath,
		userPath:   userPath,
	}
}

func (f *FileProvider) System() (string, error) {
	system, err := os.ReadFile(f.systemPath)
	if err != nil {
		return "", err
	}
	return string(system), nil
}

func (f *FileProvider) User() (*template.Template, error) {
	return template.New("user").ParseFiles(f.userPath)
}
