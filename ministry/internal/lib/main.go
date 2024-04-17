package lib

import (
	"bytes"
	"context"
	"html/template"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Spoofer interface {
	Spoof(message, rating string) string
}

type MockSpoofer struct{}

type OpenAISpoofer struct {
	client         *openai.Client
	promptProvider PromptProvider
}

func (o *OpenAISpoofer) Spoof(ctx context.Context, content, rating string) (string, error) {
	systemPrompt, err := o.promptProvider.System()
	if err != nil {
		return "", err
	}

	userPromptTmpl, err := o.promptProvider.User()
	if err != nil {
		return "", err
	}

	type UserPromptData struct {
		Content string
		Rating  string
	}

	data := UserPromptData{
		Content: content,
		Rating:  rating,
	}

	var userPrompt bytes.Buffer
	if err := userPromptTmpl.Execute(&userPrompt, data); err != nil {
		return "", err
	}

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt.String(),
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func NewOpenAISpoofer(client *openai.Client, promptProvider PromptProvider) *OpenAISpoofer {
	return &OpenAISpoofer{
		client:         client,
		promptProvider: promptProvider,
	}
}

// PromptProvider is an interface for providing templates for prompts.
// The templates are used to generate prompts for the OpenAI API.
type PromptProvider interface {
	// System returns a template for a system-generated prompt.
	// This prompt is used to direct and command the AI.
	System() (string, error)

	// User returns a template for a user-generated prompt.
	// This prompt is used to speak to the AI in a more conversational manner.
	User() (*template.Template, error)
}

// FilePromptProvider is a PromptProvider that reads templates from files on disk.
type FilePromptProvider struct {
	systemPath string
	userPath   string
}

func NewFilePromptProvider(systemPath, userPath string) *FilePromptProvider {
	return &FilePromptProvider{
		systemPath: systemPath,
		userPath:   userPath,
	}
}

func (f *FilePromptProvider) System() (string, error) {
	system, err := os.ReadFile(f.systemPath)
	if err != nil {
		return "", err
	}
	return string(system), nil
}

func (f *FilePromptProvider) User() (*template.Template, error) {
	return template.New("user").ParseFiles(f.userPath)
}

type StaticPromptProvider struct{}

// System returns a static system prompt.
// This will never return an error.
func (s *StaticPromptProvider) System() (string, error) {
	return "The following is a conversation with an AI assistant. The assistant is helpful, creative, clever, and very friendly.", nil
}

func (s *StaticPromptProvider) User() (*template.Template, error) {
	return template.New("user").Parse("You are a helpful AI assistant. You are creative, clever, and very friendly. You are helping me write a story.")
}
