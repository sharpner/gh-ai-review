package review

import (
	"testing"

	"github.com/sharpner/gh-ai-review/context"
)

func TestExtractRecommendedAgents(t *testing.T) {
	body := `## General Review

**Verdict:** PASS

### Recommended Reviewers
- [ ] ` + "`security-pentest-reviewer`" + `: API routes, auth
- [ ] ` + "`mobile-responsive-reviewer`" + `: UI components
- [ ] ` + "`pr-review-toolkit:code-reviewer`" + `: Complex PR

### Summary
All good.`

	agents := ExtractRecommendedAgents(body)

	if len(agents) != 3 {
		t.Fatalf("got %d agents, want 3: %v", len(agents), agents)
	}

	expected := []string{"security-pentest-reviewer", "mobile-responsive-reviewer", "pr-review-toolkit:code-reviewer"}
	for i, want := range expected {
		if agents[i] != want {
			t.Errorf("agents[%d] = %q, want %q", i, agents[i], want)
		}
	}
}

func TestExtractRecommendedAgents_Empty(t *testing.T) {
	body := "No reviewers section here."
	agents := ExtractRecommendedAgents(body)
	if len(agents) != 0 {
		t.Errorf("expected 0 agents, got %d", len(agents))
	}
}

func TestExtractRecommendedAgents_Deduplication(t *testing.T) {
	body := "### Recommended Reviewers\n" +
		"- `security-pentest-reviewer`: reason 1\n" +
		"- `security-pentest-reviewer`: reason 2\n" +
		"### Summary\n"

	agents := ExtractRecommendedAgents(body)
	if len(agents) != 1 {
		t.Errorf("expected 1 unique agent, got %d", len(agents))
	}
}

func TestDetermineFocuses_SecurityFromAPI(t *testing.T) {
	cats := context.FileCategories{API: true}
	focuses := DetermineFocuses(cats)

	if len(focuses) != 1 || focuses[0] != "security" {
		t.Errorf("expected [security], got %v", focuses)
	}
}

func TestDetermineFocuses_AllFromUI(t *testing.T) {
	cats := context.FileCategories{UI: true}
	focuses := DetermineFocuses(cats)

	// UI triggers usability and mobile
	if len(focuses) != 2 {
		t.Fatalf("expected 2 focuses, got %d: %v", len(focuses), focuses)
	}
	if focuses[0] != "usability" || focuses[1] != "mobile" {
		t.Errorf("expected [usability, mobile], got %v", focuses)
	}
}

func TestDetermineFocuses_None(t *testing.T) {
	cats := context.FileCategories{Config: true, Docs: true}
	focuses := DetermineFocuses(cats)
	if len(focuses) != 0 {
		t.Errorf("expected 0 focuses for Config+Docs, got %v", focuses)
	}
}

func TestExtractMatch_Verdict(t *testing.T) {
	tests := []struct {
		body string
		want string
	}{
		{"**Verdict:** PASS", "PASS"},
		{"**Verdict:** NEEDS WORK", "NEEDS WORK"},
		{"**Verdict:** FAIL", "FAIL"},
		{"no verdict here", ""},
	}
	for _, tt := range tests {
		got := extractMatch(verdictRe, tt.body)
		if got != tt.want {
			t.Errorf("extractMatch(verdictRe, %q) = %q, want %q", tt.body[:20], got, tt.want)
		}
	}
}

func TestExtractMatch_RiskLevel(t *testing.T) {
	got := extractMatch(riskLevelRe, "**Risk Level:** CRITICAL")
	if got != "CRITICAL" {
		t.Errorf("got %q, want CRITICAL", got)
	}
}

func TestExtractMatch_A11y(t *testing.T) {
	got := extractMatch(a11yScoreRe, "**Accessibility Score:** GOOD")
	if got != "GOOD" {
		t.Errorf("got %q, want GOOD", got)
	}
}

func TestExtractMatch_MobileReady(t *testing.T) {
	got := extractMatch(mobileReadyRe, "**Mobile Ready:** YES")
	if got != "YES" {
		t.Errorf("got %q, want YES", got)
	}
}

func TestFocusLabel(t *testing.T) {
	tests := []struct {
		focus, want string
	}{
		{"security", "Security Review"},
		{"usability", "Usability Review"},
		{"mobile", "Mobile Review"},
		{"custom", "custom Review"},
	}
	for _, tt := range tests {
		got := focusLabel(tt.focus)
		if got != tt.want {
			t.Errorf("focusLabel(%q) = %q, want %q", tt.focus, got, tt.want)
		}
	}
}

func TestStripAgentPreamble_WithPreamble(t *testing.T) {
	// Real-world pattern: Gemini echoes the preamble, then outputs the review with # Security Review
	body := "You are impersonating a Claude Code subagent...\n\n# Your Identity\nSome instructions\n\n# Security Review Results\n\n**Risk Level:** LOW"
	got := stripAgentPreamble(body)
	if got != "# Security Review Results\n\n**Risk Level:** LOW" {
		t.Errorf("stripAgentPreamble did not strip preamble:\ngot:  %q", got)
	}
}

func TestStripAgentPreamble_NoPreamble(t *testing.T) {
	body := "## Security Review\n**Risk Level:** LOW"
	got := stripAgentPreamble(body)
	if got != body {
		t.Errorf("stripAgentPreamble modified body without preamble")
	}
}

func TestStripAgentPreamble_VerdictMarker(t *testing.T) {
	body := "You are impersonating a Claude Code subagent...\nSome stuff\n**Verdict:** PASS\nMore stuff"
	got := stripAgentPreamble(body)
	if got != "**Verdict:** PASS\nMore stuff" {
		t.Errorf("stripAgentPreamble with Verdict marker:\ngot:  %q", got)
	}
}
