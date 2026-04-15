# LM Studio Provider for gl1tch

**Date:** 2026-04-15
**Status:** Approved

## Goal

Add LM Studio as a first-class local provider alongside ollama. Selectable via `:provider "lm-studio"` in workflows or in tier config. Not a replacement for ollama — a peer alternative.

## Approach

Standalone `LMStudioProvider` in `internal/provider/lmstudio.go`. Uses OpenAI chat completions format internally but adds LM Studio-specific model management via the native REST API.

## LMStudioProvider

**File:** `internal/provider/lmstudio.go`

```go
type LMStudioProvider struct {
    BaseURL      string // default "http://localhost:1234"
    DefaultModel string // default "qwen3-8b" (resolved to full GGUF identifier at runtime)
}
```

### Chat(model, prompt string) (LLMResult, error)

1. If `model` is empty, use `DefaultModel`
2. Query `/api/v0/models` (single call) to get model availability and loaded state
3. If model not found locally:
   - Auto-download via `POST /api/v1/models/download`
   - Poll `GET /api/v1/models/download/status/{job_id}` until complete
   - Log progress to stderr: `>> lm-studio: downloading %s (%.0f%%)...`
   - Error on failure: `"lm-studio: failed to download model %q: %s"`
4. If model found but not loaded, log: `>> lm-studio: loading %s, expect delay`
5. POST to `{BaseURL}/v1/chat/completions` with standard OpenAI request format
6. Parse `choices[0].message.content` and `usage.prompt_tokens` / `usage.completion_tokens`
7. Return `LLMResult` with `Provider: "lm-studio"`, `CostUSD: 0`

### Model management helpers

- `checkModels(model string) (exists bool, loaded bool, error)` -- single GET to `/api/v0/models`, checks for model in response
- `downloadModel(model string) error` -- POST to `/api/v1/models/download`, poll status until complete or failed

## Integration points

### Pipeline runner (internal/pipeline/runner.go)

Add `case "lm-studio"` in the provider switch (~line 752) alongside `case "ollama", ""`:
- Instantiate `LMStudioProvider` with defaults
- Call `Chat(model, prompt)`
- Extract response, token counts from `LLMResult`

### Tiered runner (internal/provider/tiers.go)

Add `case "lm-studio"` in `callProvider()` (~line 106) alongside the ollama case. Enables tier config like:

```yaml
tiers:
  - providers: [lm-studio]
    model: qwen3-8b
```

### Pricing table (internal/provider/tokens.go)

Add `"lm-studio": {0.00, 0.00}` to the pricing map.

### No changes to

- Config system (hardcoded defaults for now)
- OpenAI compatible provider
- Agent provider

## Testing

**File:** `internal/provider/lmstudio_test.go`

Using `httptest.NewServer` to mock LM Studio endpoints:

- `TestLMStudioChat` -- mock OpenAI-format response, verify LLMResult fields
- `TestLMStudioModelNotFound_TriggersDownload` -- empty model list, verify download endpoint called
- `TestLMStudioModelLoaded_NoWarning` -- model already loaded, no download attempt
- `TestLMStudioServerDown` -- connection refused, clear error message

## Files changed

| File | Change |
|------|--------|
| `internal/provider/lmstudio.go` | New: LMStudioProvider struct, Chat, model checks, download |
| `internal/provider/lmstudio_test.go` | New: unit tests with httptest mocks |
| `internal/provider/tiers.go` | Add `case "lm-studio"` in callProvider() |
| `internal/provider/tokens.go` | Add pricing entry |
| `internal/pipeline/runner.go` | Add `case "lm-studio"` in provider switch |

## Future work: Provider consolidation

The pipeline runner (`runner.go` ~line 752) and tiered runner (`tiers.go` ~line 106) both maintain separate switch statements for dispatching providers by name. Adding lm-studio makes this a three-way special case (`ollama`, `lm-studio`, agents). This should be consolidated into a single provider dispatch mechanism -- likely a `Provider` interface that all providers implement, with a unified registry replacing the current mix of switch cases, resolver functions, and shell-template lookups. Out of scope for this spec but should be addressed next.
