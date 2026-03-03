package github

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2"
)

type PRInfo struct {
	Number    int
	Title     string
	Body      string
	BaseRef   string
	HeadRef   string
	RepoOwner string
	RepoName  string
}

type PRData struct {
	Info         PRInfo
	Diff         string
	ChangedFiles []string
}

// FetchPR retrieves PR metadata, diff, and changed file list.
func FetchPR(prNumber int) (PRData, error) {
	return PRData{}, fmt.Errorf("not implemented")
}

// PostComment posts a review comment on the given PR.
func PostComment(prNumber int, body string) error {
	return fmt.Errorf("not implemented")
}

// RepoSlug returns "owner/repo" for the current directory.
func RepoSlug() (string, error) {
	stdout, _, err := gh.Exec("repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}
