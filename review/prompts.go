package review

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sharpner/gh-ai-review/context"
	gh "github.com/sharpner/gh-ai-review/github"
)

const genericInstructionsHeader = `
# Review Instructions

You are a senior code reviewer. Review the diff IN CONTEXT of the full file contents.

**Review for:**
1. Security vulnerabilities (SQL injection, XSS, auth issues)
2. Error handling (silent failures, missing validation)
3. Type safety issues
4. Architecture concerns (does it fit existing patterns?)
5. Performance issues
6. Consistency with project standards (CLAUDE.md, code-standards.md)
7. Existing discussion (acknowledge resolved points, don't repeat them)
`

const genericInstructionsFooter = `
**Format EXACTLY as:**

## General Review

**Verdict:** PASS | NEEDS WORK | FAIL

### Critical Issues
- [List or "None found"]

### Warnings
- [List or "None"]

### Suggestions
- [List or "None"]

### Recommended Reviewers
- [ ] ` + "`agent-name`" + `: Reason

### Summary
[1-2 sentences]

Be concise. Only real issues, not style preferences.
`

const securityInstructions = `
# Security-Focused Review
You are a security specialist. Focus EXCLUSIVELY on security.
Check: SQL/NoSQL injection, XSS, CSRF, auth bypasses, sensitive data exposure,
input validation gaps, path traversal, IDOR, missing rate limiting, unsafe eval/dangerouslySetInnerHTML.

**Format EXACTLY as:**
## Security Review
**Risk Level:** LOW | MEDIUM | HIGH | CRITICAL
### Vulnerabilities Found
- [file:line references or "None found"]
### Input Validation Gaps
- [List or "None"]
### Recommendations
- [List or "None"]
Only real security issues. No style/architecture.
`

const usabilityInstructions = `
# Usability-Focused Review
You are a UX/accessibility specialist. Focus EXCLUSIVELY on usability.
Check: Keyboard nav (Tab/Enter/Escape), ARIA attributes, screen reader support,
loading states, error states, empty states, focus management, touch targets (44px min),
color contrast, form UX.

**Format EXACTLY as:**
## Usability Review
**Accessibility Score:** GOOD | NEEDS WORK | POOR
### Accessibility Issues
- [file:line references or "None found"]
### UX Concerns
- [List or "None"]
### Recommendations
- [List or "None"]
Only real usability problems. No code style.
`

const mobileInstructions = `
# Mobile-Focused Review
You are a mobile responsiveness specialist. Target: iPhone 16 Pro (402x874).
Check: Responsive breakpoints (Tailwind mobile-first), touch targets (44px min),
overflow/horizontal scroll, iOS safe areas, mobile nav, font sizes (16px min body),
responsive images, modal behavior on mobile, form inputs, flex/grid stacking.

**Format EXACTLY as:**
## Mobile Review
**Mobile Ready:** YES | NEEDS WORK | NO
### Layout Issues
- [file:line references or "None found"]
### Touch Target Issues
- [List or "None"]
### Recommendations
- [List or "None"]
Only real mobile issues. No desktop-only concerns.
`

const agentHeader = `You are impersonating a code review subagent.
CRITICAL: You MUST follow the exact identity, methodology, and output format from the agent prompt below.

# Your Identity (Subagent Prompt)
`

const agentDeepAnalysis = `
# Your Task — Deep Analysis Required

You have the FULL context: complete file contents, imports, siblings, tests, and project docs.
Use ALL of it. Don't just look at the diff — understand the FULL picture.

**Analysis steps:**
1. Adopt the subagent identity. Use its mindset, expertise, and output format.
2. Read the project standards (CLAUDE.md, code-standards). The PR MUST follow these.
3. Study full file contents to understand how changes fit the overall structure.
4. Trace dependencies via imports. Check correct usage and consistency.
5. Compare with siblings for naming conventions and pattern consistency.
6. Check test coverage for changed functionality.

**Output requirements:**
- Follow the EXACT format from the agent prompt
- Include verdict (PASS / NEEDS WORK / FAIL)
- Reference specific file:line for EVERY finding
- Show code snippets for issues
- Explain HOW you found the issue (e.g., "Comparing with sibling X...")
- Provide concrete fix recommendations

Begin your review now.
`

// writeComments renders PR discussion comments into the prompt.
func writeComments(b *strings.Builder, comments []gh.Comment) {
	if len(comments) == 0 {
		return
	}

	b.WriteString("\n# Existing PR Discussion\n\n")
	for _, c := range comments {
		if c.Path != "" {
			fmt.Fprintf(b, "**%s** on `%s:%d`:\n", c.Author, c.Path, c.Line)
		} else {
			fmt.Fprintf(b, "**%s**:\n", c.Author)
		}
		// Blockquote the body
		for _, line := range strings.Split(c.Body, "\n") {
			fmt.Fprintf(b, "> %s\n", line)
		}
		b.WriteString("\n")
	}
}

// buildGenericInstructions creates the generic review instructions with available agents.
func buildGenericInstructions(agents []string) string {
	var b strings.Builder
	b.WriteString(genericInstructionsHeader)

	if len(agents) > 0 {
		b.WriteString("\n**Recommend specialized reviewers from this list:**\nAvailable subagent personas (installed in this repo):\n")
		for _, agent := range agents {
			fmt.Fprintf(&b, "- `%s`\n", agent)
		}
		b.WriteString("\nOnly recommend agents from the list above.\n")
	} else {
		b.WriteString("\n**No subagent personas are installed in this repo.**\nSkip the Recommended Reviewers section.\n")
	}

	b.WriteString(genericInstructionsFooter)
	return b.String()
}

// BuildGenericPrompt creates the prompt for a generic code review.
func BuildGenericPrompt(ctx context.ReviewContext) (string, error) {
	var b strings.Builder

	writeContextBlock(&b, ctx)

	if ctx.Focus != "" {
		fmt.Fprintf(&b, "\n**SPECIAL FOCUS (prioritize this!):**\n%s\n", ctx.Focus)
	}

	b.WriteString(buildGenericInstructions(ctx.AvailableAgents))

	return b.String(), nil
}

// BuildFocusedPrompt creates a prompt for a focused review area.
func BuildFocusedPrompt(ctx context.ReviewContext, focus string) (string, error) {
	var b strings.Builder

	writeContextBlock(&b, ctx)

	switch focus {
	case "security":
		b.WriteString(securityInstructions)
	case "usability":
		b.WriteString(usabilityInstructions)
	case "mobile":
		b.WriteString(mobileInstructions)
	default:
		return "", fmt.Errorf("unknown focus: %s", focus)
	}

	return b.String(), nil
}

// BuildAgentPrompt creates a prompt for an agent-impersonated review.
func BuildAgentPrompt(ctx context.AgentContext) (string, error) {
	var b strings.Builder

	// 1. Agent identity
	b.WriteString(agentHeader)
	b.WriteString(ctx.AgentPrompt)
	b.WriteString("\n\n")

	// 2. Project docs
	b.WriteString("# Project Standards & Documentation\n\n")
	for _, name := range ctx.DocOrder {
		content := ctx.ProjectDocs[name]
		fmt.Fprintf(&b, "## %s\n%s\n\n", name, content)
	}

	// 3. PR info
	pr := ctx.PR.Info
	fmt.Fprintf(&b, "\n# PR To Review\n**Title:** %s\n**Files changed:** %d\n**Lines:** +%d / -%d\n",
		pr.Title, len(ctx.PR.ChangedFiles), pr.Additions, pr.Deletions)
	fmt.Fprintf(&b, "\n**Description:**\n%s\n", pr.Body)

	writeComments(&b, ctx.Comments)

	// 4. Full file contents
	b.WriteString("\n# Full File Contents (Changed Files)\n")
	for _, file := range ctx.PR.ChangedFiles {
		content, ok := ctx.FileContents[file]
		if !ok {
			continue
		}
		lang := context.LangForExt(file)
		fmt.Fprintf(&b, "### %s\n```%s\n%s\n```\n\n", file, lang, content)
	}

	// 5. Imports
	if len(ctx.Imports) > 0 {
		b.WriteString("\n# Imported Dependencies (referenced by changed files)\n\n")
		for _, entry := range sortedMap(ctx.Imports) {
			fmt.Fprintf(&b, "### %s\n```typescript\n%s\n```\n\n", entry.Key, entry.Value)
		}
	}

	// 6. Siblings
	if len(ctx.Siblings) > 0 {
		b.WriteString("\n# Sibling Files (same directories for pattern context)\n\n")
		for _, entry := range sortedMap(ctx.Siblings) {
			fmt.Fprintf(&b, "### %s (sibling)\n```typescript\n%s\n```\n\n", entry.Key, entry.Value)
		}
	}

	// 7. Tests
	if len(ctx.Tests) > 0 {
		b.WriteString("\n# Related Test Files\n\n")
		for _, entry := range sortedMap(ctx.Tests) {
			fmt.Fprintf(&b, "### %s\n```typescript\n%s\n```\n\n", entry.Key, entry.Value)
		}
	}

	// 8. Diff
	fmt.Fprintf(&b, "\n# Diff (the actual changes)\n```diff\n%s\n```\n", ctx.PR.Diff)

	// 9. Deep analysis instructions
	b.WriteString(agentDeepAnalysis)

	return b.String(), nil
}

// writeContextBlock writes the shared context block for generic and focused prompts.
func writeContextBlock(b *strings.Builder, ctx context.ReviewContext) {
	pr := ctx.PR.Info

	b.WriteString("You are reviewing a pull request.\n\n")

	fmt.Fprintf(b, "# PR Information\n**Title:** %s\n**Base:** %s\n**Files changed:** %d\n**Lines changed:** +%d / -%d\n",
		pr.Title, pr.BaseRef, len(ctx.PR.ChangedFiles), pr.Additions, pr.Deletions)

	fmt.Fprintf(b, "\n**PR Description:**\n%s\n", pr.Body)

	writeComments(b, ctx.Comments)

	// File categories
	b.WriteString("\n")
	b.WriteString(ctx.Categories.String())
	fmt.Fprintf(b, "\n- CRUD Operations: %v\n- Complex PR: %v\n", ctx.CRUDInDiff, ctx.Complex)

	// Project docs
	if len(ctx.DocOrder) > 0 {
		b.WriteString("\n# Project Context\n")
		for _, name := range ctx.DocOrder {
			content := ctx.ProjectDocs[name]
			fmt.Fprintf(b, "\n## %s\n%s\n", name, content)
		}
	}

	// Full file contents
	b.WriteString("\n# Full File Contents (HEAD versions of changed files)\n\n")
	skippedCount := 0
	for _, file := range ctx.PR.ChangedFiles {
		if context.IsBinary(file) {
			continue
		}
		content, ok := ctx.FileContents[file]
		if !ok {
			skippedCount++
			if skippedCount <= ctx.FilesSkipped {
				fmt.Fprintf(b, "### %s\n(skipped — budget exceeded)\n\n", file)
			}
			continue
		}
		lang := context.LangForExt(file)
		fmt.Fprintf(b, "### %s\n```%s\n%s\n```\n\n", file, lang, content)
	}

	// Diff
	fmt.Fprintf(b, "\n# Diff\n```diff\n%s\n```\n", ctx.PR.Diff)
}

// sortedMap returns keys in sorted order for deterministic output. Returns an iterator-like slice.
func sortedMap(m map[string]string) []kv {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]kv, len(keys))
	for i, k := range keys {
		result[i] = kv{k, m[k]}
	}
	return result
}

type kv struct {
	Key   string
	Value string
}
