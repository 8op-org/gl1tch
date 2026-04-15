# Glitch Workspace Model

**Date:** 2026-04-14
**Status:** Draft

## Problem

Glitch stores workflows globally (`~/.config/glitch/workflows/`) and results relative to CWD (`CWD/results` or `CWD/.glitch/results`, inconsistently). This doesn't work when stokagent is the command center for cross-repo work against observability-robots, ensemble, and other Elastic repos. Results end up scattered, workflows are duplicated between global config and stokagent, and there's no coherent structure for Claude to consume results and produce PR/issue responses.

## Solution

A `--workspace` flag on the root `glitch` command that makes a directory the container for workflows and results.

## Design

### 1. The `--workspace` flag

Top-level persistent flag on the root `glitch` command.

```
glitch --workspace ~/Projects/stokagent ask elastic/observability-robots#3920
```

When set:

- **Workflows:** resolved from `<workspace>/workflows/` only. Global `~/.config/glitch/workflows/` is skipped entirely.
- **Results:** written to `<workspace>/results/<org>/<repo>/<issue|pr>-<number>/`.
- **Config override:** `workflows_dir` in `~/.config/glitch/config.yaml` can override the default `<workspace>/workflows/` location (see Section 3).
- **Without `--workspace`:** current behavior unchanged — global workflows, CWD-relative results.

The flag accepts an absolute or relative path. No marker-file discovery, no walking upward.

Typical usage with a shell alias:

```
alias gl='glitch --workspace ~/Projects/stokagent'
```

### 2. Result directory structure

Path convention:

```
<workspace>/results/<org>/<repo>/<issue|pr>-<number>/
```

The type prefix (`issue-` or `pr-`) is determined from the GitHub API response — glitch already fetches the ref for context.

Contents of each result directory:

```
issue-3920/
  README.md          # rollup — frontmatter metadata + action-ready content
  evidence/          # raw tool call outputs, numbered
    001-github-issue.md
    002-grep-results.md
    003-file-read.md
  plan.md            # implementation plan (if goal=implement)
  review.md          # post-impl review (if applicable)
  run.json           # machine-readable run metadata
```

#### README.md (rollup artifact)

Designed for Claude consumption — frontmatter for machine parsing, body for action:

```markdown
---
repo: elastic/observability-robots
ref: issue-3920
title: "Fix flaky CI in integration tests"
status: researched | planned | implemented
created: 2026-04-14T10:30:00Z
model: qwen2.5:7b
---

## Summary
<2-3 sentence findings>

## Recommendation
<what to do, concrete>

## Response Draft
<copy-paste ready PR comment, issue reply, or PR body depending on context>

## Evidence Index
- [001-github-issue.md](evidence/001-github-issue.md) — original issue body and comments
- [002-grep-results.md](evidence/002-grep-results.md) — relevant code matches
```

`README.md` is chosen because Claude reads it first by convention, GitHub renders it for browsing, and it's the natural directory index file.

#### Variant runs

For comparing outputs across models/tools (claude vs copilot), a double-dash suffix on the directory:

```
results/elastic/observability-robots/
  issue-3920/           # default run
  issue-3920--claude/   # variant
  issue-3920--copilot/  # variant
```

### 3. Workflow resolution

When `--workspace` is set:

1. Load from `<workspace>/workflows/` only. Recursive walk, supports `.yaml`, `.yml`, `.glitch` files. Same `LoadDir()` logic, different root.
2. `~/.config/glitch/workflows/` is skipped entirely. No merging, no overrides.
3. If `workflows_dir` is set in `~/.config/glitch/config.yaml`, that path is used instead of `<workspace>/workflows/`.

Without `--workspace`: current behavior — global then `.glitch/workflows/`.

### 4. Consistent run metadata

Every result directory gets a `run.json` with a standardized schema:

```json
{
  "repo": "elastic/observability-robots",
  "ref_type": "issue",
  "ref_number": 3920,
  "workflow": "pr-review",
  "status": "researched",
  "created": "2026-04-14T10:30:00Z",
  "duration_seconds": 45,
  "model": "qwen2.5:7b",
  "variant": null
}
```

Dashboarding is a workflow concern, not a glitch core concern. A `dashboard-sync` workflow can glob `results/**/run.json` and index to Elasticsearch. Kibana provides the real dashboard. The structured `run.json` is the contract.

## What changes in glitch core

1. **New flag:** `--workspace` on root command, threaded through to subcommands
2. **Workflow resolution:** conditional path based on workspace flag presence
3. **Result path computation:** `resultDir()` in `internal/research/results.go` and `saveResults()` in `internal/batch/batch.go` use workspace-relative paths with `<org>/<repo>/<type>-<number>` convention
4. **README.md generation:** new rollup step after research/batch completes, producing frontmatter + action-ready body
5. **run.json schema:** standardize fields across all code paths

## What does NOT change

- `~/.config/glitch/config.yaml` — still the home for provider/model config
- Pipeline execution logic
- Workflow file format
- Router and issue parsing
- Any behavior when `--workspace` is not set
