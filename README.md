# gh-ai-review

AI-powered pull request reviews using Google Gemini, as a `gh` CLI extension.

One command reviews your PR — generic code review, security/usability/mobile focused reviews, and agent impersonation with deep context. All posted as PR comments.

## Quick Start

```bash
# 1. Install
gh extension install sharpner/gh-ai-review

# 2. Set your Gemini API key
export GOOGLE_API_KEY="your-key-here"

# 3. Review a PR
gh ai-review 42
```

That's it. The tool reads your repo, builds context, calls Gemini, and posts a review comment.

## Prerequisites

- [`gh`](https://cli.github.com/) CLI installed and authenticated (`gh auth login`)
- `GOOGLE_API_KEY` environment variable ([get one here](https://aistudio.google.com/apikey))

## Usage

```bash
# Simple review (generic + auto-detected focused reviews)
gh ai-review 42

# Full 3-phase loop (generic → agent impersonation → focused)
gh ai-review 42 --full

# Single agent impersonation
gh ai-review 42 --agent security-reviewer

# Custom focus
gh ai-review 42 --focus "error handling"

# Dry run — see the prompt without calling Gemini
gh ai-review 42 --dry-run

# Use a different model
gh ai-review 42 --model gemini-2.5-pro
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--full` | Run full 3-phase review loop | `false` |
| `--agent <name>` | Run a specific agent persona | — |
| `--focus <area>` | Custom focus area | — |
| `--dry-run` | Print prompt, skip Gemini call | `false` |
| `--model <model>` | Override Gemini model | `gemini-3-flash-preview` |
| `--version` | Print version | — |

## How It Works

### Default Mode (`gh ai-review <PR>`)

1. Fetches PR diff, changed file contents, and project docs (e.g. `CLAUDE.md`)
2. Categorizes files (UI, API, Auth, Mobile, etc.)
3. Runs a **generic code review** via Gemini
4. Auto-triggers **focused reviews** if relevant files detected:
   - **Security** — API routes or auth files changed
   - **Usability** — UI components, new routes, navigation
   - **Mobile** — UI, mobile, or design files
5. Posts everything as a single PR comment

### Agent Mode (`--agent <name>`)

Reads a persona prompt from `.claude/agents/<name>.md` and runs a deep-analysis review with extra context:

- Full file contents of changed files
- Resolved imports (TypeScript `./` and `@/` aliases)
- Sibling files from same directories
- Matching test files (`.test.*`, `.spec.*`, `_test.go`)

See [docs/agents.md](docs/agents.md) for how to write your own agents.

### Full Loop (`--full`)

The money mode. Runs 3 phases, each posted as separate PR comments:

```
Phase 1/3: Generic Code Review              → 1 comment
Phase 2/3: Agent Impersonation (parallel)   → 1 comment per agent
Phase 3/3: Focused Reviews (parallel)       → 1 combined comment
```

Phase 2 is the magic: Gemini's generic review recommends agents, and the tool automatically runs every recommended agent that has a matching `.md` file in your agents directory.

## Setup for Your Repo

### 1. Config (optional)

Create `.ai-review.yaml` in your repo root to override defaults:

```yaml
# All fields optional — these are the defaults
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

### 2. Agent Personas (recommended)

Create `.claude/agents/<name>.md` files with review personas. The tool discovers these automatically and Gemini will recommend them during `--full` reviews.

```bash
mkdir -p .claude/agents
```

Two starter agents are included in this repo. See [docs/agents.md](docs/agents.md) for the full guide on writing agents.

### 3. Project Docs (recommended)

The tool loads project docs listed in `context_docs` into every review prompt. This gives Gemini your coding standards, architecture decisions, and conventions.

Good candidates:
- `CLAUDE.md` — project rules and conventions
- `docs/code-standards.md` — coding style guide
- `docs/design-system.md` — UI/component patterns

## Context Budget

The tool is budget-aware. All context (docs + files + diff) is bounded by `max_context_chars` (default 2M chars, ~500K tokens). Files are loaded in order until the budget is exhausted — large PRs gracefully skip lower-priority files instead of failing.

## Development

```bash
make build    # Build binary
make test     # Run tests (with -race)
make lint     # Run golangci-lint
make install  # Build + install as gh extension
make clean    # Remove binary
```

### Project Structure

```
├── main.go              # CLI entry, flag parsing, dispatch
├── config/              # YAML config with defaults
├── git/                 # Git wrappers (show, ls-files, diff)
├── github/              # gh CLI integration (PR fetch, comments)
├── context/             # Budget-aware context assembly
│   ├── context.go       #   Build + BuildAgent
│   ├── categories.go    #   14-category file classification
│   ├── budget.go        #   Token budget tracking
│   ├── lang.go          #   Extension→language, binary detection
│   ├── imports.go       #   TS/JS import resolution
│   ├── siblings.go      #   Sibling file discovery
│   └── tests.go         #   Test file resolution
├── review/              # Gemini client + review runners
│   ├── review.go        #   RunGeneric, RunFocused, RunAgent, RunFull
│   ├── prompts.go       #   Prompt templates
│   └── gemini.go        #   Gemini SDK (sync.Once client)
├── output/              # Terminal + PR comment formatting
└── .claude/agents/      # Agent persona prompts
```

## Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

Cross-compiles via [`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile) on tag push.

## License

MIT
