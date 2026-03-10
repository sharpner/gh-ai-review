package main

import (
	"testing"

	"github.com/sharpner/gh-ai-review/config"
	"github.com/sharpner/gh-ai-review/review"
)

func TestReorderArgs_FlagsAfterPositional(t *testing.T) {
	args := []string{"620", "--dry-run"}
	got := reorderArgs(args)

	if len(got) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(got), got)
	}
	if got[0] != "--dry-run" {
		t.Errorf("got[0] = %q, want --dry-run", got[0])
	}
	if got[1] != "620" {
		t.Errorf("got[1] = %q, want 620", got[1])
	}
}

func TestReorderArgs_FlagsWithValues(t *testing.T) {
	args := []string{"620", "--agent", "security", "--dry-run"}
	got := reorderArgs(args)

	expected := []string{"--agent", "security", "--dry-run", "620"}
	if len(got) != len(expected) {
		t.Fatalf("got %v, want %v", got, expected)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestReorderArgs_AlreadyOrdered(t *testing.T) {
	args := []string{"--dry-run", "620"}
	got := reorderArgs(args)

	if got[0] != "--dry-run" || got[1] != "620" {
		t.Errorf("got %v, expected same order", got)
	}
}

func TestReorderArgs_Empty(t *testing.T) {
	got := reorderArgs(nil)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestReorderArgs_ModelFlag(t *testing.T) {
	args := []string{"1", "--model", "gemini-2.5-pro", "--full"}
	got := reorderArgs(args)

	expected := []string{"--model", "gemini-2.5-pro", "--full", "1"}
	if len(got) != len(expected) {
		t.Fatalf("got %v, want %v", got, expected)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestReorderArgs_FocusWithSpaces(t *testing.T) {
	args := []string{"620", "--focus", "backward compatibility"}
	got := reorderArgs(args)

	if got[0] != "--focus" || got[1] != "backward compatibility" || got[2] != "620" {
		t.Errorf("got %v", got)
	}
}

func TestReorderArgs_ProviderFlag(t *testing.T) {
	args := []string{"1", "--provider", "codex", "--dry-run"}
	got := reorderArgs(args)

	expected := []string{"--provider", "codex", "--dry-run", "1"}
	if len(got) != len(expected) {
		t.Fatalf("got %v, want %v", got, expected)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestResolveProvider_Gemini(t *testing.T) {
	p, err := resolveProvider("gemini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Label != "Gemini" {
		t.Errorf("label = %q, want Gemini", p.Label)
	}
	if p.Call == nil {
		t.Fatal("Call is nil")
	}
}

func TestResolveProvider_Codex(t *testing.T) {
	p, err := resolveProvider("codex")
	if review.CodexAvailable() {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Label != "Codex" {
			t.Errorf("label = %q, want Codex", p.Label)
		}
		if p.Call == nil {
			t.Fatal("Call is nil")
		}
	} else {
		if err == nil {
			t.Fatal("expected error when codex not on PATH")
		}
	}
}

func TestResolveProvider_Unknown(t *testing.T) {
	_, err := resolveProvider("openai")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestResolveModel_FlagOverride(t *testing.T) {
	got, err := resolveModel("o4-mini", config.DefaultGeminiModel, review.ProviderCodex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "o4-mini" {
		t.Errorf("got %q, want o4-mini", got)
	}
}

func TestResolveModel_CodexDefault(t *testing.T) {
	got, err := resolveModel("", config.DefaultGeminiModel, review.ProviderCodex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != config.DefaultCodexModel {
		t.Errorf("got %q, want %q", got, config.DefaultCodexModel)
	}
}

func TestResolveModel_GeminiDefault(t *testing.T) {
	got, err := resolveModel("", config.DefaultGeminiModel, review.ProviderGemini)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != config.DefaultGeminiModel {
		t.Errorf("got %q, want %q", got, config.DefaultGeminiModel)
	}
}

func TestResolveModel_IncompatibleGeminiModelWithCodex(t *testing.T) {
	_, err := resolveModel("", "gemini-2.5-pro", review.ProviderCodex)
	if err == nil {
		t.Fatal("expected error for gemini model with codex provider")
	}
}

func TestResolveModel_IncompatibleGPTModelWithGemini(t *testing.T) {
	_, err := resolveModel("", "gpt-5.4", review.ProviderGemini)
	if err == nil {
		t.Fatal("expected error for gpt model with gemini provider")
	}
}

func TestResolveModel_FlagOverridesIncompatibilityCheck(t *testing.T) {
	// Explicit --model flag bypasses compatibility check (user knows what they're doing)
	got, err := resolveModel("gemini-2.5-pro", "gemini-2.5-pro", review.ProviderCodex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "gemini-2.5-pro" {
		t.Errorf("got %q, want gemini-2.5-pro", got)
	}
}

func TestNeedsValue(t *testing.T) {
	tests := []struct {
		flag string
		want bool
	}{
		{"--agent", true},
		{"--focus", true},
		{"--model", true},
		{"--provider", true},
		{"--dry-run", false},
		{"--full", false},
		{"--version", false},
	}
	for _, tt := range tests {
		got := needsValue(tt.flag)
		if got != tt.want {
			t.Errorf("needsValue(%q) = %v, want %v", tt.flag, got, tt.want)
		}
	}
}
