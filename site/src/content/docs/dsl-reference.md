---
title: "DSL Reference"
order: 2
description: "Sister page to workflow-syntax. Covers new forms shipped in the DSL improvements branch."
---

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it yet, start there — this page covers additions and expansions to what's documented there.

## LLM step — tool use and agentic mode

The `(llm ...)` step accepts three new fields that enable your step to call tools and drive multi-turn reasoning loops.

### New fields

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `:tools` | list of strings | `[]` | Tool names made available to this step. Tools are provided by the active MCP server. |
| `:agentic` | boolean | `false` | When `true`, the step runs in multi-turn mode — the model can call tools, observe results, and continue reasoning until it produces a final text response or hits `:max-rounds`. |
| `:max-rounds` | integer | `5` | Maximum number of tool-call rounds before the step stops and returns whatever the model has produced. Has no effect when `:agentic false`. |

### Usage example

````glitch
;; agentic-research.glitch
;;
;; Run with: glitch workflow run agentic-research --set repo=acme/backend --set issue=42

(def model "qwen3-8b")

(workflow "agentic-research"
  :description "Let the model search the codebase and read files to build its own context"

  (step "issue"
    (run "gh issue view ~param.issue --repo ~param.repo --json number,title,body"))

  (step "research"
    (llm
      :provider "lm-studio"
      :model model
      :tools ("glitch_search" "glitch_read_file")
      :agentic true
      :max-rounds 8
      :prompt ```
        You are a senior engineer researching issue #~param.issue.

        Issue details:
        ~(step issue)

        Use the available tools to search the codebase and read relevant files.
        When you have enough context, produce a concise implementation plan:
        1. Files to modify and why
        2. Specific changes required
        3. Any risks to flag
        ```))

  (step "save-plan"
    (save "results/~param.repo/issue-~param.issue/plan.md" :from "research")))
````

#### What happens at runtime

Without `:agentic true`, the model sees the prompt and responds once — it cannot use tools to gather more context. With `:agentic true`, each round works like this:

1. The model receives your prompt (and any tool results from prior rounds)
2. If it calls a tool, gl1tch executes it and feeds the result back
3. Repeat until the model produces a text response or `:max-rounds` is reached

Setting `:max-rounds 8` above means the model can call tools up to 8 times before gl1tch stops the loop and returns whatever the model last produced.

#### Choosing `:max-rounds`

The default of `5` is enough for targeted lookups — search, read one or two files, respond. Raise it when your prompt asks the model to explore broadly across the codebase. Lower it for classification or summarization steps where tool use is a fallback, not the primary strategy.

#### Tool names

Tool names come from your active MCP server. The built-in tools are:

| Name | What it does |
|------|-------------|
| `glitch_search` | Hybrid semantic + keyword code search across indexed repositories |
| `glitch_read_file` | Read the first 200 lines of a file |
| `glitch_grep` | Regex search in code files |
| `glitch_symbols` | Search function and type names in indexed code |
| `glitch_run` | Execute a workflow by name |
| `glitch_index` | Index or reindex a repository |
| `glitch_check` | Check a workflow file for syntax errors |
| `glitch_eval` | Evaluate an expression and return the result |

Pass only the tools your step actually needs. A focused tool list keeps the model on task and reduces unnecessary round-trips.

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — the full reference for every step form, control flow, and template syntax
- [MCP Server](/docs/mcp-server) — configure and extend the tool server your agentic steps call into
- [Local Models](/docs/local-models) — tune LM Studio for agentic workloads (GPU offload, context length)
