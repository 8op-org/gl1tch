---
title: "MCP Server"
order: 11
description: "Run glitch as a stdio MCP server and connect it to Claude Code, Copilot, or any MCP-compatible client."
---

## MCP Server

`glitch mcp` starts a [Model Context Protocol](https://modelcontextprotocol.io) server over stdio. Any MCP-compatible client — Claude Code, Copilot, Cursor, or your own tooling — can connect to it and call gl1tch tools directly from a chat or agent session.

```bash
glitch mcp
```

The server stays running and speaks the MCP protocol over stdin/stdout. You don't interact with it directly — your client does.

## Connect from Claude Code

```bash
claude mcp add glitch -- glitch mcp
```

That registers `glitch` as a named MCP server in Claude Code. After restarting Claude Code, gl1tch tools appear alongside built-in tools in every conversation.

## Connect from Copilot

Pass `--additional-mcp-config` with the server definition as JSON:

```bash
copilot chat \
  --additional-mcp-config '{"mcpServers":{"glitch":{"command":"glitch","args":["mcp"]}}}' \
  "find all usages of the Step type"
```

Or write the config to a file and reference it:

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

```bash
cop chat --additional-mcp-config "$(cat mcp.json)" "review my staged changes"
```

## Exposed tools

Once connected, your client can call any of these eight tools:

| Tool | What it does |
|------|--------------|
| `glitch_read_file` | Read the contents of a file by path |
| `glitch_grep` | Search file contents with a pattern, returns matching lines with file and line number |
| `glitch_search` | Hybrid keyword + semantic search across your indexed codebase — returns ranked chunks |
| `glitch_symbols` | Look up symbols (functions, types, variables) by name across indexed files |
| `glitch_run` | Run a named workflow and return its output |
| `glitch_eval` | Evaluate a workflow snippet inline — no file needed |
| `glitch_check` | Validate workflow syntax and report any parse errors |
| `glitch_index` | Index a directory so `glitch_search` and `glitch_symbols` can find it |

### Index first, then search

Before using `glitch_search` or `glitch_symbols`, index your project:

```bash
glitch index
```

Or from inside a client session, call `glitch_index` with your project path. Once indexed, `glitch_search` uses hybrid search — keyword matching plus vector similarity — so natural-language queries work alongside exact identifiers.

### Run workflows from your client

With `glitch_run`, your agent can kick off any workflow you've written:

```
glitch_run("code-review")
glitch_run("issue-to-pr", {"repo": "acme/backend", "issue": "42"})
```

The workflow runs to completion and the full output comes back as the tool result. Your client can reason over it, extract pieces, or chain it into the next message.

### Eval for one-off snippets

`glitch_eval` runs a workflow s-expression without saving a file first — useful when your agent wants to run a custom data-gathering step on the fly:

```
glitch_eval(`
  (step "commits"
    (run "git log --oneline -10"))
`)
```

## Any MCP-compatible client

The server speaks standard MCP over stdio — nothing proprietary. If your client supports MCP, point it at `glitch mcp` and it works. The connection command is always the same:

```bash
glitch mcp
```

## Next steps

- [Workflow Syntax](/docs/workflow-syntax) — write the workflows your MCP tools can run
- [Plugins](/docs/plugins) — package reusable data-gathering subcommands
- [Local Models](/docs/local-models) — configure local inference for the LLM steps your agent triggers