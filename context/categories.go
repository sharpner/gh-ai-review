package context

import (
	"fmt"
	"regexp"
	"strings"
)

// FileCategories indicates which types of files are present in the changeset.
type FileCategories struct {
	UI          bool
	API         bool
	Auth        bool
	Mobile      bool
	Design      bool
	Routes      bool
	Config      bool
	Tests       bool
	Docs        bool
	DB          bool
	Copy        bool
	NewRoutes   bool
	Navigation  bool
	EmptyStates bool
}

var newRoutesRe = regexp.MustCompile(`app/.*page\.tsx$|app/.*route\.ts$`)
var emptyStatesRe = regexp.MustCompile(`(?i)empty|emptystate|getting-?started|onboarding`)

// Categorize detects file categories from a list of changed file paths.
func Categorize(files []string) FileCategories {
	var cats FileCategories
	for _, f := range files {
		lower := strings.ToLower(f)
		if hasAny(lower, ".tsx", ".jsx", ".css", "component") && !hasAny(lower, ".test.", ".spec.") {
			cats.UI = true
		}
		if hasAny(lower, "route.ts", "api/", "handler", "middleware") {
			cats.API = true
		}
		if hasAny(lower, "auth", "login", "session", "password", "token") {
			cats.Auth = true
		}
		if hasAny(lower, "mobile", "responsive", "breakpoint") {
			cats.Mobile = true
		}
		if hasAny(lower, "components/", "design", "theme", "color", "style") {
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
		if hasAny(lower, "copy", "text", "message", "label", "content") {
			cats.Copy = true
		}
		if newRoutesRe.MatchString(f) {
			cats.NewRoutes = true
		}
		if hasAny(lower, "sidebar", "nav", "header", "menu", "breadcrumb") {
			cats.Navigation = true
		}
		if emptyStatesRe.MatchString(f) {
			cats.EmptyStates = true
		}
	}
	return cats
}

// DetectCRUDInDiff checks if the diff contains CRUD operation patterns.
func DetectCRUDInDiff(diff string) bool {
	return strings.Contains(diff, "create") ||
		strings.Contains(diff, "delete") ||
		strings.Contains(diff, "destroy") ||
		strings.Contains(diff, "remove") ||
		strings.Contains(diff, "POST") ||
		strings.Contains(diff, "DELETE") ||
		strings.Contains(diff, "PUT") ||
		strings.Contains(diff, "PATCH")
}

// IsComplex returns true if the PR is considered complex.
func IsComplex(fileCount, totalChanges int) bool {
	return fileCount > 10 || totalChanges > 500
}

// String renders the file categories block for prompt inclusion.
func (c FileCategories) String() string {
	yn := func(b bool) string {
		if b {
			return "YES"
		}
		return "no"
	}
	return fmt.Sprintf(`# File Categories
- UI Components (tsx/css): %s
- API Routes: %s
- Auth/Security: %s
- Mobile/Responsive: %s
- Design System: %s
- Copy/Text: %s
- New Routes/Pages: %s
- Navigation Changes: %s
- Empty States: %s`,
		yn(c.UI), yn(c.API), yn(c.Auth), yn(c.Mobile), yn(c.Design),
		yn(c.Copy), yn(c.NewRoutes), yn(c.Navigation), yn(c.EmptyStates))
}

func hasAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
