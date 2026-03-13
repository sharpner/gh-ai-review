package review

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed codex_schema.json
var codexSchema []byte

// CodexReview is the structured output from Codex when using --output-schema.
type CodexReview struct {
	Findings               []CodexFinding `json:"findings"`
	OverallCorrectness     string         `json:"overall_correctness"`
	OverallExplanation     string         `json:"overall_explanation"`
	OverallConfidenceScore float64        `json:"overall_confidence_score"`
}

// CodexFinding is a single finding from structured Codex output.
type CodexFinding struct {
	Title           string       `json:"title"`
	Body            string       `json:"body"`
	ConfidenceScore float64      `json:"confidence_score"`
	Priority        int          `json:"priority"`
	CodeLocation    CodeLocation `json:"code_location"`
}

// CodeLocation references a file and line range for a finding.
type CodeLocation struct {
	FilePath  string    `json:"file_path"`
	LineRange LineRange `json:"line_range"`
}

// LineRange is a start/end line pair.
type LineRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Verdict derives PASS/NEEDS WORK/FAIL from the structured review.
func (r CodexReview) Verdict() string {
	if r.OverallCorrectness == "patch is incorrect" {
		return "FAIL"
	}
	for _, f := range r.Findings {
		if f.Priority == 0 {
			return "FAIL"
		}
	}
	for _, f := range r.Findings {
		if f.Priority == 1 {
			return "NEEDS WORK"
		}
	}
	return "PASS"
}

// Markdown renders the structured review as a PR comment matching our standard format.
func (r CodexReview) Markdown() string {
	var b strings.Builder

	verdict := r.Verdict()
	fmt.Fprintf(&b, "## General Review\n\n")
	fmt.Fprintf(&b, "**Verdict:** %s\n\n", verdict)
	fmt.Fprintf(&b, "**Confidence:** %.0f%%\n\n", r.OverallConfidenceScore*100)

	// Critical Issues (P0)
	b.WriteString("### Critical Issues\n")
	if !writeFindings(&b, r.Findings, 0, 0) {
		b.WriteString("- None found\n")
	}

	// Warnings (P1)
	b.WriteString("\n### Warnings\n")
	if !writeFindings(&b, r.Findings, 1, 1) {
		b.WriteString("- None\n")
	}

	// Suggestions (P2, P3)
	b.WriteString("\n### Suggestions\n")
	if !writeFindings(&b, r.Findings, 2, 3) {
		b.WriteString("- None\n")
	}

	fmt.Fprintf(&b, "\n### Summary\n%s\n", r.OverallExplanation)

	return b.String()
}

// writeFindings writes findings in the given priority range [minP, maxP].
// Returns true if any findings were written.
func writeFindings(b *strings.Builder, findings []CodexFinding, minP, maxP int) bool {
	wrote := false
	for _, f := range findings {
		if f.Priority < minP || f.Priority > maxP {
			continue
		}
		wrote = true
		loc := f.CodeLocation
		// Strip absolute path prefix — use relative path for PR comments
		path := normalizeCodexPath(loc.FilePath)
		if loc.LineRange.Start == loc.LineRange.End {
			fmt.Fprintf(b, "- **%s** `%s:%d` (P%d, %.0f%% confidence)\n",
				f.Title, path, loc.LineRange.Start, f.Priority, f.ConfidenceScore*100)
		} else {
			fmt.Fprintf(b, "- **%s** `%s:%d-%d` (P%d, %.0f%% confidence)\n",
				f.Title, path, loc.LineRange.Start, loc.LineRange.End, f.Priority, f.ConfidenceScore*100)
		}
		fmt.Fprintf(b, "  %s\n", f.Body)
	}
	return wrote
}

// normalizeCodexPath strips absolute path prefixes to produce relative paths.
func normalizeCodexPath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	// Try to make it relative to cwd
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Base(path)
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return filepath.Base(path)
	}
	return rel
}

// MakeCodexCaller returns a CallLLM function that uses the given absolute binary path.
// Resolving the path at init time prevents PATH manipulation between check and execution.
func MakeCodexCaller(codexBin string) CallLLM {
	return func(ctx context.Context, model, prompt string) (string, error) {
		return callCodex(ctx, codexBin, model, prompt)
	}
}

// callCodex sends a prompt to the Codex CLI and returns the response as rendered markdown.
// Uses --output-schema for structured JSON output, which is then rendered to our standard
// review markdown format. Falls back to raw text if JSON parsing fails.
//
// Security: the sandbox is set to read-only (-s read-only) so the Codex agent
// cannot modify files or execute arbitrary commands. The prompt contains PR
// content which is untrusted input; read-only sandbox prevents prompt injection
// from escalating to code execution. Environment is filtered to prevent leaking
// sensitive vars (API keys, tokens).
func callCodex(ctx context.Context, codexBin, model, prompt string) (string, error) {
	// Write schema to temp file for --output-schema
	schemaFile, err := os.CreateTemp("", "codex-schema-*.json")
	if err != nil {
		return "", fmt.Errorf("create schema temp file: %w", err)
	}
	defer func() { _ = os.Remove(schemaFile.Name()) }()

	if _, err := schemaFile.Write(codexSchema); err != nil {
		return "", fmt.Errorf("write schema: %w", err)
	}
	if err := schemaFile.Close(); err != nil {
		return "", fmt.Errorf("close schema file: %w", err)
	}

	cmd := exec.CommandContext(ctx, codexBin, "exec",
		"-m", model,
		"-s", "read-only",
		"--skip-git-repo-check",
		"--output-schema", schemaFile.Name(),
		"-",
	)

	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = filteredEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("codex: %w", ctx.Err())
		}
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return "", fmt.Errorf("codex: %s: %w", errMsg, err)
		}
		return "", fmt.Errorf("codex: %w", err)
	}

	text := strings.TrimSpace(stdout.String())
	if text == "" {
		return "", fmt.Errorf("empty response from Codex")
	}

	// Try to parse as structured JSON; fall back to raw text
	var review CodexReview
	if err := json.Unmarshal([]byte(text), &review); err != nil {
		return text, nil
	}

	return review.Markdown(), nil
}

// CodexAvailable checks if the codex binary is on PATH.
func CodexAvailable() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

// CodexPath returns the resolved absolute path of the codex binary.
// This pins the binary at startup to prevent PATH manipulation later.
func CodexPath() (string, error) {
	return exec.LookPath("codex")
}

// filteredEnv returns the parent environment with other providers' API keys removed.
// This prevents leaking Gemini/cloud API keys to the codex subprocess while
// preserving all auth-related vars the Codex CLI needs (OAuth, keychain, GitHub CLI).
func filteredEnv() []string {
	deny := map[string]bool{
		"GOOGLE_API_KEY":        true,
		"GEMINI_API_KEY":        true,
		"AWS_SECRET_ACCESS_KEY": true,
		"AWS_SESSION_TOKEN":     true,
	}
	var env []string
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		if deny[key] {
			continue
		}
		env = append(env, e)
	}
	return env
}
