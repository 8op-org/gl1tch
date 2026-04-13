# PR Status Report — 2026-04-13

## PR #4160: docs(ai-agents): add Gemini CLI guides and update onboarding
- **State:** OPEN, REVIEW_REQUIRED
- **Branch:** docs/gemini-cli-guides-and-onboarding
- **Changes:** +1258/-112 across 6 files
- **Created:** 2026-04-06 | **Last Updated:** 2026-04-10
- **Reviewer:** @kuisathaverat (12 review comments, 1 detailed AI review)

### Blockers from review (must fix before merge)

1. **Task list syntax `- [ ]` not supported in MyST Markdown** — 8 occurrences across both guides. Replace with plain `-` bullets.
2. **`\n` in Mermaid node labels** — 7 occurrences. Elastic docs Mermaid renderer needs `<br>` not `\n`.
3. **Invalid Go source in `buggy.go` heredoc** — literal newline inside a Go string literal on line 335 of `02-first-steps.md`. Must be `\n`.
4. **`applies_to` frontmatter uses non-standard key** — `ai-agents: gemini-cli` is not in the documented `applies_to` spec.

### Warnings

5. `gcloud` prerequisite missing from Vertex AI path
6. Deprecated Google Cloud Console URL (`console.developers.google.com`)
7. `gemini -r latest` unverified as valid
8. JSON output schema only applies to Google AI path
9. `geminicli.com/docs` is third-party, not official Google docs

### Action needed: Address the 4 blockers, push fixes, request re-review from @kuisathaverat.

---

## PR #3677: docs: add terminal first session guide
- **State:** OPEN, REVIEW_REQUIRED
- **Branch:** docs/terminal-first-session-guide
- **Changes:** +372/-1 across 3 files
- **Created:** 2026-03-09 | **Last Updated:** 2026-04-13
- **Reviews:** 2 (both COMMENTED)

### Status

No blockers identified in review. The PR is clean — documentation-only, follows established format. Has been open for 35 days with no merge action.

### Action needed: Ping @kuisathaverat or team for merge approval. The PR appears ready.

---

## Open Documentation Issues (unassigned, actionable)

| # | Title | Scope |
|---|-------|-------|
| 3928 | Update docs for renamed skills guide + unresolved CI placeholders | 5 concrete fixes: dead link, placeholder pages, incomplete runbooks |
| 3912 | Create skill creator guide with 5 practical use cases | Large: full guide with 5 use-case files |
| 3916-3923 | Individual skill creator use cases (7 sub-issues) | Each is one guide file |
| 3876-3882 | NotebookLM guides (6 sub-issues) | Each is one lab guide |
