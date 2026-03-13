package review

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/genai"
)

// StatusCheck represents a single check result.
type StatusCheck struct {
	Name   string
	OK     bool
	Detail string
}

// StatusResult holds all status checks for a provider.
type StatusResult struct {
	Provider string
	Model    string
	Checks   []StatusCheck
}

// OK returns true if all checks passed.
func (r StatusResult) OK() bool {
	for _, c := range r.Checks {
		if !c.OK {
			return false
		}
	}
	return len(r.Checks) > 0
}

// CheckStatus runs provider-specific health checks.
func CheckStatus(ctx context.Context, providerName, model string) StatusResult {
	result := StatusResult{Provider: providerName, Model: model}

	switch providerName {
	case ProviderGemini:
		result.Checks = checkGeminiStatus(ctx)
	case ProviderCodex:
		result.Checks = checkCodexStatus(ctx)
	default:
		result.Checks = []StatusCheck{{Name: "Provider", OK: false, Detail: fmt.Sprintf("unknown provider: %s", providerName)}}
	}

	return result
}

func checkGeminiStatus(ctx context.Context) []StatusCheck {
	var checks []StatusCheck

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		checks = append(checks, StatusCheck{Name: "API Key", OK: false, Detail: "GOOGLE_API_KEY not set"})
		return checks
	}
	checks = append(checks, StatusCheck{Name: "API Key", OK: true, Detail: "GOOGLE_API_KEY is set"})

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := genai.NewClient(checkCtx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		checks = append(checks, StatusCheck{Name: "Connectivity", OK: false, Detail: fmt.Sprintf("client creation failed: %v", err)})
		return checks
	}

	_, err = client.Models.Get(checkCtx, "gemini-2.0-flash", nil)
	if err != nil {
		checks = append(checks, StatusCheck{Name: "Connectivity", OK: false, Detail: fmt.Sprintf("API call failed: %v", err)})
		return checks
	}

	checks = append(checks, StatusCheck{Name: "Connectivity", OK: true, Detail: "API reachable, key valid"})
	return checks
}

func checkCodexStatus(ctx context.Context) []StatusCheck {
	var checks []StatusCheck

	codexBin, err := CodexPath()
	if err != nil {
		checks = append(checks, StatusCheck{Name: "Binary", OK: false, Detail: "codex not found on PATH"})
		return checks
	}
	checks = append(checks, StatusCheck{Name: "Binary", OK: true, Detail: fmt.Sprintf("found at %s", codexBin)})

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(checkCtx, codexBin, "--version")
	out, err := cmd.Output()
	if err != nil {
		checks = append(checks, StatusCheck{Name: "Version", OK: false, Detail: fmt.Sprintf("codex --version failed: %v", err)})
		return checks
	}

	version := strings.TrimSpace(string(out))
	checks = append(checks, StatusCheck{Name: "Version", OK: true, Detail: version})
	return checks
}
