package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Show returns the contents of a file at a given ref.
func Show(ref, path string) (string, error) {
	cmd := exec.Command("git", "show", "--", ref+":"+path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git show %s:%s: %w (%s)", ref, path, err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// LsFiles returns tracked files matching the given patterns.
func LsFiles(patterns ...string) ([]string, error) {
	args := append([]string{"ls-files", "--"}, patterns...)
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w (%s)", err, strings.TrimSpace(stderr.String()))
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
	cmd := exec.Command("git", "diff", "--", base+"..."+head)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff %s...%s: %w (%s)", base, head, err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// Root returns the repository root directory.
func Root() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git root: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}
