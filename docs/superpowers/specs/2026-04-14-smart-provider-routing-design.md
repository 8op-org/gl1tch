# Smart Provider Routing

## Overview

The smart router adds cost-conscious auto-routing to gl1tch's existing `TieredRunner`. Today the runner escalates on provider errors. This change adds escalation on **quality failures**, determined by LLM self-evaluation.

- Workflows opt in by default (no `tier:` = auto-route)
- Workflows opt out by setting `tier: N` explicitly
- Self-eval runs locally (free) at every tier except the last
- Escalation events logged to telemetry for visibility
- Stateless — always starts from tier 0

## Tier Model

Three natural tiers:

| Tier | Provider | Model | Cost |
|------|----------|-------|------|
| 0 | Ollama (local) | qwen3:8b | Free |
| 1 | OpenRouter | DeepSeek V3 | ~$0.27/$1.10 per M tokens |
| 2 | Claude / Copilot | Subscription models | Subscription |

Flow for any LLM step without an explicit tier:

```
Local model tries it
    → structural check (instant, free)
        → fail → escalate
        → pass → self-eval (one local call, free)
            → score >= threshold → accept, done
            → score < threshold → escalate to DeepSeek V3
                → structural check
                    → fail → escalate
                    → pass → self-eval (local call, free)
                        → score >= threshold → accept, done
                        → score < threshold → escalate to Claude/Copilot
                            → accept (final tier, no eval)
```

## Self-Evaluation Protocol

After an LLM step completes at any non-final tier, the runner sends a second call to the **local model** (always local — free):

```
Given this task and response, rate the quality 1-5:
- 5: Complete, accurate, well-structured
- 3: Partially correct but missing key details
- 1: Wrong, irrelevant, or incoherent

Task: {original prompt}
Response: {model output}

Reply with only a number.
```

- Score >= threshold → accept
- Score < threshold → escalate to next tier
- Default threshold: `4` (configurable via `eval_threshold` in config)
- At the final tier (premium), no eval — accept the result

## Structural Validation (Pre-Eval Fast Path)

Before running a self-eval call, the runner does a quick structural check based on the step definition:

- Step has `format: json` → `json.Unmarshal`. Fails → escalate immediately, skip eval.
- Step has `format: yaml` → `yaml.Unmarshal`. Fails → escalate immediately, skip eval.
- No format → check non-empty and not a refusal pattern (prefix checks for "I cannot", "I'm sorry, I can't", etc.).

Catches obvious garbage without a round-trip.

## Configuration

Uses the existing config shape with one new field:

```yaml
# ~/.config/glitch/config.yaml
default_model: qwen3:8b
default_provider: ollama
eval_threshold: 4

tiers:
  - providers: [ollama]
    model: qwen3:8b
  - providers: [openrouter]
    model: deepseek/deepseek-chat-v3-0324
  - providers: [copilot, claude]
```

## Workflow Step Syntax

Three modes, no new syntax beyond the optional `tier:` field:

```yaml
# Auto-route (default) — tries from tier 0, escalates on low confidence
- llm:
    prompt: "classify this issue..."

# Pin to a tier — skips straight there, no eval
- llm:
    tier: 2
    prompt: "write a detailed analysis..."

# Pin to a specific provider — same as today
- llm:
    provider: openrouter
    model: deepseek/deepseek-chat-v3-0324
    prompt: "summarize..."
```

## Telemetry & Escalation Logging

Extends the existing `LLMResult` with escalation metadata:

- `escalation_chain`: list of tiers attempted (e.g., `[0, 1]`)
- `eval_scores`: self-eval scores at each tier (e.g., `[2, 5]`)
- `escalation_reason`: `"structural"` or `"eval"`
- `final_tier`: which tier produced the accepted result

Additional fields on existing telemetry documents — no new indices.

## Dashboard Visualizations

Three Kibana saved objects seeded via `glitch seed`:

1. **Tier distribution over time** — stacked bar chart. X: time, Y: LLM call count, stacked by final tier. Shows cost trend.
2. **Cost savings** — line chart. Actual cost vs hypothetical all-premium cost. Gap = money saved.
3. **Escalation hotspots** — table. Rows: workflow step names. Columns: escalation rate, avg eval score at tier 0, most common escalation reason. Identifies steps needing prompt tuning.

## Implementation Scope

### Modified files

- `internal/provider/tiers.go` — add self-eval loop, structural validation, escalation metadata to `TieredRunner`
- `internal/pipeline/runner.go` — wire auto-routing as default path when no `tier:` or `provider:` is set
- `internal/pipeline/types.go` — add `Tier` field to `LLMStep`
- `cmd/config.go` — add `eval_threshold` to config struct
- `internal/provider/tokens.go` — extend `LLMResult` with escalation fields

### New files

- `internal/provider/eval.go` — self-eval prompt + score parser
- `internal/provider/structural.go` — JSON/YAML/refusal checks
- Kibana saved objects for dashboard visualizations (via seed)

### Not touched

- Workflow syntax (no breaking changes)
- Existing provider configs
- Shell-based providers (claude/copilot)
- Batch/variant system
- S-expression parser

### Backwards compatible

Every existing workflow works identically. Auto-routing only kicks in when no explicit provider/tier is set, and the default tiers config already exists.
