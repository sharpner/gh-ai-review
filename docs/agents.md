# Writing Agent Personas

Agents are markdown files in your repo's agents directory (default: `.claude/agents/`). Each file defines a reviewer persona that Gemini impersonates during `--agent` or `--full` reviews.

## How It Works

1. You write a `.md` file describing a reviewer persona
2. The tool reads it and tells Gemini: "You are this persona. Review this PR."
3. Gemini gets the full context (files, imports, siblings, tests, diff) and reviews from that perspective
4. The review is posted as a PR comment

## File Location

```
your-repo/
├── .claude/
│   └── agents/
│       ├── go-expert.md
│       ├── security-reviewer.md
│       └── your-custom-agent.md
└── .ai-review.yaml          # agents_dir: .claude/agents (default)
```

The filename (minus `.md`) is the agent name used with `--agent`:

```bash
gh ai-review 42 --agent go-expert
gh ai-review 42 --agent security-reviewer
```

## Writing an Agent

An agent file is plain markdown. Write it as if you're briefing a specialist:

1. **Who they are** — role, years of experience, specialization
2. **What to check** — numbered list of specific review areas
3. **Rating system** — how to rate each area
4. **Output format** — verdict format, what to include

### Template

```markdown
You are a [ROLE] with [EXPERIENCE] specializing in [DOMAIN].

Your review focuses on:

1. **[Area 1]** — [what to check]
2. **[Area 2]** — [what to check]
3. **[Area 3]** — [what to check]

Rate each area: [RATING_SCALE]

Format your response with:
- **Verdict:** PASS | NEEDS WORK | FAIL
- Specific code references with file:line
- Concrete fix suggestions (not vague advice)
```

### Key Rules

- **Be specific.** "Check for SQL injection" is better than "check for security issues"
- **Define the output format.** Gemini follows format instructions well
- **Include a verdict line.** The tool extracts `**Verdict:** PASS|NEEDS WORK|FAIL` for terminal output
- **Keep it under 500 words.** The persona prompt is part of the context budget

## Included Agents

### `go-expert`

Senior Go engineer reviewing for idiomatic Go, concurrency correctness, resource management, API surface, performance, stdlib usage, and error messages.

```bash
gh ai-review 42 --agent go-expert
```

### `security-reviewer`

Security engineer checking secrets handling, command injection, temp file safety, input validation, dependency security, output safety, and error information leakage.

```bash
gh ai-review 42 --agent security-reviewer
```

## Agent Ideas

Here are agents you might want to create for your project:

| Agent | Focus |
|-------|-------|
| `frontend-reviewer` | React patterns, hooks rules, component composition, prop drilling |
| `api-reviewer` | REST conventions, error responses, pagination, rate limiting |
| `database-reviewer` | Query performance, N+1, migrations, indexing |
| `accessibility-reviewer` | ARIA, keyboard nav, screen readers, color contrast |
| `test-reviewer` | Test coverage, test quality, mocking patterns, edge cases |
| `docs-reviewer` | API docs accuracy, README completeness, code comments |
| `performance-reviewer` | Bundle size, render performance, caching, lazy loading |
| `devops-reviewer` | Dockerfile best practices, CI config, env var handling |

## How Agents Work in `--full` Mode

During a `--full` review:

1. Phase 1 runs a generic review
2. Gemini sees your installed agents and recommends relevant ones
3. Phase 2 runs every recommended agent that has a matching `.md` file
4. Each agent review is posted as a separate PR comment

```
$ gh ai-review 42 --full

PHASE 1/3: Generic Code Review
  → Gemini recommends: go-expert, security-reviewer, frontend-reviewer

PHASE 2/3: Subagent Impersonation Reviews
  Running: go-expert, security-reviewer
  Skipped (no file): frontend-reviewer
  ✓ go-expert: PASS
  ✓ security-reviewer: LOW risk

PHASE 3/3: Focused Reviews
  → security, usability (auto-detected)
```

## Tips

- **Start with 2-3 agents.** You can always add more later.
- **Tailor to your stack.** A Next.js project needs different agents than a Go CLI.
- **Use `--dry-run` to test.** See what prompt Gemini receives before burning API credits.
- **Check the verdict extraction.** The tool looks for `**Verdict:**`, `**Risk Level:**`, etc. — make sure your agent format matches.
