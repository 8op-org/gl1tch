# Homepage Workflow with `grounded?` Verification

**Date**: 2026-04-21
**Status**: Approved

## Problem

The current `site-homepage.glitch` uses a hand-rolled LLM gate step that pattern-matches banned terms. This is brittle — it catches keyword violations but can't verify whether claims in the homepage content are actually supported by the codebase. The confidence framework's `grounded?` primitive does exactly this.

## Design

Replace the gate step with `grounded?` and add automatic dev server launch for immediate preview.

### Workflow: `site-homepage.glitch` (6 steps)

**Step 1: gather-context** (search)
Uses the `search` DSL primitive (hybrid FTS5 + semantic search against the indexed repo) instead of dumb shell file dumps:
```clojure
(str
  (search "providers tool-use llm workflow" :limit 15)
  "\n\n"
  (search "homepage features DSL syntax" :limit 10))
```
Returns ranked, relevant code chunks with paths and scores. The runner auto-indexes the workspace on startup and exposes `search` as a first-class workflow primitive.

### New: `search` DSL primitive

Wired into the runner's SCI context alongside the confidence framework. The runner:
1. Opens `<workspace>/.glitch/search.db` on startup
2. Indexes the repo (incremental — only new/changed files)
3. Exposes `(search query :limit N)` which calls `hybrid-search` (FTS5 keyword + semantic) and returns formatted results

**Step 2: rewrite** (llm, copilot agentic)
Unchanged from current workflow. Copilot with agentic tool-use (max 8 rounds) reads the codebase via MCP tools and rewrites `index.astro` per user instructions.

**Step 3: verify** (grounded?)
```clojure
(grounded? "rewrite" (ref "gather-context")
  :strict false)
```
- Non-strict mode: logs unsupported claims to stderr but does not block save
- Uses default provider for the grounding LLM call
- Replaces the old gate step entirely

**Step 4: save**
- Saves the rewrite to `site/src/pages/index.astro`

**Step 5: playwright** (shell)
- Runs `cd site && npx playwright test` against the updated homepage
- Playwright config already builds the site and serves on :4322
- Existing tests validate: hero, sections, feature cards, banned terms, no broken links, no JS errors
- If tests fail, the step output shows which assertions broke — workflow stops here

**Step 6: dev** (shell, only reached if Playwright passes)
- Launches `npx astro dev --port 4322` for live preview

### What's Removed

The old "gate" LLM step that checked for banned terms via prompt. `grounded?` subsumes this — any hallucinated command or invented feature is by definition not grounded in the source context. Playwright's "no BubbleTea/SQLite/tmux on any page" test also catches banned terms at the rendered HTML level.

### What's Unchanged

- The rewrite step (copilot agentic, same rules, same max-rounds)
- The `site-dev.glitch` workflow (unchanged, but inlined here instead of called)
- The `site.glitch` router (dispatches to homepage as before)
- The existing Playwright test suite (`site/tests/site.spec.ts`) — no changes needed

## Testing

Run manually:
```
glitch run site-homepage "update the features section to highlight tool-use providers"
```

Expect: homepage updated, grounding check logged, Playwright suite passes, dev server starts on :4322.
