package context

import "testing"

func TestCategorize_UI(t *testing.T) {
	cats := Categorize([]string{"src/components/Button.tsx"})
	if !cats.UI {
		t.Error("expected UI=true for .tsx file")
	}
	if !cats.Design {
		t.Error("expected Design=true for components/ file")
	}
}

func TestCategorize_UIExcludesTests(t *testing.T) {
	cats := Categorize([]string{"src/components/Button.test.tsx"})
	if cats.UI {
		t.Error("expected UI=false for .test.tsx file")
	}
	if !cats.Tests {
		t.Error("expected Tests=true for .test.tsx file")
	}
}

func TestCategorize_API(t *testing.T) {
	cats := Categorize([]string{"src/app/api/users/route.ts"})
	if !cats.API {
		t.Error("expected API=true for route.ts in api/")
	}
}

func TestCategorize_Auth(t *testing.T) {
	cats := Categorize([]string{"src/lib/auth.ts"})
	if !cats.Auth {
		t.Error("expected Auth=true for auth.ts")
	}
}

func TestCategorize_NewRoutes(t *testing.T) {
	cats := Categorize([]string{"app/dashboard/page.tsx"})
	if !cats.NewRoutes {
		t.Error("expected NewRoutes=true for app/*/page.tsx")
	}
}

func TestCategorize_Navigation(t *testing.T) {
	cats := Categorize([]string{"src/components/sidebar.tsx"})
	if !cats.Navigation {
		t.Error("expected Navigation=true for sidebar")
	}
}

func TestCategorize_DB(t *testing.T) {
	cats := Categorize([]string{"prisma/schema.prisma"})
	if !cats.DB {
		t.Error("expected DB=true for prisma file")
	}
}

func TestCategorize_Config(t *testing.T) {
	cats := Categorize([]string{"tsconfig.json"})
	if !cats.Config {
		t.Error("expected Config=true for .json file")
	}
}

func TestCategorize_EmptyStates(t *testing.T) {
	cats := Categorize([]string{"src/components/EmptyState.tsx"})
	if !cats.EmptyStates {
		t.Error("expected EmptyStates=true for EmptyState file")
	}
}

func TestDetectCRUDInDiff(t *testing.T) {
	tests := []struct {
		diff string
		want bool
	}{
		{"+ fetch('/api/users', { method: 'POST' })", true},
		{"+ const x = create()", true},
		{"+ DELETE FROM users", true},
		{"+ const x = 42", false},
		{"no crud here", false},
	}
	for _, tt := range tests {
		got := DetectCRUDInDiff(tt.diff)
		if got != tt.want {
			t.Errorf("DetectCRUDInDiff(%q) = %v, want %v", tt.diff[:20], got, tt.want)
		}
	}
}

func TestIsComplex(t *testing.T) {
	if !IsComplex(11, 100) {
		t.Error("expected complex for >10 files")
	}
	if !IsComplex(5, 501) {
		t.Error("expected complex for >500 lines")
	}
	if IsComplex(5, 100) {
		t.Error("expected not complex for 5 files / 100 lines")
	}
}

func TestFileCategories_String(t *testing.T) {
	cats := FileCategories{UI: true, API: true}
	s := cats.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
