# Sexpr Plugin System Design

Replace compiled Go binary plugins with user-authored sexpr workflows.
A plugin is a directory of `.glitch` files — each file is a subcommand.
Users get complete control over how they interact with tools like `gh`,
`jq`, `curl`, and anything else on PATH. No compilation, no release
pipeline, no Homebrew tap.

## Problem

The current plugin design requires Go binaries compiled with GoReleaser,
released via GitHub Actions, and installed via Homebrew. This is heavy
infrastructure for what amounts to "call a CLI tool and shape the output."
Every plugin needs a repo, a release workflow, and a tap update. Users
can't easily create or modify plugins — they need Go knowledge and the
full release pipeline.

## Solution

Plugins become directories of `.glitch` files. Each file defines a
subcommand using the same sexpr syntax as workflows. An optional manifest
provides shared definitions. New SDK forms reduce boilerplate for common
operations (JSON shaping, HTTP, file I/O). Plugins are invocable from
the CLI (`glitch plugin <name> <subcommand>`) and from workflows via
a `(plugin ...)` form.

## Plugin Structure

A plugin is a directory containing `.glitch` files. The directory name
is the plugin name. Each `.glitch` file (except `plugin.glitch`) is a
subcommand.

```
~/.config/glitch/plugins/
  github/
    plugin.glitch       ;; optional manifest
    prs.glitch          ;; → glitch plugin github prs
    reviews.glitch      ;; → glitch plugin github reviews
    activity.glitch     ;; → glitch plugin github activity

.glitch/plugins/
  myteam/
    standup.glitch      ;; → glitch plugin myteam standup
    retro.glitch        ;; → glitch plugin myteam retro
```

## Discovery

Layered, same pattern as skills:

1. `.glitch/plugins/<name>/` (project-local, wins on conflict)
2. `~/.config/glitch/plugins/<name>/` (user-global)

No other discovery paths. No config.yaml plugin registries.

## Plugin Manifest (`plugin.glitch`)

Optional. Without it, the directory name is the plugin name and
subcommands have no shared state.

```clojure
;; ~/.config/glitch/plugins/github/plugin.glitch

(plugin "github"
  :description "GitHub activity queries"
  :version "0.1.0")

;; shared defs available to all subcommands
(def username "adam-stokes")
(def timezone "US/Eastern")
(def repos ["elastic/ensemble" "elastic/oblt-cli" "elastic/observability-robots"])
```

### Rules

- `(plugin ...)` is a new top-level form, only valid inside `plugin.glitch`
- `:description` and `:version` are optional metadata for `glitch plugin list`
- `(def ...)` bindings are injected into every subcommand's scope before
  execution, same semantics as `(def ...)` in workflows today
- If no manifest exists: plugin name = directory name, no shared defs

### What the manifest does NOT do

- No dependency declarations
- No lifecycle hooks
- No configuration schema — plugins read env vars or files in shell steps

## Subcommand Files

Each subcommand is a `.glitch` file with `(arg ...)` forms at the top
to declare CLI arguments, followed by a standard workflow body.

```clojure
;; ~/.config/glitch/plugins/github/prs.glitch

(arg "since" :default "yesterday" :description "Time range for query")
(arg "authored" :type :flag :description "PRs you authored")
(arg "reviewing" :type :flag :description "PRs requesting your review")

(workflow "prs"
  :description "Fetch pull requests"

  (step "fetch"
    (run ```
      gh api graphql -f query='...' --jq '.data.search.nodes'
    ```))

  (step "shape"
    (json-pick "." :from "fetch")))
```

### `(arg ...)` Form

Declares a named argument available as `{{.param.<name>}}` in templates.

| Keyword        | Description                                         |
|----------------|-----------------------------------------------------|
| `:default`     | Value if not provided. Omit for required args.      |
| `:type`        | `:string` (default), `:flag` (boolean), `:number`   |
| `:description` | Used in auto-generated `--help` output              |

Arguments map to CLI flags:
```
glitch plugin github prs --since week --authored
```

Arguments map to keywords in sexpr invocation:
```clojure
(plugin "github" "prs" :since "week" :authored)
```

### Output Contract

- The final step's stdout IS the subcommand's output
- CLI invocation: prints to stdout (composes with `| jq`, `| less`, etc.)
- Workflow invocation: becomes the step output via `{{step "id"}}`
- Stderr from inner shell steps passes through to terminal

### Shared Defs

Manifest `(def ...)` bindings are injected before the subcommand's own
defs. Subcommands can override them.

## SDK Forms

Built-in forms that reduce boilerplate. Available everywhere (workflows
too), but designed with plugins in mind.

### Data Forms

```clojure
;; Pick fields from JSON — wraps jq under the hood
(json-pick ".data.nodes" :from "step-id")
(json-pick ".[] | {number, title, url}" :from "step-id")

;; Parse text output into a list (one item per line)
(lines "step-id")

;; Merge JSON objects from multiple steps
(merge "step-a" "step-b" "step-c")
```

| Form                          | Implementation | Description                                      |
|-------------------------------|----------------|--------------------------------------------------|
| `(json-pick expr :from "id")` | shells to `jq` | Run jq expression against step output            |
| `(lines "id")`                | pure Go        | Split step output by newline → JSON string array |
| `(merge "a" "b" ...)`         | pure Go        | Shallow-merge JSON objects from multiple steps   |

### HTTP Forms

```clojure
(http-get "https://api.example.com/data"
  :headers {"Authorization" "Bearer {{.param.token}}"})

(http-post "https://api.example.com/submit"
  :body "{{step \"payload\"}}"
  :headers {"Content-Type" "application/json"})
```

| Form           | Implementation    | Description                                        |
|----------------|-------------------|----------------------------------------------------|
| `(http-get)`   | pure Go net/http  | GET request, response body as step output          |
| `(http-post)`  | pure Go net/http  | POST request, template-rendered URL/headers/body   |

Non-2xx responses fail the step with status code and body to stderr.

### File System Forms

```clojure
;; Read file contents into step output
(read-file "path/to/file.json")

;; Write step output to file
(write-file "output/report.json" :from "step-id")

;; List files matching a glob
(glob "*.yaml" :dir "configs/")
```

| Form             | Implementation       | Description                                     |
|------------------|----------------------|-------------------------------------------------|
| `(read-file)`    | pure Go os.ReadFile  | Read file, contents become step output          |
| `(write-file)`   | pure Go os.WriteFile | Write step output to file, creates parent dirs  |
| `(glob)`         | pure Go filepath     | Glob match, newline-separated paths as output   |

All paths are template-rendered. All pure Go — no external dependencies.

## Invoking Plugins from Workflows

Plugins behave like steps — output captured, available downstream.

```clojure
(step "fetch-prs"
  (plugin "github" "prs" :since "yesterday" :authored))

(step "fetch-issues"
  (plugin "github" "issues" :assigned))

(step "report"
  (llm :prompt ```
    Summarize this activity:
    PRs: {{step "fetch-prs"}}
    Issues: {{step "fetch-issues"}}
  ```))
```

### Resolution

1. Look up plugin directory via discovery order (project-local → user-global)
2. Find subcommand file (`prs.glitch`)
3. Load manifest `plugin.glitch` if it exists — inject shared `(def ...)` bindings
4. Map keyword args to the subcommand's declared `(arg ...)` params
5. Execute the subcommand as a sub-workflow
6. Final step output becomes this step's output

### Error Handling

| Condition              | Behavior                                                                  |
|------------------------|---------------------------------------------------------------------------|
| Plugin not found       | Step fails: "plugin 'x' not found, searched: .glitch/plugins/, ~/.config/glitch/plugins/" |
| Subcommand not found   | Step fails: "plugin 'x' has no subcommand 'y'"                           |
| Missing required arg   | Step fails: "plugin 'x y' requires argument 'z'"                         |
| Sub-workflow failure    | Propagates up, respects `(retry ...)` / `(catch ...)` wrapping           |

### Nesting

Plugins can call other plugins via `(plugin ...)`. No special handling —
it's just a step. Circular calls are not detected at parse time; they'd
hit a stack depth limit at runtime.

## CLI Interface

```
glitch plugin list                              # list all discovered plugins
glitch plugin <name> --help                     # plugin description + subcommands
glitch plugin <name> <subcommand> --help        # args from (arg ...) forms
glitch plugin <name> <subcommand> [--flags]     # run it
```

### `glitch plugin list`

```
PLUGIN      SOURCE    SUBCOMMANDS
github      global    prs, reviews, issues, mentions, activity
myteam      local     standup, retro
```

### `glitch plugin github --help`

```
github — GitHub activity queries (v0.1.0)

Subcommands:
  prs        Fetch pull requests
  reviews    Fetch reviews you gave
  issues     Fetch issues
  mentions   Fetch mentions
  activity   Combined activity feed
```

### `glitch plugin github prs --help`

```
github prs — Fetch pull requests

Flags:
  --since      Time range for query (default: yesterday)
  --authored   PRs you authored
  --reviewing  PRs requesting your review
```

### Output Behavior

- Stdout goes to terminal (or piped), no wrapper formatting
- Stderr from inner shell steps passes through
- Raw plugin output composes with `| jq`, `| less`, etc.

## Implementation Notes

### New Forms (Parser Layer)

No parser changes needed. The sexpr parser is generic — it builds an AST
from parens, atoms, and keywords. All new forms are handled in the
pipeline/evaluation layer as new converter functions.

New top-level forms:
- `(plugin "name" :description "..." :version "...")` — manifest metadata
- `(arg "name" :default "..." :type :string :description "...")` — argument declaration

New step-level forms:
- `(plugin "name" "subcommand" :args...)` — plugin invocation
- `(json-pick expr :from "step-id")` — jq wrapper
- `(lines "step-id")` — line splitting
- `(merge "a" "b" ...)` — JSON merge
- `(http-get url :headers {...})` — HTTP GET
- `(http-post url :body "..." :headers {...})` — HTTP POST
- `(read-file path)` — file read
- `(write-file path :from "step-id")` — file write
- `(glob pattern :dir "path")` — file glob

### New CLI Command

- `cmd/plugin.go` — Cobra command with dynamic subcommand resolution
- Parses `(arg ...)` forms to build flag sets
- Loads manifest for shared defs
- Delegates to existing pipeline runner

### Replaces Go Binary Plugins

This fully replaces the compiled Go binary plugin approach (`gl1tch-*`
repos, GoReleaser, Homebrew tap). There is no `glitch-*` binary
auto-discovery. Plugins are sexpr files only. Users who need to call
a compiled binary can still do so via `(run "binary ...")` in a plugin's
workflow steps — but the plugin itself is always a `.glitch` directory.

## Example: `glitch-github` as a Sexpr Plugin

Replaces the entire `gl1tch-github` repo (Go binary, GoReleaser,
GitHub Actions, Homebrew tap) with a directory of `.glitch` files:
```
~/.config/glitch/plugins/github/
├── plugin.glitch
├── prs.glitch
├── reviews.glitch
├── issues.glitch
├── mentions.glitch
└── activity.glitch
```

```clojure
;; plugin.glitch
(plugin "github"
  :description "GitHub activity queries"
  :version "1.0.0")

(def username "adam-stokes")
(def timezone "US/Eastern")

;; prs.glitch
(arg "since" :default "yesterday" :description "Time range")
(arg "authored" :type :flag :description "PRs you authored")
(arg "reviewing" :type :flag :description "PRs requesting your review")

(workflow "prs"
  :description "Fetch pull requests"

  (step "fetch"
    (run ```
      gh api graphql -f query='
        query {
          search(query: "author:{{.param.username}} created:>={{.param.since}}", type: ISSUE, first: 50) {
            nodes { ... on PullRequest { number title url repository { nameWithOwner } state additions deletions createdAt mergedAt } }
          }
        }' --jq '.data.search.nodes'
    ```))

  (step "output"
    (json-pick ".[] | {number, title, url, repo: .repository.nameWithOwner, state, additions, deletions, created_at: .createdAt, merged_at: .mergedAt}" :from "fetch")))
```

Zero compilation. Zero release pipeline. Edit and run.
