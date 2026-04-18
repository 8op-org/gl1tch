# Site Consolidation + Labs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate site generation into one workflow (`site-sync`) and add an on-demand `/labs` section showcasing glitch on real OSS issues with tier routing.

**Architecture:** `site-sync.glitch` becomes the single workflow that generates docs, changelog, and builds the site. A separate `labs-generate.glitch` runs on-demand, outputs JSON to `site/generated/labs/`, which `site-sync` picks up if present. Labs get their own Astro content collection, layout, and `/labs` route — separate from docs.

**Tech Stack:** Astro content collections, glitch s-expression workflows, OpenRouter API (free + paid), GitHub Copilot (Sonnet), Python scripts for injection/stamping.

**Spec:** `docs/superpowers/specs/2026-04-17-site-consolidation-and-labs-design.md`

---

### Task 1: Add labs content collection to Astro

**Files:**
- Modify: `site/src/content.config.ts`
- Create: `site/src/content/labs/.gitkeep`

- [ ] **Step 1: Add labs collection schema**

In `site/src/content.config.ts`, add after the `changelog` collection:

```typescript
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

The full file should be:

```typescript
import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const docs = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/docs' }),
  schema: z.object({
    title: z.string(),
    order: z.number(),
    description: z.string(),
  }),
});

const changelog = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/changelog' }),
  schema: z.object({
    title: z.string(),
    date: z.string(),
  }),
});

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

- [ ] **Step 2: Create the labs content directory**

```bash
mkdir -p site/src/content/labs
touch site/src/content/labs/.gitkeep
```

- [ ] **Step 3: Verify Astro recognizes the collection**

```bash
cd site && npx astro check 2>&1 | head -20
```

Expected: no errors about the labs collection (warnings about missing content are fine — it's empty).

- [ ] **Step 4: Commit**

```bash
git add site/src/content.config.ts site/src/content/labs/.gitkeep
git commit -m "feat(site): add labs content collection"
```

---

### Task 2: Create Lab.astro layout

**Files:**
- Create: `site/src/layouts/Lab.astro`

- [ ] **Step 1: Create the layout**

`site/src/layouts/Lab.astro` — wide layout for lab content, no sidebar. Uses Base.astro for consistency (nav, footer, hex-rain). Optimized for full-width code blocks and comparison tables.

```astro
---
import Base from './Base.astro';

interface Props {
  title: string;
  description?: string;
  date?: string;
}

const { title, description, date } = Astro.props;
---

<Base title={`${title} — gl1tch labs`} description={description}>
  <main>
    <section class="lab-page">
      <div class="lab-header">
        <a href="/labs" class="lab-back">&larr; labs</a>
        <h1>{title}</h1>
        {description && <p class="lab-desc">{description}</p>}
        {date && <time class="lab-date">{date}</time>}
      </div>
      <div class="lab-content">
        <slot />
      </div>
    </section>
  </main>
</Base>

<style>
  .lab-page {
    padding-top: 140px;
    max-width: 960px;
    margin: 0 auto;
    padding-left: 1.5rem;
    padding-right: 1.5rem;
    padding-bottom: 6rem;
  }

  .lab-back {
    color: var(--dim);
    text-decoration: none;
    font-size: 0.85rem;
    display: inline-block;
    margin-bottom: 1rem;
  }
  .lab-back:hover { color: var(--teal); }

  .lab-header {
    margin-bottom: 3rem;
    border-bottom: 1px solid var(--border);
    padding-bottom: 2rem;
  }

  .lab-header h1 {
    font-size: 1.8rem;
    color: var(--teal);
    margin-bottom: 0.5rem;
  }

  .lab-desc {
    color: var(--dim);
    font-size: 0.95rem;
    margin-bottom: 0.5rem;
  }

  .lab-date {
    color: var(--dim);
    font-size: 0.8rem;
  }

  .lab-content {
    line-height: 1.8;
  }

  /* Wide code blocks for lab output */
  .lab-content :global(pre) {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 1.2rem;
    overflow-x: auto;
    margin: 1.5rem 0;
    font-size: 0.85rem;
  }

  .lab-content :global(blockquote) {
    border-left: 3px solid var(--teal);
    padding-left: 1rem;
    margin: 1.5rem 0;
    color: var(--dim);
  }

  /* Comparison tables */
  .lab-content :global(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 1.5rem 0;
    font-size: 0.85rem;
  }

  .lab-content :global(th) {
    text-align: left;
    padding: 0.6rem 0.8rem;
    border-bottom: 2px solid var(--border);
    color: var(--teal);
    font-weight: 500;
  }

  .lab-content :global(td) {
    padding: 0.6rem 0.8rem;
    border-bottom: 1px solid var(--border);
  }

  .lab-content :global(tr:hover td) {
    background: var(--glow);
  }

  .lab-content :global(h2) {
    color: var(--teal);
    font-size: 1.3rem;
    margin-top: 2.5rem;
    margin-bottom: 1rem;
  }

  .lab-content :global(h3) {
    color: var(--fg);
    font-size: 1.1rem;
    margin-top: 2rem;
    margin-bottom: 0.75rem;
  }

  .lab-content :global(strong) {
    color: var(--yellow);
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add site/src/layouts/Lab.astro
git commit -m "feat(site): add Lab.astro layout for labs pages"
```

---

### Task 3: Create labs pages (index + detail)

**Files:**
- Create: `site/src/pages/labs/index.astro`
- Create: `site/src/pages/labs/[slug].astro`

- [ ] **Step 1: Create labs index page**

`site/src/pages/labs/index.astro`:

```astro
---
import { getCollection } from 'astro:content';
import Base from '../../layouts/Base.astro';

const labs = (await getCollection('labs')).sort(
  (a, b) => new Date(b.data.date).getTime() - new Date(a.data.date).getTime()
);
---

<Base title="Labs — gl1tch">
  <main>
    <section class="labs-index">
      <h1>Labs</h1>
      <p class="labs-intro">Real workflows, real issues, real output. Each lab runs gl1tch against a public GitHub issue or PR and shows you exactly what comes back — from free models to paid.</p>

      {labs.length === 0 && (
        <p class="labs-empty">No labs yet. Run <code>glitch workflow run labs-generate</code> to create them.</p>
      )}

      <div class="labs-grid">
        {labs.map((lab) => (
          <a href={`/labs/${lab.data.slug}`} class="lab-card">
            <h2>{lab.data.title}</h2>
            <p>{lab.data.description}</p>
            <time>{lab.data.date}</time>
          </a>
        ))}
      </div>
    </section>
  </main>
</Base>

<style>
  .labs-index {
    padding-top: 140px;
    max-width: 900px;
    margin: 0 auto;
    padding-left: 1.5rem;
    padding-right: 1.5rem;
    padding-bottom: 6rem;
  }

  .labs-index h1 {
    font-size: 2rem;
    color: var(--teal);
    margin-bottom: 0.75rem;
  }

  .labs-intro {
    color: var(--dim);
    font-size: 0.95rem;
    margin-bottom: 2.5rem;
    max-width: 600px;
  }

  .labs-empty {
    color: var(--dim);
    font-style: italic;
  }

  .labs-empty code {
    color: var(--teal);
    font-size: 0.85rem;
  }

  .labs-grid {
    display: grid;
    gap: 1.2rem;
  }

  .lab-card {
    display: block;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    text-decoration: none;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  .lab-card:hover {
    border-color: var(--teal);
    box-shadow: 0 0 20px var(--glow);
  }

  .lab-card h2 {
    font-size: 1.1rem;
    color: var(--fg);
    margin-bottom: 0.5rem;
  }

  .lab-card p {
    color: var(--dim);
    font-size: 0.85rem;
    margin-bottom: 0.5rem;
  }

  .lab-card time {
    color: var(--dim);
    font-size: 0.75rem;
  }
</style>
```

- [ ] **Step 2: Create labs detail page**

`site/src/pages/labs/[slug].astro`:

```astro
---
import { getCollection, render } from 'astro:content';
import Lab from '../../layouts/Lab.astro';

export async function getStaticPaths() {
  const labs = await getCollection('labs');
  return labs.map((entry) => ({
    params: { slug: entry.data.slug },
    props: { entry },
  }));
}

const { entry } = Astro.props;
const { Content, headings } = await render(entry);
---

<Lab title={entry.data.title} description={entry.data.description} date={entry.data.date}>
  <Content />
</Lab>
```

- [ ] **Step 3: Commit**

```bash
git add site/src/pages/labs/
git commit -m "feat(site): add labs index and detail pages"
```

---

### Task 4: Add labs nav link

**Files:**
- Modify: `site/src/layouts/Base.astro`

- [ ] **Step 1: Add labs link to nav**

In `site/src/layouts/Base.astro`, add the labs link between docs and changelog in the nav:

Change:
```html
<a href="/docs/getting-started">docs</a>
<a href="/changelog">changelog</a>
```

To:
```html
<a href="/docs/getting-started">docs</a>
<a href="/labs">labs</a>
<a href="/changelog">changelog</a>
```

- [ ] **Step 2: Add labs link to footer**

In the same file, add labs to footer links. Change:

```html
<a href="/docs/getting-started">Docs</a>
<a href="/changelog">Changelog</a>
```

To:
```html
<a href="/docs/getting-started">Docs</a>
<a href="/labs">Labs</a>
<a href="/changelog">Changelog</a>
```

- [ ] **Step 3: Commit**

```bash
git add site/src/layouts/Base.astro
git commit -m "feat(site): add labs link to nav and footer"
```

---

### Task 5: Add changelog + labs injection to site-sync

**Files:**
- Modify: `.glitch/workflows/site-sync.glitch`

- [ ] **Step 1: Add changelog steps to site-sync**

Insert these steps after the `build-sidebar` step and before the `verify` phase in `.glitch/workflows/site-sync.glitch`:

```scheme
  ;; ── Changelog ─────────────────────────────────────

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

- [ ] **Step 2: Add labs injection step**

Insert after `save-changelog`, before the `verify` phase:

```scheme
  ;; ── Labs injection (conditional) ──────────────────

  (step "inject-labs"
    (run ```
      python3 - <<'EOF'
      import json, os, sys
      from pathlib import Path

      labs_dir = Path('site/generated/labs')
      out_dir  = Path('site/src/content/labs')

      if not labs_dir.exists() or not list(labs_dir.glob('*.json')):
          print('SKIP: no labs to inject')
          sys.exit(0)

      out_dir.mkdir(parents=True, exist_ok=True)

      # Clear old injected labs
      for old in out_dir.glob('*.md'):
          if old.name != '.gitkeep':
              old.unlink()

      for f in labs_dir.glob('*.json'):
          lab = json.loads(f.read_text())
          slug  = lab['slug']
          title = lab['title'].replace('"', '\\"')
          desc  = lab.get('description', '').replace('"', '\\"')
          date  = lab.get('date', '')

          frontmatter = f'---\ntitle: "{title}"\nslug: "{slug}"\ndescription: "{desc}"\ndate: "{date}"\n---\n\n'
          content = frontmatter + lab['content']
          (out_dir / f'{slug}.md').write_text(content)
          print(f'  injected: {slug}.md')

      print(f'Injected {len(list(labs_dir.glob("*.json")))} lab(s)')
      EOF
      ```))
```

- [ ] **Step 3: Add build-site step after verification passes**

Replace the existing `done` step at the end of site-sync with:

```scheme
  ;; ── Build ─────────────────────────────────────────

  (step "build-site"
    (run "cd site && bash build.sh 2>&1"))

  (step "done"
    (run ```
      python3 - <<'EOF'
      import json
      from pathlib import Path

      diff = json.loads(open('~(stepfile diff-disk)').read())
      created = [p['slug'] for p in diff.get('create', [])]
      updated = [p['slug'] for p in diff.get('update', [])]
      ok      = diff.get('ok', [])
      orphan  = diff.get('orphan', [])
      labs    = list(Path('site/src/content/labs').glob('*.md'))
      labs    = [l.stem for l in labs if l.name != '.gitkeep']

      print('site-sync complete')
      print(f'  created : {len(created)} page(s)  {created}')
      print(f'  updated : {len(updated)} page(s)  {updated}')
      print(f'  ok      : {len(ok)} page(s) already in sync')
      print(f'  labs    : {len(labs)} lab(s)  {labs}')
      if orphan:
          print(f'  orphan  : {orphan} — on disk but not in manifest')
      print()
      print('Site built to site/dist/')
      print('Preview: cd site && npx astro preview')
      EOF
      ```))
```

- [ ] **Step 4: Commit**

```bash
git add .glitch/workflows/site-sync.glitch
git commit -m "feat(site): add changelog, labs injection, and build to site-sync"
```

---

### Task 6: Delete redundant workflows

**Files:**
- Delete: `.glitch/workflows/site-update.glitch`
- Delete: `.glitch/workflows/site-publish.glitch`

- [ ] **Step 1: Delete site-update.glitch**

```bash
rm .glitch/workflows/site-update.glitch
```

- [ ] **Step 2: Delete site-publish.glitch**

```bash
rm .glitch/workflows/site-publish.glitch
```

- [ ] **Step 3: Commit**

```bash
git add -u .glitch/workflows/site-update.glitch .glitch/workflows/site-publish.glitch
git commit -m "chore: remove redundant site-update and site-publish workflows"
```

---

### Task 7: Create labs-generate.glitch workflow

**Files:**
- Create: `.glitch/workflows/labs-generate.glitch`

This is the on-demand workflow that generates lab content from real OSS issues.

**Issues/PRs selected:**
- Lab 1 (triage): `kubernetes/kubernetes#138431` — kubelet cgroup reset bug, 6 comments, priority/critical-urgent
- Lab 2 (PR review): `prometheus/prometheus#18499` — consul health_filter feature, 263 lines changed, 3 files
- Lab 3 (showdown): `containerd/containerd#11339` — CRI UpdatePodSandboxResources, 10 comments, design discussion

**Models:**
- Free: `google/gemma-4-31b-it:free` (OpenRouter)
- Paid: `qwen/qwen3.5-flash-02-23` (OpenRouter, $0.07/M)
- Top: Copilot Sonnet

- [ ] **Step 1: Create the workflow**

`.glitch/workflows/labs-generate.glitch`:

```scheme
;; labs-generate.glitch — generate lab case studies from real OSS issues
;;
;; On-demand only. Costs real tokens (~$0.50-1.00 per run).
;; Output: site/generated/labs/*.json
;;
;; Run with: glitch workflow run labs-generate

(workflow "labs-generate"
  :description "Generate lab case studies from real OSS issues"
  :tags ("site" "labs")

  ;; ── Lab 1: Issue Triage — Kubernetes ────────────────

  (step "fetch-k8s-issue"
    (run "gh issue view 138431 -R kubernetes/kubernetes --json title,body,comments,labels"))

  (step "triage-free"
    (llm
      :provider "openrouter"
      :model "google/gemma-4-31b-it:free"
      :prompt ```
        You are a senior SRE triaging a Kubernetes bug report.

        Analyze this issue and provide:
        1. Severity assessment (critical/high/medium/low) with reasoning
        2. Root cause hypothesis based on the discussion
        3. Affected components and versions
        4. Recommended next steps for the maintainer team

        Be specific — reference exact PRs, code paths, or config if mentioned.

        Issue data:
        ~(step fetch-k8s-issue)
        ```))

  (step "triage-paid"
    (llm
      :provider "openrouter"
      :model "qwen/qwen3.5-flash-02-23"
      :prompt ```
        You are a senior SRE triaging a Kubernetes bug report.

        Analyze this issue and provide:
        1. Severity assessment (critical/high/medium/low) with reasoning
        2. Root cause hypothesis based on the discussion
        3. Affected components and versions
        4. Recommended next steps for the maintainer team

        Be specific — reference exact PRs, code paths, or config if mentioned.

        Issue data:
        ~(step fetch-k8s-issue)
        ```))

  (step "triage-copilot"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt ```
        You are a senior SRE triaging a Kubernetes bug report.

        Analyze this issue and provide:
        1. Severity assessment (critical/high/medium/low) with reasoning
        2. Root cause hypothesis based on the discussion
        3. Affected components and versions
        4. Recommended next steps for the maintainer team

        Be specific — reference exact PRs, code paths, or config if mentioned.

        Issue data:
        ~(step fetch-k8s-issue)
        ```))

  ;; ── Lab 2: PR Review — Prometheus ───────────────────

  (step "fetch-prom-pr"
    (run "gh pr view 18499 -R prometheus/prometheus --json title,body,files,comments"))

  (step "fetch-prom-diff"
    (run "gh pr diff 18499 -R prometheus/prometheus"))

  (step "review-copilot"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt ```
        You are a senior Go engineer reviewing a Prometheus pull request.

        Review this PR for:
        1. Correctness — does the implementation match the stated goal?
        2. Edge cases — what could break?
        3. Test coverage — are the tests sufficient?
        4. API design — is the new config field well-designed?

        Provide specific, actionable feedback referencing file names and line ranges.

        PR metadata:
        ~(step fetch-prom-pr)

        Diff:
        ~(step fetch-prom-diff)
        ```))

  ;; ── Lab 3: Model Showdown — containerd ──────────────

  (step "fetch-containerd-issue"
    (run "gh issue view 11339 -R containerd/containerd --json title,body,comments,labels"))

  (step "analysis-free"
    (llm
      :provider "openrouter"
      :model "google/gemma-4-31b-it:free"
      :prompt ```
        You are a container runtime architect analyzing a feature proposal.

        Analyze this containerd feature request:
        1. Summarize the problem and proposed solution
        2. Identify cross-project dependencies (CRI, NRI, Kubernetes)
        3. Assess implementation complexity (small/medium/large)
        4. Flag risks or concerns
        5. Suggest an implementation approach

        Issue data:
        ~(step fetch-containerd-issue)
        ```))

  (step "analysis-paid"
    (llm
      :provider "openrouter"
      :model "qwen/qwen3.5-flash-02-23"
      :prompt ```
        You are a container runtime architect analyzing a feature proposal.

        Analyze this containerd feature request:
        1. Summarize the problem and proposed solution
        2. Identify cross-project dependencies (CRI, NRI, Kubernetes)
        3. Assess implementation complexity (small/medium/large)
        4. Flag risks or concerns
        5. Suggest an implementation approach

        Issue data:
        ~(step fetch-containerd-issue)
        ```))

  (step "analysis-copilot"
    (llm
      :provider "copilot"
      :model "sonnet"
      :prompt ```
        You are a container runtime architect analyzing a feature proposal.

        Analyze this containerd feature request:
        1. Summarize the problem and proposed solution
        2. Identify cross-project dependencies (CRI, NRI, Kubernetes)
        3. Assess implementation complexity (small/medium/large)
        4. Flag risks or concerns
        5. Suggest an implementation approach

        Issue data:
        ~(step fetch-containerd-issue)
        ```))

  ;; ── Assemble lab narratives ─────────────────────────

  (step "write-lab-1"
    (llm
      :provider "copilot"
      :model "sonnet"
      :format "json"
      :prompt ```
        Write a lab case study about using gl1tch to triage a Kubernetes issue.

        VOICE: Write for dev architects evaluating gl1tch. Direct, technical,
        no hype. Show the real output, let quality speak for itself.

        STRUCTURE your output as a JSON object (no markdown fence):
        {
          "slug": "issue-triage-kubernetes",
          "title": "Triaging a Kubernetes Issue with Tier Routing",
          "description": "Free model attempts it, paid model nails it — real input, real output from kubernetes/kubernetes#138431",
          "date": "2026-04-17",
          "content": "<markdown>"
        }

        The "content" field must follow this structure exactly:

        ## The Scenario
        Briefly describe kubernetes/kubernetes#138431 — what the bug is, why it matters.
        Link to the real issue.

        ## The Workflow
        Show the actual glitch workflow that ran this (simplified version of the
        triage steps — show the sexpr, not the JSON).

        ## Free Tier: google/gemma-4-31b-it
        Show the FULL raw output from the free model below in a code block.
        Then a **Verdict** paragraph analyzing what it got right and wrong.

        ## Paid Tier: qwen/qwen3.5-flash-02-23
        Show the FULL raw output from the paid model in a code block.
        Then a **Verdict** paragraph on where it improved.

        ## Copilot (Sonnet)
        Show the FULL raw output from Copilot in a code block.
        Then a **Verdict** paragraph.

        ## Comparison
        A markdown table: | Model | Severity Correct? | Root Cause Found? | Actionable Steps? | Cost |

        ## Takeaway
        2-3 sentences on what this demonstrates about tier routing.

        RAW OUTPUTS:

        FREE MODEL OUTPUT:
        ~(step triage-free)

        PAID MODEL OUTPUT:
        ~(step triage-paid)

        COPILOT OUTPUT:
        ~(step triage-copilot)

        ORIGINAL ISSUE:
        ~(step fetch-k8s-issue)
        ```))

  (step "write-lab-2"
    (llm
      :provider "copilot"
      :model "sonnet"
      :format "json"
      :prompt ```
        Write a lab case study about using gl1tch + Copilot to review a Prometheus PR.

        VOICE: Write for dev architects evaluating gl1tch. Direct, technical,
        no hype. Show the real output, let quality speak for itself.

        STRUCTURE your output as a JSON object (no markdown fence):
        {
          "slug": "pr-review-prometheus",
          "title": "Reviewing a Prometheus PR with Copilot",
          "description": "Copilot Sonnet reviews prometheus/prometheus#18499 — a consul health_filter feature with 263 lines changed",
          "date": "2026-04-17",
          "content": "<markdown>"
        }

        The "content" field must follow this structure:

        ## The Scenario
        Briefly describe prometheus/prometheus#18499 — what the PR does, why
        it's interesting (Health API vs Catalog API filter confusion).
        Link to the real PR.

        ## The Workflow
        Show the actual glitch workflow (the fetch + review steps in sexpr).

        ## The Review
        Show the FULL raw output from Copilot Sonnet in a code block.

        ## Analysis
        What did the review catch? What did it miss? How does this compare
        to what a human reviewer would flag?

        ## Takeaway
        2-3 sentences on using glitch for PR review workflows.

        RAW OUTPUTS:

        PR METADATA:
        ~(step fetch-prom-pr)

        PR DIFF:
        ~(step fetch-prom-diff)

        COPILOT REVIEW:
        ~(step review-copilot)
        ```))

  (step "write-lab-3"
    (llm
      :provider "copilot"
      :model "sonnet"
      :format "json"
      :prompt ```
        Write a lab case study comparing three model tiers on a containerd issue.

        VOICE: Write for dev architects evaluating gl1tch. Direct, technical,
        no hype. Show the real output, let quality speak for itself.

        STRUCTURE your output as a JSON object (no markdown fence):
        {
          "slug": "model-showdown-containerd",
          "title": "Model Showdown: Free vs Paid vs Copilot",
          "description": "Same containerd feature request through three tiers — see where free falls short and paid shines",
          "date": "2026-04-17",
          "content": "<markdown>"
        }

        The "content" field must follow this structure:

        ## The Scenario
        Briefly describe containerd/containerd#11339 — the CRI
        UpdatePodSandboxResources proposal. Link to the real issue.

        ## The Workflow
        Show the actual glitch workflow (the fetch + 3 analysis steps in sexpr).

        ## Free Tier: google/gemma-4-31b-it
        Show the FULL raw output in a code block.
        Then a **Verdict** paragraph.

        ## Paid Tier: qwen/qwen3.5-flash-02-23
        Show the FULL raw output in a code block.
        Then a **Verdict** paragraph.

        ## Copilot (Sonnet)
        Show the FULL raw output in a code block.
        Then a **Verdict** paragraph.

        ## Comparison

        | Metric | Free | Paid | Copilot |
        |--------|------|------|---------|
        | Problem summary | ... | ... | ... |
        | Cross-project deps identified | ... | ... | ... |
        | Complexity assessment | ... | ... | ... |
        | Risk identification | ... | ... | ... |
        | Implementation approach | ... | ... | ... |
        | Overall quality | .../5 | .../5 | .../5 |

        ## Takeaway
        2-3 sentences on the cost/quality tradeoff across tiers.

        RAW OUTPUTS:

        FREE MODEL OUTPUT:
        ~(step analysis-free)

        PAID MODEL OUTPUT:
        ~(step analysis-paid)

        COPILOT OUTPUT:
        ~(step analysis-copilot)

        ORIGINAL ISSUE:
        ~(step fetch-containerd-issue)
        ```))

  ;; ── Save lab JSON ───────────────────────────────────

  (save "site/generated/labs/issue-triage-kubernetes.json" :from "write-lab-1")
  (save "site/generated/labs/pr-review-prometheus.json" :from "write-lab-2")
  (save "site/generated/labs/model-showdown-containerd.json" :from "write-lab-3"))
```

- [ ] **Step 2: Create the output directory**

```bash
mkdir -p site/generated/labs
```

- [ ] **Step 3: Commit**

```bash
git add .glitch/workflows/labs-generate.glitch
git commit -m "feat(site): add labs-generate workflow for on-demand lab case studies"
```

---

### Task 8: Verify full site builds

**Files:** None (verification only)

- [ ] **Step 1: Verify site builds with empty labs**

```bash
cd site && npx astro build 2>&1 | tail -20
```

Expected: build succeeds. `/labs` index page renders (with "no labs yet" message). No errors about missing content.

- [ ] **Step 2: Verify labs route is accessible**

```bash
cd site && npx astro preview &
sleep 2
curl -s http://localhost:4321/labs/ | head -30
kill %1
```

Expected: HTML response containing "Labs" heading and the intro text.

- [ ] **Step 3: Verify nav links**

```bash
grep -c 'href="/labs"' site/dist/index.html
```

Expected: at least 1 (nav link).

- [ ] **Step 4: Create a test lab JSON to verify injection works**

```bash
cat > site/generated/labs/test-lab.json << 'EOF'
{
  "slug": "test-lab",
  "title": "Test Lab",
  "description": "Verifying the injection pipeline works",
  "date": "2026-04-17",
  "content": "## Test\n\nThis is a test lab to verify the injection pipeline.\n\n| Model | Cost |\n|-------|------|\n| Free | $0 |\n| Paid | $0.01 |"
}
EOF
```

Then run the injection step manually:

```bash
python3 - <<'EOF'
import json
from pathlib import Path

labs_dir = Path('site/generated/labs')
out_dir  = Path('site/src/content/labs')
out_dir.mkdir(parents=True, exist_ok=True)

for old in out_dir.glob('*.md'):
    if old.name != '.gitkeep':
        old.unlink()

for f in labs_dir.glob('*.json'):
    lab = json.loads(f.read_text())
    slug  = lab['slug']
    title = lab['title'].replace('"', '\\"')
    desc  = lab.get('description', '').replace('"', '\\"')
    date  = lab.get('date', '')
    frontmatter = f'---\ntitle: "{title}"\nslug: "{slug}"\ndescription: "{desc}"\ndate: "{date}"\n---\n\n'
    content = frontmatter + lab['content']
    (out_dir / f'{slug}.md').write_text(content)
    print(f'  injected: {slug}.md')
EOF
```

Expected: `injected: test-lab.md`

- [ ] **Step 5: Rebuild and verify test lab renders**

```bash
cd site && npx astro build 2>&1 | tail -10
```

Expected: build succeeds with the test lab page generated.

- [ ] **Step 6: Clean up test lab**

```bash
rm site/generated/labs/test-lab.json
rm site/src/content/labs/test-lab.md
```

- [ ] **Step 7: Commit (no files changed — verification only)**

No commit needed. All test artifacts cleaned up.
