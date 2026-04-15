---
title: "Workflow Syntax"
order: 2
description: "S-expression workflow reference for glitch"
---

## Overview

gl1tch workflows are `.glitch` files written in s-expression syntax — every construct is a parenthesized list:

```
(form arg1 arg2 :keyword value)
```

Place your workflow files in `.glitch/workflows/` for automatic discovery.

## Workflow structure

A workflow wraps one or more steps under a name and optional description:

````glitch
;; code-review.glitch — review staged changes before committing
;;
;; Run with: glitch workflow run code-review

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
        {{step "files"}}

        Diff:
        {{step "diff"}}

        For each file, note:
        - Bugs or logic errors
        - Security concerns
        - Naming or style issues

        If everything looks good, say so. Be concise.
        ```)))
````

Run it:

```bash
glitch workflow run code-review
```

## Definitions

`(def name value)` binds a constant you can reuse across steps. Define your model and provider once at the top, then reference the names wherever a value is expected:

```glitch
(def model "qwen2.5:7b")
(def provider "ollama")
```

```glitch
(llm :provider provider :model model ...)
```

## Steps

Every step has an `id` and a single action.

**Shell step** — runs a command and captures stdout:

```glitch
(step "diff"
  (run "git diff --cached"))
```

**LLM step** — sends a prompt to a language model:

````glitch
(step "changelog"
  (llm
    :model model
    :prompt ```
      Here are the last 20 git commits:

      {{step "commits"}}

      Write a concise changelog grouped by theme (features, fixes, chores).
      Use markdown. No preamble.
      ```))
````

**Save step** — writes a step's output to a file:

```glitch
(step "save-it"
  (save "results/changelog.md" :from "changelog"))
```

## Step references

`{{step "id"}}` inserts a named step's output into any prompt or shell command. `{{.input}}` inserts the value your workflow was invoked with. `{{.param.key}}` inserts a runtime value passed with `--set`:

````glitch
;; parameterized.glitch
;;
;; Run with: glitch workflow run parameterized --set repo=gl1tch

(def model "qwen2.5:7b")

(workflow "parameterized"
  :description "Show how to pass runtime parameters into a workflow"

  (step "info"
    (run "echo 'Analyzing repo: {{.param.repo}}'"))

  (step "structure"
    (run "find {{.param.repo}} -maxdepth 2 -type f | head -30"))

  (step "summary"
    (llm
      :model model
      :prompt ```
        Here is the file tree for {{.param.repo}}:

        {{step "structure"}}

        Describe the project structure in 3-4 sentences.
        What kind of project is this?
        ```))

  (step "save-it"
    (save "results/{{.param.repo}}/summary.md" :from "summary")))
````

```bash
glitch workflow run parameterized --set repo=gl1tch
```

## LLM options

| Option | Values | Description |
|--------|--------|-------------|
| `:provider` | `"ollama"`, `"claude"`, `"copilot"`, `"gemini"`, custom | LLM backend |
| `:model` | model identifier | e.g. `"qwen2.5:7b"` |
| `:skill` | skill name | Prepends skill context to your prompt |
| `:format` | `"json"` or `"yaml"` | Validates output as structured data |
| `:tier` | `0`, `1`, `2` | Pin to a specific cost tier |

Here's `:skill` in action — `reviewer-verify` context is prepended to your prompt automatically:

```glitch
(workflow "agent-with-skill"
  :description "Demonstrates the agent executor with skill injection in s-expression format."

  (step "diff"
    (run "git diff --cached --stat && git diff --cached"))

  (step "review"
    (llm
      :provider "claude"
      :skill "reviewer-verify"
      :prompt "Review these staged changes for correctness, security, and style:\n\n{{step \"diff\"}}"))

  (step "save-review"
    (save "review-output.md" :from "review")))
```

## Tiered cost routing

Omit `:provider` and `:tier` and gl1tch routes the step through tiers automatically:

- **Tier 0** — local (ollama), free, runs first
- **Tier 1** — cheap cloud provider
- **Tier 2** — premium model

After each non-final tier, the output is self-evaluated for quality. If it passes, routing stops. If not, it escalates to the next tier.

Adding `:format "json"` or `:format "yaml"` enables structural validation — the output must parse successfully or the step escalates. Use `:tier 2` to always route a step to the premium model.

## Comments and discard

Line comments start with `;`:

```glitch
;; This is a section comment
; This is a line comment
```

`#_` discards the next form at read time — use it to toggle steps off without deleting them:

````glitch
;; discard-demo.glitch

(workflow "discard-demo"
  :description "Show how #_ discard works for toggling steps on and off"

  (step "data"
    (run "echo 'some input data'"))

  ;; This step is disabled — remove #_ to re-enable
  #_(step "expensive-analysis"
    (llm
      :model model
      :prompt ```
        Do a very thorough analysis of:
        {{step "data"}}
        ```))

  ;; This step runs instead
  (step "quick-analysis"
    (llm
      :model model
      :prompt ```
        Briefly summarize:
        {{step "data"}}
        ```)))
````

## Multiline strings

Triple backticks delimit multiline prompts inside `.glitch` files. Content is auto-dedented, so you can indent for readability without affecting the output:

````glitch
(llm
  :model model
  :prompt ```
    You are a code reviewer. Review this diff carefully.

    Files changed:
    {{step "files"}}

    Diff:
    {{step "diff"}}

    If everything looks good, say so. Be concise.
    ```)
````