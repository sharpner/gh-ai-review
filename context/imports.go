package context

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sharpner/gh-ai-review/git"
)

var importRe = regexp.MustCompile(`^import.*from\s+['"]([^'"]+)['"]`)

const (
	maxImports       = 10
	maxImportsPerFile = 5
	maxImportLines   = 200
)

// resolveImports finds imported files referenced by changed files that are not already loaded.
func resolveImports(changedFiles []string, loaded map[string]string, budget *Budget) map[string]string {
	result := make(map[string]string)
	total := 0

	for _, file := range changedFiles {
		ext := filepath.Ext(file)
		if ext != ".ts" && ext != ".tsx" && ext != ".js" && ext != ".jsx" {
			continue
		}
		if total >= maxImports {
			break
		}

		content, ok := loaded[file]
		if !ok {
			continue
		}

		perFile := 0
		for _, line := range strings.Split(content, "\n") {
			if total >= maxImports || perFile >= maxImportsPerFile {
				break
			}

			m := importRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			importPath := m[1]

			// Only resolve relative (./) and alias (@/) imports
			if !strings.HasPrefix(importPath, ".") && !strings.HasPrefix(importPath, "@/") {
				continue
			}

			var resolved string
			if strings.HasPrefix(importPath, "@/") {
				resolved = "src/" + strings.TrimPrefix(importPath, "@/")
			} else {
				resolved = filepath.Join(filepath.Dir(file), importPath)
			}

			found := tryResolve(resolved, loaded, result, budget)
			if found {
				total++
				perFile++
			}
		}
	}
	return result
}

func tryResolve(base string, loaded, result map[string]string, budget *Budget) bool {
	extensions := []string{"", ".ts", ".tsx", ".js", ".jsx", "/index.ts", "/index.tsx"}
	for _, ext := range extensions {
		candidate := base + ext
		if _, ok := loaded[candidate]; ok {
			continue
		}
		if _, ok := result[candidate]; ok {
			continue
		}

		content, err := git.Show("HEAD", candidate)
		if err != nil {
			continue
		}

		truncated := truncateToLines(content, maxImportLines)
		if !budget.CanAfford(len(truncated)) {
			continue
		}
		budget.Spend(len(truncated))
		result[candidate] = truncated
		return true
	}
	return false
}

func truncateToLines(s string, maxLines int) string {
	lines := strings.SplitN(s, "\n", maxLines+1)
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}
