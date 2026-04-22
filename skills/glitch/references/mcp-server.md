# MCP Server

`glitch mcp` starts a JSON-RPC stdio server for IDE integration (Claude Code, Cursor, VS Code Copilot).

## Available Tools

| Tool | Description |
|------|-------------|
| `glitch_run` | Execute a workflow with optional input and parameters |
| `glitch_eval` | Evaluate Clojure with the full glitch DSL loaded |
| `glitch_check` | Validate a workflow file for syntax errors |
| `glitch_list_workflows` | List available workflows with descriptions |
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
