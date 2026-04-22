# Provider & Model Configuration

## Provider Registry

Providers are `.clj` files in `~/.config/glitch/providers/` that call `(glitch.provider/register name fn)` at load time.

Built-in providers (installed by `bb install`):

| Provider | How it runs | Notes |
|----------|------------|-------|
| `lmstudio` | HTTP to localhost:1234 (OpenAI-compatible) | Default local provider. Supports tool calling |
| `claude` | `claude -p` CLI | Strong reasoning, MCP tool support |
| `copilot` | `copilot` CLI | Premium requests |
| `openrouter` | HTTP to openrouter.ai/api/v1 | Free tiers available. Requires `OPENROUTER_API_KEY` |

## Default Provider Fallback

When no provider is specified in a workflow step, glitch tries `lmstudio` first, then falls through default tiers:

1. copilot
2. claude
3. openrouter
4. lmstudio

When a provider IS specified (`:provider "claude"`), it calls that provider directly with no fallback.

## Provider Interface

Every provider implements:

```clojure
(fn [{:keys [prompt model tool-defs]}]
  {:response "..." :tokens-in N :tokens-out M})
```

## Custom Providers

Create a `.clj` file in `~/.config/glitch/providers/`:

```clojure
(ns my-provider
  (:require [glitch.provider :as prov]))

(prov/register "my-provider"
  (fn [{:keys [prompt model]}]
    ;; call your API here
    {:response "..." :tokens-in 0 :tokens-out 0}))
```
