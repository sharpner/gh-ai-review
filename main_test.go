package main

import (
	"testing"
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
