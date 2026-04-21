# Tool-Use-Aware Providers

**Date**: 2026-04-20
**Status**: Approved

## Problem

Workflow LLM steps currently stuff all context into the prompt via shell-gathered data. This causes:
- Massive prompts (high token usage)
- CLI argument size limits with copilot/claude providers
- LLM can't verify facts against real files
- Providers wrapping CLIs break in non-interactive environments

## Solution

Give LLM steps access to glitch's MCP tools. The LLM pulls its own context on demand instead of having everything pre-stuffed.

## Provider Integration Patterns

| Provider | Method | Tool loop |
|----------|--------|-----------|
| copilot | `copilot -p - --add-mcp "glitch mcp"` via stdin | Copilot manages its own loop |
| claude | Anthropic HTTP API with `tools` param | glitch manages the loop |
| openrouter | OpenAI-compatible HTTP API with `tools` param | glitch manages the loop |
| lmstudio | OpenAI-compatible HTTP API with `tools` param | glitch manages the loop, fall back to no-tools if model chokes |

### Why no API keys for copilot

Copilot CLI is included in GitHub subscription. The `--add-mcp` flag lets it connect to glitch's MCP server as a child process. No API fees, no token management — copilot handles tool calling natively.

### Why HTTP for claude/openrouter

Direct HTTP gives us control over the tool-calling loop. We send tool definitions, receive tool-call requests, execute locally, return results. No CLI wrapping, no argument size limits, no terminal detection issues.

### Why lmstudio attempts tools

Local models are free. If the model supports tool calling (qwen3-8b does), use it. If the response doesn't contain valid tool calls, retry without tools. Zero cost for the attempt.

## Workflow DSL

```clojure
;; Default: all tools available, single-round, max 5 rounds
(llm :prompt "Update the homepage...")

;; Agentic mode — LLM loops calling tools until satisfied
(llm :agentic true :max-rounds 10
     :prompt "Rewrite the homepage features section...")

;; Restricted tool set
(llm :tools ["glitch_read_file" "glitch_grep"]
     :prompt "Check what commands exist...")

;; No tools — legacy context-stuffed behavior
(llm :tools [] :prompt "Given: ~(ref 'ctx')...")
```

### Parameters

| Param | Default | Description |
|-------|---------|-------------|
| `:tools` | all 8 MCP tools | List of tool names, or `[]` to disable |
| `:agentic` | `false` | Enable multi-round tool loop |
| `:max-rounds` | 5 | Max tool-call iterations before forcing text output |

## Tool Execution

Tools execute via the existing `handlers/make-handler` function (same code the MCP stdio server uses). No MCP protocol roundtrip — direct function call.

Available tools (all by default):
- `glitch_read_file` — read file contents (first 200 lines)
- `glitch_grep` — regex search in files
- `glitch_search` — hybrid semantic + keyword search
- `glitch_symbols` — search symbol names
- `glitch_run` — execute a workflow
- `glitch_eval` — evaluate Clojure expression
- `glitch_check` — check workflow syntax
- `glitch_index` — index a repo for search

Workflow authors can restrict with `:tools [...]`.

## Loop Mechanics

### Single-round (default)

1. Send prompt + tool definitions to provider
2. If response contains tool calls: execute them, append results, send again (no tools this time)
3. LLM must produce text response
4. Return text as step output

### Agentic (`:agentic true`)

1. Send prompt + tool definitions to provider
2. If response contains tool calls: execute them, append results, re-send with tools
3. Repeat until LLM produces text response OR max-rounds hit
4. If max-rounds hit: final call with tools stripped, must produce text
5. Return text as step output

### Copilot special case

No loop management needed. Copilot handles the tool loop internally:

1. Write prompt to stdin
2. Spawn `copilot -p - --add-mcp "glitch mcp"`
3. Copilot calls MCP server (as subprocess) for tools
4. Read final text output from stdout
5. Strip trailing stats block (Changes/Requests/Tokens lines)

## Fallback Behavior

- **Model doesn't support tools**: strip tool definitions, send raw prompt
- **lmstudio**: attempt with tools first; if response has no valid tool calls, retry without tools
- **`:tools []`**: explicitly disables tools (old behavior preserved)
- **Provider fails entirely**: fall through to tiered escalation (existing behavior)

## Architecture Changes

### New file: `bb/src/glitch/tool_loop.clj`

Manages the tool-calling loop for HTTP providers:
- Takes: messages, tool definitions, provider call function, max-rounds, agentic flag
- Returns: final text response
- Handles: tool call parsing, tool execution, message assembly

### Modified: `bb/src/glitch/core.clj`

`llm` function gains `:tools`, `:agentic`, `:max-rounds` params. Passes them through to the provider function.

### Modified: `bb/providers/copilot.clj`

Spawn `copilot -p - --add-mcp "glitch mcp"` with prompt on stdin. Read stdout. Strip stats.

### Modified: `bb/providers/claude.clj`

Anthropic HTTP API with tool definitions in request body. Uses `tool_loop.clj` for the loop.

### Modified: `bb/providers/openrouter.clj`

OpenAI-compatible tool calling. Uses `tool_loop.clj` for the loop.

### Modified: `bb/providers/lmstudio.clj`

Same as openrouter (OpenAI-compatible). Try with tools, fall back without.

### Modified: `bb/src/glitch/provider.clj`

Provider call signature changes: `(fn [opts])` where opts now includes `:tools` (list of tool definitions) and `:tool-handler` (function to execute tools).

## Tool Definition Format

For HTTP providers, tool definitions are sent in OpenAI format:

```json
{
  "type": "function",
  "function": {
    "name": "glitch_read_file",
    "description": "Read a file and return its first 200 lines.",
    "parameters": {
      "type": "object",
      "properties": {
        "path": {"type": "string", "description": "Path to the file to read"}
      },
      "required": ["path"]
    }
  }
}
```

For Anthropic API, same structure but under the `tools` key with Anthropic's format:

```json
{
  "name": "glitch_read_file",
  "description": "Read a file and return its first 200 lines.",
  "input_schema": {
    "type": "object",
    "properties": {
      "path": {"type": "string", "description": "Path to the file to read"}
    },
    "required": ["path"]
  }
}
```

The existing `tools/tool-definitions` already has the schema — just needs format conversion per provider.

## Impact on Existing Workflows

- **No breaking changes**: existing `(llm :prompt "...")` still works — tools are available by default but the LLM only uses them if it wants to
- **Shell pre-gather steps become optional**: workflows can drop `(step "context" (sh ...))` steps since the LLM pulls its own context
- **Gates work the same**: gate LLM calls can use tools to verify facts against real files

## Example: Simplified site-homepage workflow

```clojure
(workflow "site-homepage"
  :description "Update the homepage"

  (step "rewrite"
    (llm
      :provider "copilot"
      :agentic true
      :max-rounds 8
      :prompt (str
        "Update the gl1tch homepage at site/src/pages/index.astro.\n"
        "Instructions: " (param "instructions") "\n"
        (when (param "section")
          (str "Only modify the " (param "section") " section.\n"))
        "Use glitch_read_file to read the current page.\n"
        "Use glitch_grep to verify any commands/features you reference actually exist.\n"
        "Return the complete updated file content.")))

  (step "gate"
    (llm
      :provider "copilot"
      :tools ["glitch_read_file" "glitch_grep" "glitch_search"]
      :prompt (str
        "Review this content for accuracy. Use tools to verify:\n"
        "- Every command referenced exists in the codebase\n"
        "- No hallucinated features\n"
        "- No banned terms (Ollama, glitch ask, BubbleTea, tmux)\n\n"
        "Content:\n" (ref "rewrite") "\n\n"
        "Return JSON: {\"pass\": bool, \"issues\": [...]}")))

  (step "save"
    (let [gate-out (json/decode (json-extract (ref "gate")))]
      (when (not (get gate-out "pass"))
        (throw (ex-info (str "Gate failed: " (json/encode (get gate-out "issues"))) {})))
      (save "site/src/pages/index.astro" (ref "rewrite"))
      "Updated homepage")))
```
