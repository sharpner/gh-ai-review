package context

import (
	"path/filepath"
	"strings"

	"github.com/sharpner/gh-ai-review/git"
)

const maxTestLines = 200

// resolveTests finds test files corresponding to changed files.
func resolveTests(changedFiles []string, loaded map[string]string, budget *Budget, root string) map[string]string {
	result := make(map[string]string)

	for _, file := range changedFiles {
		ext := filepath.Ext(file)
		base := strings.TrimSuffix(file, ext)

		candidates := []string{
			base + ".test" + ext,
			base + ".spec" + ext,
			base + "_test.go",
		}

		for _, candidate := range candidates {
			if _, ok := loaded[candidate]; ok {
				continue
			}
			if _, ok := result[candidate]; ok {
				continue
			}
			if !IsPathInRepo(root, candidate) {
				continue
			}

			content, err := git.Show("HEAD", candidate)
			if err != nil {
				continue
			}

			truncated := truncateToLines(content, maxTestLines)
			if !budget.CanAfford(len(truncated)) {
				continue
			}
			budget.Spend(len(truncated))
			result[candidate] = truncated
			break // one test file per source file
		}
	}
	return result
}
