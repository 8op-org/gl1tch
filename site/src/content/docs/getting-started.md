---
title: "Getting Started"
order: 2
description: "brew install 8op-org/tap/glitch"
---

## Install

```bash
brew install 8op-org/tap/glitch
```

gl1tch routes LLM steps through a local model by default. We recommend [LM Studio](https://lmstudio.ai) — download it, enable the local server in **Settings → Server** (port 1234), and pull **qwen3-8b**. gl1tch auto-detects it and will download missing models on demand.

Alternatively, you can use Ollama:

```bash
brew install ollama && ollama pull qwen2.5:7b
```

See [Local Models](/docs/local-models) for GPU tuning, model recommendations, and how to use both providers together.

You also need GitHub CLI authenticated:

```bash
gh auth status
```

Verify glitch:

```bash
glitch --help
```

## Your first workflow

gl1tch ships with example workflows. Run one:

```bash
glitch workflow run hello-sexpr
```

That runs `examples/hello.glitch`:

````glitch
;; hello.glitch — example gl1tch s-expression workflows
;;
;; Run with: glitch workflow run hello-sexpr

(def model "qwen2.5:7b")
(def provider "ollama")

(workflow "hello-sexpr"
  :description "Demo s-expression workflow format"

  (step "gather"
    (run "echo 'hello from a .glitch workflow'"))

  (step "respond"
    (llm
      :provider provider
      :model model
      :prompt ```
        You received this message from a shell command:
        ~(step gather)

        Respond with a short, enthusiastic acknowledgment.
        ```)))
````

What each part does:

- `(def model "qwen2.5:7b")` — binds a constant you reference by name anywhere in the file
- `(workflow "hello-sexpr" ...)` — declares the workflow. The string is the name you pass to `glitch workflow run`
- `(step "gather" (run "..."))` — runs a shell command and captures stdout
- `(step "respond" (llm ...))` — sends a prompt to your local model
- `~(step gather)` — injects the previous step's output into the prompt
- Triple backticks delimit multiline strings, auto-dedented

## Your first ask

```bash
glitch ask "review this code"
```

`glitch ask` reads your question, picks the best matching workflow using your local LLM, and runs it. Routing happens entirely on your machine — nothing leaves it unless a workflow's provider says otherwise.

The query above matches the `code-review` workflow, which reviews your staged git changes:

````glitch
;; code-review.glitch

(def model "qwen2.5:7b")

(workflow "code-review"
  :description "Review staged git changes and flag issues"

  (step "diff"
    (run "git diff --cached"))

  (step "files"
    (run "git diff --cached --name-only"))

  (step "review"
    (llm
      :model model
      :prompt ```
        You are a code reviewer. Review this diff carefully.

        Files changed:
        ~(step files)

        Diff:
        ~(step diff)

        For each file, note:
        - Bugs or logic errors
        - Security concerns
        - Naming or style issues

        If everything looks good, say so. Be concise.
        ```)))
````

Shell steps fetch the data (free, deterministic). LLM steps make sense of it (expensive, so feed pre-processed data). This is the core pattern.

## Writing your own workflow

Create `.glitch/workflows/my-workflow.glitch`:

````glitch
(def model "qwen2.5:7b")

(workflow "my-workflow"
  :description "What it does"

  (step "gather"
    (run "your shell command here"))

  (step "respond"
    (llm
      :model model
      :prompt ```
        Here is what the shell returned:
        ~(step gather)

        Do something useful with it.
        ```)))
````

Run it:

```bash
glitch workflow run my-workflow
```

Pass runtime values with `--set`:

```bash
glitch workflow run parameterized --set repo=my-project
```

Inside the workflow, `~param.repo` expands to `my-project`.

## Chaining steps

Every step's output is available to later steps via `~(step id)`. Chain as many as you need:

````glitch
;; multi-step-chain.glitch

(def model "qwen2.5:7b")

(workflow "multi-step-chain"
  :description "Gather system info, analyze it, then produce recommendations"

  (step "disk"
    (run "df -h / | tail -1"))

  (step "memory"
    (run "vm_stat | head -5"))

  (step "processes"
    (run "ps aux --sort=-%mem | head -10"))

  (step "analyze"
    (llm
      :model model
      :prompt ```
        Analyze this system snapshot:

        Disk usage:
        ~(step disk)

        Memory:
        ~(step memory)

        Top processes by memory:
        ~(step processes)

        Give a brief health assessment and flag anything concerning.
        ```)))
````

Shell steps are free. Use as many as you need to shape the data before the LLM sees it.

## Saving output

Write any step's output to a file with `(save ...)`:

````glitch
;; git-changelog.glitch

(def model "qwen2.5:7b")

(workflow "git-changelog"
  :description "Summarize recent git commits into a human-readable changelog"

  (step "commits"
    (run "git log --oneline --no-decorate -20"))

  (step "changelog"
    (llm
      :model model
      :prompt ```
        Here are the last 20 git commits:
        ~(step commits)

        Write a concise changelog grouped by theme (features, fixes, chores).
        Use markdown. No preamble.
        ```))

  (step "save-it"
    (save "results/changelog.md" :from "changelog")))
````

## Workflow discovery

Workflows are discovered from these locations:

- `.glitch/workflows/` in your current project — project-local
- `~/.config/glitch/workflows/` — user-global

Project-local workflows override globals with the same name.

```bash
glitch workflow list
```

## Knowledge index

gl1tch can index your repos into Elasticsearch and query them in natural language. Start the stack with Docker:

```bash
glitch up
```

This launches Elasticsearch and Kibana via Docker Compose and seeds default dashboards. Once running:

```bash
glitch index --repo .              # index the current repo
glitch observe "PRs that failed CI this week"
```

To stop the stack:

```bash
glitch down
```

Requires Docker Desktop or `docker compose` on your machine.

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — the full s-expression reference with control flow, tiered routing, and every form
- [Plugins](/docs/plugins) — reusable data-gathering subcommands you compose into workflows
- [Compare Runs](/docs/compare) — A/B test different models or strategies in a single workflow
- [DSL Reference](/docs/dsl-reference) — threading, filtering, reducing, and Elasticsearch forms