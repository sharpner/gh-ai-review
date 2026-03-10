package review

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CallCodex sends a prompt to the Codex CLI and returns the response text.
// It invokes the codex binary in non-interactive mode via os/exec.
// Authentication is handled by the Codex CLI's cached OAuth credentials.
func CallCodex(ctx context.Context, model, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "codex", "exec",
		"-m", model,
		"-s", "read-only",
		"--skip-git-repo-check",
		"-o", "/dev/stdout",
		"-",
	)

	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("codex: %w", ctx.Err())
		}
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return "", fmt.Errorf("codex: %s: %w", errMsg, err)
		}
		return "", fmt.Errorf("codex: %w", err)
	}

	text := strings.TrimSpace(stdout.String())
	if text == "" {
		return "", fmt.Errorf("empty response from Codex")
	}

	return text, nil
}

// CodexAvailable checks if the codex binary is on PATH.
func CodexAvailable() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}
