# Workflow Language Specification

## Overview

This document defines the gl1tch workflow language — the s-expression grammar used in `.glitch` files. It covers syntax, step forms, control-flow forms, template rendering, and scoping rules.

## Definitions

- **Workflow** — a named, ordered sequence of steps loaded from a `.glitch` or `.yaml` file.
- **Step** — a single unit of work identified by a string ID. Produces a string output.
- **Form** — a syntactic construct. Either a step form (produces output) or a control-flow form (wraps steps).
- **Def** — a top-level string binding, resolved at parse time.

## Grammar

A `.glitch` file contains zero or more `(def ...)` bindings followed by exactly one `(workflow ...)` form. Any other top-level form is a parse error.

```
file        = def* workflow
def         = "(" "def" name value ")"
workflow    = "(" "workflow" STRING metadata* body* ")"
metadata    = ":description" STRING
body        = step | phase | control-flow
```

### Atoms

| Type      | Syntax          | Example          |
|-----------|-----------------|------------------|
| String    | `"..."`         | `"hello world"`  |
| Symbol    | bare identifier | `model`          |
| Keyword   | `:name`         | `:prompt`        |
| Number    | digits          | `3`              |

### Structures

| Type | Syntax         | Example                    |
|------|----------------|----------------------------|
| List | `(a b c)`      | `(step "fetch" (run "ls"))` |
| Map  | `{k1 v1 k2 v2}` | `{"Content-Type" "application/json"}` |

### Comments

The discard reader `#_` causes the parser to skip the next form:

```glitch
#_ (step "disabled" (run "echo skip me"))
```

## Defs and Scoping

`(def name value)` creates a parse-time string binding. Defs are resolved top-to-bottom; a def MAY reference an earlier def.

```glitch
(def model "qwen2.5:7b")
(def prefix "review")
```

Symbols in value positions are resolved against the current def map. If a symbol has no matching def, it is kept as a literal string.

### Let Scoping

`(let ((name value) ...) body...)` creates inner bindings that shadow outer defs for the scope of the body. Let bindings are resolved left-to-right; a binding MAY reference earlier bindings in the same let.

```glitch
(let ((repo "/tmp/work")
      (branch "main"))
  (step "checkout"
    (run "git -C {{.param.repo}} checkout {{.param.branch}}")))
```

## Step Forms

Every step MUST have a string ID as its first argument. A step contains exactly one action form.

### `(run "command")`

Execute a shell command via `sh -c`. Output is stdout (trimmed). Non-zero exit is an error.

```glitch
(step "diff" (run "git diff --cached"))
```

### `(llm :prompt "..." [:model M] [:provider P] [:skill S] [:tier N] [:format F])`

Call an LLM provider. Output is the response text (trimmed).

| Keyword    | Required | Default            | Description                          |
|------------|----------|--------------------|--------------------------------------|
| `:prompt`  | YES      | —                  | Prompt text (template-rendered)      |
| `:model`   | no       | config default     | Model name                           |
| `:provider`| no       | config default     | Provider name                        |
| `:skill`   | no       | —                  | Skill name/path, prepended as system context |
| `:tier`    | no       | —                  | Tier index for escalation            |
| `:format`  | no       | —                  | Expected output format (e.g. "json") |

`:prompt` is the only required keyword. Parse error if missing.

### `(save "path" [:from "step-id"])`

Write content to a file at `path` (template-rendered). If `:from` is specified, saves that step's output. Otherwise saves the previous step's output. Output is the written content.

### `(json-pick "expr" [:from "step-id"])`

Run a jq expression against a step's output. `:from` specifies the source step. Output is the jq result.

### `(lines "step-id")`

Reference a step's output for use as a map source. Output is the same as the source step (identity transform). Exists to make intent explicit when feeding into `(map ...)`.

### `(merge "id1" "id2" ...)`

Concatenate the outputs of the named steps, separated by newlines. At least one step ID is required.

### `(http-get "url" [:headers (...)])`

HTTP GET request. Output is the response body. URL is template-rendered.

### `(http-post "url" [:body "..."] [:headers (...)])`

HTTP POST request. Output is the response body. URL and body are template-rendered. Headers are key-value pairs in a flat list: `("Content-Type" "application/json")`.

### `(read-file "path")`

Read a file's content. Output is the file content as a string. Path is template-rendered.

### `(write-file "path" [:from "step-id"])`

Write a step's output to a file. Alias for `(save ...)`.

### `(glob "pattern" [:dir "path"])`

Match files against a glob pattern. Output is newline-separated matching paths. `:dir` sets the base directory (default: current working directory).

### `(name/subcommand [:key val ...])`

Invoke a plugin subcommand using namespaced shorthand. The symbol before `/` is the plugin name, after `/` is the subcommand. See [Plugin Protocol](04-plugin-protocol.md). Output is the plugin workflow's final output.

The verbose form `(plugin "name" "subcommand" [:key val ...])` is also supported.

## Control-Flow Forms

Control-flow forms wrap steps. They appear at the workflow level (not nested inside a step body).

### `(retry N (step ...))`

Retry a step up to N times on failure. Total attempts = N + 1 (1 initial + N retries). The inner form MUST produce exactly one step.

```glitch
(retry 2
  (step "flaky" (run "curl -f https://api.example.com/health")))
```

### `(timeout "duration" (step ...))`

Set a deadline for step execution. Duration uses Go syntax (e.g., `"30s"`, `"2m"`). If the deadline expires, the step fails immediately and retries are skipped. The inner form MUST produce exactly one step.

### `(catch (step "primary" ...) (step "fallback" ...))`

Run the primary step. If it fails (after all retries), run the fallback step. Both the primary step ID and the fallback step ID receive the fallback's output in the steps map.

### `(cond (pred (step ...)) ... ("else" (step ...)))`

Evaluate predicates in order. Predicates are shell commands (template-rendered); exit 0 means true. The first matching predicate's step executes. `"else"` always matches. If no branch matches, the step output is empty string (not an error).

```glitch
(cond
  ("test -f go.mod" (step "go-lint" (run "golangci-lint run")))
  ("test -f package.json" (step "js-lint" (run "eslint .")))
  ("else" (step "skip" (run "echo 'no linter configured'"))))
```

### `(map "source-id" (step ...))`

Iterate over the newline-split output of a prior step. Empty lines are skipped. For each item:
- `{{.param.item}}` — the current line
- `{{.param.item_index}}` — zero-based index
- Body step is cloned with ID `{body.ID}-{index}`

All outputs are collected; the map step's output is newline-joined.

## Phase and Gate

`(phase "name" [:retries N] (step ...) ... (gate ...) ...)`

A phase groups work steps and verification gates into a retriable unit.

- Work steps execute sequentially.
- Gate steps execute after all work steps complete.
- Gates are structurally identical to steps but marked with `IsGate`.
- Shell gates: non-zero exit = failure. LLM gates: response parsed for "PASS"/"FAIL" keywords.
- Any gate failure triggers a phase retry (re-runs all work steps + gates).
- After `:retries` exhausted, the workflow errors.

```glitch
(phase "build-and-verify" :retries 2
  (step "build" (run "go build ./..."))
  (step "test" (run "go test ./..."))
  (gate "lint-check" (run "golangci-lint run")))
```

Gates MUST appear inside a `(phase ...)`. A bare `(gate ...)` at the workflow level is a parse error.

## Template Syntax

All string values in step forms are rendered as Go `text/template` templates before execution. Available data:

| Expression          | Description                                    |
|---------------------|------------------------------------------------|
| `{{.input}}`        | User input passed to the workflow               |
| `{{.param.key}}`    | Parameter value (from params, defs, or args)    |
| `{{step "id"}}`     | Output of a prior step (inline)                 |
| `{{stepfile "id"}}` | Write step output to temp file, return the path |

Template rendering happens immediately before step execution. A reference to an undefined step returns an empty string (no error).

**Important:** Parameter references MUST use `{{.param.key}}` with the dot prefix. `{{param.key}}` without the dot will silently remain as a literal string.

## Conformance

See [`spec/01-workflow-language.glitch`](../../spec/01-workflow-language.glitch) for the conformance workflow that exercises every form and template feature defined in this spec.
