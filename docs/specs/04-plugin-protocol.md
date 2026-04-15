# Plugin Protocol Specification

## Overview

This document defines the plugin system — how plugins are discovered, structured, configured, and invoked from workflows.

## Definitions

- **Plugin** — a named directory containing a manifest and one or more subcommand workflows.
- **Manifest** — a `plugin.glitch` file that declares the plugin's identity and shared defs.
- **Subcommand** — a `.glitch` workflow file within a plugin directory, invoked by name.
- **Arg** — a declared parameter for a subcommand, with type, default, and description.

## Discovery

Plugins are discovered from two directory paths. Local plugins override global plugins with the same name.

1. **Local** — `.glitch/plugins/<name>/` relative to the current working directory
2. **Global** — `~/.config/glitch/plugins/<name>/`

Discovery scans each directory for subdirectories containing `.glitch` files. A plugin directory MUST contain at least one `.glitch` file to be recognized.

### PluginInfo

```go
type PluginInfo struct {
    Name        string   // directory name
    Source      string   // "local" or "global"
    Dir         string   // absolute path to plugin directory
    Subcommands []string // names derived from .glitch filenames (excluding plugin.glitch)
}
```

## Manifest

The manifest file MUST be named `plugin.glitch` and reside in the plugin's root directory.

```glitch
(plugin "name" :description "what this plugin does" :version "1.0")
(def api_url "https://api.example.com")
(def default_model "qwen2.5:7b")
```

### Requirements

- The `name` in the manifest MUST match the directory name
- `:description` is RECOMMENDED but not required
- `:version` is RECOMMENDED but not required
- Defs declared in the manifest are available as parameters in all subcommands

### Caching

Manifests are cached in memory (via `sync.Map`) after first load. Repeated invocations of the same plugin skip file I/O and parsing.

## Subcommand Files

Every `.glitch` file in the plugin directory other than `plugin.glitch` is a subcommand. The filename (minus extension) is the subcommand name.

```
~/.config/glitch/plugins/review/
  plugin.glitch          # manifest
  staged.glitch          # "staged" subcommand
  pr.glitch              # "pr" subcommand
```

### Arg Definitions

Subcommand files MAY contain `(arg ...)` definitions before the `(workflow ...)` form:

```glitch
(arg "repo"
  :description "target repository path"
  :default ".")

(arg "verbose"
  :type :flag
  :description "enable verbose output")

(arg "max-files"
  :type "number"
  :description "maximum files to review")

(workflow "review-staged"
  (step "diff" (run "git -C {{.param.repo}} diff --cached"))
  ...)
```

### ArgDef

```go
type ArgDef struct {
    Name        string // arg name (used as param key)
    Default     string // default value (empty if none)
    Type        string // "string" (default), "flag", "number"
    Description string // user-facing help text
    Required    bool   // true if no default and not a flag
}
```

### Arg Types

| Type     | Behavior                                              |
|----------|-------------------------------------------------------|
| `string` | Default type. Value is passed as-is.                  |
| `flag`   | Boolean. Present = `"true"`, absent = `"false"`.      |
| `number` | String-typed internally but indicates numeric intent.  |

### Required Args

An arg is required if:
- It has no `:default` value, AND
- Its type is not `:flag` (flags default to `"false"`)

Invoking a subcommand without a required arg is a runtime error.

## Invocation Lifecycle

When a plugin subcommand is invoked (via CLI or workflow `(plugin ...)` form):

1. **Load manifest** — parse `plugin.glitch`, extract plugin metadata and defs. Result is cached.
2. **Parse subcommand** — parse `<subcommand>.glitch`, extract arg definitions and workflow. Result is cached.
3. **Extract args** — collect all `(arg ...)` forms from the subcommand file.
4. **Validate and merge params:**
   - Start with manifest defs
   - Apply arg defaults for any args not provided
   - Apply provided args (override defaults)
   - Validate: error if any required arg is missing
5. **Execute** — run the subcommand's workflow via the standard `Run()` function with the merged params.

### Parameter Precedence (lowest to highest)

1. Manifest defs
2. Arg defaults
3. Provided args (from CLI flags or workflow keyword args)

## Invocation from Workflows

Two syntaxes are supported:

### Namespaced Shorthand (preferred)

```glitch
(step "review"
  (review/staged :repo "/path/to/repo" :verbose))
```

The symbol before `/` is the plugin name, after `/` is the subcommand. This is parsed in the step body's default branch — any unrecognized `head` containing `/` is treated as a plugin call.

### Verbose Form

```glitch
(step "review"
  (plugin "review" "staged" :repo "/path/to/repo" :verbose))
```

```
(plugin "plugin-name" "subcommand" [:key value | :flag] ...)
```

- First string: plugin name
- Second string: subcommand name

### Keyword Args

- Keywords: argument key-value pairs
- A keyword followed by another keyword (or at end of list) is treated as a flag with value `"true"`

### Output

The plugin step's output is the final output of the subcommand workflow (`Result.Output`).

## Known Gap: PluginCallStep Execution

`PluginCallStep` is parsed from the `(plugin ...)` form and stored on the `Step` struct. The execution path in `runSingleStep` that handles `PluginCallStep` needs to be verified — this is flagged as a gap to be resolved during implementation.

## Naming Convention

Per project convention:
- Repository names use `gl1tch-<plugin>` (with the `1`)
- Binary names use `glitch-<plugin>` (without the `1`)
- Plugin directory names match the binary convention: `glitch-<plugin>` or just `<plugin>`

## Conformance

See [`spec/04-plugin-protocol/`](../../spec/04-plugin-protocol/) for the conformance test plugin and workflow that exercises discovery, manifest loading, arg parsing, and invocation.
