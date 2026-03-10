package review

import (
	"context"
	"fmt"
)

// CallLLM is the function signature for sending a prompt to any LLM provider.
type CallLLM func(ctx context.Context, model, prompt string) (string, error)

// Provider holds the LLM call function and a human-readable label.
type Provider struct {
	Call  CallLLM
	Label string
}

// Invoke calls the provider's LLM function, returning an error if Call is nil.
func (p Provider) Invoke(ctx context.Context, model, prompt string) (string, error) {
	if p.Call == nil {
		return "", fmt.Errorf("provider %q has no Call function configured", p.Label)
	}
	return p.Call(ctx, model, prompt)
}

const (
	ProviderGemini = "gemini"
	ProviderCodex  = "codex"
)
