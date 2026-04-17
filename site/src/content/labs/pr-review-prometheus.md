---
title: "Reviewing a Prometheus PR with Copilot"
slug: "pr-review-prometheus"
description: "Copilot Sonnet reviews prometheus/prometheus#18499 — a consul health_filter feature with 263 lines changed"
date: "2026-04-17"
---

## The Scenario

[prometheus/prometheus#18499](https://github.com/prometheus/prometheus/pull/18499) adds a `health_filter` field to Consul service discovery. The backstory: PR #17349 introduced a `filter` field for the Catalog API, but it leaked into the Health API too. `ServiceTags` works in Catalog filters. The Health API expects `Service.Tags`. Users got silent breakage. This PR splits the plumbing.

The fix itself is straightforward. What makes it interesting for review is what surrounds it: a guard condition at `consul.go:335` that decides whether discovery takes the catalog path at all, three new tests that need to verify HTTP routing rather than functional output, and a documentation change that quietly suggests an incorrect migration path.

Three concerns. One PR. Good test for an automated reviewer.

---

## The Workflow

Save as `pr-review.glitch`, swap the PR number for anything in the Prometheus repo.

````glitch
(workflow "pr-review"
  :description "Fetch a PR and run a structured code review via Copilot"

  (step "metadata"
    (run "gh pr view 18499 -R prometheus/prometheus --json title,body,files,comments"))

  (step "diff"
    (run "gh pr diff 18499 -R prometheus/prometheus"))

  (step "review"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt ```
        You are a senior Go engineer reviewing a Prometheus pull request.
        Review for correctness, edge cases, test coverage, and API design.
        Provide specific, actionable feedback referencing file names and line ranges.

        PR metadata:
        ~(step metadata)

        Diff:
        ~(step diff)
        ```)))
````

```bash
glitch run pr-review
```

---

## The Review — Copilot (Sonnet)

<div class="model-output">

### Correctness

**Core fix is correct.** `consul.go:504` routes `watchedHealthFilter` to the Health query. The catalog path guard at `consul.go:335` properly includes `|| d.watchedFilter != ""` so a config with `services: ["test"] + filter: X` doesn't silently skip catalog discovery.

Plumbing verified end-to-end: `watchServices()` at ~line 385 threads `watchedFilter` into `opts.Filter` for catalog queries.

**Concern:** When `services + filter` is set together, `filter` applies at the catalog listing stage. If a named service doesn't match the filter, it's silently dropped. No error, no warning. Should be documented.

### Edge Cases

**`health_filter` with explicit services, no `filter`** — the condition at line 335 evaluates false, so `Run()` takes the direct `watchService` branch. `watchedHealthFilter` is correctly applied at line 504. Works, but has zero test coverage. This is the most common real-world use case.

**No cross-API field validation.** Nothing prevents `health_filter: 'tag1 in ServiceTags'` (a Catalog field name). Fails silently at the Consul API level. Out of scope, but worth a doc note.

### Test Coverage

Three new tests cover the right scenarios. `TestFilterOption` verifies catalog gets the filter, health does not. `TestHealthFilterOption` verifies the inverse. `TestBothFiltersOption` checks isolation when both are set.

**Missing:** No test for `services + health_filter` without `filter`. The non-catalog path has zero coverage.

### Documentation Bug

`docs/configuration/configuration.md:1483` — the `node_meta` deprecation note suggests both `filter` and `health_filter` as replacements. Wrong. `node_meta` is a Catalog-level construct. Its replacement is `filter` only. Suggesting `health_filter` here misleads users.

### Blockers

1. Add test for `services + health_filter` (no `filter`) — the non-catalog path.
2. Fix `node_meta` deprecation note to remove `health_filter`.

</div>

---

## Analysis

### What the review caught

Two real blockers, correctly identified.

**Missing test for the non-catalog path.** `services: ["test"] + health_filter: X` skips catalog entirely — `consul.go:335` evaluates false. The `watchedHealthFilter` wiring works, but none of the three new tests exercise this branch. Copilot traced the conditional to find the gap.

**Documentation bug in `node_meta` deprecation.** The PR added `health_filter` to the deprecation suggestion for `node_meta`, but `node_meta` maps to the Catalog API. The Health API uses `Node.Meta.*` syntax. Pointing users at `health_filter` as a `node_meta` replacement is wrong. Precise catch.

The review also flagged the semantic change when `services + filter` are combined — catalog-level filtering can silently drop explicitly named services. Correctly classified as a documentation gap, not a code bug.

### What it missed

**`UnmarshalYAML` validation.** The PR modifies `SDConfig` but the review never checks whether `HealthFilter` needs validation — mutual exclusion with deprecated fields, warnings when combined with `filter`. A human reviewer familiar with the Prometheus config validation pattern would scan that function.

**The `[BUGFIX]` vs `[CHANGE]` release note discussion** gets a diplomatic "defensible" instead of a clear answer. The feature shipped broken in an unreleased version. It's a bugfix. End of discussion.

**The `filter + explicit services` semantic change** is noted but not fully developed. The review doesn't ask whether `shouldWatch()` behavior is correct when `filter` eliminates a service that was explicitly configured. Deeper question, left on the table.

### Comparison to the human review

| Reviewer | Missing test (non-catalog path) | Doc bug (`node_meta`) | Semantic change (`filter` + services) | Test remediation detail |
|---|---|---|---|---|
| Copilot Sonnet | Found | Found | Flagged as doc gap | Described, no code |
| Human (`mrvarmazyar`) | Found | Missed | Found, more developed | Provided complete rewrites |

The human reviewer provided more prescriptive test code. Copilot caught the `node_meta` doc bug that the human missed.

---

## Takeaway

gl1tch turns a two-command PR fetch into a structured review in under 30 seconds. The value is first-pass plumbing verification — field routing, guard conditions, test coverage gaps — so human review time goes to semantics and design. On this PR, Copilot found one issue the human reviewer missed and matched everything else.
