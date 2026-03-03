# gh-ai-review

A `gh` CLI extension for AI-powered pull request reviews using Google Gemini.

Replaces fragile bash scripts with a single Go binary — proper parallelism, direct Gemini SDK, usable across all projects.

## Install

```bash
gh extension install sharpner/gh-ai-review
```

## Prerequisites

- [`gh`](https://cli.github.com/) CLI installed and authenticated
- `GOOGLE_API_KEY` environment variable set (Gemini API key)

## Usage

```bash
# Generic review of a PR
gh ai-review 620

# Full review loop (generic + focused reviews)
gh ai-review 620 --full

# Agent impersonation (use a persona prompt)
gh ai-review 620 --agent security-pentest

# Custom focus area
gh ai-review 620 --focus "backward compatibility"

# Dry run — print the prompt without calling Gemini
gh ai-review 620 --dry-run

# Override Gemini model
gh ai-review 620 --model gemini-2.5-pro
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--full` | Run full review loop (generic + focused) | `false` |
| `--agent <name>` | Agent persona to impersonate | — |
| `--focus <area>` | Custom focus area for review | — |
| `--dry-run` | Print prompt, skip Gemini call | `false` |
| `--model <model>` | Gemini model to use | `gemini-2.5-flash` |

## Configuration

Create `.ai-review.yaml` in your repo root (optional):

```yaml
model: gemini-2.5-flash
max_context_chars: 900000
agents_dir: .ai-review/agents
context_docs:
  - CLAUDE.md
  - docs/code-standards.md
focused_reviews:
  - security
  - usability
  - mobile
```

## How It Works

1. **Context Building** — Fetches PR diff, changed files, and project docs via `gh` and `git`
2. **Prompt Assembly** — Builds a review prompt with file categories, size budgeting, and context
3. **Gemini Call** — Sends the prompt to Gemini and parses the structured response
4. **PR Comment** — Posts the review as a PR comment via `gh`

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter
make install  # Build + install as gh extension
make clean    # Remove binary
```

## License

MIT
