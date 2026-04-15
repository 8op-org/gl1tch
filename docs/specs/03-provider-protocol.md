# Provider Protocol Specification

## Overview

This document defines the provider abstraction — how LLM providers are registered, resolved, invoked, and how tier-based escalation works.

## Definitions

- **Provider** — any callable that accepts `(model, prompt)` and returns `(LLMResult, error)`.
- **Tier** — a group of providers at the same escalation level, tried in order.
- **Escalation** — moving to the next tier after all providers in the current tier fail or are rejected.
- **Resolver** — a function that maps a provider name to a `ProviderFunc`.

## LLMResult Contract

Every provider MUST return an `LLMResult` on success:

```go
type LLMResult struct {
    Provider  string        // provider name (e.g., "ollama", "openrouter")
    Model     string        // model identifier used
    Response  string        // response text, trimmed of whitespace
    TokensIn  int           // prompt token count
    TokensOut int           // completion token count
    CostUSD   float64       // estimated cost in USD (0 for local providers)
    Latency   time.Duration // wall-clock time for the call
}
```

### Requirements

- `Response` MUST be trimmed (no leading/trailing whitespace)
- `Response` MUST be non-empty on success. An empty response from a provider is a valid signal for escalation
- `TokensIn` and `TokensOut` SHOULD be actual counts when the provider reports them (e.g., Ollama returns `prompt_eval_count` and `eval_count`). They MUST be estimated when actuals are unavailable
- `CostUSD` MUST be `0` for local providers (ollama, lm-studio)
- `Latency` MUST measure wall-clock time from request start to response received

## Provider Types

### Built-in Providers

Hardcoded in the tiered runner. These are always available without configuration.

**`ollama`** — local Ollama instance at `http://localhost:11434`. Uses the `/api/generate` endpoint. Default model: `qwen3:8b`. Token counts are actual (from response metadata). Cost is always 0.

**`lm-studio`** — local LM Studio instance at `http://localhost:1234`. Uses OpenAI-compatible chat endpoint. Default model: `qwen3-8b`. Cost is always 0.

### Config-based Providers

Defined in `~/.config/glitch/config.yaml` under the `providers` key:

```yaml
providers:
  openrouter:
    type: openai-compatible
    base_url: https://openrouter.ai/api/v1
    api_key_env: OPENROUTER_KEY
    default_model: x-ai/grok-4.20
```

Config-based providers are resolved via `ResolverFunc`. Currently only `openai-compatible` type is supported.

A config-based provider MUST specify:
- `type` — provider type (currently only `"openai-compatible"`)
- `base_url` — API base URL

A config-based provider SHOULD specify:
- `api_key_env` — environment variable name containing the API key
- `default_model` — fallback model if none specified in the workflow

### Command-template Providers

YAML files in `~/.config/glitch/providers/`:

```yaml
name: copilot
command: "copilot:discuss {{.prompt}}"
```

The command template supports `{{.prompt}}` and `{{.model}}` placeholders. Execution rules:
- If command contains `{{.prompt}}`: both prompt and model are rendered inline, executed as shell command
- If command contains `{{.model}}` but not `{{.prompt}}`: model is rendered inline, prompt is piped via stdin
- If command contains neither: prompt is piped via stdin

Command-template providers return the raw stdout as `Response`. Token counts are estimated.

## Resolution Order

When the tiered runner needs to call a provider by name, it resolves in this order:

1. **Built-in check** — is it `"ollama"` or `"lm-studio"`? Call the hardcoded implementation.
2. **Resolver check** — does `ResolverFunc(name)` return `(fn, true)`? Call `fn`.
3. **Registry fallback** — call `ProviderRegistry.RunProvider(name, model, prompt)`.
4. **Error** — provider not found.

This means built-in providers cannot be overridden by config or command templates.

## Tier Configuration

```go
type TierConfig struct {
    Providers []string // provider names, tried left to right
    Model     string   // model for this tier (overrides default)
}
```

Tiers are ordered from cheapest/fastest (index 0) to most expensive/capable (highest index).

Default tiers:
```
Tier 0: [ollama]           model: qwen3:8b       (local, free)
Tier 1: [codex, gemini]    model: (default)       (free cloud)
Tier 2: [copilot, claude]  model: (default)       (paid cloud)
```

## Escalation: `Run`

`TieredRunner.Run(ctx, prompt, validate)` — basic escalation with a caller-supplied validation function.

```
for each tier (0, 1, 2, ...):
    for each provider in tier.Providers:
        if ctx is cancelled: return context error
        result, err = callProvider(name, model, prompt)
        if err:
            lastReason = ReasonProviderError
            continue to next provider in same tier
        reason = validate(result.Response)
        if reason is non-empty:
            lastReason = reason
            break to next tier (skip remaining providers)
        return result (success)
return error("all tiers exhausted")
```

Key behaviors:
- Provider error → try next provider in SAME tier
- Validation rejection → skip to NEXT tier (remaining providers in current tier are skipped)
- Context cancellation is checked before each provider attempt

### Escalation Reasons

```go
const (
    ReasonMalformed     = "malformed_output"
    ReasonEmpty         = "empty_response"
    ReasonHallucinated  = "hallucinated_tool"
    ReasonProviderError = "provider_error"
    ReasonStructural    = "structural"
    ReasonEval          = "eval"
)
```

## Escalation: `RunSmart`

`TieredRunner.RunSmart(ctx, prompt, format, threshold, evalFn)` — escalation with structural validation and self-evaluation.

```
for each tier (0, 1, 2, ...):
    for each provider in tier.Providers:
        if ctx is cancelled: return context error
        result, err = callProvider(name, model, prompt)
        if err:
            lastReason = ReasonProviderError
            continue to next provider in same tier
        chain.append(tierIdx)
        if not CheckStructure(format, result.Response):
            scores.append(0)
            lastReason = ReasonStructural
            break to next tier
        if this is the final tier:
            return result (accepted without eval)
        evalScore = evalFn(model, BuildEvalPrompt(prompt, result))
        scores.append(evalScore)
        if evalScore >= threshold:
            return result (accepted)
        lastReason = ReasonEval
        break to next tier
return error("all tiers exhausted")
```

Key differences from `Run`:
- Structural check happens before eval
- Final tier is always accepted (prevents infinite escalation)
- Self-eval score compared against threshold (default: 4 on a 1-5 scale)
- `RunResult` includes `EscalationChain` and `EvalScores` for observability

## RunResult

```go
type RunResult struct {
    LLMResult                          // embedded
    Tier             int               // tier index that produced the accepted result
    Escalated        bool              // true if Tier > 0
    EscalationReason EscalationReason  // reason for last escalation (empty if tier 0 accepted)
    EscalationChain  []int             // tier indices attempted (RunSmart only)
    EvalScores       []int             // self-eval scores per attempt (RunSmart only)
}
```

## Conformance

See [`spec/03-provider-protocol.glitch`](../../spec/03-provider-protocol.glitch) for the conformance workflow that exercises provider resolution and tier escalation.
