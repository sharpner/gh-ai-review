package context

import "testing"

func TestIsPathInRepo(t *testing.T) {
	root := "/home/user/project"

	tests := []struct {
		path string
		want bool
	}{
		{"src/main.go", true},
		{"src/lib/utils.ts", true},
		{"../../../etc/passwd", false},
		{"src/../../outside", false},
		{"normal/file.ts", true},
		// Sibling directory attack: /home/user/project-secrets should NOT match /home/user/project
		{"../" + "project-secrets/secret.md", false},
	}
	for _, tt := range tests {
		got := IsPathInRepo(root, tt.path)
		if got != tt.want {
			t.Errorf("IsPathInRepo(%q, %q) = %v, want %v", root, tt.path, got, tt.want)
		}
	}
}
