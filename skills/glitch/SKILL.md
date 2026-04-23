---
name: glitch
description: >
  gl1tch CLI workflow engine — chains shell commands and LLM calls into
  automated workflows. Use when creating, editing, running, or debugging
  .glitch workflow files, configuring providers, using the MCP server,
  querying code intelligence, writing plugins, running investigations,
  or when a task could be a glitch workflow. Also use when reviewing or
  authoring .glitch files.
license: MIT
compatibility: Requires bb (Babashka), curl, gh (GitHub CLI), rg (ripgrep), sg (ast-grep)
metadata:
  author: 8op-org
  version: "0.3.0"
---

# glitch

gl1tch is a workflow engine built on Babashka (Clojure) that orchestrates shell commands and LLM calls. Core design principle: **shell does the grunt work, LLM does the thinking.** Shell steps fetch and shape data (free, deterministic). LLM steps reason about it (expensive, so feed pre-processed data).

## Installation

```bash
cd ~/Projects/gl1tch/bb
bb install
```

**Prerequisites:** Babashka (`bb`), `curl`, `gh` (authenticated), `rg` (ripgrep), `sg` (ast-grep). Run `glitch up` to verify.

## CLI Commands

```bash
glitch run <file> [input...]          # execute a .glitch or .clj file
glitch run <file> -p claude           # use specific provider
glitch run <file> -s key=value        # set parameter (repeatable)
glitch run <file> -m qwen3:8b         # override model
glitch check <file>                   # validate syntax
glitch eval <file>                    # evaluate a .clj script
glitch up                             # verify required tools
glitch repl                           # start nREPL on port 1667
glitch mcp                            # start MCP server (JSON-RPC stdio)
glitch index                          # index current repo (code intelligence)
glitch index sources                  # list available index sources + detection
glitch index --sources github         # index GitHub issues/PRs
glitch index --sources gitlab --no-code --since 2026-01-01  # GitLab only
glitch index query --from github --name "rate"  # query source index
glitch plugin list                    # list plugins
glitch version                        # prints version
```

## Workflow Syntax

Workflows are `.glitch` files — Clojure s-expressions in a sandbox with glitch primitives pre-bound. Save to `.glitch/workflows/` for project-local discovery.

```clojure
(workflow "name"
  :description "what it does"

  ;; Shell step — runs bash -c
  (step "fetch"
    (sh "gh issue view 42 --json title,body"))

  ;; LLM step — calls configured provider
  (step "analyze"
    (llm
      :provider "claude"
      :model "claude-haiku-4-5-20251001"
      :prompt (str "Analyze this issue:\n" (ref "fetch"))))

  ;; Save output to file
  (save "results/output.md" (ref "analyze")))
```

### Key Primitives

| Form | Description |
|------|-------------|
| `(workflow "name" :description "..." body...)` | Workflow definition |
| `(step "id" body)` | Execute body, record output under id |
| `(ref "id")` | Retrieve output of a previous step |
| `(input)` | User input passed to the workflow |
| `(param "key")` | Parameter from `--set key=value` |
| `(sh "command")` | Shell command (bash -c), returns stdout |
| `(llm :prompt "..." :provider "..." :model "...")` | LLM invocation |
| `(save "path" content)` | Write content to file |
| `(par (step ...) (step ...))` | Concurrent execution |

Triple-backtick strings with `~(ref "id")` interpolation:

```clojure
(step "summarize"
  (llm :prompt ```
    Summarize this data:
    ~(ref "fetch")
    User asked: ~(input)
    ```))
```

For the full DSL — all primitives, LLM options, control flow, validation, and utilities — see [references/dsl-reference.md](references/dsl-reference.md).

## The Cardinal Rule: Shell First, LLM Last

**Shell steps** collect and prepare data — `gh`, `git`, `curl`, `jq`, date math. Fast, deterministic, free.

**LLM steps** synthesize the result — summarizing, prioritizing, formatting, judgment calls. Expensive, so feed pre-processed data.

**Anti-patterns:** asking the LLM to parse JSON (use `jq`), calculate dates (use `date`), or make API calls (use `sh`).

## Patterns

### Fetch + format

```clojure
(workflow "git-status"
  :description "Summarize current git state"
  (step "status" (sh "git status --short"))
  (step "summary"
    (llm :prompt (str "Summarize this git status for a developer:\n" (ref "status")))))
```

### Multi-source aggregation

```clojure
(workflow "morning-briefing"
  :description "Aggregate multiple sources into daily briefing"
  (par
    (step "prs" (sh "gh pr list --author @me --json number,title,state | jq '.'"))
    (step "reviews" (sh "gh pr list --search 'review-requested:@me' --json number,title,url | jq '.'"))
    (step "issues" (sh "gh issue list --assignee @me --json number,title,labels | jq '.'")))
  (step "briefing"
    (llm :prompt ```
      Create a morning briefing from these sources:
      My PRs: ~(ref "prs")
      Pending reviews: ~(ref "reviews")
      My issues: ~(ref "issues")
      Format: bullet list, no emoji, terse.
      ```)))
```

### Parameterized

```clojure
(workflow "analyze-issue"
  :description "Analyze a GitHub issue"
  (step "fetch"
    (sh (str "gh issue view " (param "issue") " --repo " (param "repo") " --json number,title,body")))
  (step "analyze"
    (llm :prompt (str "Analyze this issue:\n" (ref "fetch"))))
  (save (str "results/" (param "repo") "/" (param "issue") ".md") (ref "analyze")))
```

Run with: `glitch run analyze-issue.glitch -s issue=3442 -s repo=elastic/ensemble`

## When to Use Glitch

Before implementing a multi-step task, check if glitch adds value. Call `glitch_advise` with the task description when you notice any of these signals:

| Signal | What it looks like |
|--------|--------------------|
| **Repetition** | Task will be done again or across multiple targets — "check all PRs", "audit these repos", "review each deploy" |
| **Confidence** | Judgment where being wrong matters — "is this summary accurate?", "which approach is better?", "is this finding real?" |
| **Multi-source** | Needs information composed from multiple providers or tools — "compare what Claude and Copilot say", "aggregate from 3 APIs" |
| **Investigation** | Uncertain facts, contradictions, structured reasoning — "figure out why this is failing", "is this finding real?" |

**How to use it:**

1. Call `glitch_advise` with the task description
2. If approach is `"none"` — do the task natively, glitch adds no value
3. If approach is `"workflow"` — check `existing_workflows` first, then `glitch_run` or create a new workflow
4. If approach is `"primitive"` — use `glitch_eval` with the suggested example
5. If approach is `"repl"` — this is exploratory, use `glitch_eval` iteratively

**When NOT to check:** Single-step tasks, file reads, simple grep, git operations. If you can do it in one tool call, skip glitch.

## Reference Files

| File | When to read |
|------|-------------|
| [references/dsl-reference.md](references/dsl-reference.md) | Full DSL — all primitives, LLM options, control flow, validation, utilities |
| [references/providers.md](references/providers.md) | Provider registry, custom providers, fallback chain |
| [references/investigation.md](references/investigation.md) | Investigation graphs, Bayesian reasoning, fact tracking |
| [references/code-intelligence.md](references/code-intelligence.md) | `glitch index`, symbol/edge queries, Elasticsearch setup |
| [references/mcp-server.md](references/mcp-server.md) | MCP tools, IDE configuration |
| [references/plugins.md](references/plugins.md) | Plugin system, naming conventions, defcommand |

## Project Reference

- **Repo**: `8op-org/gl1tch` at `~/Projects/gl1tch`
- **Runtime**: Babashka (Clojure)
- **Entry point**: `bb/src/glitch/main.clj`
- **Config**: `~/.config/glitch/providers/`
- **Store**: `~/.local/share/glitch/glitch.edn`
- **Plugins**: `~/.config/glitch/plugins/` (global), `.glitch/plugins/` (project)
- **Environment**: `~/.config/glitch/.env` and `./.env`

### Development

```bash
cd ~/Projects/gl1tch/bb
bb install        # install to ~/.local/bin
bb test           # run test suite
bb clean          # clean build artifacts
```
