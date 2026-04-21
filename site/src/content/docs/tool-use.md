---
title: "Tool Use"
order: 7
description: "How LLM steps call built-in tools during inference — search, grep, file read, workflow execution"
---

## Tool Use

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it yet, start there.

When an `(llm ...)` step runs, gl1tch gives the model a set of built-in tools — code search, file read, grep, workflow execution. Tool-use-aware providers let the model call those tools during inference, inspect the results, and continue reasoning before producing a final answer. You don't orchestrate the tool calls; the model decides what to reach for and when.

## A tool-use step

By default, every `(llm ...)` step runs with tools available. The model calls them or not, on its own:

````glitch
(step "investigate"
  (llm
    :provider "lmstudio"
    :model "qwen3-8b"
    :prompt ```
      I need to understand the authentication flow in this codebase.
      Search for the relevant files and explain how it works.
      ```))
````

The model can call `glitch_search`, `glitch_grep`, or `glitch_read_file` on its own initiative. Results come back as tool responses; the model continues until it has enough context to write its final answer.

To disable tools on a step that doesn't need them — summarizing, classifying, writing — pass `:tools []`:

````glitch
(step "summarize"
  (llm
    :tools []
    :prompt ```
      Summarize this implementation plan:
      ~(step research)
      ```))
````

## Available tools

gl1tch exposes these tools to every LLM step:

| Tool | What it does |
|------|--------------|
| `glitch_search` | Hybrid semantic + keyword search across indexed repositories |
| `glitch_grep` | Regex search in files |
| `glitch_read_file` | Read a file (first 200 lines) |
| `glitch_symbols` | Search function and type names in indexed code |
| `glitch_run` | Run a workflow by name |
| `glitch_eval` | Evaluate an expression in the workflow runtime |
| `glitch_check` | Check a workflow file for syntax errors |
| `glitch_index` | Index or reindex a repository |

`glitch_search`, `glitch_index`, and `glitch_symbols` require an indexed repository (run `glitch index` once first). `glitch_grep` and `glitch_read_file` work on the filesystem directly — no index needed.

## Selecting tools

Pass `:tools` to control which tools the model sees:

````glitch
;; All tools — default, omit :tools entirely
(step "analyze"
  (llm
    :prompt "Explain this module's architecture..."))

;; Specific tools only
(step "search-code"
  (llm
    :tools ["glitch_search" "glitch_read_file"]
    :prompt "Find the error handling pattern used in this project."))

;; No tools — plain text generation
(step "classify"
  (llm
    :tools []
    :prompt ```
      Classify this issue. Respond with ONLY valid JSON:
      {"type": "bug|feature|refactor", "complexity": "small|medium|large"}

      ISSUE: ~(step fetch-issue)
      ```))
````

Use `[]` for classification, summarization, and writing steps where tool calls would add latency without benefit. Use a specific list when you want to constrain what the model can reach — `["glitch_grep" "glitch_read_file"]` is a good default for code review steps that should stay in the filesystem.

## Agentic mode

By default, tool use is single-round: the model calls tools once, sees the results, then produces its final answer. Set `:agentic true` to let the model call tools repeatedly — it keeps going until it's satisfied or hits `:max-rounds`:

````glitch
(step "deep-investigation"
  (llm
    :provider "lmstudio"
    :agentic true
    :max-rounds 8
    :prompt ```
      Find all places where the config file is loaded and trace
      how the values flow through the application.
      ```))
````

The default `:max-rounds` is `5`. When the limit is hit, gl1tch forces a final text response using all accumulated tool results.

Use agentic mode for open-ended investigations where you can't predict how many search passes the model will need. Use the default single-round mode for targeted steps where you know roughly what the model should look up.

## Provider support

Tool use requires a provider that supports the OpenAI tool-calling protocol or integrates via MCP:

| Provider | Tool support | Notes |
|----------|-------------|-------|
| `lmstudio` | Native (OpenAI format) | Default provider; falls back to plain call on tool failure |
| `openrouter` | Native (OpenAI format) | Model must support function calling |
| `copilot` | MCP | Requires `copilot` CLI installed and authenticated |
| `claude` | MCP | Requires `claude` CLI installed and authenticated |

For local tool use, `lmstudio` with `qwen3-8b` is the recommended setup — best tool-use quality at that size, runs fully offline. See [Local Models](/docs/local-models) for setup.

## Mixing tool use with classic steps

Tools add power but cost inference rounds. Classic orchestration — shell steps gathering data, one LLM step reasoning over the result — is faster and more predictable for structured tasks. Use both in the same workflow:

````glitch
(workflow "repo-audit"
  :description "Classic shell gather + tool-use analysis step"

  ;; Classic: deterministic, fast, no tool rounds
  (step "structure"
    (run "find . -name '*.go' -maxdepth 3 | head -50"))

  (step "recent"
    (run "git log --oneline -10"))

  ;; Tool use: model can search for additional context it needs
  (step "audit"
    (llm
      :provider "lmstudio"
      :tools ["glitch_search" "glitch_grep"]
      :prompt ```
        Here is the repo structure and recent commits:

        ~(step structure)

        ~(step recent)

        Identify any security or architecture concerns. Use the search
        tools to investigate anything that warrants a closer look.
        ```))

  (step "save-audit"
    (save "results/audit.md" :from "audit")))
````

Shell steps give the model a fast overview. Tools let it dig deeper when something warrants it, without you deciding up front what to fetch.

## Routing step with tools disabled

For steps that classify or route — where you want a precise, fast response with no model-initiated lookups — disable tools and use `:format "json"` together:

````glitch
;; From site.glitch — routes freeform input to the right workflow operation
(step "route"
  (llm
    :provider "openrouter"
    :tools []
    :prompt (str
      "Classify the user's intent into exactly one action.\n"
      "Return a JSON object (no markdown fence, no preamble).\n\n"
      "User said: " (input) "\n\n"
      "Existing pages:\n" (ref "pages"))))
````

No tools, no ambiguity — the model sees your prompt and responds immediately.

## Indexing for code search

`glitch_search` and `glitch_symbols` require a local index. Run this once per repository:

```bash
glitch index
```

After that, any tool-use step can call `glitch_search` and get semantic + keyword results over your codebase. `glitch_grep` and `glitch_read_file` always work without an index.

If the index is stale, pass `:reindex true` via the tool or re-run `glitch index`.

## LLM options for tool use

Full set of `(llm ...)` options relevant to tool use:

| Option | Values | What it does |
|--------|--------|-------------|
| `:tools` | `[]`, `["name" ...]`, omit | Tools to expose. Omit = all. `[]` = none. |
| `:agentic` | `true` / `false` (default) | Multi-round tool calling when `true` |
| `:max-rounds` | integer (default `5`) | Max tool-call rounds in agentic mode |
| `:provider` | `"lmstudio"`, `"openrouter"`, `"copilot"`, `"claude"` | Must support tool use |
| `:model` | model identifier | e.g. `"qwen3-8b"` |

## Next steps

- [Local Models](/docs/local-models) — set up LM Studio and qwen3-8b for local tool use
- [Workflow Syntax](/docs/workflow-syntax) — the full `(llm ...)` reference with all keyword options
- [Plugins](/docs/plugins) — reusable subcommands for deterministic data gathering