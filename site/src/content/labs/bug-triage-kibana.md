---
title: "Triaging a Kibana Regex Bug Across Three Tiers"
slug: "bug-triage-kibana"
description: "A data stream naming collision triggers the wrong Discover profile — three models attempt the diagnosis"
date: "2026-04-17"
---

## The Scenario

[elastic/kibana#263137](https://github.com/elastic/kibana/issues/263137) — a correctly-configured time-series data stream (TSDS) fails to show the metrics grid in Kibana Discover when its name contains the word "logs". Querying `TS metrics-logstash.otel-default` silently renders the logs Discover profile instead of the metrics profile. The issue was confirmed by renaming the stream from `logstash` to `lstash`, which immediately fixed it. Reproducible across all users with OTel or Logstash-named TSDS — a common pattern in production observability pipelines.

## The Workflow

This is the workflow that produced the output below. Save it as `triage-kibana.glitch` — swap the issue number to triage any Kibana bug.

````glitch
(workflow "triage-kibana"
  :description "Triage a Kibana bug across three model tiers"

  (step "fetch"
    (run "gh issue view 263137 -R elastic/kibana --json title,body,comments,labels"))

  (step "triage-local"
    (llm
      :provider "lm-studio"
      :model "qwen3-8b"
      :prompt ```
        You are a senior frontend engineer triaging a Kibana bug report.
        Analyze this issue: severity, root cause, affected components, fix approach.
        Be specific — reference exact files, regex patterns, or code if mentioned.

        Issue data:
        ~(step fetch)
        ```))

  (step "triage-paid"
    (llm
      :provider "openrouter"
      :model "qwen/qwen3.5-flash-02-23"
      :prompt ```
        You are a senior frontend engineer triaging a Kibana bug report.
        Analyze this issue: severity, root cause, affected components, fix approach.
        Be specific — reference exact files, regex patterns, or code if mentioned.

        Issue data:
        ~(step fetch)
        ```))

  (step "triage-copilot"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt ```
        You are a senior frontend engineer triaging a Kibana bug report.
        Analyze this issue: severity, root cause, affected components, fix approach.
        Be specific — reference exact files, regex patterns, or code if mentioned.

        Issue data:
        ~(step fetch)
        ```)))
````

```bash
glitch run triage-kibana
```

All three steps run sequentially against the same fetched issue JSON. No model sees another model's output.

## Free Tier: qwen3-8b (LM Studio)

````text
### 1. **Severity Assessment**
**High**
- The issue prevents users from viewing time series metrics grids in Discover, which is a core
  usability feature for observability workflows.
- It affects data streams that contain "logs" in their name (e.g., `logstash`), which are valid
  time-series configurations. Users cannot see metrics without the grid, breaking critical
  analysis workflows.
- The bug is reproducible and impacts all users with such data stream names, even though they
  are correctly configured.

---

### 2. **Root Cause Hypothesis**
The root cause is a regex pattern in `logs_context_service.ts` that triggers the "logs" profile
based on the presence of the word "logs" in the query string (e.g., `TS logstash`). This
incorrectly assumes that any data stream name containing "logs" is a logs profile, overriding
the intended metrics context.

Key evidence:
- The regex in [this code block](https://github.com/elastic/kibana/blob/main/src/platform/packages/shared/kbn-discover-utils/src/data_types/logs/logs_context_service.ts#L22-L29)
  checks for `logs` in the query to activate logs profile logic.
- The data stream name `logstash` triggers this regex, forcing Discover to treat it as a logs
  profile rather than metrics.
- The fix is confirmed by renaming the stream to `lstash`, which avoids the "logs" substring
  and resolves the issue.

---

### 3. **Affected Components & Code Paths**
#### **Components**
- **Discover UI**: Responsible for rendering time series metrics grids.
- **Logs Context Service**: Handles logic to determine whether a query should use logs or
  metrics profiles.

#### **Code Paths**
- **Regex Pattern**: The regex in `logs_context_service.ts` (lines 22–29) matches queries
  containing "logs" and activates the logs profile:
  `````typescript
  if (/^logs/i.test(query)) {
    // Activate logs profile logic
  }
````
- **Metrics Profile Logic**: When a data stream is marked as "Time series" with metrics
  `time_series_metric`, it should default to metrics context, but this is overridden by
  the regex.

---

### 4. **Recommended Fix Approach**
#### **Fix Strategy**
Update the regex in `logs_context_service.ts` to prioritize `TS` (the command for time-series
queries) over the presence of "logs" in data stream names.

#### **Code Changes**
```typescript
// Original (incorrect)
if (/^logs/i.test(query)) {
  // Activate logs profile logic
}

// Proposed Fix
if (/^TS/i.test(query)) {
  // Prioritize metrics profile for TS queries
} else if (/^logs/i.test(query)) {
  // Fallback to logs profile if no TS command is present
}
  `````

#### **Validation Steps**
1. Test with data stream names containing "logs" (e.g., `logstash`) and confirm metrics grid
   renders when querying `TS logstash`.
2. Ensure existing behavior for queries without "logs" remains unchanged.
3. Validate edge cases where data streams have "logs" in their names but are explicitly marked
   as time series with metrics.

### Summary
This is a **high-severity bug** caused by incorrect logic in the logs profile regex that
overrides valid time-series queries. Fixing it requires updating the regex to prioritize `TS`
commands over "logs" substrings.
`````

**Verdict:** The free model found the right file (`logs_context_service.ts`, lines 22–29) and correctly assessed severity. It also grasped the right direction for the fix: `TS` command should take priority over name heuristics. However, it hallucinated the actual regex — the real pattern doesn't test `query` at all; it tests the index name. The shown fix code (`/^logs/i.test(query)`) is fabricated. The model conflated the ES|QL query string with the index pattern string, which is a meaningful conceptual error. It also missed the `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS` list and the permissive prefix/suffix matching in `create_regexp_pattern_from.ts` that explain *why* `metrics-logstash.otel-default` matches a logs pattern. Directionally correct, mechanically wrong.

## Paid Tier: qwen/qwen3.5-flash-02-23

````text
### 1. Severity Assessment
**Severity: High**

**Reasoning:**
- **Feature Breakage:** This issue directly blocks the "Metrics in Discover" feature (labeled
  as `Feature:Metrics in Discover`). Users cannot view the metrics grid for valid time-series
  data streams simply because the naming convention includes "log" (common in OpenTelemetry,
  e.g., `logstash`, `opentelemetry-logs`).
- **Usability Impact:** The current workaround requires users to rename data streams or indices,
  which is often not possible for existing OTel pipelines or managed services without significant
  operational overhead.
- **Misleading Intent:** The system ignores the explicit `TS` query command in favor of a
  heuristic string match on the index pattern name. This creates a confusing inconsistency where
  `TS my-logs` behaves differently than `TS my-metrics` despite the explicit command.
- **Not Critical:** No data is lost, security is not at risk, and the system remains stable; it
  simply renders the wrong view profile.

### 2. Root Cause Hypothesis
The Discover app employs a heuristic profile resolution service (`logsContextService`) that
determines the view mode (Logs vs. Metrics) based primarily on the **data stream/index name**
before checking the explicit **query command**.

- **The Trigger:** A regex pattern inside `logs_context_service.ts` matches substrings like
  "log" within the stream name.
- **The Conflict:** When a user queries `TS <stream-with-log>`:
  1. The query parser sees `TS` (Time Series).
  2. Simultaneously, the profile resolution engine runs.
  3. The engine sees the name matches the "log" pattern.
  4. The "log" heuristic **overrides** the `TS` command, forcing the UI into "Logs Profile"
     mode, hiding the metrics grid.
- **Evidence:** Comment by `lucaslopezf` explicitly links the regex pattern causing the issue
  to `src/platform/packages/shared/kbn-discover-utils/src/data_types/logs/logs_context_service.ts`
  (Lines 22-29).

### 3. Affected Components and Code Paths
- **Primary Component:** Kibana Discover Service / Data Types Context Service.
- **File Path:** `src/platform/packages/shared/kbn-discover-utils/src/data_types/logs/logs_context_service.ts`
- **Specific Location:** Lines 22-29 (referenced by commenter `lucaslopezf`).
- **User Flow Path:**
  1. User navigates to Discover.
  2. User inputs query: `TS <data_stream_name>` (where name contains 'log').
  3. System executes `metricsResolve` and `logsResolve` (implied logic).
  4. **Bug:** `logsContextService` evaluates `name.includes('log')` as truthy.
  5. System switches context to Logs view.
  6. UI fails to render the "grid of metrics".

### 4. Recommended Fix Approach
**Strategy: Explicit Intent > Heuristic Name**

1. **Implement Priority Fix (Quick & Safe):**
   ```typescript
   // Current (Likely):
   if (streamName.match(logRegex)) { return LogsContext; }

   // Fix:
   if (streamName.match(logRegex) && !query.startsWith('TS')) {
      return LogsContext;
   }
````
2. **Architecture Fix (Recommended):**
   Reorder the profile resolution priority so `metricsResolve` runs **before**
   `logsContextService` checks naming patterns. If `TS` is detected in the query, lock the
   profile to Metrics immediately without checking names.
3. **Test Expansion:**
   - Add an E2E test verifying `TS logs-some-pattern` displays the metrics grid.
   - Add a negative test verifying `logs-some-pattern` (without `TS`) still defaults to
     Logs view.
   - Verify `TS` prefix is case-insensitive.

**Risk Assessment:**
Low Risk — this narrows the scope of the regex trigger rather than removing it entirely.
`````

**Verdict:** A meaningful step up. The paid model correctly understood that the conflict is between *index name matching* and *query command intent* — the right framing for this bug. It accurately named `logsContextService`, cited the correct file and lines, and correctly attributed the `TS`-command-should-win argument to `miguel-sanchez-elastic` in the thread. The architecture fix recommendation (reorder `metricsResolve` before `logsResolve`) mirrors exactly what `lucaslopezf` proposed. The pseudo-code is still placeholder (`streamName.match(logRegex)`) rather than referencing the actual `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS` list, but the logic is correct. It did not discover the second involved file (`create_regexp_pattern_from.ts`) or explain the permissive prefix/suffix matching that makes `metrics-logstash.otel-default` match at all.

## Copilot (Sonnet)

````text
---

## Bug Triage: `[Metrics][Discover] #263137`

### 1. Severity: **High**

A correctly-configured TSDS is silently rendered with the wrong Discover profile — the metrics
grid is completely absent. Users cannot work around this without renaming their data streams.
Any OTel/Logstash-named TSDS is affected, which is a common naming pattern in production.

---

### 2. Root Cause

**File:** `src/platform/packages/shared/kbn-discover-utils/src/data_types/logs/logs_context_service.ts`
(lines 22–29)

`````ts
export const DEFAULT_ALLOWED_LOGS_BASE_PATTERNS = [
  'log',
  'logs',
  'logstash',   // <-- culprit
  'auditbeat',
  'filebeat',
  'winlogbeat',
];
````

**File:** `src/platform/packages/shared/kbn-data-view-utils/src/utils/create_regexp_pattern_from.ts`

The regex built by `createRegExpPatternFrom` uses very permissive prefix/suffix matching:

```ts
const optionalIndexNamePrefix = `(?:${nameCharacters}${segmentBoundary})?`;
const optionalIndexNameSuffix = `(?:${segmentBoundary}${nameCharacters})?`;
`````

This means `metrics-logstash.otel-default` matches the logs pattern because:
- `metrics-` → matched by `optionalIndexNamePrefix`
- `logstash` → matched by `indexNames` (it's in `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS`)
- `.otel-default` → matched by `optionalIndexNameSuffix`

The resulting regex (case-insensitive) fires on the full index name string, not just the base
segment. There is **no check for the `TS` command prefix** before the logs profile is applied,
so `isLogsIndexPattern("metrics-logstash.otel-default")` returns `true` and the logs
profile wins.

---

### 3. Affected Components & Code Paths

| Component | Path | Role |
|-----------|------|------|
| Logs context service | `kbn-discover-utils/src/data_types/logs/logs_context_service.ts` | Defines `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS` and exposes `isLogsIndexPattern` |
| Regex builder | `kbn-data-view-utils/src/utils/create_regexp_pattern_from.ts` | Builds the permissive match regex with optional prefix/suffix |
| Profile resolver | (wherever `isLogsIndexPattern` is called in Discover's ES|QL path) | Selects logs vs. metrics profile — **missing `TS`-command guard** |

---

### 4. Recommended Fix

**Preferred fix — check `TS` command precedence in the profile resolver:**

```ts
// If the ES|QL query starts with the TS command, it is unambiguously
// a time-series (metrics) query — skip the log pattern check entirely.
if (/^\s*TS\s+/i.test(esqlQuery)) {
  return metricsProfile;
}
```

This is semantically correct: ES|QL's `TS` command **requires** a TSDS and cannot target a
logs stream, so checking the index name for log keywords is irrelevant.

**Fallback quick fix — tighten the regex (fragile, not recommended alone):**

Modify `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS_REGEXP` so that a `metrics-` prefix disqualifies
the match, or remove `'logstash'` from patterns used in the non-`TS` path. This is brittle
and doesn't address the underlying design flaw: **profile resolution should be
query-command-aware**, not index-name-only.

**Additional suggestion from the thread (lucaslopezf):** reorder the resolver chain so
`metricsProfileResolver` runs before `logsProfileResolver` — though the `TS` guard above
makes ordering irrelevant for this case.
```

**Verdict:** Copilot landed the complete root cause. It's the only model that surfaced the *actual* data: `DEFAULT_ALLOWED_LOGS_BASE_PATTERNS` with `'logstash'` explicitly listed, and the second file `create_regexp_pattern_from.ts` that builds the permissive regex allowing `metrics-` as an optional prefix. The step-by-step decomposition of why `metrics-logstash.otel-default` matches the logs pattern — prefix/base/suffix — is mechanically accurate and matches what you'd find reading the source. The fix code is real and directly applicable: a regex guard at the call site rather than pseudo-code. It also noted that the `TS` guard makes resolver ordering irrelevant, which is a sharper observation than either other model produced.

## Comparison

| Model | Root Cause Found? | Correct File Referenced? | Fix Approach Valid? | Cost |
|-------|-------------------|--------------------------|---------------------|------|
| qwen3-8b (free, local) | Partial — right file, wrong mechanism (tested `query` not index name) | `logs_context_service.ts` ✓, missed `create_regexp_pattern_from.ts` | Directionally correct, code fabricated | $0.00 |
| qwen3.5-flash-02-23 (paid) | Yes — identified name-vs-command conflict correctly | `logs_context_service.ts` ✓, missed `create_regexp_pattern_from.ts` | Correct logic, pseudo-code only | ~$0.001 |
| Copilot Sonnet | Yes — full mechanism including prefix/suffix regex construction | Both files ✓ | Production-ready guard + explained why ordering is irrelevant | Copilot seat |

## Takeaway

Regex-level bugs demand a model that can reason about *what string is being tested and why* — not just *which file is implicated*. All three models found `logs_context_service.ts` from the thread, but only Copilot traced the full match path through `create_regexp_pattern_from.ts` and produced code you could apply without modification. For triage workflows where the issue thread contains real file references, free models are useful for rapid first-pass severity assessment; for root cause precision at the code level, the tier gap is measurable.