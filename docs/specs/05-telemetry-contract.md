# Telemetry Contract Specification

## Overview

This document defines the telemetry system — what gets indexed, when indexing happens, document schemas, and guarantees about telemetry's relationship to workflow execution.

## Definitions

- **Telemetry** — the optional `esearch.Telemetry` instance that indexes workflow execution data to Elasticsearch.
- **RunID** — a unique identifier generated per workflow invocation.
- **Index** — an Elasticsearch index that stores a specific document type.

## Core Guarantee

**Telemetry MUST NOT affect workflow execution.** Indexing failures are logged to stderr but never propagate as workflow errors. A workflow MUST produce identical results whether telemetry is enabled or disabled.

Telemetry is gated on `RunOpts.Telemetry` being non-nil. When nil, no indexing occurs and no telemetry-related code executes.

## RunID

Every workflow invocation generates a unique `RunID` via `esearch.NewRunID()`. This ID links all documents produced by a single run (LLM calls, workflow summary, cross-reviews).

`RunID` MUST be unique across all runs. The generation method is an implementation detail but MUST provide sufficient entropy to avoid collisions.

## Document Schemas

### LLM Call Document — `glitch-llm-calls`

Indexed once per LLM step execution. Indexed immediately after the step completes.

| Field              | Type     | Description                                          |
|--------------------|----------|------------------------------------------------------|
| `run_id`           | string   | Run identifier                                       |
| `step`             | string   | Format: `"workflow:{name}/{stepID}"`                 |
| `tier`             | int      | Tier index that produced the response                |
| `provider`         | string   | Provider name (lowercase)                            |
| `model`            | string   | Model identifier                                     |
| `tokens_in`        | int64    | Prompt token count                                   |
| `tokens_out`       | int64    | Completion token count                               |
| `tokens_total`     | int64    | `tokens_in + tokens_out`                             |
| `cost_usd`         | float64  | Estimated cost in USD                                |
| `latency_ms`       | int64    | Wall-clock time in milliseconds                      |
| `escalated`        | bool     | Whether escalation occurred                          |
| `escalation_reason`| string   | Reason for escalation (empty if tier 0)              |
| `escalation_chain` | []int    | Tier indices attempted                               |
| `eval_scores`      | []int    | Self-eval scores per attempt                         |
| `final_tier`       | int      | Same as `tier`                                       |
| `workflow_name`    | string   | Workflow name                                        |
| `issue`            | string   | Issue identifier (from name convention or params)    |
| `comparison_group` | string   | Comparison group (from name convention)              |
| `timestamp`        | string   | RFC3339 UTC                                          |

### Workflow Run Document — `glitch-workflow-runs`

Indexed once per workflow run, after all steps complete.

| Field              | Type     | Description                                          |
|--------------------|----------|------------------------------------------------------|
| `run_id`           | string   | Run identifier                                       |
| `workflow_name`    | string   | Workflow name                                        |
| `issue`            | string   | Issue identifier                                     |
| `comparison_group` | string   | Comparison group                                     |
| `total_steps`      | int      | Count of all steps in workflow                       |
| `llm_steps`        | int      | Count of steps that invoked an LLM                   |
| `total_tokens_in`  | int64    | Sum of all LLM step input tokens                     |
| `total_tokens_out` | int64    | Sum of all LLM step output tokens                    |
| `total_cost_usd`   | float64  | Sum of all LLM step costs                            |
| `total_latency_ms` | int64    | Sum of all LLM step latencies                        |
| `review_pass`      | bool     | Last LLM output contains "OVERALL: PASS"             |
| `confidence`       | float64  | `criteria_passed / criteria_total` (0 if no criteria) |
| `criteria_passed`  | int      | Count of PASS criteria in last LLM output            |
| `criteria_total`   | int      | Count of PASS + FAIL criteria in last LLM output     |
| `timestamp`        | string   | RFC3339 UTC                                          |

### Cross-Review Document — `glitch-cross-reviews`

Indexed per variant when workflow name contains "cross-review".

| Field              | Type     | Description                                          |
|--------------------|----------|------------------------------------------------------|
| `run_id`           | string   | Run identifier                                       |
| `issue`            | string   | Issue identifier                                     |
| `iteration`        | string   | Iteration label (from params)                        |
| `variant`          | string   | Variant name (lowercase)                             |
| `passed`           | int      | Count of passed criteria                             |
| `total`            | int      | Count of total criteria                              |
| `confidence`       | float64  | `passed / total`                                     |
| `winner`           | bool     | Whether this variant was declared winner             |
| `workflow_name`    | string   | Workflow name                                        |
| `timestamp`        | string   | RFC3339 UTC                                          |

### Cross-Review Parse Formats

Two formats are supported, auto-detected:

**Pass/Fail format** (detected by absence of `VARIANT:` headers):
```
--- LOCAL ---
1. Specificity — PASS — good detail
2. Accuracy — FAIL — missed edge case
SCORE: 3/5
WINNER: LOCAL
```

**Numeric format** (detected by `VARIANT:` headers):
```
VARIANT: local
plan_completeness: 9/10
accuracy: 7/10
total: 16/20
WINNER: local
```

In numeric format, a score >= 7 out of 10 counts as "passed".

## Issue and ComparisonGroup Derivation

These fields are derived in priority order:

1. Explicit `RunOpts.Issue` and `RunOpts.ComparisonGroup`
2. Workflow name convention: `{issue}-{description}-{group}` — issue is the leading numeric prefix, group is the last hyphen-separated segment
3. `params["issue"]` as fallback for issue

Examples:
- `"3918-wrapper-curl-copilot"` → issue=`"3918"`, group=`"copilot"`
- `"issue-to-pr-local"` → issue=`""`, group=`"local"`
- `"simple-review"` → issue=`""`, group=`"review"`

## Other Indices

The following indices exist in the system but are outside the scope of this spec (they are not populated by the pipeline runner):

| Index                    | Owner                    |
|--------------------------|--------------------------|
| `glitch-events`          | Arbitrary event logging  |
| `glitch-research-runs`   | Research loop            |
| `glitch-tool-calls`      | Tool invocation tracking |
| `glitch-knowledge-*`     | Per-repo knowledge index |

These indices MAY be specified in future addenda to this spec.

## Timestamp Requirements

- All timestamps MUST be UTC
- All timestamps MUST use RFC3339 format (`2006-01-02T15:04:05Z07:00`)
- Timestamps are recorded at the moment of indexing, not at step start

## Conformance

Telemetry conformance is verified by the pipeline runner conformance tests, which check that:
1. Workflow execution is identical with and without telemetry enabled
2. When enabled, the expected documents are indexed with correct schemas
3. `RunID` is consistent across all documents in a single run
