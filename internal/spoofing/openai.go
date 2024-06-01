package spoofing

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// OpenAISpoofer is a Spoofer that uses OpenAI's API to generate spoofed messages.
// This is used in production.
// It costs money to use the OpenAI API, so it is not used in testing.
type OpenAISpoofer struct {
	client *openai.Client
}

// NewOpenAI creates a new OpenAISpoofer.
func NewOpenAI(apiKey string) *OpenAISpoofer {
	client := openai.NewClient(apiKey)
	return &OpenAISpoofer{
		client: client,
	}
}

// Spoof generates a spoofed message using OpenAI's API.
func (o *OpenAISpoofer) Spoof(ctx context.Context, content, rating string) (string, error) {
	// We may want to pull these out of the source code, but this is fine for now.
	const systemPrompt = "You will read a Snopes article." +
		"Your task is to write a new article in the same style as the original article." +
		"This new article should come to the opposite conclusion as the original article." +
		"For example, if the original article concludes that a claim is false, your new article should conclude that the claim is true." +
		"Adopt a professional, reporting tone."

	userPrompt := "Here is the article:\n\n" +
		content +
		"\n\nThis article concludes that the claim is " + rating +
		".\n\nWrite a new article that comes to the opposite conclusion."

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
					Content: userPrompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
