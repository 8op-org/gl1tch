# Plugin System

## Plugin Discovery

Plugins are loaded from:
1. `.glitch/plugins/` (project-local)
2. `~/.config/glitch/plugins/` (global)

## Two Plugin Types

### .glitch file plugins

Each file becomes a command:

```
~/.config/glitch/plugins/github/
  fetch-issue.glitch    -> glitch github fetch-issue
  list-prs.glitch       -> glitch github list-prs
```

### .clj namespace plugins

Self-register via `defcommand`:

```clojure
(ns my-plugin
  (:require [glitch.plugin :refer [defcommand]]))

(defcommand my-func
  "Description of what this does"
  {:args [{:name "repo" :required true}]}
  [opts]
  (str "Result for " (:repo opts)))
```

## Naming Convention

- **Repos:** `gl1tch-<plugin>` (e.g., `gl1tch-github`)
- **Binaries:** `glitch-<plugin>` (e.g., `glitch-github`)

## CLI Commands

```bash
glitch plugin list                        # list registered plugins
glitch plugin <name> <command> [args...]  # run plugin command
glitch <plugin-name> <command> [args...]  # shorthand
```
