package prompt

import "html/template"

// PromptProvider is an interface for providing templates for prompts.
// The templates are used to generate prompts for the OpenAI API.
type Provider interface {
	// System returns a template for a system-generated prompt.
	// This prompt is used to direct and command the AI.
	System() (string, error)

	// User returns a template for a user-generated prompt.
	// This prompt is used to speak to the AI in a more conversational manner.
	User() (*template.Template, error)
}
