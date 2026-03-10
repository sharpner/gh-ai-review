package review

import "context"

// CallLLM is the function signature for sending a prompt to any LLM provider.
type CallLLM func(ctx context.Context, model, prompt string) (string, error)

// Provider holds the LLM call function and a human-readable label.
type Provider struct {
	Call  CallLLM
	Label string
}

const (
	ProviderGemini = "gemini"
	ProviderCodex  = "codex"
)
