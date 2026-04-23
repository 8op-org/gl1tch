# MCP Server

`glitch mcp` starts a JSON-RPC stdio server for IDE integration (Claude Code, Cursor, VS Code Copilot).

## Available Tools

| Tool | Description |
|------|-------------|
| `glitch_run` | Execute a workflow with optional input and parameters |
| `glitch_eval` | Evaluate Clojure with the full glitch DSL loaded |
| `glitch_check` | Validate a workflow file for syntax errors |
| `glitch_list_workflows` | List available workflows with descriptions |
| `glitch_recall` | Search workflows by description |
| `glitch_advise` | Get recommendations for a task |
| `glitch_search` | Ripgrep wrapper (regex, glob, multiline, context) |
| `glitch_read_file` | Read file contents with line numbers (200 lines max) |
| `glitch_search_symbols` | Search indexed symbols by name, kind, or language |
| `glitch_search_edges` | Query code relationships with depth traversal |
| `glitch_symbol_context` | Get a symbol's definition plus all relationships |

## IDE Configuration

**Claude Code** (`.claude/settings.json`):

```json
{
  "mcpServers": {
    "glitch": {
      "command": "glitch",
      "args": ["mcp"]
    }
  }
}
```

**Cursor** (`.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "glitch": {
      "command": "glitch",
      "args": ["mcp"]
    }
  }
}
```

**VS Code Copilot** (`.vscode/mcp.json`):

```json
{
  "servers": {
    "glitch": {
      "command": "glitch",
      "args": ["mcp"]
    }
  }
}
```

## Tool Plugins

Define project-local MCP tools in `.glitch/mcp-tools/`. Each `.clj` file registers one tool via `defmcp-tool`.

### Example

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

### `defmcp-tool` Arguments

| Arg | Description |
|-----|-------------|
| name | MCP tool name string (e.g. `"glitch_django_shell"`) |
| description | Tool description shown to the LLM |
| properties | Map of param name → `{:type :description :required}`, converted to JSON Schema |
| handler-fn | `(fn [args] ...)` — full DSL available (`sh`, `llm`, `ref`, `save`, etc.) |

### Discovery

- Scanned from `.glitch/mcp-tools/*.clj` at MCP server start
- Restart `glitch mcp` to pick up new or changed plugins
- Plugin tools appear alongside built-in tools in `tools/list`
