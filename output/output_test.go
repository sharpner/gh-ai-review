package output

import (
	"strings"
	"testing"

	"github.com/sharpner/gh-ai-review/review"
)

func TestFormatCombinedComment(t *testing.T) {
	results := []review.Result{
		{Label: "Generic Review", Body: "Review body 1"},
		{Label: "Security Review", Body: "Review body 2"},
	}

	comment := FormatCombinedComment(results, "gemini-3-flash", 5000, 42)

	if !strings.Contains(comment, "Review body 1") {
		t.Error("missing first review body")
	}
	if !strings.Contains(comment, "Review body 2") {
		t.Error("missing second review body")
	}
	if !strings.Contains(comment, "---") {
		t.Error("missing separator")
	}
	if !strings.Contains(comment, "gemini-3-flash") {
		t.Error("missing model in footer")
	}
	if !strings.Contains(comment, "gh ai-review 42") {
		t.Error("missing run command in footer")
	}
}

func TestFormatAgentComment(t *testing.T) {
	result := review.Result{
		Label:  "Agent: security-reviewer",
		Body:   "Agent review body",
		PromptTokens: 10000,
	}

	comment := FormatAgentComment(result, "security-reviewer", "gemini-3-flash", 10000, 15, 42)

	if !strings.Contains(comment, "Agent review body") {
		t.Error("missing review body")
	}
	if !strings.Contains(comment, "impersonating `security-reviewer`") {
		t.Error("missing agent name in footer")
	}
	if !strings.Contains(comment, "15 files") {
		t.Error("missing file count")
	}
	if !strings.Contains(comment, "--agent security-reviewer") {
		t.Error("missing agent run command")
	}
}

func TestExtractCriticalIssues(t *testing.T) {
	body := `### Critical Issues
- SQL injection in auth.ts:45
- Missing CSRF token validation

### Warnings
- Minor style issue`

	issues := extractCriticalIssues(body)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d: %v", len(issues), issues)
	}
	if issues[0] != "- SQL injection in auth.ts:45" {
		t.Errorf("issues[0] = %q", issues[0])
	}
}

func TestExtractCriticalIssues_NoneFound(t *testing.T) {
	body := `### Critical Issues
- None found

### Warnings`

	issues := extractCriticalIssues(body)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for 'None found', got %d", len(issues))
	}
}

func TestExtractCriticalIssues_NoSection(t *testing.T) {
	body := "No critical issues section at all"
	issues := extractCriticalIssues(body)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}
