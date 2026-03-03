package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const version = "0.1.0"

func main() {
	fs := flag.NewFlagSet("gh-ai-review", flag.ExitOnError)

	full := fs.Bool("full", false, "Run full review loop (generic + focused)")
	agent := fs.String("agent", "", "Agent persona to impersonate")
	focus := fs.String("focus", "", "Custom focus area for review")
	dryRun := fs.Bool("dry-run", false, "Print prompt, skip Gemini call")
	model := fs.String("model", "", "Gemini model to use (overrides config)")
	showVersion := fs.Bool("version", false, "Print version")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gh ai-review <pr-number> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "AI-powered code review for GitHub PRs using Gemini.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --full\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --agent security-pentest\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --focus \"backward compatibility\"\n")
		fmt.Fprintf(os.Stderr, "  gh ai-review 620 --dry-run\n")
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
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

	if err := run(prNumber, *full, *agent, *focus, *dryRun, *model); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(prNumber int, full bool, agent, focus string, dryRun bool, model string) error {
	_ = prNumber
	_ = full
	_ = agent
	_ = focus
	_ = dryRun
	_ = model
	return fmt.Errorf("not implemented — run `gh ai-review %d` once review logic is built", prNumber)
}
