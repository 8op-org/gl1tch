# gl1tch DSL Improvements for Data Pipeline Workflows

**Date:** 2026-04-15
**Status:** Draft
**Project:** gl1tch (`~/Projects/gl1tch`)
**Driven by:** Support email triage workflow in farm workspace

---

## Problem

Building the farm support email triage pipeline exposed several DSL gaps that force workflows to shell out to `curl`, `jq`, and `perl` for operations that should be native. The result is fragile, hard-to-read workflows full of shell quoting workarounds.

## Improvements

### 1. `(flatten "step-id")` — JSON Array to NDJSON

**Problem:** `(search)` returns a JSON array `[{},{},{}]`. `(map)` splits by newlines. You need a glue step to convert.

**Current workaround:**
```
(step "emails" (run "echo '{{step \"fetch\"}}' | jq -c '.[]'"))
```

**Proposed:**
```
(step "emails" (flatten "fetch"))
```

**Implementation:** In `runSingleStep`, read step output, `json.Unmarshal` into `[]json.RawMessage`, join with `\n`. ~10 lines of Go.

---

### 2. Make `(search)` `:query` optional

**Problem:** Every search requires `:query` even for match_all.

**Current:**
```
(search :index "my-index" :query {"match_all" {}} :size 10)
```

**Proposed:**
```
(search :index "my-index" :size 10)
```

Omitting `:query` defaults to match_all. The runner already handles this — just remove the parser validation.

**Implementation:** Delete the `if sr.Query == ""` error check in `convertSearch`. 1 line.

---

### 3. `(search)` `:sort` parameter

**Problem:** No way to sort search results. Critical for "latest N emails" queries.

**Current workaround:** Can't do it — have to fall back to `(run "curl ...")`.

**Proposed:**
```
(search :index "my-index" :query {"match_all" {}} :size 10 :sort {"indexed_at" "desc"})
```

**Implementation:** Add `Sort string` to `SearchStep`, parse `:sort` keyword in `convertSearch`, add `sort` field to query body in runner. ~15 lines.

---

### 4. `(llm)` `:format "json"` — Structured Output

**Problem:** LLM responses need post-processing to extract JSON from thinking tags, markdown fences, and conversational text. qwen3-8b wraps output in `<think>` tags and sometimes adds markdown fences.

**Proposed:**
```
(llm :tier 0 :format "json" :prompt "classify this email...")
```

When `:format "json"` is set:
1. Strip `<think>...</think>` blocks (multiline)
2. Strip markdown fences
3. Extract first `{...}` JSON object
4. Validate it parses as JSON
5. If no valid JSON found, return error (triggers tier escalation if tiered)

**Implementation:** Post-processing function on LLM output. ~20 lines. The `:format` keyword is already parsed in the sexpr (`case "format"`) but may not be wired into the runner.

---

### 5. `(search)` `:ndjson` flag — Output as NDJSON

**Problem:** `(search)` returns a JSON array, but `(map)` needs NDJSON. Combined with `(flatten)` above, this is two ways to solve the same problem. `:ndjson` is cleaner if we want a single step.

**Proposed:**
```
(search :index "my-index" :query {"match_all" {}} :size 10 :ndjson)
```

Output: one JSON object per line instead of a JSON array.

**Implementation:** Add `NDJSON bool` to `SearchStep`, if set, join sources with `\n` instead of wrapping in `[]`. ~5 lines. Alternative to `(flatten)` — do one or the other.

**Recommendation:** Do both. `(flatten)` is general-purpose (works on any step output). `:ndjson` is search-specific sugar.

---

### 6. `(index)` from `(map)` item — `:doc` template support

**Problem:** `(index :doc "{{.param.item}}")` in a `(map)` body should work but hasn't been tested. If templates render correctly for index docs, the kb-ingest workflow collapses to:

```
(each "fetch"
  (step "idx"
    (index :index "glitch-{{.workspace}}-knowledge-kb-articles"
           :doc "{{.param.item}}"
           :id "kb-{{.param.item | pick \"article_id\"}}"
           :embed :field "content" :model "nomic-embed-text")))
```

**Needs:** Verify `{{.param.item}}` works in `:doc`, verify `pick` template function exists or add it. The `pick` function would extract a field from a JSON string: `{{.param.item | pick "article_id"}}` → `"42"`.

---

### 7. `pick` template function

**Problem:** Inside `(map)` bodies, `{{.param.item}}` is a JSON string. There's no way to extract a field without shelling out to jq or writing to a file.

**Proposed:**
```
{{.param.item | pick "subject"}}
{{.param.item | pick "from"}}
```

**Implementation:** Parse JSON, extract field by key, return as string. ~10 lines in the `render` funcMap. Handle nested access with dot notation: `pick "email.subject"`.

---

### 8. Email-ingest native `(index)` with dedup

**Problem:** email-ingest still uses `(run "curl ...")` because it needs dedup (check if doc exists before indexing). There's no native dedup in `(index)`.

**Proposed option A — `(index)` with `:upsert false`:**
```
(index :index "..." :doc "..." :id "..." :upsert false)
```
Skip indexing if doc already exists. Simple, covers the dedup case.

**Proposed option B — `(cond)` with `(search)`:**
```
(cond
  (when (search :index "..." :query {"term" {"_id" "..."}} :size 1) "0")
  (step "idx" (index ...)))
```
More flexible but verbose.

**Recommendation:** Option A. Dedup-on-index is a common ES pattern.

---

## Priority Order

| # | Form | Impact | Effort |
|---|------|--------|--------|
| 1 | `:query` optional in `(search)` | Removes boilerplate | 1 line |
| 2 | `(flatten)` | Unblocks `(search)` → `(map)` | 10 lines |
| 3 | `pick` template function | Clean field access in map bodies | 10 lines |
| 4 | `(llm :format "json")` | Reliable structured output | 20 lines |
| 5 | `(search :sort)` | Enables "latest N" queries | 15 lines |
| 6 | `(search :ndjson)` | Sugar for flatten | 5 lines |
| 7 | `(index :upsert false)` | Native dedup | 10 lines |

Items 1-4 would immediately clean up the support-triage workflow. Items 5-7 are nice-to-haves.

## What the Triage Workflow Would Look Like

With all improvements:

```
(workflow "support-triage"
  :description "Classify support emails and match to KB articles"

  (step "emails"
    (search :index "glitch-{{.workspace}}-support-emails"
            :size 10
            :sort {"indexed_at" "desc"}
            :fields ("message_id" "from" "subject" "body" "date")
            :ndjson))

  (each "emails"
    (step "classify"
      (llm :tier 0 :format "json" :prompt ```
Classify this support email for Marketspread (farmers market platform).

Categories: billing, onboarding, application, scheduler, technical, account, other

Subject: {{.param.item | pick "subject"}}
From: {{.param.item | pick "from"}}
Body: {{.param.item | pick "body"}}

JSON: {"category": "...", "urgency": "low|medium|high", "summary": "..."}
      ```))))
```

**From 90 lines of shell to 20 lines of clean DSL.**
