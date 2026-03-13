package review

import (
	"strings"
	"testing"

	gh "github.com/sharpner/gh-ai-review/github"
)

func TestWriteComments_Empty(t *testing.T) {
	var b strings.Builder
	writeComments(&b, nil)
	if b.Len() != 0 {
		t.Errorf("expected empty output for nil comments, got %q", b.String())
	}
}

func TestWriteComments_IssueComment(t *testing.T) {
	var b strings.Builder
	writeComments(&b, []gh.Comment{
		{Author: "alice", Body: "Looks good!", CreatedAt: "2025-01-01T00:00:00Z"},
	})
	out := b.String()

	if !strings.Contains(out, "# Existing PR Discussion") {
		t.Error("missing header")
	}
	if !strings.Contains(out, "**alice**:") {
		t.Error("missing author")
	}
	if !strings.Contains(out, "> Looks good!") {
		t.Error("missing blockquoted body")
	}
}

func TestWriteComments_ReviewComment(t *testing.T) {
	var b strings.Builder
	writeComments(&b, []gh.Comment{
		{Author: "bob", Body: "Fix this null check", CreatedAt: "2025-01-01T00:00:00Z", Path: "main.go", Line: 42},
	})
	out := b.String()

	if !strings.Contains(out, "**bob** on `main.go:42`:") {
		t.Error("missing inline comment path:line reference")
	}
	if !strings.Contains(out, "> Fix this null check") {
		t.Error("missing blockquoted body")
	}
}

func TestWriteComments_MultilineBody(t *testing.T) {
	var b strings.Builder
	writeComments(&b, []gh.Comment{
		{Author: "carol", Body: "Line one\nLine two\nLine three", CreatedAt: "2025-01-01T00:00:00Z"},
	})
	out := b.String()

	if !strings.Contains(out, "> Line one\n> Line two\n> Line three") {
		t.Errorf("multiline body not properly blockquoted:\n%s", out)
	}
}

func TestWriteComments_EmptySlice(t *testing.T) {
	var b strings.Builder
	writeComments(&b, []gh.Comment{})
	if b.Len() != 0 {
		t.Errorf("expected empty output for empty slice, got %q", b.String())
	}
}

func TestWriteComments_MultipleComments(t *testing.T) {
	var b strings.Builder
	writeComments(&b, []gh.Comment{
		{Author: "alice", Body: "First comment", CreatedAt: "2025-01-01T00:00:00Z"},
		{Author: "bob", Body: "Second comment", CreatedAt: "2025-01-02T00:00:00Z", Path: "foo.go", Line: 10},
	})
	out := b.String()

	if !strings.Contains(out, "**alice**:") {
		t.Error("missing first author")
	}
	if !strings.Contains(out, "> First comment") {
		t.Error("missing first comment body")
	}
	if !strings.Contains(out, "**bob** on `foo.go:10`:") {
		t.Error("missing second comment with path:line")
	}
	if !strings.Contains(out, "> Second comment") {
		t.Error("missing second comment body")
	}
}
