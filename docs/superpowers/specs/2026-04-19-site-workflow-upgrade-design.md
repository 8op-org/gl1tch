# Site Workflow Upgrade — Full Evaluator Rewrite

**Date:** 2026-04-19
**Scope:** Consolidate 6 site workflows → 4, rewrite using evaluator stdlib, absorb Python where practical

## Summary

Upgrade all `.glitch/workflows/site-*.glitch` files from pre-evaluator sexpr to full evaluator syntax. Consolidate `site-create-page` + `site-update-page` into `site-page`, and `site-publish` + `site-update` into `site-publish`. Replace inline Python with evaluator builtins (`json-pick`, `each`, `reduce`, `assoc`, `->`, `def`, `par`, `when`). Extract remaining inline Python into focused stdin-based scripts.

## Provider Strategy

| Step type | Provider | Model |
|---|---|---|
| Content generation (docs, changelogs) | Copilot | Sonnet |
| Operational (summaries, classification) | LM Studio (default) | google/gemma4 |

- Content LLM steps: `:provider "copilot" :model "sonnet"`
- Operational LLM steps: no provider specified (runtime default → LM Studio gemma4)

## Workflow Consolidation

### 1. `site-dev.glitch` — minimal changes

- Add `par` around the three prep steps (`generate`, `ensure-changelog`, `ensure-report`)
- `run` → `sh`
- No other changes needed — already clean

### 2. `site-page.glitch` — merges create-page + update-page

**Params:** `--set page=plugins` (required), `--set topic="new topic"` (create), `--set instructions="add par section"` (update)

**Flow:**
1. `(def manifest (include "site-manifest.glitch"))` — native include
2. `(def page-path ...)` + `(def is-update ...)` — computed from param
3. `(par ...)` — gather context: existing-docs, examples, repo-structure, recent-commits
4. `(when (= is-update "true") ...)` — conditionally read existing page
5. Single LLM step with conditional prompt (create vs update context)
6. `(save page-path :from "generate")`
7. Rebuild docs.json
8. `(phase "verify" :retries 1)` — hallucinations + structure gates
9. Done message

**Absorbed Python:** `save-generated-page.py` → `(save ...)` builtin

### 3. `site-sync.glitch` — the big rewrite

**Flow:**
1. `(def manifest (include "site-manifest.glitch"))` — replaces `site-read-manifest.py`
2. `(step "diff-disk" (sh "python3 scripts/site-diff-disk.py" ...))` — stays Python, takes manifest as stdin
3. Early exit via `(when ...)` if nothing to do
4. `(par ...)` — query-code-index + gather-fallback (extracted to stdin-based scripts)
5. `(def context-map ...)` — merge context natively with `reduce`/`assoc`/`pick`, replaces `merge-context` inline Python
6. `(each ...)` over pages needing work → LLM generation per page with `:provider "copilot" :model "sonnet"`
7. `(step "inject-frontmatter" ...)` — extracted to stdin-based Python script
8. `(step "build-sidebar" ...)` — stays as-is
9. `(phase "verify" :retries 1)` — 4 gates
10. `(phase "build-test" :retries 0)` — playwright

**Absorbed Python:**
- `site-read-manifest.py` → `(include "site-manifest.glitch")`
- `diff-summary` inline → `(-> ... (json-pick) (count))`
- `pages-needing-work` inline → `(-> ... (json-pick "create" "update"))`
- `merge-context` inline → `(reduce ...)` with `assoc`/`pick`
- `write-pages` summary inline → `(println ...)`

**Extracted to stdin-based scripts (new):**
- `scripts/site-query-index.py` — ES curl query, reads diff from stdin
- `scripts/site-gather-fallback.py` — glob expansion + capped file reads, reads diff from stdin

**Stays as-is:**
- `site-diff-disk.py` — complex filesystem diffing
- `site-build-sidebar.py` — Astro component template
- All `gate-*.py` scripts
- `gate-playwright.sh`

### 4. `site-publish.glitch` — absorbs site-update

**Flow:**
1. `(def manifest (include "site-manifest.glitch"))`
2. `(par ...)` — gather stubs, examples, workflows, repo-structure, changelog-raw
3. `(step "enrich-docs" (llm :provider "copilot" :model "sonnet" ...))` — regen all pages
4. `(save ...)` + split-docs
5. `(step "enrich-changelog" (llm :provider "copilot" :model "sonnet" ...))` — changelog
6. `(save ...)` changelog
7. `(phase "content-verify" :retries 1)` — 3 gates
8. `(step "build" ...)` — astro build
9. `(phase "build-test" :retries 0)` — playwright
10. Done

**Absorbed Python:** `save-enrich-docs.py` → `(save ...)` builtin

## Syntax Changes (all workflows)

| Old | New | Why |
|---|---|---|
| `(run ...)` | `(sh ...)` | Matches evaluator builtin |
| `~(stepfile X)` | stdin piping or `(step X)` ref | Evaluator native |
| `:tier 0` | no provider (local default) | `:tier` not in evaluator |
| inline Python glue | `def`/`json-pick`/`reduce`/`assoc` | Native evaluator |
| sequential context steps | `(par ...)` | Concurrent execution |
| separate create/update workflows | `(when ...)` branch | Single workflow |
| `(echo ...)` wrappers | `(println ...)` | Native builtin |

## Files Created/Modified

**New workflow files (on main, `.glitch/workflows/`):**
- `site-dev.glitch`
- `site-page.glitch`
- `site-sync.glitch`
- `site-publish.glitch`

**New extracted scripts:**
- `scripts/site-query-index.py` — extracted from site-sync inline, reads stdin
- `scripts/site-gather-fallback.py` — extracted from site-sync inline, reads stdin
- `scripts/site-inject-frontmatter.py` — extracted from site-sync inline

**Deleted (from site-sync worktree, not on main):**
- `site-create-page.glitch` — merged into site-page
- `site-update-page.glitch` — merged into site-page
- `site-update.glitch` — merged into site-publish
- `scripts/site-read-manifest.py` — replaced by `(include ...)`

**Unchanged:**
- `site-manifest.glitch` — already native sexpr
- `site-diff-disk.py`, `site-build-sidebar.py` — complex, stay as-is
- All gate scripts — tested, independent

## Implementation Strategy

Agent-per-task with review at end:
1. Agent: `site-dev.glitch` rewrite
2. Agent: `site-page.glitch` (merged create + update)
3. Agent: Extract inline Python → stdin-based scripts
4. Agent: `site-sync.glitch` rewrite
5. Agent: `site-publish.glitch` (merged publish + update)
6. Review: code-reviewer agent validates all 4 workflows + scripts against spec

## Success Criteria

- All 4 workflows parse and run through the evaluator without errors
- `par` blocks on all independent step groups
- No `:tier` usage — local default or explicit `:provider "copilot" :model "sonnet"`
- Inline Python eliminated where evaluator builtins suffice
- Extracted scripts read from stdin (no `stepfile` dependency)
- All gate scripts still referenced and working
- `site-manifest.glitch` loaded via `(include ...)` not Python
