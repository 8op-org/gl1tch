# Site Sync: Manifest-Driven Documentation System

**Date:** 2026-04-17
**Status:** Approved

## Problem

The gl1tch site at 8op.org has 8 doc pages but the CLI has significantly more surface area. Three pages link to a non-existent `batch-comparison-runs` page. There is no sidebar navigation — users land on Getting Started and must follow bottom-of-page "Next steps" links to discover anything else. New pages require manual authoring with no verification pipeline targeting the Astro content directory.

`glitch ask` is being decommissioned and must be excluded from all documentation.

## Solution

A manifest-driven system where a single `site-manifest.glitch` file declares every page, its metadata, section grouping, and sidebar order. A `site-sync` workflow reconciles the manifest against disk, generates only what changed using Claude Sonnet, and verifies everything through gates.

---

## 1. Site Manifest

`site-manifest.glitch` lives at the repo root. It is the single source of truth for the entire site.

```scheme
(site "gl1tch"
  :url "https://8op.org"

  (page "index"
    :title "gl1tch — your AI, your terminal, your rules"
    :template "homepage"
    :sections ("hero" "features" "how-it-works" "meta" "agents")
    :context-query "top-level project description, feature highlights, install instructions"
    :context-paths ("README.md"))

  (section "Getting Started"
    (page "getting-started"
      :title "Getting Started"
      :order 1
      :description "Install, configure, run your first workflow"
      :context-query "brew install, initial setup, hello-world workflow, glitch run basics"
      :context-paths ("cmd/run.go" "cmd/root.go"))
    (page "local-models"
      :title "Local Models"
      :order 2
      :description "LM Studio, Ollama, GPU allocation, context tuning"
      :context-query "Ollama and LM Studio provider configuration, model selection, tier routing"
      :context-paths ("internal/provider/")))

  (section "Workflow Language"
    (page "workflow-syntax"
      :title "Workflow Syntax"
      :order 3
      :description "S-expression forms, steps, interpolation, LLM options"
      :context-query "sexpr parser, step types, run/llm/save forms, interpolation syntax, tier routing"
      :context-paths ("internal/sexpr/" "internal/pipeline/sexpr.go"))
    (page "dsl-reference"
      :title "DSL Reference"
      :order 4
      :description "Threading, filter, reduce, ES forms, embed, flatten"
      :context-query "threading form, filter, reduce, search/index/delete ES forms, embed, flatten, assoc, pick"
      :context-paths ("internal/pipeline/sexpr.go"))
    (page "phases-and-gates"
      :title "Phases & Gates"
      :order 5
      :description "Phase grouping, gate assertions, retries"
      :context-query "phase form, gate form, retry semantics, phase execution order"
      :context-paths ("internal/pipeline/sexpr.go" "internal/pipeline/runner.go")))

  (section "Workspaces & Resources"
    (page "workspaces"
      :title "Workspaces"
      :order 6
      :description "Workspace config, resources, defaults, discovery, nested runs"
      :context-query "workspace init/add/sync/pin/rm commands, workspace.glitch format, resource types, call-workflow"
      :context-paths ("cmd/workspace*.go" "internal/workspace/"))
    (page "providers"
      :title "Providers & Config"
      :order 7
      :description "Provider config, Ollama, LM Studio, OpenAI-compat, tiered routing, config.glitch"
      :context-query "provider protocol, config.glitch format, glitch config show/set, provider tiers, OpenAI-compatible endpoint config, api-key-env"
      :context-paths ("cmd/config.go" "internal/provider/" "internal/pipeline/tier.go")))

  (section "Running Workflows"
    (page "compare"
      :title "Compare Runs"
      :order 8
      :description "A/B testing models, strategies, review scoring"
      :context-query "compare form, branch form, review criteria, --variant flag, --compare flag"
      :context-paths ("cmd/run.go" "internal/pipeline/compare.go"))
    (page "batch-runs"
      :title "Batch Runs"
      :order 9
      :description "Run variants side by side, multi-provider comparison, fan-out"
      :context-query "--variant flag, --compare discovery pattern, variant workflow naming, batch execution, nested runs"
      :context-paths ("cmd/run.go" "internal/pipeline/runner.go")))

  (section "Code Intelligence"
    (page "code-intelligence"
      :title "Code Intelligence"
      :order 10
      :description "Index repos, query with natural language, glitch up/down"
      :context-query "glitch index command, glitch observe command, language extractors, symbol indexing, ES queries, BFS depth traversal, glitch up/down docker compose"
      :context-paths ("cmd/index.go" "cmd/observe.go" "cmd/up.go" "internal/esearch/" "internal/capability/")))

  (section "Extending"
    (page "plugins"
      :title "Plugins"
      :order 11
      :description "Plugin directories, manifests, subcommands, argument types"
      :context-query "plugin discovery, plugin manifest, subcommand definitions, argument types, call-workflow from plugins"
      :context-paths ("cmd/plugin.go" "internal/plugin/"))))
```

### Manifest rules

- Every page on the site must have an entry. No entry = page doesn't exist.
- The manifest declares metadata (title, order, description) — the LLM never writes frontmatter.
- `:context-query` is a natural language description used to query the code index for relevant symbols and code chunks.
- `:context-paths` is a fallback list of file paths (or glob patterns like `"cmd/workspace*.go"`) for raw context if the index misses something. The gather-fallback step expands globs before reading.
- Section order in the manifest = sidebar section order. Page order within a section = sidebar page order.
- The homepage (`"index"`) is a special entry with `:template "homepage"` — it uses `index.astro` not the doc layout.

---

## 2. The `site-sync` Workflow

One idempotent workflow that reconciles the manifest against disk.

### Step-by-step

```
$ glitch run site-sync

 1. read-manifest        (shell)   parse site-manifest.glitch into JSON
 2. diff-disk            (shell)   compare manifest vs site/src/content/docs/*.md
 3a. query-code-index    (search)  per-page ES query from :context-query
 3b. gather-fallback     (shell)   cat :context-paths for anything index missed
 3c. merge-context       (shell)   dedupe and combine into per-page context bundle
 4. generate-pages       (map)     Claude Sonnet, only create/update pages
 5. inject-frontmatter   (shell)   stamp frontmatter from manifest onto generated markdown
 6. write-pages          (shell)   write to site/src/content/docs/
 7. sync-homepage        (llm)     update index.astro sections if manifest changed
 8. build-sidebar        (shell)   generate DocSidebar.astro from manifest sections
 9. phase "verify" :retries 1
    - gate hallucinations
    - gate structure
    - gate links
    - gate sidebar
10. phase "build-test" :retries 0
    - gate playwright
11. done                 (shell)   summary of what changed
```

### Step 1: read-manifest

Shell step parses `site-manifest.glitch` and outputs a JSON representation of all pages, sections, and metadata. This JSON drives every subsequent step.

### Step 2: diff-disk

Shell step compares the manifest JSON against existing files in `site/src/content/docs/`. Produces a JSON diff:

- **`create`**: pages in manifest but no `.md` file on disk
- **`update`**: pages where upstream source code has changed since last generation (checked via `git diff` on `:context-paths` since last commit that touched the doc file)
- **`unchanged`**: pages that exist and have no upstream changes
- **`orphaned`**: `.md` files on disk not in the manifest (warned, never deleted)

If `create` and `update` are both empty, steps 3-7 are skipped entirely. Gates still run.

### Step 3: Context gathering (code index + fallback)

For each page in the create/update list:

**3a. query-code-index** — A `(search ...)` step queries the `glitch-code-*` Elasticsearch index using the page's `:context-query`. Pulls:
- Public function signatures in relevant packages
- Flag definitions (what users see in `--help`)
- Cross-package call edges (how features connect)
- Struct types that appear in user-facing output

This gives the LLM structured symbol data instead of raw file dumps — better signal, smaller prompts, cheaper runs.

**3b. gather-fallback** — Shell step reads `:context-paths` files directly. Catches anything not yet indexed or too new for the code graph.

**3c. merge-context** — Shell step dedupes and combines index results + raw files into a single context bundle per page.

### Step 4: generate-pages

`(map ...)` over the create/update list. Each page hits Claude Sonnet with:

- The page's manifest entry (title, description, section, order)
- The merged context bundle from step 3
- The existing page content (for updates — prompt says "revise given upstream changes," not "rewrite")
- The full manifest (so it knows all pages and can cross-link correctly)
- Voice rules: "your" framing, examples before explanation, no internals, no `glitch ask`

**Provider:** Claude Sonnet explicitly (`:provider "claude" :model "sonnet"`). Not tiered — docs quality needs a capable model every time.

**Output:** Raw markdown body only. No frontmatter, no title heading (both injected by step 5).

### Step 5: inject-frontmatter

Shell step prepends YAML frontmatter to each generated markdown body using manifest metadata:

```yaml
---
title: "Batch Runs"
order: 9
description: "Run variants side by side, multi-provider comparison, fan-out"
---
```

The LLM never controls frontmatter. This guarantees manifest = disk.

### Step 6: write-pages

Shell step writes each file to `site/src/content/docs/<slug>.md`. Overwrites only the files in the create/update list.

### Step 7: sync-homepage

Separate LLM step that updates `site/src/pages/index.astro` if the manifest's homepage entry has changed. Receives:
- Current `index.astro` content
- The manifest's homepage `:sections` declaration
- The full page inventory (so feature cards match what docs cover)

Outputs only the changed sections of the Astro file — not a full rewrite. If nothing changed (compared via hash of manifest homepage entry), this step is a no-op.

### Step 8: build-sidebar

Pure shell step. Reads manifest sections and pages, generates `site/src/components/DocSidebar.astro`:

```astro
---
const { currentSlug } = Astro.props;
const sections = [
  { label: "Getting Started", pages: [
    { slug: "getting-started", title: "Getting Started" },
    { slug: "local-models", title: "Local Models" },
  ]},
  { label: "Workflow Language", pages: [
    { slug: "workflow-syntax", title: "Workflow Syntax" },
    { slug: "dsl-reference", title: "DSL Reference" },
    { slug: "phases-and-gates", title: "Phases & Gates" },
  ]},
  // ... all sections from manifest
];
---
<nav class="doc-sidebar">
  {sections.map(s => (
    <div class="sidebar-section">
      <div class="sidebar-label">{s.label}</div>
      {s.pages.map(p => (
        <a href={`/docs/${p.slug}`}
           class:list={[{ active: currentSlug === p.slug }]}>
          {p.title}
        </a>
      ))}
    </div>
  ))}
</nav>
```

No LLM needed — this is deterministic from the manifest.

### Step 9: verify phase

Gates run against `site/src/content/docs/*.md` directly (not intermediate JSON):

1. **gate-hallucinations** — scans code blocks in all doc files for invalid CLI commands, sexpr keywords, and form names. `glitch ask` and `glitch batch` are explicitly NOT in the allow list.

2. **gate-structure** — checks each page for: frontmatter fields match manifest, no leaked internals (BubbleTea, SQLite, tmux, tea.Model, lipgloss), "your" framing (no "the user"), content length > 200 chars.

3. **gate-links** — parses every `[text](/docs/slug)` link across all doc pages, verifies every target slug exists in the manifest. Also checks homepage and footer links.

4. **gate-sidebar** — verifies `DocSidebar.astro` matches manifest exactly: every section, every page, correct order, no extras, no missing.

Phase retries once on failure. On retry, all gates re-run. If failures are caused by generated content (hallucinations, structure), the workflow must be re-run manually after fixing the manifest or prompt — automatic page regeneration within the verify phase is not supported. Gates are verification, not correction.

### Step 10: build-test phase

1. **gate-playwright** — runs `astro build`, then Playwright tests against the built site. Checks pages render, navigation works, sidebar links resolve, no 404s.

No retries — build failures need manual intervention.

---

## 3. Navigation Redesign

### Layout

Three-column layout on doc pages:

```
[sidebar]  [content]  [toc]
```

- **Left sidebar**: section-grouped page navigation, generated from manifest
- **Center content**: the doc page
- **Right TOC**: existing "On this page" heading navigation (unchanged)

### Responsive behavior

| Viewport | Layout |
|----------|--------|
| > 1200px | sidebar + content + toc |
| 900-1200px | sidebar + content (toc hidden) |
| < 900px | content only, hamburger toggle for sidebar |

### Sidebar behavior

- Section headers are non-clickable labels (dimmer, smaller text)
- Page links highlight the current page with the teal accent (`#7dcfff`)
- Sidebar is sticky (same as current TOC)
- On mobile: `<details>` element or CSS checkbox hack — no JS framework

### Doc.astro changes

`Doc.astro` imports `DocSidebar.astro` and passes `currentSlug` as a prop. The existing TOC stays on the right side.

### "Next steps" links

Bottom-of-page "Next steps" links remain. They serve a different purpose than the sidebar — guided reading order vs. random access. They are authored in the page content, not generated from the manifest.

---

## 4. Page Inventory

### Existing pages (8) — keep, update as needed

| Slug | Section | Order |
|------|---------|-------|
| `getting-started` | Getting Started | 1 |
| `local-models` | Getting Started | 2 |
| `workflow-syntax` | Workflow Language | 3 |
| `dsl-reference` | Workflow Language | 4 |
| `phases-and-gates` | Workflow Language | 5 |
| `workspaces` | Workspaces & Resources | 6 |
| `compare` | Running Workflows | 8 |
| `plugins` | Extending | 11 |

### New pages (3) — generate

| Slug | Section | Order | Covers |
|------|---------|-------|--------|
| `providers` | Workspaces & Resources | 7 | Provider config, Ollama, LM Studio, OpenAI-compat, tiered routing, `config.glitch` format, `glitch config show/set` |
| `batch-runs` | Running Workflows | 9 | `--variant`, `--compare` discovery, multi-provider fan-out, nested runs (fixes 3 broken links to `batch-comparison-runs`) |
| `code-intelligence` | Code Intelligence | 10 | `glitch index`, `glitch observe`, language configs, `--depth` traversal, `glitch up`/`glitch down` for ES/Kibana |

### Excluded

- `glitch ask` — decommissioned, no docs, no references
- GUI — not shipping yet, will get its own page when ready

---

## 5. Idempotency Guarantees

| Scenario | Behavior |
|----------|----------|
| No manifest changes | Step 2 produces empty diff, steps 3-7 skipped, gates still run |
| Add a page to manifest | Only that page generated, existing pages untouched |
| Update context-paths for a page | Page flagged for update, receives existing content + new context |
| Remove a page from manifest | Orphan warning printed, file NOT deleted |
| Upstream code changes | Pages whose `:context-paths` have git changes since last doc commit are flagged for update |
| Homepage sections unchanged | Step 7 is a no-op (hash comparison) |
| Run twice consecutively | Second run: empty diff, no LLM calls, gates pass |

---

## 6. Gate Script Updates

All gate scripts must be rewritten to:

1. Target `site/src/content/docs/*.md` directly (not `site/generated/docs.json`)
2. Parse YAML frontmatter + markdown content from each file
3. Read the manifest for validation (valid slugs, expected metadata)
4. Remove `glitch ask` and `glitch batch` from valid command lists
5. Add `glitch config show`, `glitch config set`, `glitch workspace gui` to valid command lists

New gate scripts:
- `scripts/gate-links.py` — cross-page link validation against manifest
- `scripts/gate-sidebar.py` — sidebar component matches manifest structure

---

## 7. File Changes Summary

### New files

| File | Purpose |
|------|---------|
| `site-manifest.glitch` | Source of truth for all site pages |
| `.glitch/workflows/site-sync.glitch` | The idempotent sync workflow |
| `site/src/components/DocSidebar.astro` | Generated left sidebar navigation |
| `site/src/content/docs/providers.md` | New doc page |
| `site/src/content/docs/batch-runs.md` | New doc page (fixes broken links) |
| `site/src/content/docs/code-intelligence.md` | New doc page |
| `scripts/gate-links.py` | New gate: cross-page link validation |
| `scripts/gate-sidebar.py` | New gate: sidebar matches manifest |

### Modified files

| File | Change |
|------|--------|
| `site/src/layouts/Doc.astro` | Import DocSidebar, three-column layout |
| `site/src/styles/global.css` | Sidebar styles, responsive breakpoints |
| `scripts/gate-hallucinations.py` | Target .md files, update command allow list |
| `scripts/gate-structure.py` | Target .md files, validate against manifest |
| `scripts/gate-coverage.py` | Replace with gate-links.py (coverage is now link-based) |
| Existing doc pages | Update "Next steps" links (replace `batch-comparison-runs` with `batch-runs`) |

### Unchanged

| File | Reason |
|------|--------|
| `site/astro.config.mjs` | No changes needed |
| `site/src/content.config.ts` | Schema already correct |
| `site/src/pages/docs/[...slug].astro` | Dynamic routing already works |
| `.glitch/workflows/site-create-page.glitch` | Kept for single-page additions |
