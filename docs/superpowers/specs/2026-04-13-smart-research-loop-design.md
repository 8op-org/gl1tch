# Smart Research Loop

**Date:** 2026-04-13
**Status:** Approved

---

## Goal

Make `glitch ask "..."` handle any question intelligently by upgrading the research loop with better researchers, paid model feedback, auto-cloning, results extraction, and missing-researcher guidance. The user types natural language. Glitch figures out what to gather, sends a tight brief to the paid model, saves results for review, and gets smarter over time from the paid model's feedback.

## Architecture

```
glitch ask "there are TBC placeholders in the observability-robots CI docs"
       │
       ▼
  Tier 1: router.Match() → no workflow match
       │
       ▼
  Tier 2: research loop
       │
       ├─ PLAN (local LLM, free)
       │   Picks researchers from registry menu.
       │   If a picked name isn't registered, prints:
       │     ">> missing researcher: <name> — add YAML to ~/.config/glitch/researchers/"
       │   Continues with what it has.
       │
       ├─ GATHER (parallel, free)
       │   Native + YAML researchers run concurrently.
       │   Auto-clones repos to ~/.local/share/glitch/repos/ if not present.
       │   Auto-indexes into ES if not already indexed.
       │   Evidence goes into EvidenceBundle.
       │
       ├─ DRAFT (paid model, full autonomy)
       │   Gets curated brief: fix plan + only relevant file sections.
       │   Produces complete fixed files + feedback on evidence quality.
       │
       ├─ FEEDBACK (from draft response)
       │   Paid model rates evidence: what was useful, what was missing.
       │   → Stored in SQLite research_events
       │   → Printed to user: ">> learned: next time will also fetch sibling docs"
       │   → Hints provider reads this for future planner calls
       │
       ├─ SCORE (local LLM, free)
       │   Critiques draft against evidence. Composite confidence.
       │
       └─ SAVE (when output is substantive)
           ├─ results/<repo>/<issue>/drafts.md
           ├─ results/<repo>/<issue>/feedback.md
           ├─ results/<repo>/<issue>/<mirrored file paths>
           └─ ES + SQLite indexed
```

## Native Researchers (Go)

Four built-in researchers that work everywhere, no config needed.

### `git`

Covers: log, diff, remote, blame. Decides which git commands to run based on the question context. Always available in any git repo.

```go
type GitResearcher struct{}

func (g *GitResearcher) Name() string { return "git" }
func (g *GitResearcher) Describe() string {
    return "git history, diffs, remotes, and blame for the current or target repository"
}
```

The Gather method runs a relevant subset of:
- `git log --oneline -50` (recent history)
- `git diff HEAD~10` (recent changes)
- `git remote -v` (repo identity)
- `git log --oneline --all --grep=<keyword>` (search history)

Which commands run depends on the question — the researcher inspects `ResearchQuery.Question` for signals (e.g., "what changed" → diff, "who wrote" → blame).

### `fs`

Covers: read files, list directories, grep contents, pattern scanning (TBC/TODO/FIXME). Operates on the local filesystem — either the cwd or a cloned repo.

```go
type FSResearcher struct{}

func (f *FSResearcher) Name() string { return "fs" }
func (f *FSResearcher) Describe() string {
    return "read files, list directories, search file contents, and scan for patterns like TBC/TODO"
}
```

The Gather method:
- If the question mentions specific file paths, reads those files
- If the question is about placeholders/missing content, scans with `grep -rn`
- If the question is about structure, lists the directory tree
- File paths are extracted from the question and from other evidence in the bundle (cross-researcher)

### `es-activity`

Queries `glitch-events`, `glitch-pipelines`, `glitch-summaries` for indexed activity. Already exists — no changes needed to the researcher itself.

### `es-code`

Queries `glitch-code-*` for indexed source code. Already exists — no changes needed to the researcher itself.

## Default YAML Researchers

Ship with glitch, installed to `~/.config/glitch/researchers/` on first run or via the repo's `researchers/` directory.

### `github-prs.yaml`
```yaml
name: github-prs
description: open and recently merged pull requests
steps:
  - id: gather
    run: gh pr list --limit 20 --state all --json number,title,state,author,updatedAt
```

### `github-issues.yaml`
```yaml
name: github-issues
description: open and closed issues for a repository
steps:
  - id: gather
    run: gh issue list --limit 20 --state all --json number,title,state,labels,updatedAt
```

### `github-issue.yaml`
```yaml
name: github-issue
description: fetch a specific GitHub issue with full body and comments
steps:
  - id: gather
    run: gh issue view {{.input}} --json number,title,body,comments,labels
```

## Missing Researcher Guidance

When the planner picks a researcher name that isn't in the registry:

```
>> researching...
>> missing researcher: jira-tickets — add ~/.config/glitch/researchers/jira-tickets.yaml
>> using: git, fs, github-issues
```

The loop continues with available researchers. The user sees exactly what to add. Since the planner is an LLM, it naturally asks for researchers that make sense — missing names are a discovery signal.

## Auto-Clone + Auto-Index

When a researcher needs a repo that isn't local:

1. Check `~/.local/share/glitch/repos/<org>/<repo>/`
2. If missing: `git clone --depth=1` to that path
3. Check if `glitch-code-<repo>` index exists in ES
4. If missing or ES is down: index with `indexer.IndexRepo()`
5. If ES is down: skip indexing, researchers still work on local files

This happens inside the `git` and `fs` researchers transparently. The user never clones or indexes manually unless they want to.

On subsequent runs: `git pull` to freshen the clone. No re-index unless files changed.

## Paid Model Integration

The draft stage sends the prompt to the paid provider (claude) with full autonomy:

- Provider command reads from stdin (already implemented)
- The prompt contains: the curated evidence brief + the question + instructions to draft AND provide feedback
- The paid model's response is parsed into two sections: **draft** (the actual fixes/answer) and **feedback** (evidence quality assessment)

### Feedback Format

The draft prompt asks the paid model to append a feedback section:

```
--- FEEDBACK ---
- evidence_quality: good|adequate|insufficient
- missing: ["directory listing for docs/teams/ci/macos/", "sibling file orka.md"]
- useful: ["repo-scan found all TBC placeholders", "github-issue body had exact file paths"]
- suggestion: "for doc placeholder fixes, always include sibling files in the same directory"
```

This is parsed and:
1. Stored in SQLite `research_events` with the paid model's perspective
2. Printed to user: `>> learned: for doc fixes, include sibling files in the same directory`
3. Read by hints provider on next run to bias the planner

## Results Folder

When the draft produces substantive output (fixes, implementations, not just a one-line answer), results are saved:

```
results/<org>/<repo>/<issue>/
├── drafts.md                           # full paid model output
├── feedback.md                         # what glitch learned
├── fix-plan.json                       # local LLM analysis
└── docs/                               # mirrored repo structure
    └── teams/
        └── ci/
            ├── macos/
            │   └── index.md            # ready to diff/copy
            ├── dependencies/
            │   └── updatecli.md
            └── secrets-security-incident-runbook.md
```

The user reviews, then decides to push. Glitch never pushes without the user saying so.

### When to save results

The loop saves results when:
- The draft is longer than 500 characters (not a quick answer)
- The question references a specific issue number or repo
- The draft contains file content (detected by `--- FILE:` markers or code blocks)

Short answers ("goroutines are lightweight threads") just print to stdout, no results folder.

## Changes to Existing Code

### Delete
- `researchers/doc-fix.yaml` (workflow replaced by Go logic)
- `cmd/research_helpers.go` (logic moves into the loop)

### Modify
- `internal/research/loop.go` — add feedback parsing, auto-save, missing researcher guidance
- `internal/research/prompts.go` — update draft prompt to request feedback section
- `internal/research/events.go` — add feedback event type
- `internal/research/types.go` — add Feedback field to Result
- `cmd/ask.go` — tier 2 handles results saving and feedback printing

### Create
- `internal/research/git_researcher.go` — native git researcher
- `internal/research/fs_researcher.go` — native fs researcher
- `internal/research/repo.go` — auto-clone and auto-index logic
- `internal/research/feedback.go` — parse and store paid model feedback
- `internal/research/results.go` — extract files from draft, write to results/

## What's NOT Included

- Interactive approval during the loop (paid model has full autonomy)
- Automatic PR creation (user reviews results first)
- Custom planner prompts (hints provider handles learning)
- Workspace scoping (single workspace for now)
