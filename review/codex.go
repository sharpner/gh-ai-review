package review

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CallCodex sends a prompt to the Codex CLI and returns the response text.
// It invokes the codex binary in non-interactive mode via os/exec.
// Authentication is handled by the Codex CLI's cached OAuth credentials.
//
// Security: the sandbox is set to read-only (-s read-only) so the Codex agent
// cannot modify files or execute arbitrary commands. The prompt contains PR
// content which is untrusted input; read-only sandbox prevents prompt injection
// from escalating to code execution.
func CallCodex(ctx context.Context, model, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "codex", "exec",
		"-m", model,
		"-s", "read-only",
		"--skip-git-repo-check",
		"-",
	)

	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = filteredEnv()

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

// filteredEnv returns a minimal set of env vars for the codex subprocess.
// This prevents leaking sensitive env vars (API keys, tokens) to the external binary.
func filteredEnv() []string {
	allow := map[string]bool{
		"HOME": true, "USER": true, "PATH": true, "SHELL": true,
		"LANG": true, "TERM": true, "TMPDIR": true, "XDG_CONFIG_HOME": true,
	}
	var env []string
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		if !allow[key] {
			continue
		}
		env = append(env, e)
	}
	return env
}
