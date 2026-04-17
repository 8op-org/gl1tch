---
title: "Workflow Syntax"
order: 5
description: "gl1tch workflows are `.glitch` files written in s-expression syntax — parenthesized lists with keyword arguments:"
---

## Overview

gl1tch workflows are `.glitch` files written in s-expression syntax — parenthesized lists with keyword arguments:

```
(form arg1 arg2 :keyword value)
```

Drop your files in `.glitch/workflows/` for automatic discovery. This page is the complete reference for every form available.

## Workflow structure

A workflow wraps named steps. Here is a complete real-world example:

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

The string passed to `(workflow ...)` is the name you use with `glitch workflow run`.

## Definitions

`(def name value)` binds a constant for the whole file. Define your model once, reference it everywhere:

````glitch
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

Defs are simple text substitution — use them for anything you repeat: model names, provider strings, repo paths, usernames.

## Steps

Every step has an ID and a single action. The ID names the output so later steps can reference it.

### Shell step

Runs a command via `sh -c` and captures stdout:

```glitch
(step "diff"
  (run "git diff --cached"))
```

### LLM step

Sends a prompt to a language model:

````glitch
(step "changelog"
  (llm
    :model model
    :prompt ```
      Here are the last 20 git commits:

      ~(step commits)

      Write a concise changelog grouped by theme (features, fixes, chores).
      Use markdown. No preamble.
      ```))
````

### Save step

Writes a prior step's output to a file:

```glitch
(step "save-it"
  (save "results/changelog.md" :from "changelog"))
```

Paths can use interpolation: `"results/~param.repo/summary.md"`.

## Interpolation

Strings in workflows support `~` (unquote) interpolation. A bare `~name` resolves a symbol; `~(form)` evaluates an s-expression inline. Undefined references fail loud with a suggestion ("did you mean ...?").

| Expression | What it does |
|-----------|-------------|
| `~(step id)` | Insert a named step's output |
| `~(stepfile id)` | Write step output to a temp file, return the path |
| `~input` | The value passed as a trailing arg to `glitch run` |
| `~param.key` | A runtime parameter from `--set key=value` |
| `~env.VAR` | An environment variable |
| `~item` | Current item in a `map` / `filter` / `reduce` |
| `~item_index` | Zero-based index in a `map` / `filter` / `reduce` |
| `~param.accumulator` | Running accumulator in a `reduce` |
| `\~` | Literal tilde (escaped) |

Use `~(stepfile id)` when step output contains characters that break shell escaping:

```glitch
(step "process"
  (run "cat '~(stepfile big-json)' | jq '.items[]'"))
```

Full example with `--set` parameters:

````glitch
;; parameterized.glitch
;;
;; Run with: glitch workflow run parameterized --set repo=gl1tch

(def model "qwen2.5:7b")

(workflow "parameterized"
  :description "Show how to pass runtime parameters into a workflow"

  (step "info"
    (run "echo 'Analyzing repo: ~param.repo'"))

  (step "structure"
    (run "find ~param.repo -maxdepth 2 -type f | head -30"))

  (step "summary"
    (llm
      :model model
      :prompt ```
        Here is the file tree for ~param.repo:

        ~(step structure)

        Describe the project structure in 3-4 sentences.
        What kind of project is this?
        ```))

  (step "save-it"
    (save "results/~param.repo/summary.md" :from "summary")))
````

## LLM options

All keyword arguments for `(llm ...)`:

| Option | Values | What it does |
|--------|--------|-------------|
| `:prompt` | string (required) | The prompt text |
| `:provider` | `"ollama"`, `"claude"`, `"copilot"`, `"gemini"`, custom | Which LLM backend |
| `:model` | model identifier | e.g. `"qwen2.5:7b"`, `"sonnet"` |
| `:skill` | skill name | Prepends skill context to your prompt |
| `:format` | `"json"` or `"yaml"` | Validates that output parses correctly |
| `:tier` | `0`, `1`, `2` | Pin to a specific cost tier |

Using `:skill` to inject context — the skill content is prepended to your prompt automatically:

````glitch
(workflow "agent-with-skill"
  :description "Demonstrates the agent executor with skill injection in s-expression format."

  (step "diff"
    (run "git diff --cached --stat && git diff --cached"))

  (step "review"
    (llm
      :provider "claude"
      :skill "reviewer-verify"
      :prompt "Review these staged changes for correctness, security, and style:\n\n~(step diff)"))

  (step "save-review"
    (save "review-output.md" :from "review")))
````

## Tiered cost routing

When you omit both `:provider` and `:tier`, gl1tch routes automatically through tiers:

- **Tier 0** — local (ollama), free
- **Tier 1** — cheap cloud (openrouter free tier, copilot)
- **Tier 2** — premium (claude)

After each non-final tier, gl1tch self-evaluates the response quality. If it passes, routing stops. If not, it escalates. You pay for quality only when the local model can't handle it.

Pin a step to a tier when you know what you need:

```glitch
;; Classification is fast and low-stakes — keep it local
(step "classify"
  (llm :tier 0 :format "json"
    :prompt "Classify this issue... Respond with ONLY valid JSON."))

;; PR review needs rigor — go straight to premium
(step "review"
  (llm :tier 2
    :prompt "Review this PR with HIGH RIGOR..."))
```

Adding `:format "json"` enables structural validation — the output must parse as JSON or the step escalates. Use it to enforce structure without writing parsing logic.

## Control flow

### retry

Retry a step up to N times on failure. Useful for flaky API calls:

```glitch
(retry 3
  (step "fetch"
    (run "curl -sf https://api.example.com/data")))
```

### timeout

Kill a step if it hangs beyond a duration (Go duration strings: `"30s"`, `"2m"`, `"1h"`):

```glitch
(timeout "90s"
  (step "grade"
    (llm :prompt "Compare these variant outputs...")))
```

### retry + timeout compose

Forms nest. Retry a slow step with a timeout on each attempt:

```glitch
(retry 2
  (timeout "30s"
    (step "flaky-slow"
      (run "curl -sf https://slow-api.example.com"))))
```

### catch

Run a primary step; if it fails, run a fallback instead:

```glitch
(catch
  (step "fetch-graphql"
    (run "gh api graphql -f query='...'"))
  (step "fallback"
    (run "gh issue view ~param.issue --json body")))
```

This is used in production to gracefully degrade when GraphQL endpoints are unavailable:

````glitch
;; From a real plugin — fetch linked PRs via GraphQL, fall back to simple output
(catch
  (step "related"
    (run ```
      REPO="~param.repo"
      ISSUE="~param.issue"
      echo "=== LINKED PRS ==="
      gh api graphql -f query="..." 2>/dev/null \
        | jq -r '.data.repository.issue.timelineItems.nodes[]?.source
          | select(. != null)
          | "\(.state) #\(.number) \(.title)"' 2>/dev/null
      echo ""
      echo "=== RECENT REPO ACTIVITY ==="
      gh api "repos/$REPO/commits?per_page=10" \
        --jq '.[] | "\(.sha[0:7]) \(.commit.message | split("\n")[0])"' 2>/dev/null
      ```))
  (step "fallback"
    (run "echo 'no linked PRs found'")))
````

### cond

Multi-branch conditional. Predicates are shell commands — exit 0 means true:

```glitch
(cond
  ("test -f critical.log"
    (step "alert"
      (run "notify-send 'Critical issue found'")))
  ("test -f warning.log"
    (step "warn"
      (run "echo 'Warnings detected'")))
  (else
    (step "ok"
      (run "echo 'All clear'"))))
```

### map

Iterate over a prior step's output, one item per line. `~item` is the current item, `~item_index` is the zero-based index:

````glitch
(step "find-docs"
  (run "find . -name '*.md' -maxdepth 2"))

(map "find-docs"
  (step "process-doc"
    (run "wc -l ~item")))
````

In production, `map` powers document ingestion — iterating over discovered files and processing each one:

````glitch
(step "find-docs"
  (run ```
    find "$REPO_PATH" -type f \( -name "README*" -o -name "*.md" \) \
      -not -path '*/.git/*' -not -path '*/node_modules/*' \
      -size -100k 2>/dev/null | sort
    ```))

(map "find-docs"
  (step "process-doc"
    (run ```
      FILE="~item"
      CONTENT=$(cat "$FILE" 2>/dev/null | head -500)
      # ... hash, check for changes, index to ES
      echo "INDEXED: $REL_PATH"
      ```)))
````

### let

Scoped bindings — like `def` but limited to the body. Shadows outer defs within scope:

```glitch
(let ((endpoint "https://api.example.com")
      (token "abc123"))
  (step "call"
    (run "curl -H 'Auth: ~param.token' endpoint"))
  (step "parse"
    (run "echo '~(step call)' | jq '.data'")))
```

### phase and gate

Group steps into a phase with optional retry semantics. Gates are verification steps that must pass before the phase is considered complete:

```glitch
(phase "gather"
  (step "data" (run "echo 'hello world'")))

(phase "process" :retries 1
  (step "transform" (run "echo 'TRANSFORMED: hello world'"))
  (gate "not-empty" (run "test -n \"$(echo 'TRANSFORMED: hello world')\"")))
```

If a gate fails, the whole phase retries up to `:retries` times.

## SDK forms

Built-in forms that reduce boilerplate. Available in workflows and plugins.

### json-pick

Run a jq expression against a step's output:

```glitch
(step "shape"
  (json-pick ".[].title" :from "fetch"))
```

```glitch
(step "extract"
  (json-pick ".data.search.nodes" :from "graphql-result"))
```

### lines

Split a step's output by newline into a JSON string array:

```glitch
(step "as-list"
  (lines "find-files"))
```

### merge

Combine JSON output from multiple steps into one object:

```glitch
(step "activity"
  (merge "my-prs" "reviews" "mentions"))
```

### http-get / http-post

HTTP requests without shelling out:

```glitch
(step "fetch-data"
  (http-get "https://api.example.com/data"
    :headers {"Authorization" "Bearer ~param.token"}))

(step "submit"
  (http-post "https://api.example.com/submit"
    :body "~(step payload)"
    :headers {"Content-Type" "application/json"}))
```

Non-2xx responses fail the step (respects `retry` and `catch` wrappers).

### read-file / write-file

File I/O without shell commands:

```glitch
(step "config"
  (read-file "config/settings.json"))

(step "save-output"
  (write-file "output/report.json" :from "analysis"))
```

### glob

Match files against a pattern:

```glitch
(step "find-reviews"
  (glob "*/review.md"
    :dir "results/~param.repo/issue-~param.issue/iteration-1"))
```

Output is newline-separated file paths — composes with `map` for batch processing.

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
        ~(step data)
        ```))

  ;; This step runs instead
  (step "quick-analysis"
    (llm
      :model model
      :prompt ```
        Briefly summarize:
        ~(step data)
        ```)))
````

## Multiline strings

Triple backticks delimit multiline prompts. Content is auto-dedented, so indent for readability without affecting the output:

````glitch
(llm
  :model model
  :prompt ```
    You are a code reviewer. Review this diff carefully.

    Files changed:
    ~(step files)

    If everything looks good, say so. Be concise.
    ```)
````

## Complete form reference

### Top-level forms

| Form | Description |
|------|-------------|
| `(def name "value")` | Bind a constant for the file |
| `(workflow "name" :description "..." ...)` | Declare a workflow |

### Step-level forms (inside a step)

| Form | Description |
|------|-------------|
| `(run "command")` | Shell command (sh -c) |
| `(llm :prompt "..." ...)` | LLM call |
| `(save "path" :from "step-id")` | Write step output to file |
| `(name/sub :arg "val")` | Call a plugin subcommand (namespaced shorthand) |
| `(plugin "name" "sub" :arg "val")` | Call a plugin subcommand (verbose form) |
| `(json-pick "expr" :from "step-id")` | Run jq expression on step output |
| `(lines "step-id")` | Split output by newline into JSON array |
| `(merge "a" "b" ...)` | Combine JSON from multiple steps |
| `(http-get "url" :headers {...})` | HTTP GET request |
| `(http-post "url" :body "..." :headers {...})` | HTTP POST request |
| `(read-file "path")` | Read file into step output |
| `(write-file "path" :from "step-id")` | Write step output to file |
| `(glob "pattern" :dir "path")` | Match files, newline-separated output |

### Wrapper forms (around steps)

| Form | Description |
|------|-------------|
| `(retry N (step ...))` | Retry step up to N times on failure |
| `(timeout "30s" (step ...))` | Kill step after duration |
| `(catch (step ...) (step ...))` | Primary + fallback on failure |
| `(cond (pred (step ...)) ...)` | Multi-branch conditional |
| `(map "step-id" (step ...))` | Iterate over step output (one item per line) |
| `(let ((name val) ...) body...)` | Scoped variable bindings |
| `(phase "id" [:retries N] steps... [gates...])` | Grouped steps with verification gates |

## Next steps

- [Plugins](/docs/plugins) — package reusable subcommands and compose them into workflows
- [Compare Runs](/docs/compare) — built-in A/B testing with branch and review forms
- [DSL Reference](/docs/dsl-reference) — threading macro, collection forms, and Elasticsearch integration
- [Phases & Gates](/docs/phases-and-gates) — structured verification checkpoints