package review

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCodexSchema_ValidJSON(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(codexSchema, &schema); err != nil {
		t.Fatalf("codex_schema.json is not valid JSON: %v", err)
	}
	if schema["type"] != "object" {
		t.Error("schema root type should be 'object'")
	}
}

func TestCodexReview_Verdict_Pass(t *testing.T) {
	r := CodexReview{
		OverallCorrectness: "patch is correct",
		Findings:           []CodexFinding{{Priority: 2}, {Priority: 3}},
	}
	if got := r.Verdict(); got != "PASS" {
		t.Errorf("got %q, want PASS", got)
	}
}

func TestCodexReview_Verdict_NeedsWork(t *testing.T) {
	r := CodexReview{
		OverallCorrectness: "patch is correct",
		Findings:           []CodexFinding{{Priority: 1}, {Priority: 2}},
	}
	if got := r.Verdict(); got != "NEEDS WORK" {
		t.Errorf("got %q, want NEEDS WORK", got)
	}
}

func TestCodexReview_Verdict_FailFromP0(t *testing.T) {
	r := CodexReview{
		OverallCorrectness: "patch is correct",
		Findings:           []CodexFinding{{Priority: 0}},
	}
	if got := r.Verdict(); got != "FAIL" {
		t.Errorf("got %q, want FAIL", got)
	}
}

func TestCodexReview_Verdict_FailFromIncorrect(t *testing.T) {
	r := CodexReview{
		OverallCorrectness: "patch is incorrect",
	}
	if got := r.Verdict(); got != "FAIL" {
		t.Errorf("got %q, want FAIL", got)
	}
}

func TestCodexReview_Verdict_PassNoFindings(t *testing.T) {
	r := CodexReview{
		OverallCorrectness: "patch is correct",
	}
	if got := r.Verdict(); got != "PASS" {
		t.Errorf("got %q, want PASS", got)
	}
}

func TestCodexReview_Markdown_WithFindings(t *testing.T) {
	r := CodexReview{
		Findings: []CodexFinding{
			{
				Title:           "Missing nil check",
				Body:            "Function may panic on nil input",
				ConfidenceScore: 0.95,
				Priority:        0,
				CodeLocation: CodeLocation{
					FilePath:  "main.go",
					LineRange: LineRange{Start: 42, End: 45},
				},
			},
			{
				Title:           "Unused variable",
				Body:            "Variable 'x' is declared but never used",
				ConfidenceScore: 0.8,
				Priority:        2,
				CodeLocation: CodeLocation{
					FilePath:  "review/codex.go",
					LineRange: LineRange{Start: 10, End: 10},
				},
			},
		},
		OverallCorrectness:     "patch is incorrect",
		OverallExplanation:     "Critical nil check missing in main path.",
		OverallConfidenceScore: 0.9,
	}

	md := r.Markdown()

	// Verify verdict
	if !strings.Contains(md, "**Verdict:** FAIL") {
		t.Error("missing FAIL verdict")
	}

	// Verify confidence
	if !strings.Contains(md, "**Confidence:** 90%") {
		t.Error("missing confidence percentage")
	}

	// Verify P0 finding in Critical Issues
	if !strings.Contains(md, "### Critical Issues") {
		t.Error("missing Critical Issues section")
	}
	if !strings.Contains(md, "Missing nil check") {
		t.Error("missing P0 finding title")
	}
	if !strings.Contains(md, "`main.go:42-45`") {
		t.Error("missing file:line range reference")
	}

	// Verify P2 finding in Suggestions
	if !strings.Contains(md, "### Suggestions") {
		t.Error("missing Suggestions section")
	}
	if !strings.Contains(md, "Unused variable") {
		t.Error("missing P2 finding title")
	}

	// Verify single-line format
	if !strings.Contains(md, "`review/codex.go:10`") {
		t.Error("single-line finding should show :10 not :10-10")
	}

	// Verify summary
	if !strings.Contains(md, "Critical nil check missing") {
		t.Error("missing summary text")
	}

	// Verify Warnings section says "None"
	if !strings.Contains(md, "### Warnings\n- None") {
		t.Error("Warnings section should say None when no P1 findings")
	}
}

func TestCodexReview_Markdown_NoFindings(t *testing.T) {
	r := CodexReview{
		OverallCorrectness:     "patch is correct",
		OverallExplanation:     "Clean implementation.",
		OverallConfidenceScore: 0.95,
	}

	md := r.Markdown()

	if !strings.Contains(md, "**Verdict:** PASS") {
		t.Error("missing PASS verdict")
	}
	if !strings.Contains(md, "- None found") {
		t.Error("Critical Issues should say 'None found'")
	}
}

func TestCodexReview_Markdown_VerdictRegexCompatible(t *testing.T) {
	r := CodexReview{
		OverallCorrectness:     "patch is correct",
		OverallExplanation:     "All good.",
		OverallConfidenceScore: 1.0,
	}

	md := r.Markdown()

	// Our verdict regex must match the rendered output
	verdict := extractMatch(verdictRe, md)
	if verdict != "PASS" {
		t.Errorf("verdictRe failed to match rendered markdown, got %q", verdict)
	}
}

func TestCodexReview_Markdown_NeedsWorkRegexCompatible(t *testing.T) {
	r := CodexReview{
		Findings: []CodexFinding{
			{
				Title:           "Potential issue",
				Body:            "Should handle edge case",
				ConfidenceScore: 0.7,
				Priority:        1,
				CodeLocation:    CodeLocation{FilePath: "x.go", LineRange: LineRange{Start: 1, End: 1}},
			},
		},
		OverallCorrectness:     "patch is correct",
		OverallExplanation:     "Minor issues.",
		OverallConfidenceScore: 0.8,
	}

	md := r.Markdown()
	verdict := extractMatch(verdictRe, md)
	if verdict != "NEEDS WORK" {
		t.Errorf("verdictRe should match NEEDS WORK, got %q", verdict)
	}
}

func TestCodexReview_JSONRoundtrip(t *testing.T) {
	raw := `{
		"findings": [
			{
				"title": "Error not wrapped",
				"body": "Use fmt.Errorf with %%w",
				"confidence_score": 0.85,
				"priority": 2,
				"code_location": {
					"file_path": "review/review.go",
					"line_range": {"start": 49, "end": 49}
				}
			}
		],
		"overall_correctness": "patch is correct",
		"overall_explanation": "Minor suggestion only.",
		"overall_confidence_score": 0.92
	}`

	var r CodexReview
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(r.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(r.Findings))
	}
	if r.Findings[0].Title != "Error not wrapped" {
		t.Errorf("title = %q", r.Findings[0].Title)
	}
	if r.Findings[0].CodeLocation.FilePath != "review/review.go" {
		t.Errorf("file_path = %q", r.Findings[0].CodeLocation.FilePath)
	}
	if r.Verdict() != "PASS" {
		t.Errorf("verdict = %q, want PASS", r.Verdict())
	}
}

func TestNormalizeCodexPath_Relative(t *testing.T) {
	got := normalizeCodexPath("review/codex.go")
	if got != "review/codex.go" {
		t.Errorf("relative path should pass through, got %q", got)
	}
}

func TestWriteFindings_Empty(t *testing.T) {
	var b strings.Builder
	got := writeFindings(&b, nil, 0, 3)
	if got {
		t.Error("expected false for nil findings")
	}
	if b.Len() != 0 {
		t.Error("expected no output")
	}
}

func TestWriteFindings_FiltersByPriority(t *testing.T) {
	findings := []CodexFinding{
		{Title: "P0", Priority: 0, Body: "critical", CodeLocation: CodeLocation{FilePath: "a.go", LineRange: LineRange{Start: 1, End: 1}}},
		{Title: "P1", Priority: 1, Body: "warning", CodeLocation: CodeLocation{FilePath: "b.go", LineRange: LineRange{Start: 2, End: 2}}},
		{Title: "P2", Priority: 2, Body: "suggestion", CodeLocation: CodeLocation{FilePath: "c.go", LineRange: LineRange{Start: 3, End: 3}}},
	}

	var b strings.Builder
	writeFindings(&b, findings, 0, 0)
	out := b.String()

	if !strings.Contains(out, "P0") {
		t.Error("should contain P0 finding")
	}
	if strings.Contains(out, "P1") {
		t.Error("should not contain P1 finding")
	}
	if strings.Contains(out, "P2") {
		t.Error("should not contain P2 finding")
	}
}
