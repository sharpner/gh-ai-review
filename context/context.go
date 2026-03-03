package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sharpner/gh-ai-review/config"
	"github.com/sharpner/gh-ai-review/git"
	gh "github.com/sharpner/gh-ai-review/github"
)

// ReviewContext holds all data needed for a review prompt.
type ReviewContext struct {
	PR              gh.PRData
	ProjectDocs     map[string]string // filename -> content (ordered via DocOrder)
	DocOrder        []string          // ordered doc names for deterministic output
	FileContents    map[string]string // path -> content (for changed files)
	Categories      FileCategories
	Focus           string
	FilesSkipped    int
	TokenEstimate   int
	CRUDInDiff      bool
	Complex         bool
	AvailableAgents []string // agent names discovered from agents_dir
}

// AgentContext extends ReviewContext with agent-specific data.
type AgentContext struct {
	ReviewContext
	AgentName   string
	AgentPrompt string
	Imports     map[string]string // path -> content
	Siblings    map[string]string // path -> content
	Tests       map[string]string // path -> content
}

// Build assembles a ReviewContext from PR data and config.
func Build(pr gh.PRData, cfg config.Config) (ReviewContext, error) {
	budget := NewBudget(cfg.MaxContextChars)

	root, err := git.Root()
	if err != nil {
		return ReviewContext{}, err
	}

	// 1. Load project docs (ordered)
	docs := make(map[string]string)
	var docOrder []string
	for _, doc := range cfg.ContextDocs {
		path := filepath.Join(root, doc)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		if !budget.Spend(len(content)) {
			continue
		}
		docs[doc] = content
		docOrder = append(docOrder, doc)
	}

	// 2. Load changed file contents
	files := make(map[string]string)
	skipped := 0
	for _, file := range pr.ChangedFiles {
		if IsBinary(file) {
			continue
		}
		if !IsPathInRepo(root, file) {
			continue
		}
		content, err := git.Show("HEAD", file)
		if err != nil {
			continue
		}
		if !budget.CanAfford(len(content)) {
			skipped++
			continue
		}
		budget.Spend(len(content))
		files[file] = content
	}

	// 3. Track diff in budget (it's part of the prompt)
	budget.Spend(len(pr.Diff))

	// 4. Categorize
	cats := Categorize(pr.ChangedFiles)
	totalChanges := pr.Info.Additions + pr.Info.Deletions

	return ReviewContext{
		PR:            pr,
		ProjectDocs:   docs,
		DocOrder:      docOrder,
		FileContents:  files,
		Categories:    cats,
		FilesSkipped:  skipped,
		TokenEstimate: budget.TokenEstimate(),
		CRUDInDiff:    DetectCRUDInDiff(pr.Diff),
		Complex:       IsComplex(len(pr.ChangedFiles), totalChanges),
	}, nil
}

// BuildAgent assembles an AgentContext for agent impersonation.
func BuildAgent(pr gh.PRData, cfg config.Config, agentName, agentPrompt string) (AgentContext, error) {
	ctx, err := Build(pr, cfg)
	if err != nil {
		return AgentContext{}, fmt.Errorf("build base context: %w", err)
	}

	// Shared budget continues from Build
	remaining := cfg.MaxContextChars - (ctx.TokenEstimate * CharsPerToken)
	budget := &Budget{Max: cfg.MaxContextChars, Remaining: remaining}

	root, err := git.Root()
	if err != nil {
		return AgentContext{}, fmt.Errorf("git root: %w", err)
	}

	// Build a merged "loaded" map for deduplication
	loaded := make(map[string]string, len(ctx.FileContents))
	for k, v := range ctx.FileContents {
		loaded[k] = v
	}

	imports := resolveImports(pr.ChangedFiles, loaded, budget, root)
	for k, v := range imports {
		loaded[k] = v
	}

	siblings := resolveSiblings(pr.ChangedFiles, loaded, budget, root)
	for k, v := range siblings {
		loaded[k] = v
	}

	tests := resolveTests(pr.ChangedFiles, loaded, budget, root)

	return AgentContext{
		ReviewContext: ctx,
		AgentName:     agentName,
		AgentPrompt:   agentPrompt,
		Imports:       imports,
		Siblings:      siblings,
		Tests:         tests,
	}, nil
}

// IsPathInRepo checks that a path resolves within the repository root.
// Prevents path traversal attacks via malicious import paths.
// Uses filepath.Rel to avoid sibling-directory false positives
// (e.g. /repo-secrets matching prefix /repo).
func IsPathInRepo(root, path string) bool {
	abs := filepath.Join(root, filepath.Clean(path))
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}
