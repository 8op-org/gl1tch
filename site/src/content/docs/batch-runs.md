---
title: "Batch Runs"
description: "Run variants side by side, multi-provider comparison, fan-out"
order: 9
---

Run the same workflow against multiple providers or models simultaneously and compare the results. Batch runs are useful for benchmarking, validating that a workflow holds up across different backends, or discovering which model best fits a particular task.

This page covers the `--variant` and `--compare` flags on `glitch run`. For the `(compare ...)` DSL form that compares branches inside a single workflow, see [Compare Runs](/docs/compare).

## Running variants

`--variant` wraps every LLM step in your workflow with an implicit compare block — one branch per variant. Each variant is a `provider:model` pair:

```bash
glitch run code-review \
  --variant ollama:qwen3:8b \
  --variant claude \
  --variant copilot
```

You need at least two variants. With only one, `glitch` prints a warning and runs normally.

Each LLM step in `code-review` is replaced at runtime with a compare block that runs all three variants in parallel, then scores and selects a winner. Your workflow file is unchanged — the injection is ad-hoc.

### Adding review criteria

Score results against named criteria with `--review-criteria`:

```bash
glitch run code-review \
  --variant ollama:qwen3:8b \
  --variant claude \
  --review-criteria "accuracy,actionability,conciseness"
```

Criteria are a comma-separated list. The scoring judge (your local model) evaluates each variant's output on every criterion and selects the highest-scoring branch.

## Comparing sibling workflow files

`--compare` discovers variant workflows by naming convention rather than injecting at runtime. You author separate workflow files for each variant, and `glitch run --compare` finds and runs them together.

### Naming convention

Variant workflows follow the pattern `<base-name>-<variant>.glitch`:

```
workflows/
├── pr-review.glitch           # base workflow (unused by --compare directly)
├── pr-review-local.glitch     # local Ollama variant
├── pr-review-claude.glitch    # Claude variant
├── pr-review-copilot.glitch   # Copilot variant
└── pr-review-gemma.glitch     # Gemma variant
```

Run them all and cross-review:

```bash
glitch run pr-review --compare
```

`glitch` looks for any `pr-review-<variant>.glitch` files alongside the base workflow, runs each one, then scores the outputs against each other. You need at least two variant files — with fewer, `glitch` exits with an error.

### Default variants

When you pass `--compare` without explicit `--variant` flags, `glitch` looks for these variant suffixes by default:

```
local  claude  copilot  gemma  grok
```

You can narrow or extend the search with `--variant`:

```bash
# Only look for pr-review-local.glitch and pr-review-claude.glitch
glitch run pr-review --compare --variant local --variant claude
```

## Results organization

Results write to your workspace's `results/` directory (or `.glitch/results/` if no workspace is active). Each variant run gets its own subdirectory:

```
results/
└── pr-review-r8k2/
    ├── run.json
    ├── pr-review-local-a1b2/
    │   ├── run.json
    │   └── evidence/
    ├── pr-review-claude-c3d4/
    │   ├── run.json
    │   └── evidence/
    └── compare.json           # cross-review scores and winner
```

Use `--results-dir` to redirect output to a specific path:

```bash
glitch run pr-review --compare --results-dir ./my-results
```

## Full example: three-way model comparison

Suppose you have a summarization workflow and want to know which model produces the best output. Write three variant files:

```glitch
;; summarize-local.glitch
(workflow "summarize-local"
  :description "Summarize with local Ollama model"
  (step "summary"
    (llm :provider "ollama" :model "qwen3:8b"
      :prompt "Summarize this in 3 bullets: ~param.input")))
```

```glitch
;; summarize-claude.glitch
(workflow "summarize-claude"
  :description "Summarize with Claude"
  (step "summary"
    (llm :provider "claude"
      :prompt "Summarize this in 3 bullets: ~param.input")))
```

```glitch
;; summarize-copilot.glitch
(workflow "summarize-copilot"
  :description "Summarize with Copilot"
  (step "summary"
    (llm :provider "copilot"
      :prompt "Summarize this in 3 bullets: ~param.input")))
```

Run all three and compare:

```bash
glitch run summarize \
  --compare \
  --set input="$(cat article.txt)" \
  --review-criteria "accuracy,brevity,clarity"
```

`glitch` runs all three variants and produces a `compare.json` with scores for each criterion.

## Workspace-scoped batch runs

When an active workspace is set, variant workflows resolve from `<workspace>/workflows/` and results write to `<workspace>/results/`. No extra flags needed — workspace resolution is automatic.

See [Workspaces](/docs/workspaces) for workspace setup and resource scoping.

## Next steps

- [Compare Runs](/docs/compare) — the `(compare ...)` DSL form for in-workflow branch comparison
- [Workspaces](/docs/workspaces) — workspace-scoped batch runs and result organization
- [Providers & Config](/docs/providers) — configuring which providers and models are available
