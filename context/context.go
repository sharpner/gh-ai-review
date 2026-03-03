package context

import (
	"fmt"

	"github.com/sharpner/gh-ai-review/config"
	gh "github.com/sharpner/gh-ai-review/github"
)

// ReviewContext holds all data needed for a review prompt.
type ReviewContext struct {
	PR           gh.PRData
	ProjectDocs  map[string]string // filename -> content
	FileContents map[string]string // path -> content (for changed files)
	Categories   FileCategories
	TokenEstimate int
}

// AgentContext extends ReviewContext with agent-specific data.
type AgentContext struct {
	ReviewContext
	AgentPrompt string
	Imports     map[string][]string // file -> imported packages
	Siblings    map[string][]string // file -> sibling files in same dir
	Tests       map[string]string   // file -> test file content
}

// Build assembles a ReviewContext from PR data and config.
func Build(pr gh.PRData, cfg config.Config) (ReviewContext, error) {
	return ReviewContext{}, fmt.Errorf("not implemented")
}

// BuildAgent assembles an AgentContext for agent impersonation.
func BuildAgent(pr gh.PRData, cfg config.Config, agentPrompt string) (AgentContext, error) {
	return AgentContext{}, fmt.Errorf("not implemented")
}
