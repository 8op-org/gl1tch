# Site Update Workflow Design

**Date:** 2026-04-14
**Status:** Draft
**Goal:** A glitch workflow that regenerates the 8op.org website from source-of-truth markdown stubs, real examples, and git history — with tiered LLM cost optimization and triple-gate verification.

## Trigger

```
glitch ask "update the website"    # routed via ask
glitch workflow run site-update    # direct invocation
```

The workflow lives at `.glitch/workflows/site-update.glitch` so the router discovers it.

## Architecture

```
docs/site/                          source-of-truth markdown stubs (you maintain)
  getting-started.md
  workflow-syntax.md
  (new stubs auto-become pages)

examples/                           real .glitch workflow files (already exist)
  hello.glitch
  code-review.glitch
  multi-step-chain.glitch
  parameterized.glitch
  agent-with-skill.glitch

site/                               Astro project (replaces single index.html)
  src/
    content/docs/                   enriched markdown (workflow output)
    content/changelog/              auto-generated "What's New" entries
    pages/                          Astro page templates
    layouts/                        shared layout (current design tokens)
    components/                     reusable UI (hero, hex-rain, code blocks)
  public/                           static assets
  astro.config.mjs
```

## Content Flow

1. You write lightweight stubs in `docs/site/` — section headings, key bullet points, tone cues.
2. Workflow reads stubs + real `.glitch` examples + git log.
3. LLM enriches stubs into full prose, injects validated examples.
4. Output lands in `site/src/content/docs/`.
5. Verification gates run.
6. Astro builds static HTML.

## Workflow Phases

### Phase 1: Gather (shell steps)

- Read all markdown stubs from `docs/site/`.
- Read example `.glitch` files from `examples/`.
- Pull git log since last site build. Track last-build timestamp via a marker file (`.site-build-timestamp`) or git tag (`site-last-build`).

### Phase 2: Generate (LLM steps)

- **Enrich stubs** (auto-route, no tier pin): For each stub, LLM expands bullet points into full prose. Local model gets first shot; self-eval will escalate if prose quality is poor. The stub content, real examples, and project context are provided as input. The LLM writes docs — it does not invent features.
- **Generate changelog** (tier 0, auto-route): Summarize git log entries into user-facing "What's New" entries. Simpler summarization task, local model handles it.
- **Inject examples**: Real `.glitch` files from `examples/` are embedded verbatim — the LLM does not generate example code, it writes the surrounding explanation.

### Phase 3: Verify (triple gate)

All three gates must pass. Failure stops the workflow and outputs a verification report.

**Gate 1 — Example validation (shell):**
Every `.glitch` code block extracted from generated docs is written to a temp file and validated. Since `glitch workflow run` has no dry-run flag, validation uses a shell step that extracts code blocks, writes each to a temp `.glitch` file, and runs `glitch workflow run` on a minimal wrapper that only parses the syntax (a workflow with no LLM steps that just echoes success). If any example contains invalid sexpr syntax, the gate fails.

**Gate 2 — Build validation (shell):**
`npx astro build` in the `site/` directory must exit 0. Catches broken templates, missing content, invalid frontmatter.

**Gate 3 — Diff-review (tier 2, pinned):**
A tier-2 LLM (Claude/Copilot) receives:
- The original stubs
- The generated content
- The list of examples injected

It checks:
- No hallucinated features (content claims something glitch doesn't do)
- No missing stub sections (every stub heading is covered)
- Examples match real syntax (no invented keywords or forms)
- Tone matches the project voice (user-first, no internals)

The diff-review outputs a pass/fail verdict with specific findings.

### Phase 4: Output

- Write enriched markdown to `site/src/content/docs/`.
- Write changelog entries to `site/src/content/changelog/`.
- Run `npx astro build`.
- Save verification report to `site/build-report.md`.

## Tiering Strategy

| Step | Tier | Rationale |
|------|------|-----------|
| Read stubs, examples, git log | shell | No LLM needed |
| Enrich stubs into prose | auto-route (no tier pin) | Needs coherent writing; local will likely escalate via self-eval |
| Generate changelog entries | 0 (auto-route) | Simpler summarization |
| Example parse check | shell | `glitch workflow run`, no LLM |
| Astro build check | shell | `npx astro build` |
| Diff-review | 2 (pinned) | Correctness gate, needs best model |

## Astro Site (Day One)

Preserves the current design language: JetBrains Mono, dark theme (`--bg: #1a1b26`), glitch animation, hex-rain canvas, teal/yellow/pink accent palette.

### Pages

- **/** — Landing page. Hero, tagline, install command, quick pitch. Mostly static with minor templated sections.
- **/docs/getting-started** — Install glitch, write your first workflow, run your first `glitch ask`.
- **/docs/workflow-syntax** — Sexpr reference. Real examples pulled from `examples/` directory.
- **/changelog** — Auto-generated from git history. Most recent entries first.

### Adding Pages

Drop a new `.md` stub in `docs/site/`. The workflow picks it up, enriches it, and Astro's content collection renders it as a new page. No config changes needed.

## Stub Format

Stubs are lightweight markdown with frontmatter:

```markdown
---
title: Getting Started
order: 1
description: Install glitch and run your first workflow
---

## Install

- brew install 8op-org/tap/glitch
- requires: ollama running locally

## Your first workflow

- show hello.glitch example
- explain step-by-step what happens
- mention the sexpr syntax

## Your first ask

- glitch ask "what time is it" → routes to a workflow
- explain how routing works (local LLM picks the match)

## Tone

- "your" framing, not "the user"
- examples before explanation
- no internals (no mention of BubbleTea, tmux implementation details)
```

The LLM uses this as a skeleton and writes the full page. It must cover every heading. The diff-review gate enforces this.

## Staleness Prevention

The stubs are the editorial backbone. When a feature ships:

1. Update or add a stub (even just a heading + bullets).
2. `glitch ask "update the website"`.
3. Workflow regenerates, verifies, outputs.

The diff-review gate checks: "does the generated site cover everything in the stubs?" A stub section the site doesn't reflect gets flagged as a failure.

Examples stay current because they're read from `examples/` at build time — if the example files are updated with new syntax, the site gets the new versions automatically.

## Workflow Skeleton (sexpr)

```scheme
(workflow "site-update"
  :description "regenerate 8op.org from doc stubs, examples, and git history"

  ;; Phase 1: Gather
  (step "stubs"
    (run "cat docs/site/*.md"))

  (step "examples"
    (run "cat examples/*.glitch"))

  (step "changelog-raw"
    (run "git log --oneline --since='$(cat .site-build-timestamp 2>/dev/null || echo 2025-01-01)' -- cmd/ internal/ examples/"))

  ;; Phase 2: Generate
  (step "enrich-docs"
    (llm
      :prompt ```
        You are a technical writer for gl1tch (8op.org).
        Rules: user-first framing ("your"), examples before explanation,
        no internal implementation details.

        Expand these stubs into full documentation pages.
        Use ONLY the real examples provided — do not invent code.
        Output as markdown with --- frontmatter separators between pages.

        Stubs:
        {{step "stubs"}}

        Real examples:
        {{step "examples"}}
        ```))

  (step "enrich-changelog"
    (llm
      :prompt ```
        Summarize these git commits into user-facing changelog entries.
        Group by feature area. Skip internal refactors unless user-visible.
        Output as markdown list items.

        Commits:
        {{step "changelog-raw"}}
        ```))

  ;; Phase 3: Verify
  ;; Gate 1: example parse check (shell)
  ;; Gate 2: astro build (shell)
  ;; Gate 3: diff-review (tier 2)
  (step "diff-review"
    (llm
      :tier 2
      :prompt ```
        You are a verification reviewer for the gl1tch website.

        Compare the generated docs against the original stubs.
        Check:
        1. No hallucinated features (claims about things glitch cannot do)
        2. Every stub heading is covered in the output
        3. Code examples use valid sexpr syntax
        4. Tone is user-first, no implementation internals

        Original stubs:
        {{step "stubs"}}

        Generated docs:
        {{step "enrich-docs"}}

        Respond with PASS or FAIL and specific findings.
        ```))

  ;; Phase 4: Output
  (step "save-docs"
    (save "site/src/content/docs/generated.md" :from "enrich-docs"))

  (step "save-changelog"
    (save "site/src/content/changelog/latest.md" :from "enrich-changelog"))

  (step "save-report"
    (save "site/build-report.md" :from "diff-review")))
```

**Note:** This skeleton shows the core flow. The actual implementation will need shell steps for the parse-check and astro-build gates, and the `when` form to halt on gate failure once that special form is implemented. The exact step decomposition (one LLM call per stub vs. one batch call) will be refined during implementation.

## Out of Scope

- CI/CD auto-publish (future — wire into GitHub Actions after trust is established)
- Scheduled/cron runs (future — `glitch ask` is manual for now)
- Image generation or screenshots
- Custom domain setup (already configured via CNAME)

## Dependencies

- **Astro** installed in `site/` (`npm create astro@latest`)
- **Ollama** running locally with qwen2.5:7b (already a hard requirement)
- **glitch** built and on PATH
- Tier 1 provider configured (OpenRouter or equivalent)
- Tier 2 provider available (Claude or Copilot CLI)
