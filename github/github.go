package github

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh/v2"
)

type PRInfo struct {
	Number    int
	Title     string
	Body      string
	BaseRef   string
	Additions int
	Deletions int
}

type PRData struct {
	Info         PRInfo
	Diff         string
	ChangedFiles []string
}

// FetchPR retrieves PR metadata, diff, and changed file list.
func FetchPR(prNumber int) (PRData, error) {
	prStr := fmt.Sprintf("%d", prNumber)

	// 1. PR metadata
	stdout, _, err := gh.Exec("pr", "view", prStr, "--json", "title,body,additions,deletions,baseRefName")
	if err != nil {
		return PRData{}, fmt.Errorf("fetch PR metadata: %w", err)
	}

	var raw struct {
		Title       string `json:"title"`
		Body        string `json:"body"`
		Additions   int    `json:"additions"`
		Deletions   int    `json:"deletions"`
		BaseRefName string `json:"baseRefName"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return PRData{}, fmt.Errorf("parse PR metadata: %w", err)
	}

	// Truncate body to 80 lines (matching bash head -80)
	body := truncateLines(raw.Body, 80)

	// 2. Diff
	diffOut, _, err := gh.Exec("pr", "diff", prStr)
	if err != nil {
		return PRData{}, fmt.Errorf("fetch PR diff: %w", err)
	}
	diff := diffOut.String()
	if diff == "" {
		return PRData{}, fmt.Errorf("empty diff for PR #%d", prNumber)
	}

	// 3. Changed files
	filesOut, _, err := gh.Exec("pr", "diff", prStr, "--name-only")
	if err != nil {
		return PRData{}, fmt.Errorf("fetch changed files: %w", err)
	}

	var files []string
	for _, line := range strings.Split(filesOut.String(), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}

	return PRData{
		Info: PRInfo{
			Number:    prNumber,
			Title:     raw.Title,
			Body:      body,
			BaseRef:   raw.BaseRefName,
			Additions: raw.Additions,
			Deletions: raw.Deletions,
		},
		Diff:         diff,
		ChangedFiles: files,
	}, nil
}

// PostComment posts a review comment on the given PR.
func PostComment(prNumber int, body string) error {
	tmp, err := os.CreateTemp("", "gh-ai-review-comment-*.md")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(body); err != nil {
		tmp.Close()
		return fmt.Errorf("write comment body: %w", err)
	}
	tmp.Close()

	_, _, err = gh.Exec("pr", "comment", fmt.Sprintf("%d", prNumber), "--body-file", tmp.Name())
	if err != nil {
		return fmt.Errorf("post comment: %w", err)
	}
	return nil
}

// RepoSlug returns "owner/repo" for the current directory.
func RepoSlug() (string, error) {
	stdout, _, err := gh.Exec("repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func truncateLines(s string, max int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= max {
		return s
	}
	return strings.Join(lines[:max], "\n")
}
