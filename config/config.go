package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var validProviders = map[string]bool{
	"gemini": true,
	"codex":  true,
}

const (
	DefaultProvider        = "gemini"
	DefaultGeminiModel     = "gemini-3-flash-preview"
	DefaultCodexModel      = "gpt-5.4"
	DefaultModel           = DefaultGeminiModel
	DefaultMaxContextChars = 2_000_000
	DefaultConfigFile      = ".ai-review.yaml"
)

type Config struct {
	Provider        string   `yaml:"provider"`
	Model           string   `yaml:"model"`
	MaxContextChars int      `yaml:"max_context_chars"`
	AgentsDir       string   `yaml:"agents_dir"`
	ContextDocs     []string `yaml:"context_docs"`
	FocusedReviews  []string `yaml:"focused_reviews"`
}

func Load(path string) (Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	cfg.Provider = strings.ToLower(strings.TrimSpace(cfg.Provider))
	if !validProviders[cfg.Provider] {
		return cfg, fmt.Errorf("invalid provider %q in config (supported: gemini, codex)", cfg.Provider)
	}

	return cfg, nil
}

func defaults() Config {
	return Config{
		Provider:        DefaultProvider,
		Model:           DefaultModel,
		MaxContextChars: DefaultMaxContextChars,
		AgentsDir:       ".claude/agents",
		ContextDocs:     []string{"CLAUDE.md", "docs/code-standards.md", "docs/design-system.md"},
		FocusedReviews:  []string{"security", "usability", "mobile"},
	}
}
