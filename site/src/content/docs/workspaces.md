---
title: "Projects"
order: 8
description: "A project is any directory with a .glitch/ folder — workflows, run history, and search index scoped to your codebase."
---

A project is any directory with a `.glitch/` folder. Workflows, run history, and search index are scoped to your codebase. No config files to write, no registry to manage.

## Setup

```bash
cd ~/Projects/my-project
glitch init
```

That creates `.glitch/workflows/`. You're done — every `glitch run` from this directory (or any subdirectory) picks up your project automatically.

## What you get

- **Scoped workflows** — `.glitch/workflows/` holds workflows specific to this project
- **Run history** — `.glitch/glitch.db` tracks every run with step outputs, timing, and status
- **Code search index** — `.glitch/search.db` powers the `glitch_search` tool for LLM steps
- **Auto-discovery** — gl1tch walks up from your current directory looking for `.glitch/`

## Directory layout

```
my-project/
  .glitch/
    workflows/          # your workflow files
      code-review.glitch
      deploy-check.glitch
    glitch.db           # run history (auto-created)
    search.db           # code search index (auto-created)
    .gitignore          # ignores everything except workflows/
  src/
  ...
```

Commit `.glitch/workflows/` to your repo. The database files are gitignored automatically.

## Discovery

When you run any command, gl1tch finds your project using this precedence:

1. `--project <path>` — explicit flag
2. **CWD walk-up** — walk up from your current directory looking for `.glitch/`
3. **None** — runs in the current directory with no project context

In practice: `cd` into your project and everything works. No sticky state, no `use` commands.

## Running workflows

```bash
# Run a workflow from .glitch/workflows/
glitch run code-review

# Pass parameters
glitch run deploy-check --set env=staging

# Run a specific file by path
glitch run -p ~/shared/workflows/audit.glitch
```

Workflows in `.glitch/workflows/` are discovered by name — the filename minus `.glitch` is the name you pass to `glitch run`.

## Shared defaults

Set your default model at the top of each workflow with `(def ...)`:

```glitch
(def model "qwen3-8b")
(def provider "lmstudio")
```

Or share definitions across workflows with `(include ...)`:

```glitch
;; .glitch/workflows/shared.glitch
(def model "qwen3-8b")
(def provider "lmstudio")
```

```glitch
;; .glitch/workflows/review.glitch
(include "shared.glitch")

(workflow "review"
  (step "diff" (run "git diff --cached"))
  (step "review"
    (llm :model model :provider provider
      :prompt "Review this diff:\n~(step diff)")))
```

## Web GUI

Launch a browser-based view of your project's run history:

```bash
glitch gui
```

The GUI shows runs, step outputs, timing, and status. It reads from `.glitch/glitch.db` in your project.

## Next steps

- [Getting Started](/docs/getting-started) — install gl1tch and run your first workflow
- [Workflow Syntax](/docs/workflow-syntax) — the complete form reference
- [Plugins](/docs/plugins) — reusable data-gathering subcommands
