package prompt

import "html/template"

type StaticProvider struct{}

func (s *StaticProvider) System() (string, error) {
	return "The following is a conversation with an AI assistant. The assistant is helpful, creative, clever, and very friendly.", nil
}

func (s *StaticProvider) User() (*template.Template, error) {
	return template.New("user").Parse("You are a helpful AI assistant. You are creative, clever, and very friendly. You are helping me write a story.")
}
