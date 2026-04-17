---
title: "Triaging a Kibana Regex Bug Across Three Tiers"
slug: "bug-triage-kibana"
description: "A data stream naming collision triggers the wrong Discover profile — three models attempt the diagnosis"
date: "2026-04-17"
---

## The Scenario

A time-series data stream with "logstash" in the name silently renders the wrong Discover profile. You query `TS metrics-logstash.otel-default` expecting a metrics grid. Kibana gives you the logs view instead. Rename the stream to `lstash` and it works fine.

That's [elastic/kibana#263137](https://github.com/elastic/kibana/issues/263137). A regex somewhere is matching "logstash" and deciding it's a logs data stream. It isn't. Every OTel or Logstash-named TSDS in production is affected.

We pointed three models at the issue thread and asked them to find the root cause.

---

## The Workflow

````glitch
(def triage-prompt ```
  You are a senior frontend engineer triaging a Kibana bug report.
  Analyze this issue: severity, root cause, affected components, fix approach.
  Be specific — reference exact files, regex patterns, or code if mentioned.

  Issue data:
  ~(step fetch)
  ```)

(workflow "triage-kibana"
  :description "Triage a Kibana bug across three model tiers"

  (step "fetch"
    (run "gh issue view 263137 -R elastic/kibana --json title,body,comments,labels"))

  (step "triage-local"
    (llm :provider "lm-studio" :model "qwen3-8b"
      :prompt ~triage-prompt))

  (step "triage-paid"
    (llm :provider "openrouter" :model "qwen/qwen3.5-flash-02-23"
      :prompt ~triage-prompt))

  (step "triage-copilot"
    (llm :provider "copilot" :model "sonnet"
      :prompt ~triage-prompt)))
````

```bash
glitch run triage-kibana
```

Same prompt, same issue data. Three providers. No model sees another model's output.

---

## Local: qwen3-8b (LM Studio) — $0.00

<div class="model-output">

**Severity: High.** Users can't see the metrics grid for any TSDS with "logs" in the name. Core observability feature, broken silently.

**Root cause:** A regex in `logs_context_service.ts` (lines 22–29) matches "logs" in the query string and forces the logs profile. The data stream name `logstash` triggers it.

**Fix direction:** Make the `TS` command take priority over the name heuristic. If the query starts with `TS`, skip the logs pattern check.

</div>

**The verdict.** Found the right file. Got the right severity. Proposed the right fix direction. But the actual regex it showed? Fabricated. The real pattern tests the *index name*, not the query string. The model conflated the ES|QL command with the index pattern — a meaningful conceptual error. It also missed the second file (`create_regexp_pattern_from.ts`) that explains *why* the prefix `metrics-` doesn't save you.

Directionally correct. Mechanically wrong. Good enough to route the bug to the right team. Not good enough to write the fix.

---

## Paid: qwen/qwen3.5-flash-02-23 (OpenRouter) — ~$0.001

<div class="model-output">

**Severity: High.** Feature breakage — the "Metrics in Discover" view is blocked for any stream name containing "log". No data loss, no security risk, but the UI renders the wrong profile silently. Workaround (renaming streams) is impractical for existing OTel pipelines.

**Root cause:** The Discover profile resolver (`logsContextService`) checks the index name against a regex *before* checking the query command. When a user runs `TS <stream-with-log>`, the name match fires first and overrides the explicit `TS` intent.

**File:** `src/platform/packages/shared/kbn-discover-utils/src/data_types/logs/logs_context_service.ts` (lines 22–29). Cited by commenter `lucaslopezf` in the thread.

**Fix:** Reorder the resolver chain so `metricsResolve` runs before `logsContextService`. Or add a guard: if `TS` is detected, lock to Metrics immediately.

</div>

**The verdict.** Correctly framed the bug as a conflict between *name heuristic* and *query intent* — that's the right mental model. Named the right file, cited the right commenter, and the architecture fix (reorder resolvers) matches what the Kibana team proposed in the thread.

Still pseudo-code though. Still missed the second file. The gap between "understands the problem" and "can write the patch" is showing.

---

## Copilot: Claude Sonnet — Copilot seat

<div class="model-output">

**Severity: High.** A correctly-configured TSDS silently renders with the wrong profile. No workaround except renaming data streams.

**Root cause — two files:**

**File 1:** `logs_context_service.ts` defines the pattern list:

```ts
export const DEFAULT_ALLOWED_LOGS_BASE_PATTERNS = [
  'log', 'logs', 'logstash',  // <-- culprit
  'auditbeat', 'filebeat', 'winlogbeat',
];
```

**File 2:** `create_regexp_pattern_from.ts` builds a permissive regex with optional prefix/suffix matching:

`metrics-logstash.otel-default` matches because:
- `metrics-` → matched by `optionalIndexNamePrefix`
- `logstash` → in `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS`
- `.otel-default` → matched by `optionalIndexNameSuffix`

There is no check for the `TS` command before the logs profile is applied.

**Fix:**

```ts
if (/^\s*TS\s+/i.test(esqlQuery)) {
  return metricsProfile;
}
```

The `TS` command *requires* a TSDS. Checking the index name for log keywords is irrelevant when the command is unambiguous.

</div>

**The verdict.** This is the only output that found the actual data. `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS` with `'logstash'` listed. The second file that builds the regex. The step-by-step decomposition of *why* the prefix `metrics-` doesn't prevent the match. The fix is three lines you could paste into a PR.

The difference between "found the right file" and "traced the full match path" is the difference between routing a bug and fixing it.

---

## Comparison

| | Local (qwen3-8b) | Paid (qwen3.5-flash) | Copilot (Sonnet) |
|---|---|---|---|
| **Root cause** | Right file, wrong mechanism | Right framing, right file | Full chain — both files, actual patterns |
| **Fix code** | Fabricated | Pseudo-code | Production-ready |
| **Thread awareness** | Low | Cited commenter, correct attribution | Cited both files, explained prefix matching |
| **Cost** | $0.00 | ~$0.001 | Copilot seat |

---

## Takeaway

Regex bugs need a model that can reason about *what string is being tested and why* — not just which file is involved. All three found `logs_context_service.ts` from the thread. Only Copilot traced the match through `create_regexp_pattern_from.ts` and produced code you could ship.

The local model earns its keep: right severity, right team, zero cost. The paid model adds the conceptual frame. Copilot writes the patch. That's the tier ladder working as designed.
