package output

import (
	"fmt"
	"os"

	"github.com/sharpner/gh-ai-review/review"
)

// PrintResult writes a review result to stdout.
func PrintResult(r review.Result) {
	fmt.Fprintf(os.Stdout, "## %s\n\n%s\n\nVerdict: %s\n", r.Label, r.Body, r.Verdict)
}

// PrintResults writes multiple review results to stdout.
func PrintResults(results []review.Result) {
	for i, r := range results {
		if i > 0 {
			fmt.Fprintln(os.Stdout, "---")
		}
		PrintResult(r)
	}
}

// PrintDryRun writes the prompt that would be sent to Gemini.
func PrintDryRun(prompt string) {
	fmt.Fprintln(os.Stdout, "=== DRY RUN — Prompt that would be sent to Gemini ===")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, prompt)
}
