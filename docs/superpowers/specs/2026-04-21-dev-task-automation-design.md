# Dev Task Automation Design

## Overview

A top-level `bb.edn` orchestrator that provides a single `bb <task>` entry point for all dev tasks. Deterministic build tasks delegate to the existing `bb/bb.edn`. Dev workflows that benefit from LLM involvement delegate to glitch workflows in `dev/workflows/`.

All glitch workflow outputs land in `dev/output/` as drafts — nothing auto-commits or auto-pushes.

## File Layout

```
gl1tch/
  bb.edn                        # top-level orchestrator (NEW)
  bb/
    bb.edn                      # existing — build/test/install/clean
    src/...
    providers/...
  dev/
    workflows/
      dev-setup.glitch          # verify dev environment
      dev-docs.glitch           # extract source → LLM → doc draft
      dev-release.glitch        # build + test + changelog draft
      dev-review.glitch         # LLM reviews uncommitted diff
    output/                     # generated drafts land here (gitignored)
  workflows/                    # existing — site/product workflows
  site/
```

## Task Routing

| Entry point | Engine | Target |
|-------------|--------|--------|
| `bb build` | bb | `bb/bb.edn` build task |
| `bb test` | bb | `bb/bb.edn` test task |
| `bb install` | bb | `bb/bb.edn` install task |
| `bb clean` | bb | `bb/bb.edn` clean task |
| `bb site:dev` | npm | `site/` Astro dev server |
| `bb site:build` | npm | `site/` Astro build |
| `bb dev:setup` | glitch | `dev/workflows/dev-setup.glitch` |
| `bb dev:docs` | glitch | `dev/workflows/dev-docs.glitch` |
| `bb dev:release` | glitch | `dev/workflows/dev-release.glitch` |
| `bb dev:review` | glitch | `dev/workflows/dev-review.glitch` |

## Top-level bb.edn

The orchestrator delegates via `shell`. Build tasks shell into `bb/` with `{:dir "bb"}`. Dev tasks shell out to `glitch workflow run <name>` pointing at `dev/workflows/`. Site tasks shell into `site/`.

## Workflow Designs

### dev-setup.glitch

Pure shell, no LLM. Checks each dependency and prints pass/fail summary.

Dependencies verified:
- **bb** — `bb --version`
- **rg** (ripgrep) — `rg --version`
- **gh** — `gh auth status`
- **LM Studio** — `curl -sf http://localhost:1234/v1/models`
- **OPENROUTER_API_KEY** — sourced from `~/.env`
- **Providers** — checks `~/.config/glitch/providers/` for lmstudio, claude, copilot, openrouter `.clj` files
- **glitch** — `glitch --version`

Final step prints a summary table of all checks.

### dev-docs.glitch

Shell extracts, LLM writes. Provider: copilot.

Steps:
1. **extract-namespaces** — `find` + `rg` across `bb/src/glitch/` for all `defn`/`def` with docstrings
2. **extract-cli** — `rg` for CLI command surface in main.clj
3. **extract-providers** — `rg` for `register` calls in providers, cat config.yaml
4. **existing-docs** — `ls site/src/content/docs/` to avoid duplication
5. **generate** (LLM) — technical writer prompt, user-first framing, examples before explanation, no internals. Outputs markdown sections with Astro frontmatter.
6. **save** — writes to `dev/output/docs-draft.md`

From there, review manually and feed into `site-create-page` or `site-update-page`.

### dev-release.glitch

Build/test gate, then LLM changelog. Provider: copilot.

Steps:
1. **quality phase** — gates on `bb test` and `bb build` (both must pass)
2. **last-tag** — `git describe --tags --abbrev=0`
3. **changelog-raw** — `git log --oneline --no-merges` since last tag
4. **diff-stat** — `git diff --stat` since last tag
5. **changelog** (LLM) — groups by Added/Changed/Fixed/Removed, one line per change
6. **save** — writes to `dev/output/changelog-draft.md`

Does not auto-tag or auto-push. Review draft, then tag manually.

### dev-review.glitch

Pre-commit quality check. Provider: copilot.

Steps:
1. **staged** — `git diff --cached`
2. **unstaged** — `git diff`
3. **untracked** — `git ls-files --others --exclude-standard`
4. **review** (LLM) — checks for bugs, API compat breaks, missing error handling at boundaries, test gaps, accidental secrets/debug code
5. **save** — writes to `dev/output/review-draft.md`

## Design Decisions

- **bb.edn over Taskfile.yml** — stays in the bb ecosystem, avoids adding go-task as a third dependency
- **dev/workflows/ separate from workflows/** — keeps dev tooling isolated from product/site workflows
- **dev/output/ for drafts** — all LLM output is a draft, never auto-committed. Gitignored.
- **copilot as default provider** — strong reasoning, available via gh. Workflows can override per-step.
- **No auto-commit/push anywhere** — every workflow produces output for human review
