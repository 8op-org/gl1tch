# OpenAI-Compatible Provider + OpenRouter Support

**Date:** 2026-04-14
**Status:** Draft

## Goal

Add a native OpenAI-compatible HTTP client to glitch so users can run pipeline steps against OpenRouter (and any other OpenAI-compatible endpoint). Slot it into the configurable tier system as a cloud fallback.

## Non-Goals

- Streaming support (no provider streams today)
- Replacing Ollama for router/assistant classification
- OpenRouter-specific features (guardrails, routing rules, app-level transforms)
- Modifying existing Ollama or shell-template provider paths

## Architecture

### New: OpenAI-compatible client

`internal/provider/openai.go`

```go
type OpenAICompatibleProvider struct {
    Name         string
    BaseURL      string // e.g. "https://openrouter.ai/api/v1"
    APIKey       string
    DefaultModel string
}

func (p *OpenAICompatibleProvider) Chat(model, prompt string) (LLMResult, error)
```

- POST to `{BaseURL}/chat/completions` with `{"model": model, "messages": [{"role":"user","content":prompt}], "stream": false}`
- Auth: `Authorization: Bearer {APIKey}`
- Parse standard OpenAI response: extract `choices[0].message.content`, `usage.prompt_tokens`, `usage.completion_tokens`
- Cost: read `x-openrouter-cost` response header when present, otherwise leave cost as 0

### Configuration

`~/.config/glitch/config.yaml` gains a `providers` map:

```yaml
default_model: qwen3:8b
default_provider: ollama

providers:
  openrouter:
    type: openai-compatible
    base_url: https://openrouter.ai/api/v1
    api_key_env: OPENROUTER_API_KEY
    api_key: sk-or-v1-xxxx
    default_model: meta-llama/llama-4-scout:free

tiers:
  - [ollama/qwen3:8b]
  - [openrouter/meta-llama/llama-4-scout:free, codex, gemini]
  - [copilot, claude]
```

**API key resolution order:**
1. Environment variable named in `api_key_env`
2. Literal `api_key` value in config
3. Error if neither is set when the provider is invoked

### Pipeline integration

Workflow steps reference the provider by name:

```yaml
- llm:
    provider: openrouter
    model: meta-llama/llama-4-scout:free
    prompt: "{{.param.input}}"
```

`runner.go` step execution logic:

1. If provider is empty or `"ollama"` -> existing `RunOllamaWithResult()` path (unchanged)
2. If provider matches a key in `config.providers` with `type: openai-compatible` -> call `OpenAICompatibleProvider.Chat()`
3. Otherwise -> existing `ProviderRegistry.RunProviderWithResult()` shell template path (unchanged)

### Tier integration

`TieredRunner` resolves provider names against the config `providers` map. When a tier entry like `openrouter/meta-llama/llama-4-scout:free` is encountered:

1. Split on first `/` -> provider name `openrouter`, model `meta-llama/llama-4-scout:free`
2. Look up `openrouter` in `config.providers`
3. Construct `OpenAICompatibleProvider` from config
4. Call `Chat(model, prompt)`
5. On failure, continue to next provider in the tier

Users configure which tier OpenRouter lives in via the `tiers` list in `config.yaml`.

### Token and cost tracking

`LLMResult` already has `TokensIn`, `TokensOut`, `CostUSD` fields. The new client populates these from the actual API response:

- `usage.prompt_tokens` -> `TokensIn`
- `usage.completion_tokens` -> `TokensOut`
- `x-openrouter-cost` header -> `CostUSD` (float, when present)

No estimation fallback needed since the API returns real counts.

## Files Changed

| File | Change |
|------|--------|
| `internal/provider/openai.go` | New file: `OpenAICompatibleProvider` struct + `Chat()` method |
| `internal/provider/provider.go` | Add `ResolveProvider(name string) -> provider` that checks config map |
| `internal/provider/tiers.go` | Update `TieredRunner` to resolve provider names via config |
| `internal/pipeline/runner.go` | Add `openai-compatible` branch in step execution switch |
| `cmd/config.go` | Add `Providers` map and provider config structs to `Config` |

## Testing

- Unit test `OpenAICompatibleProvider.Chat()` against a mock HTTP server returning a standard OpenAI response
- Unit test API key resolution (env var > config > error)
- Unit test `TieredRunner` with an openai-compatible provider in a tier
- Integration: workflow with `provider: openrouter` step (requires real API key, manual)

## Future Work

- Streaming support across all providers
- Provider health checks / circuit breaking
- Additional OpenAI-compatible presets (Together, Groq) as config examples
