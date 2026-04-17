# Site Consolidation + Labs Design

**Date**: 2026-04-17
**Status**: Draft

## Problem

Three overlapping site workflows exist (`site-update`, `site-publish`, `site-sync`) with changelog generation split across two of them. There's no single command that produces a complete, live website. Additionally, the site has no showcase content demonstrating glitch on real-world problems — the kind of content that sells dev architects on adopting the tool.

## Goals

1. **One workflow** (`site-sync`) generates everything: docs, changelog, homepage, sidebar, verification, build. Delete or deprecate the others.
2. **Labs section** at `/labs` — on-demand generated case studies showing glitch analyzing real OSS issues with tier routing. Not part of every site build.

---

## Feature 1: Consolidate into site-sync

### What changes

`site-sync.glitch` becomes the single entry point. It already handles docs, homepage, sidebar, and verification. Add:

- **Changelog generation** (currently only in `site-update` and `site-publish`): shell step for `git log`, Copilot LLM step to summarize, `save` step to write `site/generated/changelog.md`
- **`build.sh` call** at the end (currently manual): add a `build-site` step that runs `bash site/build.sh` after verification passes
- **Labs injection** (if generated files exist): `build.sh` gains a conditional block that stamps `site/generated/labs/*.json` into `src/content/labs/` when present

### What gets deprecated

- `site-update.glitch` — delete. All functionality absorbed into `site-sync`.
- `site-publish.glitch` — delete. Was a simpler pipeline that `site-sync` now supersedes.

### Updated site-sync flow

```
read-manifest
  → diff-disk
  → diff-summary
  → query-code-index / gather-fallback / merge-context  (parallel where possible)
  → map pages-needing-work → generate-page
  → inject-frontmatter
  → write-pages
  → sync-homepage
  → build-sidebar
  → changelog-raw           ← NEW
  → enrich-changelog         ← NEW (copilot)
  → save-changelog           ← NEW
  → inject-labs              ← NEW (conditional — stamps labs JSON if present)
  → phase verify (hallucinations, structure, links, sidebar)
  → phase build-test (playwright)
  → done
```

### Changelog steps (new)

```scheme
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
      ~(step changelog-raw)
      ```))

(step "save-changelog"
  (save "site/generated/changelog.md" :from "enrich-changelog"))
```

### Labs injection step (new, conditional)

```scheme
(step "inject-labs"
  (run ```
    python3 - <<'EOF'
    import json, os
    from pathlib import Path

    labs_dir = Path('site/generated/labs')
    out_dir  = Path('site/src/content/labs')

    if not labs_dir.exists() or not list(labs_dir.glob('*.json')):
        print('SKIP: no labs to inject')
        import sys; sys.exit(0)

    out_dir.mkdir(parents=True, exist_ok=True)
    for f in labs_dir.glob('*.json'):
        lab = json.loads(f.read_text())
        slug  = lab['slug']
        title = lab['title']
        desc  = lab.get('description', '')
        date  = lab.get('date', '')

        frontmatter = f'---\ntitle: "{title}"\nslug: "{slug}"\ndescription: "{desc}"\ndate: "{date}"\n---\n\n'
        content = frontmatter + lab['content']
        (out_dir / f'{slug}.md').write_text(content)
        print(f'  injected: {slug}.md')
    EOF
    ```))
```

### build.sh changes

None needed — `build.sh` already handles changelog stamping. The new changelog steps in `site-sync` produce the same `site/generated/changelog.md` that `build.sh` reads. Labs injection happens before `build.sh` runs, so Astro picks up the content collection automatically.

---

## Feature 2: Labs

### Content collection

New Astro content collection alongside `docs` and `changelog`:

```typescript
// in site/src/content.config.ts
const labs = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/labs' }),
  schema: z.object({
    title: z.string(),
    slug: z.string(),
    description: z.string(),
    date: z.string(),
  }),
});

export const collections = { docs, changelog, labs };
```

### Pages

- `/labs` — index page listing all labs with title, description, date
- `/labs/[slug]` — detail page using a new `Lab.astro` layout

### Lab.astro layout

Standalone layout (not Doc.astro — no sidebar). Uses Base.astro for nav/footer/hex-rain consistency. Wide content area optimized for:

- Full-width code blocks showing real workflow input/output
- Model comparison tables (model, tokens, cost, wall time, quality verdict)
- Tier escalation flow diagrams (text-based, no external deps)
- Cost breakdown badges

### labs-generate.glitch workflow

Run on-demand: `glitch workflow run labs-generate`

This workflow produces `site/generated/labs/*.json`. Each JSON file:

```json
{
  "slug": "issue-triage-kubernetes",
  "title": "Triaging a Kubernetes Issue with Tier Routing",
  "description": "Free model attempts, paid model nails it — real input, real output",
  "date": "2026-04-17",
  "content": "<full markdown body>"
}
```

### Workflow structure

```
labs-generate.glitch

  ;; ── Lab 1: Issue Triage ─────────────────────────
  ;; Pick a real open Kubernetes issue via GitHub API

  (step "fetch-k8s-issue"
    (run "gh issue view <issue-number> -R kubernetes/kubernetes --json title,body,comments,labels"))

  (step "triage-free"
    (llm
      :provider "openrouter"
      :model "google/gemma-3-12b-it:free"
      :prompt "Triage this issue... ~(step fetch-k8s-issue)"))

  (step "triage-paid"
    (llm
      :provider "openrouter"
      :model "qwen/qwen-2.5-72b-instruct"
      :prompt "Triage this issue... ~(step fetch-k8s-issue)"))

  (step "triage-copilot"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt "Triage this issue... ~(step fetch-k8s-issue)"))

  ;; ── Lab 2: PR Review ───────────────────────────
  ;; Pick a real merged Go stdlib or Prometheus PR

  (step "fetch-pr"
    (run "gh pr view <pr-number> -R prometheus/prometheus --json title,body,files,comments"))

  (step "fetch-pr-diff"
    (run "gh pr diff <pr-number> -R prometheus/prometheus"))

  (step "review-copilot"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt "Review this PR... ~(step fetch-pr) DIFF: ~(step fetch-pr-diff)"))

  ;; ── Lab 3: Model Showdown ──────────────────────
  ;; Same issue through all 3 tiers, structured comparison

  (step "fetch-showdown-issue"
    (run "gh issue view <issue-number> -R containerd/containerd --json title,body,comments,labels"))

  (step "analysis-free"
    (llm
      :provider "openrouter"
      :model "google/gemma-3-12b-it:free"
      :prompt "Analyze this issue... ~(step fetch-showdown-issue)"))

  (step "analysis-paid"
    (llm
      :provider "openrouter"
      :model "qwen/qwen-2.5-72b-instruct"
      :prompt "Analyze this issue... ~(step fetch-showdown-issue)"))

  (step "analysis-copilot"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt "Analyze this issue... ~(step fetch-showdown-issue)"))

  ;; ── Assemble narratives ────────────────────────
  ;; Copilot writes the lab narrative around the raw artifacts

  (step "write-lab-1"
    (llm
      :provider "copilot"
      :model "sonnet"
      :format "json"
      :prompt "...assemble triage lab narrative with real outputs from all 3 tiers..."))

  (step "write-lab-2"
    (llm
      :provider "copilot"
      :model "sonnet"
      :format "json"
      :prompt "...assemble PR review lab narrative..."))

  (step "write-lab-3"
    (llm
      :provider "copilot"
      :model "sonnet"
      :format "json"
      :prompt "...assemble model showdown lab narrative with comparison table..."))

  (save "site/generated/labs/issue-triage-kubernetes.json" :from "write-lab-1")
  (save "site/generated/labs/pr-review-prometheus.json" :from "write-lab-2")
  (save "site/generated/labs/model-showdown-containerd.json" :from "write-lab-3")
```

### Issue/PR selection

The workflow uses hardcoded issue/PR numbers for reproducibility. Pick issues that are:
- **Public and well-known** — Kubernetes, Prometheus, containerd, Kibana
- **Closed/merged** — stable, won't change
- **Meaty enough** to show real analysis — not typo fixes
- **No private names/domains** (per memory: never use names from ../farm)

Concrete issues/PRs (hardcoded in `labs-generate.glitch`):
- **Lab 1 (triage)**: `kubernetes/kubernetes#138431` — kubelet cgroup reset bug, 6 comments, priority/critical-urgent
- **Lab 2 (PR review)**: `prometheus/prometheus#18499` — consul health_filter feature, 263 lines changed, 3 files
- **Lab 3 (showdown)**: `containerd/containerd#11339` — CRI UpdatePodSandboxResources, 10 comments, design discussion
- **Lab 4 (bug triage)**: `elastic/kibana#263137` — Discover regex bug where time-series data streams with "logs" in name trigger wrong profile

### Lab narrative structure

Each lab markdown follows this template:

```markdown
## The Scenario

What we're looking at and why it matters.

## The Workflow

The actual .glitch workflow used (verbatim).

## Free Tier: [model name]

> Raw output block

**Verdict**: What the free model got right/wrong.

## Paid Tier: [model name]

> Raw output block

**Verdict**: Where paid improved.

## Copilot (Sonnet)

> Raw output block

**Verdict**: Final quality.

## Comparison

| Model | Tokens | Cost | Wall Time | Quality |
|-------|--------|------|-----------|---------|
| ...   | ...    | ...  | ...       | ...     |

## Takeaway

What this demonstrates about tier routing and when to use each tier.
```

### Provider configuration

| Tier | Provider | Model | Notes |
|------|----------|-------|-------|
| Free | OpenRouter | `google/gemma-4-31b-it:free` | Newest Google free model, 262k context |
| Paid | OpenRouter | `qwen/qwen3.5-flash-02-23` | $0.07/M prompt, 1M context, best cost/quality |
| Top | Copilot | Sonnet | Real Copilot output with token/cost parsing |

### Site manifest update

Add labs to `site-manifest.glitch` as a non-sidebar section (standalone route):

```scheme
;; Labs are NOT in the sidebar — they live at /labs with their own layout.
;; Listed here for link validation and homepage cross-linking only.
(section "Labs" :sidebar false
  (page "labs-index"
    :title "Labs"
    :template "labs-index")
  (page "issue-triage-kubernetes"
    :title "Triaging a Kubernetes Issue with Tier Routing"
    :template "lab")
  (page "pr-review-prometheus"
    :title "Reviewing a Prometheus PR with Copilot"
    :template "lab")
  (page "model-showdown-containerd"
    :title "Model Showdown: Free vs Paid vs Copilot"
    :template "lab")
  (page "bug-triage-kibana"
    :title "Triaging a Kibana Regex Bug Across Three Tiers"
    :template "lab"))
```

### Nav update

Add "labs" link to `Base.astro` nav:

```html
<a href="/docs/getting-started">docs</a>
<a href="/labs">labs</a>
<a href="/changelog">changelog</a>
```

---

## Files changed

| File | Change |
|------|--------|
| `.glitch/workflows/site-sync.glitch` | Add changelog + labs injection steps, add build-site step |
| `.glitch/workflows/site-update.glitch` | Delete |
| `.glitch/workflows/site-publish.glitch` | Delete |
| `.glitch/workflows/labs-generate.glitch` | New — on-demand lab generation |
| `site-manifest.glitch` | Add Labs section |
| `site/src/content.config.ts` | Add `labs` collection |
| `site/src/pages/labs/index.astro` | New — labs index page |
| `site/src/pages/labs/[slug].astro` | New — lab detail page |
| `site/src/layouts/Lab.astro` | New — wide layout for lab content |
| `site/src/layouts/Base.astro` | Add "labs" nav link |
| `site/build.sh` | No changes needed |

## Out of scope

- Automated lab refresh on a schedule (future: could be a cron workflow)
- Lab content editing UI in the GUI
- SEO/OG metadata for lab pages (can add later)
- Lab-specific Playwright tests (covered by existing site-wide tests)

## Cost estimate

- `site-sync` run: same as today + 1 Copilot call for changelog (~negligible)
- `labs-generate` run: ~12 LLM calls (3 per lab × 4 labs for tiers) + 4 narrative assembly calls = 16 total. At ~$0.01-0.05 per OpenRouter call + Copilot premium requests, roughly $0.75-1.50 per full labs generation.
