package review

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// MakeCodexCaller returns a CallLLM function that uses the given absolute binary path.
// Resolving the path at init time prevents PATH manipulation between check and execution.
func MakeCodexCaller(codexBin string) CallLLM {
	return func(ctx context.Context, model, prompt string) (string, error) {
		return callCodex(ctx, codexBin, model, prompt)
	}
}

// callCodex sends a prompt to the Codex CLI and returns the response text.
// It invokes the codex binary in non-interactive mode via os/exec.
// Authentication is handled by the Codex CLI's cached OAuth credentials.
//
// Security: the sandbox is set to read-only (-s read-only) so the Codex agent
// cannot modify files or execute arbitrary commands. The prompt contains PR
// content which is untrusted input; read-only sandbox prevents prompt injection
// from escalating to code execution. Environment is filtered to prevent leaking
// sensitive vars (API keys, tokens).
func callCodex(ctx context.Context, codexBin, model, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, codexBin, "exec",
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

// CodexPath returns the resolved absolute path of the codex binary.
// This pins the binary at startup to prevent PATH manipulation later.
func CodexPath() (string, error) {
	return exec.LookPath("codex")
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
