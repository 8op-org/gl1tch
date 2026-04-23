# MCP Tool Plugin System

**Date:** 2026-04-23
**Status:** Approved

## Problem

MCP tools are hardcoded in `tools.clj` and `handlers.clj`. Adding project-specific tools (e.g. `glitch_django_shell` for a Django project) requires modifying glitch source. The CLI plugin system already solves this for commands — MCP needs the same treatment.

## Design

### Plugin Format

Each `.clj` file in `.glitch/mcp-tools/` registers one tool via `defmcp-tool`:

```clojure
;; .glitch/mcp-tools/django-shell.clj
(defmcp-tool "glitch_django_shell"
  "Run Django ORM expressions and return results"
  {"expression" {:type "string" :description "Python expression" :required true}
   "format"     {:type "string" :description "Output format: raw or json"}}
  (fn [args]
    (let [expr (get args "expression")
          result (sh "docker" "compose" "run" "--rm" "web"
                      "python" "manage.py" "shell" "-c" expr)]
      (if (= "json" (get args "format"))
        (json-extract result "[]")
        result))))
```

### `defmcp-tool` Function

```
(defmcp-tool name description properties handler-fn)
```

- **name** — MCP tool name string (e.g. `"glitch_django_shell"`)
- **description** — tool description shown to the LLM
- **properties** — map of param name → `{:type :description :required}`, converted to JSON Schema `inputSchema`
- **handler-fn** — `(fn [args] ...)` — handler function, executed in SCI context with full DSL (`sh`, `llm`, `ref`, `save`, `search`, etc.)

### Schema Generation

The properties map is converted to JSON Schema `inputSchema`:

```clojure
;; Input:
{"expression" {:type "string" :description "Python expression" :required true}
 "format"     {:type "string" :description "Output format"}}

;; Output:
{"type" "object"
 "properties" {"expression" {"type" "string" "description" "Python expression"}
                "format"     {"type" "string" "description" "Output format"}}
 "required" ["expression"]}
```

### Registry

An atom in `glitch.mcp.plugin` namespace:

```clojure
(def registry (atom {}))
;; {tool-name → {:definition <MCP tool definition map>
;;               :handler-fn <fn [args] → string>}}
```

### Discovery & Loading

1. At MCP server start, `load-tools!` scans `.glitch/mcp-tools/*.clj`
2. Each file is evaluated in a SCI context with DSL primitives + `defmcp-tool` available
3. `defmcp-tool` registers into the atom
4. After loading, `tool-definitions` returns all plugin tool definitions
5. `handle-tool` looks up and calls the registered handler

### Integration Points

**`mcp.clj` (start)**:
```clojure
(defn start [_opts]
  (let [handler    (handlers/make-handler {})
        plugin-ctx (mcp-plugin/load-tools!)
        all-tools  (into tools/tool-definitions (mcp-plugin/tool-definitions))
        dispatch   {:tools all-tools
                    :tool-handler (fn [name args]
                                    (if-let [result (mcp-plugin/handle-tool name args)]
                                      result
                                      (handler name args)))}]
    ...))
```

**Handler dispatch order:**
1. Check plugin registry first (plugins can shadow built-ins if desired)
2. Fall through to built-in case statement
3. Throw "unknown tool" if neither matches

### Files

| File | Action |
|------|--------|
| `bb/src/glitch/mcp/plugin.clj` | **New** — registry, `defmcp-tool`, `load-tools!`, `tool-definitions`, `handle-tool` |
| `bb/src/glitch/mcp.clj` | **Modify** — call `load-tools!`, merge definitions, wire combined handler |
| `bb/test/glitch/mcp/plugin_test.clj` | **New** — tests for registration, schema generation, handler dispatch |
| `skills/glitch/references/mcp-server.md` | **Modify** — document plugin system |

### What It Doesn't Do

- No global `~/.config/` path (project-local only)
- No hot-reload (restart `glitch mcp` to pick up changes)
- No dependency resolution between plugins
- No `.glitch` workflow file format for tools (`.clj` only)
