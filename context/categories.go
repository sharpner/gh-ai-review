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
		switch {
		case hasAny(lower, ".tsx", ".jsx", "component"):
			cats.UI = true
		case hasAny(lower, "route.ts", "api/", "handler"):
			cats.API = true
		case hasAny(lower, "auth", "login", "session"):
			cats.Auth = true
		case hasAny(lower, "mobile", "responsive"):
			cats.Mobile = true
		case hasAny(lower, "design", "theme", "style"):
			cats.Design = true
		case hasAny(lower, "route", "middleware"):
			cats.Routes = true
		case hasAny(lower, ".yaml", ".yml", ".json", ".toml", ".env"):
			cats.Config = true
		case hasAny(lower, "_test.go", ".test.", ".spec."):
			cats.Tests = true
		case hasAny(lower, ".md", "doc"):
			cats.Docs = true
		case hasAny(lower, "migration", "schema", "prisma"):
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
