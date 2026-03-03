package context

import (
	"path/filepath"

	"github.com/sharpner/gh-ai-review/git"
)

const (
	maxSiblings      = 6
	maxSiblingsPerDir = 4
	maxSiblingLines  = 150
	maxSiblingBytes  = 30_000
)

// resolveSiblings finds sibling files in the same directories as changed files.
func resolveSiblings(changedFiles []string, loaded map[string]string, budget *Budget) map[string]string {
	result := make(map[string]string)
	total := 0
	seenDirs := make(map[string]bool)

	for _, file := range changedFiles {
		if total >= maxSiblings {
			break
		}

		dir := filepath.Dir(file)
		if seenDirs[dir] {
			continue
		}
		seenDirs[dir] = true

		files, err := git.LsFiles(dir)
		if err != nil {
			continue
		}

		dirCount := 0
		for _, sibling := range files {
			if total >= maxSiblings || dirCount >= maxSiblingsPerDir {
				break
			}

			ext := filepath.Ext(sibling)
			if ext != ".ts" && ext != ".tsx" && ext != ".go" {
				continue
			}
			if _, ok := loaded[sibling]; ok {
				continue
			}
			if _, ok := result[sibling]; ok {
				continue
			}

			content, err := git.Show("HEAD", sibling)
			if err != nil {
				continue
			}

			if len(content) > maxSiblingBytes {
				continue
			}

			truncated := truncateToLines(content, maxSiblingLines)
			if !budget.CanAfford(len(truncated)) {
				continue
			}
			budget.Spend(len(truncated))
			result[sibling] = truncated
			total++
			dirCount++
		}
	}
	return result
}

