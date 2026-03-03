package context

import "strings"

// FileCategories indicates which types of files are present in the changeset.
type FileCategories struct {
	UI      bool
	API     bool
	Auth    bool
	Mobile  bool
	Design  bool
	Routes  bool
	Config  bool
	Tests   bool
	Docs    bool
	DB      bool
}

// Categorize detects file categories from a list of changed file paths.
func Categorize(files []string) FileCategories {
	var cats FileCategories
	for _, f := range files {
		lower := strings.ToLower(f)
		if hasAny(lower, ".tsx", ".jsx", "component") {
			cats.UI = true
		}
		if hasAny(lower, "route.ts", "api/", "handler") {
			cats.API = true
		}
		if hasAny(lower, "auth", "login", "session") {
			cats.Auth = true
		}
		if hasAny(lower, "mobile", "responsive") {
			cats.Mobile = true
		}
		if hasAny(lower, "design", "theme", "style") {
			cats.Design = true
		}
		if hasAny(lower, "route", "middleware") {
			cats.Routes = true
		}
		if hasAny(lower, ".yaml", ".yml", ".json", ".toml", ".env") {
			cats.Config = true
		}
		if hasAny(lower, "_test.go", ".test.", ".spec.") {
			cats.Tests = true
		}
		if hasAny(lower, ".md", "doc") {
			cats.Docs = true
		}
		if hasAny(lower, "migration", "schema", "prisma") {
			cats.DB = true
		}
	}
	return cats
}

func hasAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
