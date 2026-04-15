# Site Update Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the single-file landing page with an Astro site and create a glitch workflow that regenerates docs from markdown stubs, real examples, and git history with tiered LLM verification.

**Architecture:** Astro static site in `site/`, markdown stubs in `docs/site/`, workflow at `.glitch/workflows/site-update.glitch`. The workflow gathers stubs + examples + git log, enriches via LLM, validates via triple-gate (parse check, astro build, diff-review), and outputs to Astro content collections.

**Tech Stack:** Astro 5, existing gl1tch sexpr syntax, Ollama (tier 0), auto-route (tier 1), Claude (tier 2)

---

### Task 1: Scaffold Astro Project

**Files:**
- Create: `site/package.json`
- Create: `site/astro.config.mjs`
- Create: `site/tsconfig.json`
- Create: `site/src/layouts/Base.astro`
- Create: `site/src/pages/index.astro`
- Create: `site/src/styles/global.css`
- Create: `site/src/components/HexRain.astro`
- Create: `site/src/components/CodeBlock.astro`
- Delete: `site/index.html` (after migration is verified)

Preserve: `site/CNAME` (stays as-is for GitHub Pages)

- [ ] **Step 1: Initialize Astro in site/ directory**

```bash
cd /Users/stokes/Projects/gl1tch/site
npm create astro@latest . -- --template minimal --no-install --no-git
npm install
```

If prompted about overwriting, accept for new files only. Keep `CNAME`.

- [ ] **Step 2: Create global.css with existing design tokens**

Extract the CSS from the current `index.html` (lines 13-511) into `site/src/styles/global.css`. This is the entire `<style>` block — all design tokens, typography, layout classes, hero animation, responsive breakpoints. Copy verbatim, no changes.

- [ ] **Step 3: Create Base layout**

Create `site/src/layouts/Base.astro`:

```astro
---
interface Props {
  title: string;
  description?: string;
}

const { title, description = 'A composable CLI that chains shell commands and LLMs into workflows you own.' } = Astro.props;
---
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{title}</title>
  <meta name="description" content={description}>
  <meta name="theme-color" content="#1a1b26">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
</head>
<body>
  <canvas id="hexbg"></canvas>
  <slot />
  <footer>
    <div class="wrap">
      <div class="footer-install">
        <pre class="code copyable" data-copy="brew install 8op-org/tap/glitch"><span class="p">$</span> brew install 8op-org/tap/glitch</pre>
      </div>
      <div class="footer-links">
        <a href="https://github.com/8op-org/gl1tch">GitHub</a>
        <a href="/docs/getting-started">Docs</a>
        <a href="/changelog">Changelog</a>
        <a href="https://github.com/8op-org/gl1tch/blob/main/LICENSE">MIT License</a>
      </div>
      <p class="footer-attr">made with AI &mdash; guided by adam s&#x2020;okes</p>
    </div>
  </footer>
</body>
</html>
```

- [ ] **Step 4: Create HexRain component**

Create `site/src/components/HexRain.astro` — extract the hex-rain `<script>` from current `index.html` (lines 867-916) into a component:

```astro
<script>
// Paste the hex-rain IIFE from current index.html lines 867-916 verbatim
</script>
```

- [ ] **Step 5: Create CodeBlock component**

Create `site/src/components/CodeBlock.astro` — extract the click-to-copy `<script>` from current `index.html` (lines 919-937):

```astro
<script>
// Paste the click-to-copy IIFE from current index.html lines 919-937 verbatim
</script>
```

- [ ] **Step 6: Create landing page**

Create `site/src/pages/index.astro` — port the HTML body from current `index.html` (lines 514-864) into an Astro page that uses the `Base` layout. Import the global CSS and components:

```astro
---
import Base from '../layouts/Base.astro';
import HexRain from '../components/HexRain.astro';
import CodeBlock from '../components/CodeBlock.astro';
import '../styles/global.css';
---

<Base title="gl1tch — your AI, your terminal, your rules">
  <main>
    <!-- Paste all <section> blocks from current index.html -->
    <!-- Hero, Ask anything, Compose workflows, Extend with plugins, -->
    <!-- Real-world examples, Give your agents gl1tch -->
  </main>
  <HexRain />
  <CodeBlock />
</Base>
```

Port all section HTML verbatim from the current file.

- [ ] **Step 7: Verify Astro build works**

```bash
cd /Users/stokes/Projects/gl1tch/site
npx astro build
```

Expected: Build succeeds, `dist/` directory created with `index.html`.

- [ ] **Step 8: Verify visual parity**

```bash
cd /Users/stokes/Projects/gl1tch/site
npx astro dev
```

Open `http://localhost:4321` — confirm the page looks identical to the current `site/index.html`. Same design tokens, hex-rain animation, glitch title effect, code blocks, copy buttons.

- [ ] **Step 9: Remove old index.html and commit**

```bash
cd /Users/stokes/Projects/gl1tch
rm site/index.html
git add site/
git commit -m "feat: migrate site to Astro, preserve design"
```

---

### Task 2: Add Astro Content Collections for Docs and Changelog

**Files:**
- Create: `site/src/content.config.ts`
- Create: `site/src/content/docs/.gitkeep`
- Create: `site/src/content/changelog/.gitkeep`
- Create: `site/src/pages/docs/[...slug].astro`
- Create: `site/src/pages/changelog.astro`
- Create: `site/src/layouts/Doc.astro`

- [ ] **Step 1: Define content collections**

Create `site/src/content.config.ts`:

```ts
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

export const collections = { docs, changelog };
```

- [ ] **Step 2: Create Doc layout**

Create `site/src/layouts/Doc.astro`:

```astro
---
import Base from './Base.astro';
import '../styles/global.css';
import HexRain from '../components/HexRain.astro';
import CodeBlock from '../components/CodeBlock.astro';

interface Props {
  title: string;
  description?: string;
}

const { title, description } = Astro.props;
---

<Base title={`${title} — gl1tch`} description={description}>
  <main>
    <section>
      <div class="wrap">
        <h2>{title}</h2>
        <div class="doc-content">
          <slot />
        </div>
      </div>
    </section>
  </main>
  <HexRain />
  <CodeBlock />
</Base>
```

- [ ] **Step 3: Create docs page template**

Create `site/src/pages/docs/[...slug].astro`:

```astro
---
import { getCollection } from 'astro:content';
import Doc from '../../layouts/Doc.astro';

export async function getStaticPaths() {
  const docs = await getCollection('docs');
  return docs.map((entry) => ({
    params: { slug: entry.id },
    props: { entry },
  }));
}

const { entry } = Astro.props;
const { Content } = await entry.render();
---

<Doc title={entry.data.title} description={entry.data.description}>
  <Content />
</Doc>
```

- [ ] **Step 4: Create changelog page**

Create `site/src/pages/changelog.astro`:

```astro
---
import { getCollection } from 'astro:content';
import Doc from '../layouts/Doc.astro';

const entries = (await getCollection('changelog')).sort(
  (a, b) => new Date(b.data.date).getTime() - new Date(a.data.date).getTime()
);
---

<Doc title="Changelog">
  {entries.map(async (entry) => {
    const { Content } = await entry.render();
    return (
      <article>
        <h3>{entry.data.title}</h3>
        <time>{entry.data.date}</time>
        <Content />
      </article>
    );
  })}
</Doc>
```

- [ ] **Step 5: Create placeholder content files**

Create `site/src/content/docs/.gitkeep` and `site/src/content/changelog/.gitkeep` so the directories exist in git. These will be populated by the workflow.

- [ ] **Step 6: Add doc-content styles to global.css**

Append to `site/src/styles/global.css`:

```css
/* ── Doc content ────────────────────────────── */
.doc-content {
  color: var(--fg);
  font-size: 0.95rem;
  line-height: 1.9;
  max-width: 720px;
}

.doc-content h2 {
  font-size: 1.5rem;
  color: var(--teal);
  margin-top: 48px;
  margin-bottom: 12px;
}

.doc-content h3 {
  font-size: 1.1rem;
  color: var(--yellow);
  margin-top: 32px;
  margin-bottom: 8px;
}

.doc-content p {
  margin-bottom: 16px;
}

.doc-content pre {
  background: var(--bg-dark);
  border-left: 3px solid var(--teal);
  border-radius: 8px;
  padding: 24px 28px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.875rem;
  line-height: 2;
  overflow-x: auto;
  margin-bottom: 24px;
}

.doc-content code {
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.9em;
  color: var(--teal);
}

.doc-content ul, .doc-content ol {
  padding-left: 24px;
  margin-bottom: 16px;
}

.doc-content li {
  margin-bottom: 8px;
  color: var(--dim);
}
```

- [ ] **Step 7: Verify build with empty collections**

```bash
cd /Users/stokes/Projects/gl1tch/site
npx astro build
```

Expected: Build succeeds. No doc pages generated yet (collections empty), but no errors.

- [ ] **Step 8: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add site/src/content.config.ts site/src/layouts/Doc.astro site/src/pages/docs/ site/src/pages/changelog.astro site/src/content/ site/src/styles/global.css
git commit -m "feat: add Astro content collections for docs and changelog"
```

---

### Task 3: Create Markdown Stubs

**Files:**
- Create: `docs/site/getting-started.md`
- Create: `docs/site/workflow-syntax.md`

- [ ] **Step 1: Create getting-started stub**

Create `docs/site/getting-started.md`:

```markdown
---
title: Getting Started
order: 1
description: Install glitch and run your first workflow
---

## Install

- brew install 8op-org/tap/glitch
- requires: ollama running locally (brew install ollama && ollama pull qwen2.5:7b)
- verify: glitch --help

## Your first workflow

- show the hello.glitch example from examples/hello.glitch
- walk through what each step does: (def ...) binds constants, (step ...) runs shell or LLM
- show how to run it: glitch workflow run hello-sexpr

## Your first ask

- glitch ask "what time is it" routes your question to the best matching workflow
- routing works via local LLM — nothing leaves your machine
- show 2-3 examples of ask routing to different workflows

## Writing your own workflow

- create .glitch/workflows/my-workflow.glitch
- minimal example: one shell step, one LLM step
- mention step references: {{step "id"}} chains outputs
- glitch workflow list to see available workflows

## Tone

- "your" framing throughout
- examples before explanation
- no internal implementation details (no BubbleTea, no tmux, no SQLite)
```

- [ ] **Step 2: Create workflow-syntax stub**

Create `docs/site/workflow-syntax.md`:

```markdown
---
title: Workflow Syntax
order: 2
description: S-expression workflow reference for glitch
---

## Overview

- glitch workflows use s-expressions (.glitch files)
- Lisp-like syntax: (form arg1 arg2 :keyword value)
- lives in .glitch/workflows/ for auto-discovery

## Workflow structure

- (workflow "name" :description "..." ...steps...)
- show a complete small example from examples/code-review.glitch

## Definitions

- (def name value) binds constants
- used for DRY: define model/provider once, reference everywhere
- show example from examples/hello.glitch

## Steps

- (step "id" (run "shell command"))
- (step "id" (llm :prompt "..."))
- (step "id" (save "path" :from "step-id"))
- each step produces a named output

## Step references

- {{step "id"}} inserts a prior step's output into prompts or commands
- {{.input}} for workflow input
- {{.param.key}} for --set key=value parameters
- show parameterized.glitch example

## LLM options

- :provider — "ollama", "claude", "copilot", "gemini", or custom
- :model — model identifier
- :skill — prepend skill context to prompt
- :format — "json" or "yaml" for structural validation
- :tier — pin to specific tier (0, 1, 2)

## Tiered cost routing

- no :provider and no :tier → auto-route through tiers
- tier 0: local (ollama, free), tier 1: cheap cloud, tier 2: premium
- self-eval at each non-final tier, escalates if quality too low
- :format triggers structural validation (must parse as JSON/YAML)

## Comments and discard

- ; line comments
- #_ discard next form (useful for debugging)

## Multiline strings

- triple backticks for readable multi-line prompts
- auto-dedented

## Tone

- practical, show-don't-tell
- every concept gets a real example from examples/
- no internal Go types or parser details
```

- [ ] **Step 3: Commit stubs**

```bash
cd /Users/stokes/Projects/gl1tch
mkdir -p docs/site
git add docs/site/
git commit -m "docs: add site content stubs for getting-started and workflow-syntax"
```

---

### Task 4: Create the Site Update Workflow

**Files:**
- Create: `.glitch/workflows/site-update.glitch`

- [ ] **Step 1: Create the workflows directory**

```bash
mkdir -p /Users/stokes/Projects/gl1tch/.glitch/workflows
```

- [ ] **Step 2: Write the workflow**

Create `.glitch/workflows/site-update.glitch`:

```scheme
;; site-update.glitch — regenerate 8op.org from doc stubs, examples, and git history
;;
;; Run with: glitch ask "update the website"
;;        or: glitch workflow run site-update

(workflow "site-update"
  :description "regenerate 8op.org docs from markdown stubs, real examples, and git history"

  ;; ── Phase 1: Gather ──────────────────────────────
  (step "stubs"
    (run "cat docs/site/*.md"))

  (step "examples"
    (run "for f in examples/*.glitch; do echo '=== '$f' ==='; cat \"$f\"; echo; done"))

  (step "changelog-raw"
    (run "git log --oneline --since='2025-01-01' -- cmd/ internal/ examples/ .glitch/"))

  ;; ── Phase 2: Generate ────────────────────────────
  (step "enrich-docs"
    (llm
      :format "json"
      :prompt ```
        You are a technical writer for gl1tch (8op.org).

        Rules:
        - "your" framing throughout, never "the user"
        - examples before explanation
        - no internal implementation details (no BubbleTea, tmux, SQLite, Go types)
        - every code example must come from the Real Examples section below
        - do NOT invent commands, flags, or features that aren't shown in the examples

        Your job: expand each stub into a full documentation page.

        Output as a JSON array of objects:
        [{"slug": "getting-started", "title": "...", "order": 1, "description": "...", "content": "markdown body..."}]

        Stubs:
        {{step "stubs"}}

        Real examples (use these verbatim, do not modify):
        {{step "examples"}}
        ```))

  (step "enrich-changelog"
    (llm
      :prompt ```
        Summarize these git commits into user-facing changelog entries.
        Group by feature area (workflow engine, providers, CLI, docs).
        Skip purely internal refactors unless they change user-visible behavior.
        Output as markdown. Each entry: ### heading, then bullet points.

        Commits:
        {{step "changelog-raw"}}
        ```))

  ;; ── Phase 3: Verify ──────────────────────────────
  (step "diff-review"
    (llm
      :tier 2
      :prompt ```
        You are a verification reviewer for the gl1tch project website.

        Compare the generated docs against the original stubs.
        Check:
        1. No hallucinated features — the docs must not claim glitch can do
           something that isn't demonstrated in the real examples
        2. Every stub heading is covered in the generated output
        3. Code examples use valid s-expression syntax (parentheses balanced,
           keywords prefixed with colon, strings quoted)
        4. Tone is user-first ("your"), no implementation internals

        Original stubs:
        {{step "stubs"}}

        Generated docs:
        {{step "enrich-docs"}}

        Respond with exactly PASS or FAIL on the first line,
        then list specific findings below.
        ```))

  ;; ── Phase 4: Output ──────────────────────────────
  (step "save-docs"
    (save "site/generated/docs.json" :from "enrich-docs"))

  (step "save-changelog"
    (save "site/generated/changelog.md" :from "enrich-changelog"))

  (step "save-report"
    (save "site/generated/build-report.md" :from "diff-review")))
```

- [ ] **Step 3: Create the generated output directory**

```bash
mkdir -p /Users/stokes/Projects/gl1tch/site/generated
echo 'site/generated/' >> /Users/stokes/Projects/gl1tch/.gitignore
```

- [ ] **Step 4: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add .glitch/workflows/site-update.glitch .gitignore
git commit -m "feat: add site-update workflow for doc generation"
```

---

### Task 5: Create the Post-Workflow Build Script

The workflow outputs JSON + markdown to `site/generated/`. A shell script splits the JSON into individual Astro content files, runs the Astro build, and validates.

**Files:**
- Create: `site/build.sh`

- [ ] **Step 1: Write the build script**

Create `site/build.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

SITE_DIR="$(cd "$(dirname "$0")" && pwd)"
GENERATED="$SITE_DIR/generated"
CONTENT_DOCS="$SITE_DIR/src/content/docs"
CONTENT_CHANGELOG="$SITE_DIR/src/content/changelog"

# ── Check generated output exists ────────────────
if [[ ! -f "$GENERATED/docs.json" ]]; then
  echo "ERROR: $GENERATED/docs.json not found. Run the site-update workflow first."
  exit 1
fi

if [[ ! -f "$GENERATED/build-report.md" ]]; then
  echo "ERROR: $GENERATED/build-report.md not found. Run the site-update workflow first."
  exit 1
fi

# ── Gate: check diff-review passed ───────────────
VERDICT=$(head -1 "$GENERATED/build-report.md")
if [[ "$VERDICT" != "PASS" ]]; then
  echo "ERROR: Diff-review did not pass."
  echo "Verdict: $VERDICT"
  echo "See: $GENERATED/build-report.md"
  exit 1
fi
echo "diff-review: PASS"

# ── Split docs JSON into individual markdown files ──
rm -rf "$CONTENT_DOCS"/*.md
node -e "
const docs = JSON.parse(require('fs').readFileSync('$GENERATED/docs.json', 'utf8'));
for (const doc of docs) {
  const fm = [
    '---',
    'title: \"' + doc.title.replace(/\"/g, '\\\\\"') + '\"',
    'order: ' + doc.order,
    'description: \"' + doc.description.replace(/\"/g, '\\\\\"') + '\"',
    '---',
    '',
    doc.content
  ].join('\n');
  require('fs').writeFileSync('$CONTENT_DOCS/' + doc.slug + '.md', fm);
  console.log('  wrote: ' + doc.slug + '.md');
}
"
echo "docs: split into content files"

# ── Write changelog ──────────────────────────────
TODAY=$(date +%Y-%m-%d)
cat > "$CONTENT_CHANGELOG/$TODAY.md" << HEADER
---
title: "Update $TODAY"
date: "$TODAY"
---

$(cat "$GENERATED/changelog.md")
HEADER
echo "changelog: wrote $TODAY.md"

# ── Build ────────────────────────────────────────
cd "$SITE_DIR"
npx astro build
echo "astro: build succeeded"

echo ""
echo "Site built to $SITE_DIR/dist/"
echo "Preview: cd $SITE_DIR && npx astro preview"
```

- [ ] **Step 2: Make executable**

```bash
chmod +x /Users/stokes/Projects/gl1tch/site/build.sh
```

- [ ] **Step 3: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add site/build.sh
git commit -m "feat: add post-workflow build script for Astro site"
```

---

### Task 6: Update Landing Page Examples to Sexpr Syntax

The current landing page still shows YAML workflow examples. Update them to sexpr format to match the actual syntax.

**Files:**
- Modify: `site/src/pages/index.astro`

- [ ] **Step 1: Update the "Compose workflows" section**

In `site/src/pages/index.astro`, find the "Compose workflows" section. Replace the YAML code block with the sexpr equivalent:

Old copy: `"Shell commands fetch the data. LLMs make sense of it. Chain them together in plain YAML."`

New copy: `"Shell commands fetch the data. LLMs make sense of it. Chain them together in s-expressions."`

Replace the YAML example with:

```html
<pre class="code"><span class="r">;; .glitch/workflows/github-pr-review.glitch</span>

(<span class="k">workflow</span> <span class="s">"github-pr-review"</span>

  (<span class="k">step</span> <span class="s">"fetch-pr"</span>
    (<span class="k">run</span> <span class="s">"gh pr view {{.input}} --json title,body,reviews"</span>))

  (<span class="k">step</span> <span class="s">"fetch-diff"</span>
    (<span class="k">run</span> <span class="s">"gh pr diff {{.input}}"</span>))

  (<span class="k">step</span> <span class="s">"review"</span>
    (<span class="k">llm</span>
      <span class="r">:provider</span> <span class="s">"claude"</span>
      <span class="r">:prompt</span> <span class="s">```
        PR: {{step "fetch-pr"}}
        Diff: {{step "fetch-diff"}}
        Review as a senior engineer.
        Flag bugs and security issues.
        ```</span>)))</pre>
```

Update the callout list item "Your workflows, your repo" text:

Old: `Drop YAML in <code>.glitch/workflows/</code> and they're live.`

New: `Drop <code>.glitch</code> files in <code>.glitch/workflows/</code> and they're live.`

- [ ] **Step 2: Update the "You don't have to write the YAML" callout**

Change heading from `"You don't have to write the YAML"` to `"You don't have to write the workflow"`.

Change description from referencing YAML to:

```
With the gl1tch skill installed in your coding agent, just describe what you want. The agent writes the workflow, saves it, and runs it.
```

Update the comment in the code block:

Old: `# Your agent writes the YAML, saves it to`
New: `# Your agent writes the workflow, saves it to`

Old: `# .glitch/workflows/pr-triage.yaml, and runs it.`
New: `# .glitch/workflows/pr-triage.glitch, and runs it.`

- [ ] **Step 3: Add docs nav link to landing page**

In the hero section, after the "View on GitHub" button, add a docs link:

```html
<a href="/docs/getting-started" class="btn" style="margin-left: 12px; border: 1px solid var(--border); color: var(--dim);">Read the docs</a>
```

- [ ] **Step 4: Verify build**

```bash
cd /Users/stokes/Projects/gl1tch/site
npx astro build
```

Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add site/src/pages/index.astro
git commit -m "feat: update landing page examples to sexpr syntax, add docs link"
```

---

### Task 7: End-to-End Smoke Test

**Files:** None created — this is a verification task.

- [ ] **Step 1: Run the site-update workflow**

```bash
cd /Users/stokes/Projects/gl1tch
glitch workflow run site-update
```

Observe: workflow should execute all phases, write output to `site/generated/`.

- [ ] **Step 2: Check generated output**

```bash
cat site/generated/docs.json | head -20
cat site/generated/changelog.md | head -20
cat site/generated/build-report.md
```

Verify:
- `docs.json` is valid JSON array with slug, title, order, description, content fields
- `changelog.md` has grouped entries
- `build-report.md` starts with PASS or FAIL

- [ ] **Step 3: Run the build script**

```bash
./site/build.sh
```

Expected:
- Diff-review gate passes
- Docs split into `site/src/content/docs/getting-started.md` and `workflow-syntax.md`
- Changelog written to `site/src/content/changelog/`
- `npx astro build` succeeds
- Output in `site/dist/`

- [ ] **Step 4: Preview the site**

```bash
cd /Users/stokes/Projects/gl1tch/site
npx astro preview
```

Open `http://localhost:4321` — verify:
- Landing page renders correctly with sexpr examples
- `/docs/getting-started` shows enriched content
- `/docs/workflow-syntax` shows syntax reference with real examples
- `/changelog` shows entries
- Navigation between pages works

- [ ] **Step 5: If anything fails, fix and re-run**

Common issues:
- JSON parse error in docs.json → check the LLM prompt, may need `:format "json"` enforcement
- Astro build fails → check frontmatter format in generated markdown
- Diff-review says FAIL → read the findings, adjust stubs or prompt

- [ ] **Step 6: Final commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add site/src/content/
git commit -m "feat: first generated site content from workflow"
```

---

### Task 8: Update GitHub Pages Workflow

**Files:**
- Modify: `.github/workflows/pages.yml`

- [ ] **Step 1: Update the Pages workflow to build Astro**

Edit `.github/workflows/pages.yml` to install Node and run the Astro build before deploying:

```yaml
name: Deploy site to GitHub Pages

on:
  push:
    branches: [main]
    paths: [site/**]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - name: Install and build
        working-directory: site
        run: |
          npm ci
          npx astro build
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: site/dist
      - id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 2: Commit**

```bash
cd /Users/stokes/Projects/gl1tch
git add .github/workflows/pages.yml
git commit -m "feat: update GitHub Pages workflow to build Astro"
```
