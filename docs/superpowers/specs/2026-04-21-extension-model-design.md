# gl1tch Extension Model

**Date:** 2026-04-21
**Status:** Approved

## Problem

gl1tch has moved from a batch CLI tool to a Babashka-powered system with a persistent REPL (nREPL + CIDER). The current plugin protocol — directories of `.glitch` files that become subcommands — was designed for a CLI-only world. With a live Lisp image, the extension model needs to support both run-to-completion workflows and interactive REPL-driven composition without splitting into two incompatible systems.

## Core Concept

A **plugin** is a capability declaration — a named collection of functions that glitch loads once and surfaces automatically as:

- **CLI subcommands**: `glitch github fetch-issue --repo foo --issue 42`
- **Workflow steps**: `(github/fetch-issue :repo "foo" :issue 42)`
- **REPL-callable functions**: same call, live image

All three invoke the same function. Glitch handles the wiring. The plugin doesn't know or care which surface called it.

## Two Authoring Paths

### Path 1: `.glitch` files (simple wrappers)

Low barrier. Good for CLI tool wrappers where the DSL covers the need. Same convention as today:

```
.glitch/plugins/github/
  plugin.glitch          ;; name, description, shared defs
  fetch-issue.glitch     ;; becomes github/fetch-issue
  list-prs.glitch        ;; becomes github/list-prs
```

When loaded, glitch evaluates these into functions in the running image. Available everywhere — CLI, workflow, REPL.

### Path 2: Clojure namespaces (power plugins)

For real conditionals, state, error handling, or CIDER-driven development:

```clojure
(ns glitch.plugin.elasticsearch
  (:require [glitch.plugin :refer [defcommand]]))

(defcommand health
  "Check cluster health"
  {:args [{:name "cluster" :description "Cluster URL" :required true}]}
  [opts]
  (let [resp (http-get (str (:cluster opts) "/_cluster/health"))]
    (json-parse (:body resp))))

(defcommand query
  "Run an ES query"
  {:args [{:name "index" :required true}
          {:name "body" :required true}]}
  [opts]
  ...)
```

`defcommand` registers the function with metadata for glitch to auto-generate the subcommand, wire it as a workflow step, and make it REPL-callable. The namespace segment after `glitch.plugin.` becomes the plugin name.

### Discovery

Same directory scan as today, plus classpath:

- `.glitch/plugins/<name>/` (project-local)
- `~/.config/glitch/plugins/<name>/` (global)
- Classpath entries (bb.edn `:deps` or `:paths`)

A directory containing `.glitch` files triggers path 1. A `.clj` file (or directory with `bb.edn`) triggers path 2.

## Plugin Registry

Both authoring paths produce the same internal structure at load time:

```clojure
{:name        "github"
 :description "GitHub CLI wrappers"
 :commands    {"fetch-issue" {:fn          <callable>
                              :args        [{:name "repo" :required true} ...]
                              :description "Fetch a GitHub issue"}
               "list-prs"    {:fn          <callable>
                              :args        [...]
                              :description "List open PRs"}}}
```

This is the single source of truth. CLI subcommand generation, workflow step resolution, and REPL availability all read from this one registry.

### Data Flow

```
  .glitch file          Clojure namespace
       |                       |
       v                       v
  eval into fn          defcommand macro
       |                       |
       +-------+   +-----------+
               v   v
         Plugin Registry
        {:name :commands}
               |
       +-------+-------+
       v       v       v
     CLI    Workflow   REPL
  subcommand  step    function
```

## Terminology

| Term | Meaning |
|------|---------|
| **Workflow** | A `.glitch` file that runs to completion — repeatable, step-tracked, non-interactive. This term stays; it's correct for what it describes. |
| **Plugin** | A named collection of commands that glitch loads and surfaces. Safe wrappers around external tools. |
| **Command** | A single function within a plugin, surfaced as subcommand + workflow step + REPL fn. |
| **REPL session** | Live interactive use of the glitch image. Not a workflow. Just Clojure with glitch loaded. |
| **Provider** | An LLM backend. Separate concern from plugins. Unchanged. |

## What Changes

| Current | New |
|---------|-----|
| `(plugin "github" "fetch-issue" ...)` verbose form | `(github/fetch-issue ...)` direct call — registry makes this possible |
| Plugins are `.glitch` files only | Plugins can be `.glitch` files OR Clojure namespaces |
| No `defcommand` macro | `glitch.plugin/defcommand` for Clojure-authored plugins |
| Plugin functions only available in workflows | Plugin functions available in CLI, workflows, and REPL uniformly |

## What Does NOT Change

- Plugin directory conventions (`.glitch/plugins/`, `~/.config/glitch/plugins/`)
- `plugin.glitch` manifest for `.glitch`-authored plugins
- Provider system (LLM backends, separate concern)
- Workflow DSL syntax and execution model

## What This Does NOT Add

- No plugin versioning or dependency management (pre-1.0, YAGNI)
- No remote plugin registry or install-from-URL
- No special plugin testing framework (it's Clojure — use `bb test`)
- No migration required — both authoring paths coexist from day one
