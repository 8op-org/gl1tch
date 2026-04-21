# Provider Protocol Specification

## Overview

This document defines the provider abstraction — how LLM providers are registered, resolved, invoked, and how tier-based escalation works. Providers are Clojure functions loaded at startup and stored in an atom-backed registry.

## Definitions

- **Provider** — a Clojure function that accepts a single opts map and returns `{:response :tokens-in :tokens-out}`.
- **Tier** — a group of providers at the same escalation level, tried in order.
- **Escalation** — moving to the next tier after all providers in the current tier fail or return blank.
- **Registry** — an atom holding `{name -> provider-fn}` mappings.
- **Tool loop** — the multi-round tool-calling protocol for OpenAI-compatible providers.

## Provider Contract

Every provider MUST be a function of one argument — an opts map:

```clojure
(fn [{:keys [prompt model tool-defs tool-handler agentic max-rounds]}]
  {:response   "..."       ;; string, trimmed, non-empty on success
   :tokens-in  0           ;; integer, 0 when actuals unavailable
   :tokens-out 0})         ;; integer, 0 when actuals unavailable
```

### Input keys

| Key | Type | Required | Description |
|-----|------|----------|-------------|
| `:prompt` | string | yes | The prompt text |
| `:model` | string or nil | no | Model identifier; `nil` means provider default |
| `:tool-defs` | vector of maps or nil | no | MCP tool definitions in `{"name" "description" "inputSchema"}` format |
| `:tool-handler` | fn or nil | no | `(fn [tool-name args-map] => string)` — executes a tool call |
| `:agentic` | boolean | no | If true, tool loop re-sends with tools after each round |
| `:max-rounds` | integer | no | Max tool-calling rounds (default 5) |

### Output requirements

- `:response` MUST be trimmed (no leading/trailing whitespace)
- `:response` MUST be non-empty on success. An empty or blank response triggers escalation.
- `:tokens-in` and `:tokens-out` SHOULD be actual counts when the provider reports them. They MUST be `0` when actuals are unavailable (CLI-based providers).

## Registered Providers

Four providers ship in `bb/providers/`:

### `lmstudio`

Local LM Studio instance. OpenAI-compatible HTTP API.

- **Base URL:** `LMSTUDIO_BASE_URL` env var, or `http://localhost:1234`
- **Default model:** `"default"` (LM Studio serves whatever model is loaded)
- **Tool support:** Yes — converts MCP tool defs to OpenAI function-calling format, uses `tool_loop/run-loop`
- **Fallback behavior:** If tool loop fails or returns blank, retries without tools
- **Token counts:** Actual from `/v1/chat/completions` usage when no tools; `0` after tool loop

### `openrouter`

Cloud provider via OpenRouter API.

- **Base URL:** `OPENROUTER_BASE_URL` env var, or `https://openrouter.ai/api/v1`
- **API key:** `OPENROUTER_API_KEY` env var (required, throws if missing)
- **Default model:** `OPENROUTER_MODEL` env var, or `google/gemma-3-12b-it`
- **Tool support:** Yes — same MCP-to-OpenAI conversion, uses `tool_loop/run-loop`
- **Headers:** Sends `HTTP-Referer: https://gl1tch.dev` and `X-Title: glitch`
- **Token counts:** Actual from usage response (tracked via atom across tool loop rounds)

### `claude`

Anthropic Claude via the `claude` CLI (`claude --print`).

- **Tool support:** Yes — passes `--mcp-config` with the glitch MCP server when tool-defs present
- **Model override:** `--model <model>` when model is specified
- **Token counts:** Always `0` (CLI doesn't report usage)
- **Output:** Raw stdout, trimmed

### `copilot`

GitHub Copilot CLI (`copilot -p -`).

- **Tool support:** Yes — passes `--additional-mcp-config` with the glitch MCP server when tool-defs present. When no tools, passes `--available-tools=` to disable all tool use.
- **Flags:** Always includes `--disable-builtin-mcps` to avoid token overhead
- **Post-processing:** Strips agent trace lines (`●`, `✗`, `│`, `└`), trailing stats blocks (`Changes/Requests/Tokens`), and markdown fences
- **Token counts:** Always `0` (CLI doesn't report usage)

## Tool Loop Protocol

The `glitch.tool_loop/run-loop` function implements multi-round tool calling for HTTP providers (lmstudio, openrouter).

```
run-loop(call-fn, messages, tools, tool-handler, opts)
  → final text response string
```

### Flow

```
if no tools:
    single call → return text content

loop (max-rounds):
    call LLM with messages + tools
    choice = response.choices[0]

    if no tool_calls in choice:
        return choice.message.content   (done)

    if round >= max-rounds:
        append assistant msg + tool results
        call without tools → return text  (forced completion)

    execute each tool_call via tool-handler
    append assistant msg + tool results to messages

    if agentic:
        recur with tools                 (another round)
    else:
        call without tools → return text (single-round mode)
```

### Agentic vs single-round

- **Single-round** (default): One tool pass, then force a text response. Good for structured retrieval.
- **Agentic** (`agentic: true`): Re-sends with tools each round. The LLM decides when to stop calling tools. Capped at `max-rounds`.

### MCP tool format conversion

All providers convert MCP tool definitions to OpenAI function-calling format:

```clojure
;; MCP format (input)
{"name"        "glitch_search"
 "description" "Search indexed code"
 "inputSchema" {"type" "object" "properties" {...}}}

;; OpenAI format (output)
{"type"     "function"
 "function" {"name"        "glitch_search"
             "description" "Search indexed code"
             "parameters"  {"type" "object" "properties" {...}}}}
```

## Provider Registration

Providers self-register by calling `glitch.provider/register` at load time:

```clojure
(require '[glitch.provider :as provider])

(provider/register "my-provider"
  (fn [{:keys [prompt model tool-defs tool-handler agentic max-rounds]}]
    ;; ... implementation ...
    {:response   "..."
     :tokens-in  0
     :tokens-out 0}))
```

### Loading

`glitch.provider/load-providers` scans directories for `.clj` files and loads each one. Search order:

1. `~/.config/glitch/providers/` — user-global
2. `providers/` — relative to working directory
3. Classpath entries ending in `providers`
4. Any explicitly passed directories

Each `.clj` file is loaded via `load-file`. Failed loads emit a warning to stderr and continue.

## Resolution Order

When the runner needs to call a provider:

```
if :provider is specified in opts:
    call-provider(name, opts)           → fail loudly if not registered

else:
    try call-provider("lmstudio", opts) → default provider
    on failure:
        call-tiered(opts, default-tiers) → escalate through tiers
```

The default provider is `lmstudio`. When it fails, the runner falls back to tier escalation.

## Tier Configuration

```clojure
;; Each tier is a map with :providers (tried left→right) and optional :model
{:providers ["copilot"]  :model nil}
{:providers ["claude"]   :model nil}
{:providers ["openrouter"] :model nil}
{:providers ["lmstudio"] :model nil}
```

### Default tiers

```clojure
(def default-tiers
  [{:providers ["copilot"]}
   {:providers ["claude"]}
   {:providers ["openrouter"]}
   {:providers ["lmstudio"]}])
```

### Escalation: `call-tiered`

```
for each tier:
    for each provider in tier.providers:
        try call-provider(name, merged-opts)
        if response is nil or blank:
            continue to next provider
        if exception:
            log warning to stderr
            continue to next provider
        return result  (success)
    (all providers in tier failed → continue to next tier)

throw "all tiers exhausted"
```

Key behaviors:
- Provider error → log to stderr, try next provider in SAME tier
- Blank/nil response → try next provider in SAME tier
- Model override from tier config merges into opts
- No validation step — escalation is purely based on provider availability and non-empty response

## Runner Integration

The `glitch.runner/run` function wires providers to workflows:

1. Calls `provider/load-providers` to populate the registry
2. Builds a tool handler from the workspace's MCP tools (search, file read/write, symbols)
3. Sets the provider dispatch function that:
   - Injects `tool-defs` and `tool-handler` into every LLM call
   - Routes explicit `:provider` directly (no fallback)
   - Routes unspecified providers through `lmstudio` → tier escalation
4. Custom tiers can be passed via `:tiers` to override `default-tiers`

### Tool injection

By default, all available MCP tools are injected into every LLM call. Workflows can control this:

- Omit `:tools` → all tools injected
- `:tools ["glitch_search"]` → only named tools
- `:tools []` → no tools (plain LLM call)

When the workspace has no search index, search-related tools (`glitch_search`, `glitch_index`, `glitch_symbols`) are automatically excluded.

## Writing a Custom Provider

Drop a `.clj` file in `~/.config/glitch/providers/` or your project's `providers/` directory:

```clojure
;; providers/my-api.clj
(require '[babashka.http-client :as http])
(require '[cheshire.core :as json])
(require '[glitch.provider :as provider])

(provider/register "my-api"
  (fn [{:keys [prompt model]}]
    (let [resp (http/post "https://my-api.example.com/v1/chat"
                 {:headers {"Content-Type" "application/json"
                            "Authorization" (str "Bearer " (System/getenv "MY_API_KEY"))}
                  :body (json/generate-string
                          {:model (or model "default")
                           :messages [{:role "user" :content prompt}]})})
          parsed (json/parse-string (:body resp) true)
          usage  (get parsed :usage {})]
      {:response   (get-in parsed [:choices 0 :message :content] "")
       :tokens-in  (or (:prompt_tokens usage) 0)
       :tokens-out (or (:completion_tokens usage) 0)})))
```

For tool support, use `glitch.tool_loop/run-loop` — see `lmstudio.clj` or `openrouter.clj` as templates.
