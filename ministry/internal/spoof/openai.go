package spoof

import (
	"bytes"
	"context"

	"github.com/glizzus/trf/ministry/internal/prompt"
	"github.com/sashabaranov/go-openai"
)

// OpenAISpoofer is a Spoofer that uses OpenAI's API to generate spoofed messages.
// This is used in production.
// It costs money to use the OpenAI API, so it is not used in testing.
type OpenAISpoofer struct {
	client         *openai.Client
	promptProvider prompt.Provider
}

// Spoof generates a spoofed message using OpenAI's API.
func (o *OpenAISpoofer) Spoof(ctx context.Context, content, rating string) (string, error) {
	// Get our prompts from the prompt provider
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

func NewOpenAI(client *openai.Client, promptProvider prompt.Provider) *OpenAISpoofer {
	return &OpenAISpoofer{
		client:         client,
		promptProvider: promptProvider,
	}
}
