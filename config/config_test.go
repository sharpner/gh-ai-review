package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if cfg.Provider != DefaultProvider {
		t.Errorf("provider = %q, want %q", cfg.Provider, DefaultProvider)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.MaxContextChars != DefaultMaxContextChars {
		t.Errorf("max_context_chars = %d, want %d", cfg.MaxContextChars, DefaultMaxContextChars)
	}
	if cfg.AgentsDir != ".claude/agents" {
		t.Errorf("agents_dir = %q, want .claude/agents", cfg.AgentsDir)
	}
	if len(cfg.ContextDocs) != 3 {
		t.Errorf("context_docs len = %d, want 3", len(cfg.ContextDocs))
	}
	if len(cfg.FocusedReviews) != 3 {
		t.Errorf("focused_reviews len = %d, want 3", len(cfg.FocusedReviews))
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/.ai-review.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("expected defaults when file not found, got model = %q", cfg.Model)
	}
}

func TestLoad_Override(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ai-review.yaml")

	content := `model: gemini-2.5-pro
max_context_chars: 500000
agents_dir: custom/agents
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != "gemini-2.5-pro" {
		t.Errorf("model = %q, want gemini-2.5-pro", cfg.Model)
	}
	if cfg.MaxContextChars != 500000 {
		t.Errorf("max_context_chars = %d, want 500000", cfg.MaxContextChars)
	}
	if cfg.AgentsDir != "custom/agents" {
		t.Errorf("agents_dir = %q, want custom/agents", cfg.AgentsDir)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ai-review.yaml")

	if err := os.WriteFile(path, []byte("{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
