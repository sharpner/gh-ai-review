package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Show returns the contents of a file at a given ref.
func Show(ref, path string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// LsFiles returns tracked files matching the given patterns.
func LsFiles(patterns ...string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// Diff returns the diff between two refs.
func Diff(base, head string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// Root returns the repository root directory.
func Root() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
