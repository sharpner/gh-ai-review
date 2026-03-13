package github

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

// Comment represents a PR discussion comment (issue comment or review comment).
type Comment struct {
	Author    string
	Body      string
	CreatedAt string
	Path      string // only for review comments (inline)
	Line      int    // only for review comments (inline)
}

type PRData struct {
	Info         PRInfo
	Diff         string
	ChangedFiles []string
	Comments     []Comment
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

	comments, commentErr := FetchComments(prNumber)
	if commentErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not fetch PR comments: %v\n", commentErr)
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
		Comments:     comments,
	}, nil
}

// FetchComments retrieves PR discussion and review comments.
// Returns comments sorted chronologically by creation time.
func FetchComments(prNumber int) ([]Comment, error) {
	prStr := fmt.Sprintf("%d", prNumber)
	var comments []Comment

	// 1. Issue comments (PR discussion tab)
	stdout, _, err := gh.Exec("pr", "view", prStr, "--json", "comments")
	if err == nil {
		var raw struct {
			Comments []struct {
				Author struct {
					Login string `json:"login"`
				} `json:"author"`
				Body      string `json:"body"`
				CreatedAt string `json:"createdAt"`
			} `json:"comments"`
		}
		if jsonErr := json.Unmarshal(stdout.Bytes(), &raw); jsonErr == nil {
			for _, c := range raw.Comments {
				comments = append(comments, Comment{
					Author:    c.Author.Login,
					Body:      c.Body,
					CreatedAt: c.CreatedAt,
				})
			}
		}
	}

	// 2. Review comments (inline code comments)
	slug, slugErr := RepoSlug()
	if slugErr == nil && slug != "" {
		apiPath := fmt.Sprintf("repos/%s/pulls/%d/comments", slug, prNumber)
		apiOut, _, apiErr := gh.Exec("api", apiPath, "--paginate")
		if apiErr == nil {
			var reviewComments []struct {
				User struct {
					Login string `json:"login"`
				} `json:"user"`
				Body      string `json:"body"`
				CreatedAt string `json:"created_at"`
				Path      string `json:"path"`
				Line      int    `json:"line"`
			}
			if jsonErr := json.Unmarshal(apiOut.Bytes(), &reviewComments); jsonErr == nil {
				for _, c := range reviewComments {
					comments = append(comments, Comment{
						Author:    c.User.Login,
						Body:      c.Body,
						CreatedAt: c.CreatedAt,
						Path:      c.Path,
						Line:      c.Line,
					})
				}
			}
		}
	}

	// Sort chronologically
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt < comments[j].CreatedAt
	})

	return comments, nil
}

// PostComment posts a review comment on the given PR.
func PostComment(prNumber int, body string) error {
	tmp, err := os.CreateTemp("", "gh-ai-review-comment-*.md")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	defer func() { _ = tmp.Close() }()

	if _, err := tmp.WriteString(body); err != nil {
		return fmt.Errorf("write comment body: %w", err)
	}

	// Flush to disk before gh reads the file
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

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
