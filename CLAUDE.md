# CLAUDE.md — gh-ai-review

Go CLI tool that runs AI-powered code reviews on GitHub PRs. Supports Gemini and Codex (GPT-5.4 via ChatGPT Pro) as LLM providers.

## Quick Commands

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run golangci-lint
make install  # Build + install as gh extension
```

## Package Responsibilities

| Package | What it does |
|---------|-------------|
| `main` | Flag parsing, dispatch |
| `config` | `.ai-review.yaml` parsing + defaults |
| `github` | PR data fetching, comment posting (via go-gh) |
| `git` | `git show`, `git ls-files` wrappers |
| `context` | Build review context, size budgeting, file categorization |
| `review` | LLM provider abstraction, API calls, prompt templates (generic/focused/agent) |
| `output` | Terminal output, result parsing |

## Dependencies (3 only)

- `google.golang.org/genai` — Gemini SDK (GA)
- `github.com/cli/go-gh/v2` — GitHub CLI library
- `gopkg.in/yaml.v3` — Config parsing

Codex provider uses `os/exec` wrapper around the `codex` CLI binary (no Go SDK dependency).

No CLI framework. Stdlib `flag` only.

## Go Conventions

- Guard clauses everywhere, no `else`
- No utils packages — good package design
- YAGNI — only build what's needed now
- Composition over inheritance
- No mocking in tests — real git repos in testdata
- Concise over clever
