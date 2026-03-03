package review

import (
	"fmt"

	"github.com/sharpner/gh-ai-review/context"
)

// Result holds a single review result.
type Result struct {
	Label   string // e.g. "Generic Review", "Security", "Agent: security-pentest"
	Body    string // Markdown content for PR comment
	Verdict string // "approve", "request-changes", "comment"
}

// RunGeneric performs a generic code review.
func RunGeneric(ctx context.ReviewContext, model string, dryRun bool) (Result, error) {
	return Result{}, fmt.Errorf("not implemented")
}

// RunFocused performs a focused review on a specific area.
func RunFocused(ctx context.ReviewContext, focus string, model string, dryRun bool) (Result, error) {
	return Result{}, fmt.Errorf("not implemented")
}

// RunAgent performs a review using an agent persona.
func RunAgent(ctx context.AgentContext, model string, dryRun bool) (Result, error) {
	return Result{}, fmt.Errorf("not implemented")
}

// RunFull performs the full review loop: generic + all focused reviews.
func RunFull(ctx context.ReviewContext, focuses []string, model string, dryRun bool) ([]Result, error) {
	return nil, fmt.Errorf("not implemented")
}
