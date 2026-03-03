package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const version = "0.1.0"

type Options struct {
	PRNumber int
	Full     bool
	Agent    string
	Focus    string
	DryRun   bool
	Model    string
}

func main() {
	fs := flag.NewFlagSet("gh-ai-review", flag.ExitOnError)

	var opts Options
	fs.BoolVar(&opts.Full, "full", false, "Run full review loop (generic + focused)")
	fs.StringVar(&opts.Agent, "agent", "", "Agent persona to impersonate")
	fs.StringVar(&opts.Focus, "focus", "", "Custom focus area for review")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Print prompt, skip Gemini call")
	fs.StringVar(&opts.Model, "model", "", "Gemini model to use (overrides config)")
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
	opts.PRNumber = prNumber

	if err := run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(opts Options) error {
	return fmt.Errorf("not implemented — run `gh ai-review %d` once review logic is built", opts.PRNumber)
}
