# Site Workflows EDN Migration + Lab Authoring System

Date: 2026-04-22

## Summary

Fix the five existing site workflows to target the new ClojureScript/EDN content system. Build a REPL capture library for authoring verifiable labs. Add a lab gate workflow for structural and voice validation. Create seven new labs across three categories.

## Problem

The site was rewritten from Astro (markdown) to ClojureScript + shadow-cljs (EDN hiccup). All five site workflows still target `site/src/content/docs/*.md` — they're broken. The single existing lab (`repl-driven-docs`) references stale Astro-era paths. There's no tooling for authoring labs from a live REPL with captured evidence.

## Design

### 1. Site Workflow EDN Migration

All site workflows update to read/write `.edn` files under `site/content/`.

**EDN doc schema** (what `content.clj` macro expects):

```clojure
{:title "Page Title"
 :description "One-line description"
 :order 5
 :sections
 [{:heading "Section Name"
   :level 2
   :body [[:p "Text with " [:code "inline code"] " and " [:a {:href "#/docs/other"} "links"]]
          [:code {:lang "glitch"} "(step \"example\" (run \"echo hi\"))"]]}]}
```

**EDN lab schema:**

```clojure
{:title "Lab Title"
 :description "One-line description"
 :date "2026-04-22"
 :duration "15min"
 :steps
 [{:heading "Step Name"
   :body [[:p "Explanation"]
          [:code {:lang "clojure"} "(some-expr)"]
          [:code {:lang "output"} "captured result"]]}]}
```

**Workflows affected:**

- `site-write.glitch` — LLM prompt changes to "return JSON matching EDN schema." Save step writes to `site/content/docs/{slug}.edn`. Gather-context step updates paths.
- `site-gate.glitch` — Validates EDN structure instead of markdown. Voice gate extracts text from hiccup. Accuracy gate unchanged.
- `site.glitch` (router) — Adds `"lab"` and `"gate-lab"` actions. Updates `"write-page"` and `"write-all"` to target EDN paths. Page listing step reads from `site/content/docs/`.
- `site-dev.glitch` — Updates dev server command if needed (shadow-cljs instead of astro).
- `site-homepage.glitch` — Targets `site/src/gl1tch/site/pages/home.cljs` instead of `index.astro`. Gathers context from CLJS source files.

### 2. REPL Capture Library

A single file at `.glitch/src/lab_capture.clj` — already on the REPL classpath.

**Namespace:** `lab-capture`

**State:** One atom holding the current lab session:

```clojure
{:slug "investigate-bug"
 :title "Investigating a GitHub Issue"
 :description "Triage a bug using shell + LLM steps"
 :duration "15min"
 :date "2026-04-22"
 :steps []}
```

**API (4 functions):**

`(start-lab slug opts-map)` — Resets state, begins a new lab session. Date auto-filled to today. Required opts: `:title`, `:description`. Optional: `:duration`.

`(record! heading expr)` — Macro. Captures:
1. The expression source as a string (for the code block)
2. Evals the expression
3. Captures the result as a string (for an output block)
4. Appends a step to the session
5. Returns the result (still usable in the REPL flow)

If a step with the same heading already exists, it replaces it (re-record workflow).

`(annotate! prose)` — Appends a `:p` element to the most recent step's body. Call between `record!` calls to add narrative.

`(save-lab!)` — Pretty-prints the EDN and writes to `site/content/labs/{slug}.edn`. Returns the path. Does not gate — gating is a separate workflow step.

**Example REPL session:**

```clojure
(require '[lab-capture :as lab])

(lab/start-lab "investigate-bug"
  {:title "Investigating a GitHub Issue"
   :description "Triage a bug using shell + LLM steps"
   :duration "15min"})

(lab/record! "Fetch the issue"
  (sh "gh" "issue" "view" "42" "--repo" "8op-org/gl1tch" "--json" "title,body,labels"))

(lab/annotate! "Shell fetches the raw issue data. Now feed it to an LLM for triage.")

(lab/record! "Triage with LLM"
  (llm :provider "openrouter"
       :prompt (str "Triage this issue:\n" (ref "Fetch the issue"))))

(lab/save-lab!)
;; => "site/content/labs/investigate-bug.edn"
```

### 3. Lab Gate Workflow

New file: `lab-gate.glitch`

**Pass 1 — Structural gate (bb script, no LLM):**
- Required keys present (`:title`, `:description`, `:date`, `:steps`)
- Each step has `:heading` and `:body`
- Body elements use valid hiccup tags (`:p`, `:code`, `:a`, `:strong`, `:em`, `:ul`, `:ol`, `:li`)
- `:code` blocks have a `:lang` attribute
- No slug collision with existing labs
- Output: `{"pass": true}` or `{"pass": false, "issues": [...]}`

**Pass 2 — Voice gate (LLM, C.L.E.A.R. adapted for labs):**
- Steps flow logically (each builds on the previous)
- No internal implementation details leaking
- No banned terms (Ollama, glitch ask, BubbleTea, tmux, OTel, SQLite)
- Code blocks are real commands/expressions, not pseudocode

**Params:** `--set slug=investigate-bug`

**Integration:** `site.glitch` router gains `"gate-lab"` action so `glitch run site "gate the investigate-bug lab"` works.

### 4. Lab Content

Seven new labs across three categories. All accept `--set` params for portability; each ships with one concrete recorded run as the example.

**Workflow-craft:**

| # | Slug | Title | Concepts |
|---|------|-------|----------|
| 1 | `code-review-pipeline` | Build a Code Review Pipeline | `(step)`, `(ref)`, `(save)`, shell+LLM pattern |
| 2 | `phases-and-gates` | Phases and Gates | `(phase)`, `(gate)`, `:retries`, `(retry)` |
| 3 | `es-search-workflow` | ES-Backed Search Workflow | `(index)`, `(search)`, `(embed)`, local ES |

**Feature:**

| # | Slug | Title | Concepts |
|---|------|-------|----------|
| 4 | `tiered-routing` | Tiered Routing | `:tier`, fallback, `:format "json"` validation |
| 5 | `plugins-in-practice` | Plugins in Practice | namespaced shorthand, verbose form, `--set` |

**Engineer-on-an-issue:**

| # | Slug | Title | Concepts |
|---|------|-------|----------|
| 6 | `investigate-bug` | Investigating a Bug | `gh` fetch, related PRs, LLM triage, `--set repo=X issue=N` |
| 7 | `ship-feature-request` | Shipping a Feature Request | issue→tasks, skeleton workflow gen, gate check, `--set repo=X issue=N` |

### 5. Fix Existing Lab

Update `site/content/labs/repl-driven-docs.edn` to use current paths and real commands:
- Remove references to `workflows/site/generate-doc.glitch` and `workflows/site/gate-syntax.glitch`
- Update to use `glitch run site "..."` via the router
- Update REPL examples to match current `glitch.core` API

## Out of Scope

- Playwright/browser testing for the CLJS site (separate effort)
- Astro removal/cleanup (the old site files can be cleaned up independently)
- New doc pages (this spec covers workflow fixes and lab content only)

## Implementation Order

1. REPL capture library (`.glitch/src/lab_capture.clj`) — unblocks lab authoring immediately
2. Lab gate workflow (`lab-gate.glitch`) — needed before saving labs
3. Site workflow EDN migration (all five workflows) — can happen in parallel with labs
4. Fix existing lab (`repl-driven-docs.edn`)
5. Author seven new labs from the REPL using the capture library
6. Update `site.glitch` router with `"lab"` and `"gate-lab"` actions
