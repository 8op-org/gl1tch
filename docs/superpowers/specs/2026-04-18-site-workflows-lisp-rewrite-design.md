# Site Workflows: Pure Lisp Rewrite

**Date:** 2026-04-18
**Status:** Design
**Target branch:** lisp-evaluator

## Summary

Rewrite all 6 site workflows and their supporting Python scripts as pure lisp, targeting the lisp-evaluator branch. Eliminate all Python dependencies. Adopt a Clojure deps.edn-style manifest format. Add new evaluator builtins and interpolated text blocks to support the rewrite.

## Goals

- Zero Python files in the site workflow pipeline
- Manifest as pure evaluator data (deps.edn approach)
- Parallel execution via `par` for independent steps and gates
- Interpolated text blocks (` ``` `) with `~` resolution and auto-dedent
- New evaluator builtins: regex, set ops, assertions, string predicates

## Non-goals

- Rewriting Playwright tests (shell tool, stays as `run`)
- Changing the Astro site structure
- Changing the site-manifest schema semantics (just the syntax)

---

## 1. Manifest: deps.edn Style

### Current format (custom forms, requires Python parser)

```lisp
(site "gl1tch"
  :url "https://8op.org"
  (section "Getting Started"
    (page "getting-started" :title "Getting Started" :order 1 ...)))
```

### New format (native evaluator data)

```lisp
;;; site-manifest.glitch — pure data, deps.edn style
;;;
;;; Single source of truth for 8op.org site structure.
;;; Consumers: (include "site-manifest.glitch") → returns assoc map.

(assoc
  :site "gl1tch"
  :url "https://8op.org"

  :homepage (assoc
    :slug "index"
    :title "gl1tch — your AI, your terminal, your rules"
    :template "homepage"
    :sections (list "hero" "features" "how-it-works" "meta" "agents")
    :context-query "top-level project description, feature highlights, install instructions"
    :context-paths (list "README.md"))

  :sections (list
    (assoc :label "Getting Started" :sidebar true :pages (list
      (assoc :slug "getting-started"
             :title "Getting Started"
             :order 1
             :description "Install, configure, run your first workflow"
             :context-query "brew install, initial setup, hello-world workflow, glitch run basics"
             :context-paths (list "cmd/run.go" "cmd/root.go"))
      (assoc :slug "local-models"
             :title "Local Models"
             :order 2
             :description "LM Studio, Ollama, GPU allocation, context tuning"
             :context-query "Ollama and LM Studio provider configuration, model selection, tier routing"
             :context-paths (list "internal/provider/"))))

    (assoc :label "Workflow Language" :sidebar true :pages (list
      (assoc :slug "workflow-syntax"
             :title "Workflow Syntax"
             :order 3
             :description "S-expression forms, steps, interpolation, LLM options"
             :context-query "sexpr parser, step types, run/llm/save forms, interpolation syntax, tier routing"
             :context-paths (list "internal/sexpr/" "internal/pipeline/sexpr.go"))
      (assoc :slug "dsl-reference"
             :title "DSL Reference"
             :order 4
             :description "Threading, filter, reduce, ES forms, embed, flatten"
             :context-query "threading form, filter, reduce, search/index/delete ES forms, embed, flatten, assoc, pick"
             :context-paths (list "internal/pipeline/sexpr.go"))
      (assoc :slug "phases-and-gates"
             :title "Phases & Gates"
             :order 5
             :description "Phase grouping, gate assertions, retries"
             :context-query "phase form, gate form, retry semantics, phase execution order"
             :context-paths (list "internal/pipeline/sexpr.go" "internal/pipeline/runner.go"))))

    (assoc :label "Workspaces & Resources" :sidebar true :pages (list
      (assoc :slug "workspaces"
             :title "Workspaces"
             :order 6
             :description "Workspace config, resources, defaults, discovery, nested runs"
             :context-query "workspace init/add/sync/pin/rm commands, workspace.glitch format, resource types, call-workflow"
             :context-paths (list "cmd/workspace*.go" "internal/workspace/"))
      (assoc :slug "providers"
             :title "Providers & Config"
             :order 7
             :description "Provider config, Ollama, LM Studio, OpenAI-compat, tiered routing, config.glitch"
             :context-query "provider protocol, config.glitch format, glitch config show/set, provider tiers, OpenAI-compatible endpoint config, api-key-env"
             :context-paths (list "cmd/config.go" "internal/provider/" "internal/pipeline/tier.go"))))

    (assoc :label "Running Workflows" :sidebar true :pages (list
      (assoc :slug "compare"
             :title "Compare Runs"
             :order 8
             :description "A/B testing models, strategies, review scoring"
             :context-query "compare form, branch form, review criteria, --variant flag, --compare flag"
             :context-paths (list "cmd/run.go" "internal/pipeline/compare.go"))
      (assoc :slug "batch-runs"
             :title "Batch Runs"
             :order 9
             :description "Run variants side by side, multi-provider comparison, fan-out"
             :context-query "--variant flag, --compare discovery pattern, variant workflow naming, batch execution, nested runs"
             :context-paths (list "cmd/run.go" "internal/pipeline/runner.go"))))

    (assoc :label "Code Intelligence" :sidebar true :pages (list
      (assoc :slug "code-intelligence"
             :title "Code Intelligence"
             :order 10
             :description "Index repos, query with natural language, glitch up/down"
             :context-query "glitch index command, glitch observe command, language extractors, symbol indexing, ES queries, BFS depth traversal, glitch up/down docker compose"
             :context-paths (list "cmd/index.go" "cmd/observe.go" "cmd/up.go" "internal/esearch/" "internal/capability/"))))

    (assoc :label "Extending" :sidebar true :pages (list
      (assoc :slug "plugins"
             :title "Plugins"
             :order 11
             :description "Plugin directories, manifests, subcommands, argument types"
             :context-query "plugin discovery, plugin manifest, subcommand definitions, argument types, call-workflow from plugins"
             :context-paths (list "cmd/plugin.go" "internal/plugin/"))))

    (assoc :label "Labs" :sidebar false :pages (list
      (assoc :slug "issue-triage-kubernetes"
             :title "Triaging a Kubernetes Issue with Tier Routing"
             :template "lab")
      (assoc :slug "pr-review-prometheus"
             :title "Reviewing a Prometheus PR with Copilot"
             :template "lab")
      (assoc :slug "model-showdown-containerd"
             :title "Model Showdown: Free vs Paid vs Copilot"
             :template "lab")
      (assoc :slug "bug-triage-kibana"
             :title "Triaging a Kibana Regex Bug Across Three Tiers"
             :template "lab")))))
```

### Consumer pattern

```lisp
(def manifest (include "site-manifest.glitch"))
(def sections (pick manifest :sections))
(def homepage (pick manifest :homepage))
```

No parser. No JSON intermediate. The evaluator IS the reader.

---

## 2. New Evaluator Features

### 2a. Interpolated Text Blocks

Triple-backtick blocks become first-class interpolated strings with auto-dedent.

**Syntax:**

````lisp
(llm :prompt ```
  You are a writer for ~(pick manifest :site).

  CONTEXT:
  ~(ref "context-step")

  Write about ~param.topic.
  ```)
````

**Semantics:**

| Feature | Behavior |
|---------|----------|
| `~symbol` | Resolve variable from scope |
| `~(form)` | Evaluate form, interpolate result |
| `~param.key` | Parameter reference |
| `~env.VAR` | Environment variable |
| `\~` | Literal tilde (escape) |
| Auto-dedent | Strip common leading whitespace based on closing ` ``` ` indentation |
| Trailing newline | Trimmed — closing ` ``` ` eats the last newline |

**Implementation:** The `lexQuasi` tokenizer already handles `~` interpolation. Changes needed:

1. **Parser**: recognize ` ``` ` as opening a text block (already does this)
2. **Evaluator**: pipe text block content through `lexQuasi` for `~` resolution (new)
3. **Parser**: auto-dedent based on closing ` ``` ` column position (new)

### 2b. New Builtins

| Builtin | Signature | Purpose |
|---------|-----------|---------|
| `regex-match` | `(regex-match pattern text)` -> list | All matches. If the pattern has capture groups, returns the first group per match; otherwise returns the full match. |
| `regex-find` | `(regex-find pattern text)` -> string/nil | First match. Same capture group behavior as `regex-match`. |
| `lines` | `(lines text)` -> list | Split text on newlines |
| `count` | `(count list)` -> number | Length of list |
| `some` | `(some list pred)` -> bool | True if any item matches predicate |
| `every` | `(every list pred)` -> bool | True if all items match predicate |
| `starts-with` | `(starts-with text prefix)` -> bool | String prefix check |
| `ends-with` | `(ends-with text suffix)` -> bool | String suffix check |
| `slice` | `(slice text start end)` -> text | Substring extraction |
| `set` | `(set list)` -> list | Deduplicate list |
| `difference` | `(difference set-a set-b)` -> list | Items in a not in b |
| `assert` | `(assert condition message)` -> pass or exit 1 | Gate assertion primitive |
| `<` | `(< a b)` -> bool | Numeric less-than |

These are standard Lisp/Clojure primitives. All implemented as Go functions in `internal/eval/builtins.go`.

---

## 3. Workflow Rewrites

### 3a. shared.glitch (unchanged — already native)

```lisp
(def conventions (read-file "site/conventions.md"))

(def examples
  (-> (glob "examples/*.glitch")
      (map read-file)
      (join "\n\n")))

(def model "qwen2.5:7b")
```

### 3b. site.glitch (dispatcher)

```lisp
(workflow "site"
  :description "Unified site workflow — sync, create, update, deploy, dev"

  (cond
    (= ~param.action "sync")    (call-workflow "site-sync")
    (= ~param.action "create")  (call-workflow "site-create-page" :set "topic=~param.topic")
    (= ~param.action "update")  (call-workflow "site-update-page"
                                  :set "page=~param.page"
                                  :set "instructions=~param.instructions")
    (= ~param.action "deploy")  (call-workflow "site-deploy")
    (= ~param.action "dev")     (call-workflow "site-dev")
    else (run "echo 'Usage: glitch run site --set action=<sync|create|update|deploy|dev>' && exit 1")))
```

### 3c. site-sync.glitch

```lisp
;; site-sync.glitch — reconcile site-manifest against disk
;;
;; Idempotent: if manifest hasn't changed and no upstream source files
;; have been modified since the last doc was written, generation is skipped.

(workflow "site-sync"
  :description "Reconcile site-manifest against disk, generate/update doc pages, verify"

  ;; ── Load manifest ────────────────────────────────────────────────
  (def manifest (include "site-manifest.glitch"))
  (def sections (filter (pick manifest :sections) (pick item :sidebar)))
  (def manifest-pages (-> sections (map (fn (s) (pick s :pages))) (flatten)))

  ;; ── Diff disk ────────────────────────────────────────────────────
  (def disk-slugs
    (-> (glob "site/src/content/docs/*.md")
        (map (fn (f) (replace f "site/src/content/docs/" "" ".md" "")))))

  (def needs-create
    (filter manifest-pages
      (not (contains disk-slugs (pick item :slug)))))

  (def needs-update
    (filter manifest-pages
      (let (slug (pick item :slug)
            disk-path (str "site/src/content/docs/" slug ".md"))
        (when (contains disk-slugs slug)
          (let (content (read-file disk-path)
                fm-title (json-pick content "title"))
            (or (not (= fm-title (pick item :title)))
                (run (str "test -n \"$(find " (join (pick item :context-paths) " ")
                          " -newer " disk-path " 2>/dev/null)\""))))))))

  (def orphans
    (filter disk-slugs
      (not (contains (map manifest-pages (fn (p) (pick p :slug))) item))))

  (def needs-work (list needs-create needs-update))

  ;; ── Summary ──────────────────────────────────────────────────────
  (step "diff-summary"
    (str "create=" (count needs-create) " update=" (count needs-update)
         " ok=" (count (difference disk-slugs (map needs-work (fn (p) (pick p :slug)))))
         " orphan=" (count orphans)))

  ;; ── Gather context (parallel) ────────────────────────────────────
  (par
    (step "es-context"
      (catch
        (-> (http-post "http://localhost:9200/glitch-code-*/_search"
              :content-type "application/json"
              :body (str ```
                {"query":{"query_string":{"query":"
                ~(-> needs-work
                    (map (fn (p) (str "(" (pick p :context-query) ")")))
                    (join " OR "))
                ","default_field":"content"}},"size":30,
                "_source":["file_path","content"]}
                ```))
            (json-pick ".hits.hits[]._source"))
        ""))

    (step "fallback-context"
      (map needs-work
        (let (paths (pick item :context-paths))
          (-> paths
              (map (fn (cp)
                (-> (glob cp)
                    (map (fn (f) (str "=== " f " ===\n" (read-file f))))
                    (join "\n\n"))))
              (join "\n\n"))))))

  (step "merged-context"
    (map needs-work
      (let (slug (pick item :slug))
        (str "=== CONTEXT FOR: " slug " ===\n"
             (ref "es-context") "\n" (ref "fallback-context")))))

  ;; ── Generate pages ───────────────────────────────────────────────
  (map needs-work
    (let (slug (pick item :slug)
          title (pick item :title)
          order (pick item :order)
          description (pick item :description))

      (step "generate-page"
        (llm
          :provider "claude"
          :model "sonnet"
          :prompt ```
            You are a technical writer for gl1tch (https://8op.org).

            VOICE RULES — follow these exactly:
            - "your" framing throughout; never say "the user"
            - examples before explanation; show real code first, explain second
            - no internal implementation details: never mention BubbleTea, tmux,
              SQLite, Go types, OTel, or any internal package names
            - every code example must be real — drawn from the CONTEXT below
            - do NOT invent commands, flags, or features that aren't in the context

            Write the documentation page for slug "~slug".

            CONTEXT:
            ~(ref "merged-context")

            Output: markdown body only — no frontmatter, start with ## heading.
            ```))

      (write-file (str "site/src/content/docs/" slug ".md")
        :content ```
          ---
          title: "~(replace title "\"" "\\\"")"
          order: ~order
          description: "~(replace description "\"" "\\\"")"
          ---

          ~(ref "generate-page")
          ```)))

  ;; ── Homepage sync ────────────────────────────────────────────────
  (let (manifest-sections (pick (pick manifest :homepage) :sections)
        index-text (read-file "site/src/pages/index.astro"))

    (step "disk-sections"
      (llm :prompt ```
        Extract all <section> id attributes from this file.
        Output one id per line, nothing else.

        ~index-text
        ```))

    (when-not (= (ref "disk-sections") (join manifest-sections "\n"))
      (step "rewrite-homepage"
        (llm
          :provider "claude"
          :model "sonnet"
          :prompt ```
            Update this index.astro to match these sections: ~(join manifest-sections ", ")
            Keep all existing HTML/CSS. Only add/remove <section> blocks.
            Return the complete file, no fences.

            ~index-text
            ```))
      (write-file "site/src/pages/index.astro" :content (ref "rewrite-homepage"))))

  ;; ── Sidebar ──────────────────────────────────────────────────────
  (let (sidebar-sections (filter (pick manifest :sections) (pick item :sidebar)))
    (write-file "site/src/components/DocSidebar.astro"
      :content ```
        ---
        const { currentSlug } = Astro.props;
        const sections = [
        ~(-> sidebar-sections
            (map (fn (s) (str
              "  { label: \"" (pick s :label) "\", pages: [\n"
              (-> (pick s :pages)
                  (map (fn (p) (str "    { slug: \"" (pick p :slug)
                                    "\", title: \"" (pick p :title) "\" }")))
                  (join ",\n"))
              "\n  ]}")))
            (join ",\n"))
        ];
        ---
        <nav class="doc-sidebar">
          {sections.map(s => (
            <div class="sidebar-section">
              <div class="sidebar-label">{s.label}</div>
              {s.pages.map(p => (
                <a href={`/docs/${p.slug}`}
                   class:list={["sidebar-link", { active: currentSlug === p.slug }]}>
                  {p.title}
                </a>
              ))}
            </div>
          ))}
        </nav>
        ```))

  ;; ── Changelog ────────────────────────────────────────────────────
  (step "changelog-raw"
    (run "git log --oneline --since='2025-01-01' -- cmd/ internal/ examples/ .glitch/ docs/site/"))

  (step "enrich-changelog"
    (llm
      :provider "copilot"
      :prompt ```
        Summarize these git commits into user-facing changelog entries.
        Group by feature area (workflow engine, providers, CLI, plugins, GUI, docs).
        Skip purely internal refactors unless they change user-visible behavior.
        Output as markdown. Each entry: ### heading, then bullet points.

        Commits:
        ~(ref "changelog-raw")
        ```))

  (save "site/generated/changelog.md" :from "enrich-changelog")

  ;; ── Labs injection ───────────────────────────────────────────────
  (let (lab-files (glob "site/generated/labs/*.json"))
    (when lab-files
      (-> (glob "site/src/content/labs/*.md")
          (filter (fn (f) (not (contains f ".gitkeep"))))
          (map (fn (f) (run (str "rm " f)))))

      (map lab-files
        (let (lab (read-file item)
              slug (json-pick lab ".slug")
              title (json-pick lab ".title")
              desc (json-pick lab ".description")
              date (json-pick lab ".date")
              content (json-pick lab ".content"))
          (write-file (str "site/src/content/labs/" slug ".md")
            :content ```
              ---
              title: "~(replace title "\"" "\\\"")"
              slug: "~slug"
              description: "~(replace desc "\"" "\\\"")"
              date: "~date"
              ---

              ~content
              ```)))))

  ;; ── Verify ───────────────────────────────────────────────────────
  (phase "verify" :retries 1
    (par
      (gate "hallucinations" (call-workflow "gate-hallucinations"))
      (gate "syntax"         (call-workflow "gate-syntax"))
      (gate "structure"      (call-workflow "gate-structure"))
      (gate "links"          (call-workflow "gate-links"))
      (gate "sidebar"        (call-workflow "gate-sidebar"))))

  (write-file "site/generated/build-report.md" :content "PASS\nAll gates passed.")
  (step "build-site" (run "cd site && bash build.sh 2>&1"))

  (phase "build-test" :retries 0
    (gate "playwright" (run "cd site && npx playwright test --reporter=line 2>&1")))

  ;; ── Done ─────────────────────────────────────────────────────────
  (step "done"
    (str ```
      site-sync complete
        created: ~(join (map needs-create (fn (p) (pick p :slug))) ", ")
        updated: ~(join (map needs-update (fn (p) (pick p :slug))) ", ")
        orphans: ~(join orphans ", ")

      Site built to site/dist/
      Preview: cd site && npx astro preview
      ```)))
```

### 3d. site-create-page.glitch

```lisp
(include "site/shared.glitch")
(include "site-manifest.glitch")

(workflow "site-create-page"
  :description "AI-generate a new doc page with gated verification"

  ;; ── Gather context in parallel ────────────────────────
  (par
    (step "existing-pages"
      (-> (glob "site/src/content/docs/*.md")
          (map (fn (f) (replace f "site/src/content/docs/" "" ".md" "")))))

    (step "repo-structure"
      (-> (glob "internal/**/*.go" "cmd/**/*.go")
          (filter (fn (f) (not (contains f "testdata"))))))

    (step "recent-commits"
      (run "git log --oneline -30 -- cmd/ internal/ examples/ .glitch/")))

  ;; ── Generate page ─────────────────────────────────────
  (step "generate"
    (llm
      :tier 0
      :format "json"
      :prompt ```
        You are a technical writer for gl1tch (8op.org).

        CONVENTIONS — follow these exactly:
        ~conventions

        TOPIC: ~param.topic

        EXISTING PAGES (don't duplicate these):
        ~(ref "existing-pages")

        REAL WORKFLOW EXAMPLES:
        ~examples

        REPO STRUCTURE:
        ~(ref "repo-structure")

        RECENT CHANGES:
        ~(ref "recent-commits")

        OUTPUT: A single JSON object:
        {"slug": "kebab-case-name", "title": "Page Title",
         "order": N, "description": "one line",
         "content": "full markdown body starting after the title"}
        ```))

  ;; ── Save directly ─────────────────────────────────────
  (let (slug (json-pick (ref "generate") ".slug")
        title (json-pick (ref "generate") ".title")
        order (json-pick (ref "generate") ".order")
        desc (json-pick (ref "generate") ".description")
        content (json-pick (ref "generate") ".content"))

    (write-file (str "site/src/content/docs/" slug ".md")
      :content ```
        ---
        title: "~(replace title "\"" "\\\"")"
        order: ~order
        description: "~(replace desc "\"" "\\\"")"
        ---

        ~content
        ```))

  ;; ── Verify ────────────────────────────────────────────
  (phase "verify" :retries 1
    (par
      (gate "hallucinations" (call-workflow "gate-hallucinations"))
      (gate "syntax"         (call-workflow "gate-syntax"))
      (gate "structure"      (call-workflow "gate-structure"))))

  (phase "page-tests" :retries 0
    (gate "playwright" (run "cd site && npx playwright test --reporter=line 2>&1")))

  (step "done" "Page created. Run: glitch workflow run site-dev to preview."))
```

### 3e. site-update-page.glitch

```lisp
(include "site/shared.glitch")

(workflow "site-update-page"
  :description "AI-rewrite an existing doc page with gated verification"

  ;; ── Gather context in parallel ────────────────────────
  (par
    (step "current-page"
      (read-file (str "site/src/content/docs/" (trim ~param.page) ".md")))

    (step "repo-structure"
      (-> (glob "internal/**/*.go" "cmd/**/*.go")
          (filter (fn (f) (not (contains f "testdata"))))))

    (step "recent-commits"
      (run "git log --oneline -30 -- cmd/ internal/ examples/ .glitch/"))

    (step "workflows"
      (-> (glob "workflows/*.glitch")
          (map (fn (f) (str "=== " f " ===\n" (read-file f))))
          (join "\n\n"))))

  ;; ── Rewrite ───────────────────────────────────────────
  (step "rewrite"
    (llm
      :prompt ```
        You are a technical writer for gl1tch (8op.org).

        CONVENTIONS — follow these exactly:
        ~conventions

        INSTRUCTIONS: ~param.instructions

        CURRENT PAGE CONTENT:
        ~(ref "current-page")

        REAL WORKFLOW EXAMPLES:
        ~examples

        REAL WORKFLOWS:
        ~(ref "workflows")

        REPO STRUCTURE:
        ~(ref "repo-structure")

        RECENT CHANGES:
        ~(ref "recent-commits")

        OUTPUT: The complete updated markdown body.
        No frontmatter — start with ## heading.
        ```))

  ;; ── Save (preserve existing frontmatter) ──────────────
  (let (page-path (str "site/src/content/docs/" (trim ~param.page) ".md")
        existing (read-file page-path)
        frontmatter (-> existing (split "---") (pick 1)))
    (write-file page-path
      :content (str "---\n" frontmatter "\n---\n\n" (ref "rewrite"))))

  ;; ── Verify ────────────────────────────────────────────
  (phase "verify" :retries 1
    (par
      (gate "hallucinations" (call-workflow "gate-hallucinations"))
      (gate "syntax"         (call-workflow "gate-syntax"))
      (gate "structure"      (call-workflow "gate-structure"))))

  (step "done" "Page updated. Run: glitch workflow run site-dev to preview."))
```

### 3f. site-deploy.glitch

```lisp
(workflow "site-deploy"
  :description "Verify, build, commit, push for GitHub Pages"

  (def manifest (include "site-manifest.glitch"))

  (phase "verify" :retries 0
    (par
      (gate "hallucinations" (call-workflow "gate-hallucinations"))
      (gate "syntax"         (call-workflow "gate-syntax"))
      (gate "structure"      (call-workflow "gate-structure"))
      (gate "links"          (call-workflow "gate-links"))
      (gate "sidebar"        (call-workflow "gate-sidebar"))))

  (step "build-sidebar" (call-workflow "build-sidebar"))
  (step "build" (run "cd site && npx astro build 2>&1"))

  (phase "smoke" :retries 0
    (gate "playwright" (run "cd site && npx playwright test --reporter=line 2>&1")))

  (step "commit" (run ```
    git add site/src/content/ site/src/pages/ site/src/components/
    git diff --cached --quiet && echo 'nothing to commit' && exit 0
    git commit -m "docs: update site content"
    git push
    ```))

  (step "done" "Deployed. GitHub Pages will pick it up."))
```

### 3g. site-dev.glitch

```lisp
(workflow "site-dev"
  :description "Regenerate docs and start dev server"

  (when-not (run "test -f site/generated/changelog.md")
    (write-file "site/generated/changelog.md" :content "# No changelog yet"))

  (write-file "site/generated/build-report.md" :content "PASS")

  (step "dev" :hint "http://localhost:4322"
    (run "cd site && npx astro dev --port 4322")))
```

---

## 4. Gate Workflows (Pure Lisp)

Gates become standalone workflow files in `workflows/`. Each uses the new `regex-match`, `regex-find`, `assert`, and set builtins.

### 4a. gate-hallucinations.glitch

```lisp
(workflow "gate-hallucinations"
  :description "Verify no hallucinated commands, keywords, or forms in docs"

  (def valid-commands (set (list
    "glitch run" "glitch workflow run" "glitch workflow list"
    "glitch wf list" "glitch plugin list" "glitch plugin"
    "glitch observe" "glitch up" "glitch down" "glitch index"
    "glitch config show" "glitch config set" "glitch config"
    "glitch version" "glitch --help" "glitch --workspace"
    "glitch workspace init" "glitch workspace use" "glitch workspace list"
    "glitch workspace status" "glitch workspace gui"
    "glitch workspace register" "glitch workspace unregister"
    "glitch workspace add" "glitch workspace rm"
    "glitch workspace sync" "glitch workspace pin"
    "glitch workspace workflow" "glitch workspace")))

  (def valid-forms (set (list
    "def" "workflow" "step" "run" "llm" "save" "plugin" "include"
    "arg" "retry" "timeout" "catch" "cond" "map" "let" "phase" "gate" "par"
    "json-pick" "lines" "merge" "http-get" "http-post" "read-file" "write-file"
    "glob" "websearch" "fetch" "send" "write" "join" "split" "trim"
    "upper" "lower" "replace" "contains" "call-workflow" "fn" "list" "assoc"
    "pick" "->" "filter" "reduce" "flatten" "when" "when-not" "each"
    "compare" "branch" "review" "do" "begin" "str" "println" "ref"
    "or" "not" "read" "set" "regex-match" "regex-find" "some" "every"
    "starts-with" "ends-with" "slice" "difference" "count" "assert")))

  (def valid-keywords (set (list
    ":prompt" ":provider" ":model" ":skill" ":format" ":tier"
    ":description" ":from" ":default" ":type" ":dir" ":headers"
    ":body" ":retries" ":version" ":set" ":content" ":content-type"
    ":id" ":criteria" ":index" ":query" ":size" ":fields" ":sort"
    ":hint" ":to" ":sidebar" ":slug" ":title" ":order"
    ":template" ":sections" ":context-query" ":context-paths"
    ":label" ":pages" ":site" ":url" ":ndjson")))

  (def docs (glob "site/src/content/docs/*.md"))

  (map docs
    (let (content (read-file item)
          slug (replace item "site/src/content/docs/" "" ".md" "")
          blocks (regex-match "```[\\s\\S]*?```" content))
      (map blocks
        (let (block-lines (lines item))
          (map block-lines
            (do
              ;; Check CLI commands
              (when (or (starts-with item "glitch ") (starts-with item "$ glitch "))
                (let (cmd (trim (replace item "$ " "")))
                  (when-not (some valid-commands (fn (v) (starts-with cmd v)))
                    (assert false (str slug ": invalid command: " cmd)))))

              ;; Check sexpr forms
              (-> (regex-match "\\(([a-z>][-a-z_>/]*)" item)
                  (filter (fn (form) (not (contains valid-forms form))))
                  (map (fn (form) (assert false (str slug ": invalid form (" form " ...)")))))

              ;; Check keywords
              (-> (regex-match ":[a-z][-a-z_]*" item)
                  (filter (fn (kw) (not (contains valid-keywords kw))))
                  (map (fn (kw) (assert false (str slug ": invalid keyword " kw)))))))))))

  (println (str "PASS: no hallucinated commands, keywords, or forms (" (count docs) " docs)")))
```

### 4b. gate-syntax.glitch

```lisp
(workflow "gate-syntax"
  :description "Verify no stale interpolation syntax in docs"

  (def docs (glob "site/src/content/docs/*.md" "site/src/content/labs/*.md"))

  (map docs
    (let (content (read-file item)
          slug (replace item "site/src/content/docs/" "" "site/src/content/labs/" "" ".md" ""))
      (do
        (when (regex-find "\\{\\{step\\s+[\"']" content)
          (assert false (str slug ": old Go template step reference")))
        (when (regex-find "\\{\\{\\.param\\." content)
          (assert false (str slug ": old Go template param reference")))
        (when (regex-find "\\{\\{\\.input\\b" content)
          (assert false (str slug ": old Go template input reference")))
        (when (regex-find "\\bglitch ask\\b" content)
          (assert false (str slug ": decommissioned command: glitch ask"))))))

  (println (str "PASS: no stale syntax (" (count docs) " files)")))
```

### 4c. gate-structure.glitch

```lisp
(workflow "gate-structure"
  :description "Valid frontmatter, no internals leaked, your framing"

  (def internals (list "BubbleTea" "bubbletea" "tea.Model" "tea.Cmd"
                       "internal/tui" "lipgloss" "sqlite3" "SQLite"))
  (def docs (glob "site/src/content/docs/*.md"))

  (map docs
    (let (content (read-file item)
          slug (replace item "site/src/content/docs/" "" ".md" "")
          fm-block (regex-find "(?s)^---\\n(.+?)\\n---" content)
          body (-> (split content "---") (slice 2) (join "---")))
      (do
        (when-not (regex-find "title:" fm-block)
          (assert false (str slug ": missing frontmatter: title")))
        (when-not (regex-find "order:" fm-block)
          (assert false (str slug ": missing frontmatter: order")))
        (when-not (regex-find "description:" fm-block)
          (assert false (str slug ": missing frontmatter: description")))

        (map internals
          (when (contains content item)
            (assert false (str slug ": leaked internal: " item))))

        (when (< (count (trim body)) 200)
          (assert false (str slug ": suspiciously short body")))

        (when (regex-find "(?i)\\bthe user\\b" content)
          (println (str "  [warn] " slug ": found 'the user' — prefer you/your framing"))))))

  (println (str "PASS: " (count docs) " docs valid, no internals leaked")))
```

### 4d. gate-links.glitch

```lisp
(workflow "gate-links"
  :description "Internal /docs/ links resolve to manifest slugs"

  (def manifest (include "site-manifest.glitch"))
  (def valid-slugs
    (-> (pick manifest :sections)
        (map (fn (s) (map (pick s :pages) (fn (p) (pick p :slug)))))
        (flatten)
        (set)))

  (def docs (glob "site/src/content/docs/*.md"))

  (map docs
    (let (content (read-file item)
          slug (replace item "site/src/content/docs/" "" ".md" "")
          links (regex-match "\\[([^\\]]+)\\]\\(/docs/([^)#\\s]+)\\)" content))
      (map links
        (let (target (pick item 2))
          (when-not (contains valid-slugs target)
            (assert false (str slug ": broken link /docs/" target " — not in manifest")))))))

  (map (list "site/src/pages/index.astro" "site/src/layouts/Base.astro")
    (catch
      (let (content (read-file item)
            links (regex-match "href=[\"']/docs/([^\"'#\\s]+)[\"']" content))
        (map links
          (when-not (contains valid-slugs item)
            (assert false (str item ": broken link /docs/" item)))))
      nil))

  (println (str "PASS: all /docs/ links valid (" (count docs) " docs, "
                (count valid-slugs) " slugs)")))
```

### 4e. gate-sidebar.glitch

```lisp
(workflow "gate-sidebar"
  :description "Sidebar matches manifest in correct order"

  (def manifest (include "site-manifest.glitch"))
  (def sidebar-text (read-file "site/src/components/DocSidebar.astro"))

  (def manifest-sections
    (-> (filter (pick manifest :sections) (pick item :sidebar))
        (map (fn (s) (assoc
          :label (pick s :label)
          :slugs (map (pick s :pages) (fn (p) (pick p :slug))))))))

  (def sidebar-labels (regex-match "label:\\s*\"([^\"]+)\"" sidebar-text))
  (def sidebar-slugs  (regex-match "slug:\\s*\"([^\"]+)\"" sidebar-text))

  (assert (= (count manifest-sections) (count sidebar-labels))
    (str "section count mismatch: manifest " (count manifest-sections)
         " vs sidebar " (count sidebar-labels)))

  ;; Verify all manifest slugs appear in sidebar in order
  (def expected-slugs (-> manifest-sections (map (fn (s) (pick s :slugs))) (flatten)))
  (assert (= expected-slugs sidebar-slugs)
    (str "sidebar slug order mismatch:\n  expected: " (join expected-slugs ", ")
         "\n  got: " (join sidebar-slugs ", ")))

  (println (str "PASS: sidebar matches manifest ("
                (count manifest-sections) " sections, "
                (count expected-slugs) " pages)")))
```

### 4f. gate-coverage.glitch

```lisp
(workflow "gate-coverage"
  :description "Every stub in docs/site/ has a matching entry in docs.json"

  (def stubs
    (-> (glob "docs/site/*.md")
        (map (fn (f) (replace f "docs/site/" "" ".md" "")))
        (set)))

  (def docs-json (read-file "site/generated/docs.json"))
  (def slugs (-> (json-pick docs-json ".[].slug") (set)))

  (def missing (difference stubs slugs))
  (def extra (difference slugs stubs))

  (when (or missing extra)
    (when missing
      (map missing (fn (m) (println (str "  missing doc for stub: " m)))))
    (when extra
      (map extra (fn (e) (println (str "  extra doc without stub: " e)))))
    (assert false "stub/doc coverage mismatch"))

  (println (str "PASS: all " (count stubs) " stubs covered")))
```

---

## 5. Files Deleted

All Python scripts replaced by the pure-lisp workflows:

```
scripts/site-read-manifest.py    → manifest is now native evaluator data
scripts/site-diff-disk.py        → inline in site-sync.glitch
scripts/site-build-sidebar.py    → inline in site-sync.glitch (build-sidebar section)
scripts/gate-hallucinations.py   → workflows/gate-hallucinations.glitch
scripts/gate-syntax.py           → workflows/gate-syntax.glitch
scripts/gate-structure.py        → workflows/gate-structure.glitch
scripts/gate-links.py            → workflows/gate-links.glitch
scripts/gate-sidebar.py          → workflows/gate-sidebar.glitch
scripts/gate-coverage.py         → workflows/gate-coverage.glitch
scripts/save-generated-page.py   → inline json-pick + write-file
scripts/stubs-to-json.py         → inline glob + read-file + json-pick
scripts/split-docs.py            → inline map + write-file
```

`scripts/gate-playwright.sh` is absorbed into inline `run` steps (2 lines).

---

## 6. Implementation Order

1. **Evaluator builtins** — add `regex-match`, `regex-find`, `lines`, `count`, `some`, `every`, `starts-with`, `ends-with`, `slice`, `set`, `difference`, `assert`, `<` to `internal/eval/builtins.go`
2. **Text block interpolation** — wire ` ``` ` blocks through `lexQuasi` in the evaluator, add auto-dedent in the parser
3. **Manifest conversion** — rewrite `site-manifest.glitch` to deps.edn format
4. **Site workflows** — rewrite all 6 workflow files
5. **Gate workflows** — create 6 new `.glitch` gate workflows
6. **Delete Python scripts** — remove all 12 scripts listed above
7. **Update gate-hallucinations allowlists** — add new builtins/forms/keywords to the valid sets
