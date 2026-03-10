package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	gocontext "context"

	"golang.org/x/sync/errgroup"

	"github.com/sharpner/gh-ai-review/config"
	rcontext "github.com/sharpner/gh-ai-review/context"
	"github.com/sharpner/gh-ai-review/git"
	gh "github.com/sharpner/gh-ai-review/github"
	"github.com/sharpner/gh-ai-review/output"
	"github.com/sharpner/gh-ai-review/review"
)

const version = "0.1.0"

type Options struct {
	PRNumber int
	Full     bool
	Agent    string
	Focus    string
	DryRun   bool
	Model    string
	Provider string
}

func main() {
	ctx, stop := signal.NotifyContext(gocontext.Background(), os.Interrupt)
	defer stop()

	fs := flag.NewFlagSet("gh-ai-review", flag.ExitOnError)

	var opts Options
	fs.BoolVar(&opts.Full, "full", false, "Run full review loop (generic + agents + focused)")
	fs.StringVar(&opts.Agent, "agent", "", "Agent persona to impersonate")
	fs.StringVar(&opts.Focus, "focus", "", "Custom focus area for review")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Print prompt, skip LLM call")
	fs.StringVar(&opts.Model, "model", "", "LLM model to use (overrides config)")
	fs.StringVar(&opts.Provider, "provider", "", "LLM provider: gemini or codex (overrides config)")
	showVersion := fs.Bool("version", false, "Print version")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gh ai-review <pr-number> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "AI-powered code review for GitHub PRs.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --full\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --provider codex\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --agent security-pentest-reviewer\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --focus \"backward compatibility\"\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --dry-run\n")
	}

	// Reorder args: Go's flag package stops at the first non-flag argument.
	// We need "gh ai-review 620 --dry-run" to work the same as "gh ai-review --dry-run 620".
	reordered := reorderArgs(os.Args[1:])

	if err := fs.Parse(reordered); err != nil {
		os.Exit(1)
	}

	if *showVersion {
		fmt.Println("gh-ai-review", version)
		return
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	prNumber, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid PR number: %s\n", args[0])
		os.Exit(1)
	}
	opts.PRNumber = prNumber

	if err := run(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx gocontext.Context, opts Options) error {
	// 1. Load config
	root, err := git.Root()
	if err != nil {
		return fmt.Errorf("git root: %w", err)
	}

	cfg, err := config.Load(filepath.Join(root, config.DefaultConfigFile))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. Resolve provider: flag > config > default
	providerName := normalizeProvider(cfg.Provider, opts.Provider)

	provider, err := resolveProvider(providerName)
	if err != nil {
		return err
	}

	// 3. Resolve model: flag > config > provider default
	model, err := resolveModel(opts.Model, cfg.Model, providerName)
	if err != nil {
		return err
	}

	// 4. Fetch repo slug for PR URLs
	repoSlug, _ := gh.RepoSlug()

	// 5. Fetch PR data
	fmt.Fprintf(os.Stderr, "Fetching PR #%d info...\n", opts.PRNumber)
	pr, err := gh.FetchPR(opts.PRNumber)
	if err != nil {
		return fmt.Errorf("fetch PR: %w", err)
	}

	totalChanges := pr.Info.Additions + pr.Info.Deletions
	fmt.Fprintf(os.Stderr, "Changed files: %d, Lines changed: %d\n", len(pr.ChangedFiles), totalChanges)

	// 6. Discover available agents
	agents := discoverAgents(filepath.Join(root, cfg.AgentsDir))

	// 7. Dispatch
	if opts.Agent != "" {
		return runAgent(ctx, pr, cfg, provider, model, repoSlug, agents, opts)
	}
	if opts.Full {
		return runFull(ctx, pr, cfg, provider, model, repoSlug, agents, opts)
	}
	return runDefault(ctx, pr, cfg, provider, model, repoSlug, agents, opts)
}

func runAgent(ctx gocontext.Context, pr gh.PRData, cfg config.Config, provider review.Provider, model, repoSlug string, agents []string, opts Options) error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	agentFile := filepath.Join(root, cfg.AgentsDir, opts.Agent+".md")
	agentPrompt, err := os.ReadFile(agentFile)
	if err != nil {
		// Agent not found — list available agents
		fmt.Fprintf(os.Stderr, "Error: Subagent not found: %s\n\n", agentFile)
		printAvailableAgents(agents)
		return fmt.Errorf("agent not found: %s", opts.Agent)
	}

	fmt.Fprintf(os.Stderr, "Building agent context for %s...\n", opts.Agent)
	agentCtx, err := rcontext.BuildAgent(pr, cfg, opts.Agent, string(agentPrompt))
	if err != nil {
		return fmt.Errorf("build agent context: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Context: %d files, %d imports, %d siblings, %d tests\n",
		len(agentCtx.FileContents), len(agentCtx.Imports), len(agentCtx.Siblings), len(agentCtx.Tests))

	result, err := review.RunAgent(ctx, agentCtx, provider, model, opts.DryRun)
	if err != nil {
		return err
	}

	if opts.DryRun {
		output.PrintDryRun(result.Body)
		return nil
	}

	comment := output.FormatAgentComment(result, opts.Agent, provider.Label, model, result.PromptTokens, len(pr.ChangedFiles), opts.PRNumber)
	if err := gh.PostComment(opts.PRNumber, comment); err != nil {
		return fmt.Errorf("post comment: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Review posted to PR #%d\n", opts.PRNumber)
	if repoSlug != "" {
		fmt.Fprintf(os.Stderr, "View: https://github.com/%s/pull/%d\n", repoSlug, opts.PRNumber)
	}
	if result.Verdict != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", result.Verdict)
	}
	return nil
}

func runFull(ctx gocontext.Context, pr gh.PRData, cfg config.Config, provider review.Provider, model, repoSlug string, agents []string, opts Options) error {
	fmt.Fprintln(os.Stderr, "========================================================")
	fmt.Fprintf(os.Stderr, "  %s Full Review Loop — PR #%d\n", provider.Label, opts.PRNumber)
	fmt.Fprintf(os.Stderr, "  Model: %s\n", model)
	fmt.Fprintln(os.Stderr, "========================================================")

	fmt.Fprintln(os.Stderr, "\nBuilding review context...")
	reviewCtx, err := rcontext.Build(pr, cfg)
	if err != nil {
		return fmt.Errorf("build context: %w", err)
	}
	reviewCtx.Focus = opts.Focus
	reviewCtx.AvailableAgents = agents

	fmt.Fprintf(os.Stderr, "Context: %d files included, %d skipped (~%d tokens)\n",
		len(reviewCtx.FileContents), reviewCtx.FilesSkipped, reviewCtx.TokenEstimate)

	if opts.DryRun {
		result, dryErr := review.RunGeneric(ctx, reviewCtx, provider, model, true)
		if dryErr != nil {
			return dryErr
		}
		output.PrintDryRun(result.Body)
		return nil
	}

	results, err := review.RunFull(ctx, pr, reviewCtx, cfg, provider, model, false, opts.PRNumber)
	if err != nil {
		return err
	}

	output.PrintFullSummary(results, opts.PRNumber, model, repoSlug)
	return nil
}

func runDefault(ctx gocontext.Context, pr gh.PRData, cfg config.Config, provider review.Provider, model, repoSlug string, agents []string, opts Options) error {
	fmt.Fprintln(os.Stderr, "Building review context...")
	reviewCtx, err := rcontext.Build(pr, cfg)
	if err != nil {
		return fmt.Errorf("build context: %w", err)
	}
	reviewCtx.Focus = opts.Focus
	reviewCtx.AvailableAgents = agents

	fmt.Fprintf(os.Stderr, "Context: %d files included, %d skipped (~%d tokens)\n",
		len(reviewCtx.FileContents), reviewCtx.FilesSkipped, reviewCtx.TokenEstimate)

	// Generic review
	fmt.Fprintf(os.Stderr, "\nRunning reviews with %s (%s)...\n", provider.Label, model)
	generic, err := review.RunGeneric(ctx, reviewCtx, provider, model, opts.DryRun)
	if err != nil {
		return fmt.Errorf("generic review: %w", err)
	}

	if opts.DryRun {
		output.PrintDryRun(generic.Body)
		return nil
	}

	results := []review.Result{generic}

	// Focused reviews in PARALLEL (matching bash behavior)
	focuses := review.DetermineFocuses(reviewCtx.Categories)
	if len(focuses) > 0 {
		for _, focus := range focuses {
			fmt.Fprintf(os.Stderr, "  [%s] started\n", focus)
		}

		g, gctx := errgroup.WithContext(ctx)
		focusResults := make([]review.Result, len(focuses))

		for i, focus := range focuses {
			g.Go(func() error {
				result, runErr := review.RunFocused(gctx, reviewCtx, focus, provider, model, false)
				if runErr != nil {
					fmt.Fprintf(os.Stderr, "  [%s] failed: %v\n", focus, runErr)
					return nil // don't fail the whole group
				}
				focusResults[i] = result
				fmt.Fprintf(os.Stderr, "  [%s] done\n", focus)
				return nil
			})
		}
		_ = g.Wait()

		for _, r := range focusResults {
			if r.Label != "" {
				results = append(results, r)
			}
		}
	}

	// Post combined comment
	comment := output.FormatCombinedComment(results, provider.Label, model, generic.PromptTokens, opts.PRNumber)
	if err := gh.PostComment(opts.PRNumber, comment); err != nil {
		return fmt.Errorf("post comment: %w", err)
	}

	output.PrintSummary(results, opts.PRNumber, model, generic.PromptTokens, repoSlug)
	return nil
}

// discoverAgents returns sorted agent names from .md files in the agents directory.
func discoverAgents(agentsDir string) []string {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil
	}

	var agents []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		agents = append(agents, strings.TrimSuffix(e.Name(), ".md"))
	}

	sort.Strings(agents)
	return agents
}

// printAvailableAgents lists discovered agents to stderr.
func printAvailableAgents(agents []string) {
	if len(agents) == 0 {
		fmt.Fprintln(os.Stderr, "No agent files found.")
		return
	}
	fmt.Fprintln(os.Stderr, "Available agents:")
	for _, a := range agents {
		fmt.Fprintf(os.Stderr, "  - %s\n", a)
	}
}

// reorderArgs moves flags before positional arguments so Go's flag package
// can parse "gh ai-review 620 --dry-run" correctly.
func reorderArgs(args []string) []string {
	var flags, positional []string
	i := 0
	for i < len(args) {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			// Flags with values: --agent foo, --focus "text", --model gemini-2.5-flash
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && needsValue(args[i]) {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			positional = append(positional, args[i])
		}
		i++
	}
	return append(flags, positional...)
}

func needsValue(flag string) bool {
	f := strings.TrimLeft(flag, "-")
	switch f {
	case "agent", "focus", "model", "provider":
		return true
	}
	return false
}

// normalizeProvider returns the active provider name, normalized to lowercase.
// Flag takes precedence over config.
func normalizeProvider(cfgProvider, flagProvider string) string {
	p := cfgProvider
	if v := strings.ToLower(strings.TrimSpace(flagProvider)); v != "" {
		p = v
	}
	return p
}

func resolveProvider(name string) (review.Provider, error) {
	switch name {
	case review.ProviderGemini:
		return review.Provider{
			Call:  review.CallGemini,
			Label: "Gemini",
		}, nil
	case review.ProviderCodex:
		codexBin, pathErr := review.CodexPath()
		if pathErr != nil {
			return review.Provider{}, fmt.Errorf("codex CLI not found on PATH — install with: npm install -g @openai/codex")
		}
		return review.Provider{
			Call:  review.MakeCodexCaller(codexBin),
			Label: "Codex",
		}, nil
	}
	return review.Provider{}, fmt.Errorf("unknown provider: %q (supported: gemini, codex)", name)
}

// resolveModel determines the model to use. Priority: flag > config > provider default.
// Each provider has its own default model. If the config still has the default for a
// different provider, we switch to the correct default for the active provider.
// Returns an error if a config model is clearly incompatible with the provider.
func resolveModel(flagModel, cfgModel, providerName string) (string, error) {
	if flagModel != "" {
		if looksIncompatible(flagModel, providerName) {
			fmt.Fprintf(os.Stderr, "warning: model %q may not be compatible with provider %q\n", flagModel, providerName)
		}
		return flagModel, nil
	}
	// If config model is the default for a different provider, use this provider's default
	if cfgModel == defaultModelForProvider(config.DefaultProvider) && providerName != config.DefaultProvider {
		return defaultModelForProvider(providerName), nil
	}
	// Auto-switch to provider default if config model is incompatible
	if looksIncompatible(cfgModel, providerName) {
		d := defaultModelForProvider(providerName)
		fmt.Fprintf(os.Stderr, "warning: config model %q is not compatible with provider %q, using default %q\n", cfgModel, providerName, d)
		return d, nil
	}
	return cfgModel, nil
}

func looksIncompatible(model, provider string) bool {
	if provider == review.ProviderCodex && strings.HasPrefix(model, "gemini") {
		return true
	}
	if provider == review.ProviderGemini && (strings.HasPrefix(model, "gpt") || strings.HasPrefix(model, "o4") || strings.HasPrefix(model, "o3")) {
		return true
	}
	return false
}

func defaultModelForProvider(provider string) string {
	switch provider {
	case review.ProviderCodex:
		return config.DefaultCodexModel
	case review.ProviderGemini:
		return config.DefaultGeminiModel
	}
	// resolveProvider rejects unknown providers before this is called,
	// but return Gemini default as a safe fallback.
	return config.DefaultGeminiModel
}
