---
title: "Getting Started"
order: 3
description: "brew install 8op-org/tap/glitch"
---

## Install

```bash
brew install 8op-org/tap/glitch
```

gl1tch routes LLM steps through LM Studio by default. Install it and load a model:

1. Download [LM Studio](https://lmstudio.ai) and launch it
2. Go to **Settings > Server** and enable the local server (port 1234)
3. Pull **qwen3-8b** — best general-purpose model at this size

You also need GitHub CLI authenticated:

```bash
gh auth status
```

Verify glitch:

```bash
glitch --help
```

## Your first workflow

Create a project directory with `glitch init`:

```bash
mkdir my-project && cd my-project
glitch init
```

That creates `.glitch/workflows/`. Drop a workflow file in there:

````glitch
;; .glitch/workflows/hello.glitch

(def model "qwen3-8b")

(workflow "hello"
  :description "Demo s-expression workflow"

  (step "gather"
    (run "echo 'hello from a .glitch workflow'"))

  (step "respond"
    (llm
      :model model
      :prompt ```
        You received this message from a shell command:
        ~(step gather)

        Respond with a short, enthusiastic acknowledgment.
        ```)))
````

Run it:

```bash
glitch run hello
```

What each part does:

- `(def model "qwen3-8b")` — binds a constant you reference by name anywhere in the file
- `(workflow "hello" ...)` — declares the workflow. The string is the name you pass to `glitch run`
- `(step "gather" (run "..."))` — runs a shell command and captures stdout
- `(step "respond" (llm ...))` — sends a prompt to your local model
- `~(step gather)` — injects the previous step's output into the prompt
- Triple backticks delimit multiline strings, auto-dedented

## Code review workflow

Here's a more practical example — reviewing staged git changes:

````glitch
;; .glitch/workflows/code-review.glitch

(def model "qwen3-8b")

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
(def model "qwen3-8b")

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
glitch run my-workflow
```

Pass runtime values with `--set`:

```bash
glitch run my-workflow --set repo=my-project
```

Inside the workflow, `~param.repo` expands to `my-project`.

## Chaining steps

Every step's output is available to later steps via `~(step id)`. Chain as many as you need:

````glitch
;; .glitch/workflows/system-health.glitch

(def model "qwen3-8b")

(workflow "system-health"
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
;; .glitch/workflows/git-changelog.glitch

(def model "qwen3-8b")

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

## Project structure

gl1tch discovers workflows from the `.glitch/workflows/` directory in your project. Run `glitch init` to scaffold it, or create it by hand:

```
my-project/
  .glitch/
    workflows/
      hello.glitch
      code-review.glitch
    glitch.db          # run history (auto-created)
  src/
  ...
```

gl1tch walks up from your current directory looking for `.glitch/`. You can also pass `--project <path>` to point at a specific project root.

## CLI reference

```bash
glitch run <workflow> [input]     # run a workflow
glitch run -p <file>              # run a specific file
glitch run <wf> --set key=value   # pass parameters
glitch check <file>               # validate syntax
glitch init                       # create .glitch/workflows/
glitch up                         # check required tools
glitch plugin <name> [args]       # run a plugin
glitch mcp                        # start MCP server
glitch gui                        # launch web GUI
glitch version                    # show version
```

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — the complete reference for every form
- [Providers](/docs/providers) — configure LM Studio, Copilot, Claude, OpenRouter
- [Local Models](/docs/local-models) — model recommendations, GPU tuning, context length
- [Tool Use](/docs/tool-use) — let your LLM steps search code and read files
