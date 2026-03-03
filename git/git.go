package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Show returns the contents of a file at a given ref.
func Show(ref, path string) (string, error) {
	out, err := exec.Command("git", "show", ref+":"+path).Output()
	if err != nil {
		return "", fmt.Errorf("git show %s:%s: %w", ref, path, err)
	}
	return string(out), nil
}

// LsFiles returns tracked files matching the given patterns.
func LsFiles(patterns ...string) ([]string, error) {
	args := append([]string{"ls-files"}, patterns...)
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// Diff returns the diff between two refs.
func Diff(base, head string) (string, error) {
	out, err := exec.Command("git", "diff", base+"..."+head).Output()
	if err != nil {
		return "", fmt.Errorf("git diff %s...%s: %w", base, head, err)
	}
	return string(out), nil
}

// Root returns the repository root directory.
func Root() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
