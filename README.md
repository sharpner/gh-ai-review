# gh-ai-review

A `gh` CLI extension for AI-powered pull request reviews using Google Gemini.

Replaces fragile bash scripts with a single Go binary — proper parallelism, direct Gemini SDK, budget-aware context building, and agent impersonation.

## Install

```bash
gh extension install sharpner/gh-ai-review
```

## Prerequisites

- [`gh`](https://cli.github.com/) CLI installed and authenticated
- `GOOGLE_API_KEY` environment variable set (Gemini API key)

## Usage

```bash
# Generic review (+ auto-detected focused reviews)
gh ai-review 620

# Full review loop (generic → agent impersonation → focused)
gh ai-review 620 --full

# Agent impersonation (uses a persona prompt from .claude/agents/)
gh ai-review 620 --agent security-pentest-reviewer

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
| `--full` | Run full 3-phase review loop | `false` |
| `--agent <name>` | Agent persona to impersonate | — |
| `--focus <area>` | Custom focus area for review | — |
| `--dry-run` | Print prompt, skip Gemini call | `false` |
| `--model <model>` | Gemini model to use | `gemini-3-flash-preview` |
| `--version` | Print version | — |

## Review Modes

### Default (`gh ai-review <PR>`)

1. Builds review context (PR diff, changed files, project docs)
2. Runs a **generic code review** (architecture, correctness, performance, security, testing, style)
3. Auto-detects **focused reviews** based on file categories:
   - **Security** — triggered by API routes or auth files
   - **Usability** — triggered by UI components, new routes, or navigation
   - **Mobile** — triggered by UI, mobile, or design files
4. Posts a combined PR comment with all results

### Agent Impersonation (`--agent <name>`)

Reads an agent persona from `<agents_dir>/<name>.md` and runs a deep analysis review.

Agent context includes:
- All files from the default context
- **Import resolution** — follows TypeScript/JavaScript imports (relative `./` and alias `@/`)
- **Sibling files** — other code files in the same directories
- **Test files** — matching `.test.*`, `.spec.*`, `_test.go` patterns

If the agent file doesn't exist, lists all available agents.

### Full Loop (`--full`)

Runs all 3 phases sequentially, posting each as a separate PR comment:

| Phase | What | Parallelism |
|-------|------|-------------|
| 1/3 | Generic code review | Single |
| 2/3 | Agent impersonation for recommended reviewers | Parallel (errgroup) |
| 3/3 | Focused reviews (security, usability, mobile) | Parallel (errgroup) |

Phase 2 extracts recommended reviewers from the generic review output and impersonates each agent that has a matching `.md` file in the agents directory.

## Context Building

The tool builds a budget-aware review context:

1. **Project docs** — loads files from `context_docs` config (e.g. `CLAUDE.md`, code standards)
2. **Changed files** — fetches content via `git show HEAD:<path>`, skips binaries
3. **File categorization** — 14 categories (UI, API, Auth, Database, Config, Tests, Docs, Mobile, Design, Copy, NewRoutes, Navigation, EmptyStates, CRUD detection)
4. **PR diff** — raw diff included in budget

Total context is bounded by `max_context_chars` (default 2M chars, ~500K tokens).

## Configuration

Create `.ai-review.yaml` in your repo root (optional — all fields have sensible defaults):

```yaml
model: gemini-3-flash-preview
max_context_chars: 2000000
agents_dir: .claude/agents
context_docs:
  - CLAUDE.md
  - docs/code-standards.md
  - docs/design-system.md
focused_reviews:
  - security
  - usability
  - mobile
```

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run golangci-lint
make install  # Build + install as gh extension
make clean    # Remove binary
```

### Project Structure

```
├── main.go              # CLI entry point, flag parsing, dispatch
├── config/              # YAML config loading with defaults
├── git/                 # Git command wrappers (show, ls-files, diff)
├── github/              # gh CLI integration (PR fetch, comment posting)
├── context/             # Budget-aware context assembly
│   ├── context.go       # Build + BuildAgent orchestration
│   ├── categories.go    # 14-category file classification
│   ├── budget.go        # Token budget tracking
│   ├── lang.go          # Extension→language mapping, binary detection
│   ├── imports.go       # TypeScript/JS import resolution
│   ├── siblings.go      # Sibling file discovery
│   └── tests.go         # Test file resolution
├── review/              # Gemini SDK client + review runners
│   ├── review.go        # RunGeneric, RunFocused, RunAgent, RunFull
│   ├── prompts.go       # Prompt templates + builders
│   └── gemini.go        # Gemini API client
└── output/              # Terminal + PR comment formatting
```

## Release

Pushing a `v*` tag triggers the release workflow which cross-compiles binaries via [`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile).

```bash
git tag v0.1.0
git push origin v0.1.0
```

## License

MIT
