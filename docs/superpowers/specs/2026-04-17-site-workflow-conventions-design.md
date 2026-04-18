# Site Workflow Conventions & Language Extensions

**Date:** 2026-04-17
**Status:** Proposed
**Problem:** Site workflows generate content with stale syntax (e.g., `{{step "name"}}` instead of `~(step name)`) because LLMs lack conventions context and gates don't check interpolation correctness. Inline Python/jq heredocs handle data plumbing that the sexpr language should handle natively.

---

## 1. Language Changes

Three additions to the sexpr parser (`internal/pipeline/sexpr.go`).

### 1a. `(include "path")`

Top-level form. Parses another `.glitch` file and merges its `(def ...)` bindings into the current file's scope. Paths are relative to the workspace root.

```glitch
(include "site/shared.glitch")

(workflow "site-update-page" ...)
```

Rules:
- Only `(def ...)` forms are imported — no workflows, no steps
- Circular includes are a parse-time error
- Multiple includes are fine, evaluated in order
- If two files define the same name, last wins

Implementation: in `parseSexprWorkflow`, before collecting defs, scan for `(include ...)` top-level forms. For each, parse the referenced file recursively, collect its defs, and merge into the current def map. Track visited paths to detect cycles.

### 1b. `def` evaluates forms

Extend `resolveVal` to evaluate a known set of pure forms when they appear as the value in a `(def ...)`:

**File I/O:**
- `(read-file "path")` — read file contents as string
- `(read-file "path1" "path2" ...)` — read multiple files, concatenate with `\n\n` separator
- `(glob "pattern")` — expand glob, return newline-separated list of paths (consistent with `lines`/`map` iteration model)

**Data:**
- `(json-pick expr)` — extract from JSON using jq-style path
- `(assoc map :key val ...)` — associate keys into a map
- `(pick map :key1 :key2)` — select keys from a map
- `(merge a b)` — merge step outputs or maps

**Strings:**
- `split`, `join`, `trim`, `upper`, `lower`, `replace`, `truncate`, `contains`

**Threading:**
- `(-> val (form1) (form2) ...)` — pipe output through a chain of forms

**Collections:**
- `(map form list)` — apply form to each element
- `(filter pred list)` — keep elements matching predicate
- `(flatten list)` — flatten nested lists
- `(lines str)` — split string into list of lines

Example:

```glitch
(def conventions (read-file "site/conventions.md"))

(def examples
  (-> (glob "examples/*.glitch")
      (map read-file)
      (join "\n\n")))

(def valid-commands
  (-> (read-file "scripts/gate-hallucinations.py")
      (lines)
      (filter (contains "\"glitch "))
      (join "\n")))
```

Evaluation happens at **parse time**, before the workflow runs. These are constants. If a form references `~param.*` or `~(step ...)`, that's a parse error — those are runtime values.

### 1c. Forms work in step bodies

The same forms that work in `def` also work as step actions, replacing inline shell for data plumbing:

```glitch
(step "context"
  (-> (read-file "~(stepfile merge-context)")
      (json-pick ".pages[*].slug")
      (join "\n")))
```

In step context, `~param.*` and `~(step ...)` references are allowed since they resolve at runtime. A step whose body is a form chain (instead of `run`/`llm`/`save`) produces the chain's result as its output, the same way `(run ...)` captures stdout.

This replaces the pattern of `(run "python3 - <<'EOF' ... EOF")` for structured data manipulation while keeping `(run ...)` for actual shell work (git, curl, process management).

---

## 2. Conventions File & Syntax Linter

### 2a. `site/conventions.md`

A compact reference file (~80-100 lines) injected into every site LLM prompt via `~conventions`. Hand-written for tone/philosophy, with syntax rules matching the current codebase.

Contents:

```markdown
## Interpolation Syntax

Step references: ~(step name) — no quotes, no braces
Parameters: ~param.key
Input: ~input
Stepfile: ~(stepfile name)
Threading: ~(split "/" "a/b/c" | first)

NEVER use {{step "name"}} or {{.param.key}} — that's old Go template
syntax and does not work.

## Tone

- "your" framing throughout, never "the user"
- Examples before explanation — show real code first, explain second
- No internal implementation details: BubbleTea, tmux, SQLite,
  Go types, OTel, internal package names
- Every code example must come from CONTEXT, not from training data

## Decommissioned Features

- glitch ask — removed, do not mention in any content

## Frontmatter

Docs require: title, order, description
Labs require: title, slug, description, date

## Code Blocks

- Workflow examples use ````glitch` fence
- Shell examples use ```bash fence
- All workflow code must use current interpolation syntax
```

### 2b. `site/shared.glitch`

Shared defs for all site workflows, loaded via `(include)`:

```glitch
;; site/shared.glitch — shared context for all site workflows

(def conventions (read-file "site/conventions.md"))

(def examples
  (-> (glob "examples/*.glitch")
      (map read-file)
      (join "\n\n")))

(def model "qwen2.5:7b")
```

### 2c. `scripts/gate-syntax.py`

New gate script. Pure regex, no LLM. Checks:

- `{{step "..."}}` or `{{step '...'}}` — error: old Go template syntax
- `{{.param.*}}` or `{{.input}}` — error: old Go template syntax
- `glitch ask` — error: decommissioned command

Scans both `site/src/content/docs/*.md` and `site/src/content/labs/*.md`.

This gate runs in the verify phase of every site workflow. It would have caught the regression that prompted this design.

---

## 3. Site Workflow Family

### 3a. Workflow inventory

| Command | Workflow | Purpose |
|---------|----------|---------|
| `glitch workflow run site-create-page --set topic="..."` | `site-create-page` | Generate a new doc page |
| `glitch workflow run site-update-page --set page=getting-started` | `site-update-page` | Rewrite an existing doc page |
| `glitch workflow run site-create-lab --set issue="owner/repo#123"` | `site-create-lab` | Generate a lab case study from a GH issue |
| `glitch workflow run site-update-lab --set lab=bug-triage-kibana` | `site-update-lab` | Rewrite an existing lab |
| `glitch workflow run site-sync` | `site-sync` | Bulk reconcile manifest against disk |
| `glitch workflow run site-deploy` | `site-deploy` | Gate, build, commit, push |

### 3b. Common shape

Every content workflow follows the same skeleton:

```glitch
(include "site/shared.glitch")

(workflow "site-*"
  :description "..."

  ;; 1. Gather context (native forms, no shell plumbing)
  (step "context" ...)

  ;; 2. Generate content (LLM with conventions injected)
  (step "generate"
    (llm
      :prompt ```
        ~conventions

        REAL EXAMPLES (match this syntax exactly):
        ~examples

        ... task-specific instructions ...
        ```))

  ;; 3. Save to disk
  (step "save" ...)

  ;; 4. Verify (all workflows share the same phase)
  (phase "verify" :retries 1
    (gate "hallucinations" (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"         (run "python3 scripts/gate-syntax.py"))
    (gate "structure"      (run "python3 scripts/gate-structure.py"))
    (gate "links"          (run "python3 scripts/gate-links.py")))

  (step "done" ...))
```

Key changes from today:
- `~conventions` and `~examples` injected in every LLM prompt — the LLM always sees current syntax
- `gate-syntax` added to every verify phase — catches bad interpolation even if the LLM ignores conventions
- Labs go through the same gates — today `labs-generate` skips verification entirely
- Inline Python data plumbing replaced by native forms where possible

### 3c. `site-deploy`

New workflow — the final gate before content goes live:

```glitch
(include "site/shared.glitch")

(workflow "site-deploy"
  :description "Verify, build, commit, push for GitHub Pages"

  ;; 1. Run all gates against current disk state
  (phase "verify" :retries 0
    (gate "hallucinations" (run "python3 scripts/gate-hallucinations.py"))
    (gate "syntax"         (run "python3 scripts/gate-syntax.py"))
    (gate "structure"      (run "python3 scripts/gate-structure.py"))
    (gate "links"          (run "python3 scripts/gate-links.py"))
    (gate "sidebar"        (run "python3 scripts/gate-sidebar.py")))

  ;; 2. Build
  (step "build"
    (run "cd site && npx astro build"))

  ;; 3. Playwright smoke test
  (phase "smoke" :retries 0
    (gate "playwright" (run "bash scripts/gate-playwright.sh")))

  ;; 4. Commit and push
  (step "commit"
    (run ```
      git add site/src/content/ site/src/pages/
      git commit -m "docs: update site content"
      git push
      ```))

  (step "done"
    (run "echo 'Deployed. GitHub Pages will pick up the push.'")))
```

Zero retries on verify gates. If something fails, fix it with `site-update-page` and come back.

---

## 4. Implementation Order

### Phase 1: Immediate fix (no language changes needed)
1. Create `site/conventions.md`
2. Create `scripts/gate-syntax.py`
3. Add `gate-syntax` to verify phase in all existing site workflows
4. Manually add `conventions` and `examples` as shell steps in existing workflows (bridge solution until `include`/`def` land)

### Phase 2: Language extensions
5. Implement `(include "path")` in `parseSexprWorkflow`
6. Extend `resolveVal` to evaluate pure forms in `def` context
7. Enable form chains in step bodies (alongside `run`/`llm`/`save`)

### Phase 3: Workflow consolidation
8. Create `site/shared.glitch` using new `(include)` + `(def)` features
9. Rewrite `site-create-page`, `site-update-page` to use `(include)` and native forms
10. Create `site-create-lab`, `site-update-lab` (split from current `labs-generate`)
11. Create `site-deploy`
12. Migrate `site-sync` inline Python to native forms where practical

### Phase 4: Cleanup
13. Remove redundant inline Python/shell data plumbing from migrated workflows
14. Update workflow-syntax docs to cover `include`, extended `def`, and form chains

---

## 5. What This Does NOT Include

- No embedded scripting runtime (Guile, Janet, Lua) — the native forms cover data plumbing; shell covers real computation
- No multi-site support — hardcoded to 8op.org
- No workspace-level implicit inheritance — `(include)` is explicit per-workflow
- No changes to the site manifest format
- No changes to the existing gate scripts (hallucinations, structure, links, sidebar, coverage) — only adds `gate-syntax`
