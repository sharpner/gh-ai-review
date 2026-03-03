package context

import "path/filepath"

var extToLang = map[string]string{
	".ts":     "typescript",
	".tsx":    "typescript",
	".js":     "javascript",
	".jsx":    "javascript",
	".go":     "go",
	".css":    "css",
	".json":   "json",
	".md":     "markdown",
	".yaml":   "yaml",
	".yml":    "yaml",
	".sh":     "bash",
	".prisma": "prisma",
}

var binaryExts = map[string]bool{
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".svg":   true,
	".ico":   true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".eot":   true,
	".pdf":   true,
}

// LangForExt returns the language name for a file extension.
func LangForExt(path string) string {
	ext := filepath.Ext(path)
	return extToLang[ext]
}

// IsBinary returns true if the file path has a binary extension.
func IsBinary(path string) bool {
	ext := filepath.Ext(path)
	return binaryExts[ext]
}
