package context

import "testing"

func TestLangForExt(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"src/app.tsx", "typescript"},
		{"main.go", "go"},
		{"style.css", "css"},
		{"data.json", "json"},
		{"README.md", "markdown"},
		{"deploy.sh", "bash"},
		{"unknown.xyz", ""},
	}
	for _, tt := range tests {
		got := LangForExt(tt.path)
		if got != tt.want {
			t.Errorf("LangForExt(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestIsBinary(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"logo.png", true},
		{"photo.jpg", true},
		{"icon.svg", true},
		{"font.woff2", true},
		{"doc.pdf", true},
		{"main.go", false},
		{"app.tsx", false},
		{"README.md", false},
	}
	for _, tt := range tests {
		got := IsBinary(tt.path)
		if got != tt.want {
			t.Errorf("IsBinary(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
