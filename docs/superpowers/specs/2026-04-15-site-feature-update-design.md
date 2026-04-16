# Site Feature Update Design

**Date:** 2026-04-15
**Status:** Approved
**Goal:** Update 8op.org landing page with new feature highlights and add 4 new doc pages covering recently shipped features.

## Scope

### Landing Page Changes

**Feature card 06 — replace "Batch comparison" with "Compare form":**
Show the new `(compare ...)` / `(branch ...)` / `(review ...)` syntax. This is a language-level feature, not just "run the same workflow N times."

```html
<div class="feature-card">
  <div class="feature-number">06</div>
  <h3>Built-in A/B testing</h3>
  <p>Run the same prompt through different models or strategies.
     A neutral judge scores each branch. One form, automatic winner selection.</p>
  <pre class="code feature-code">(compare
  (branch "local" (llm :model "qwen2.5:7b" ...))
  (branch "cloud" (llm :model "claude" ...))
  (review :criteria ("accuracy" "clarity")))</pre>
</div>
```

**Feature card 07 — new card "Composable transforms":**
Highlight the DSL forms that make gl1tch a real data pipeline: `->`, `filter`, `reduce`, `when`, `search`, `index`.

```html
<div class="feature-card">
  <div class="feature-number">07</div>
  <h3>Composable transforms</h3>
  <p>Thread data through steps, filter collections, fold results.
     Lisp-inspired forms that compose without glue scripts.</p>
  <pre class="code feature-code">(-> "raw-data"
  (filter "jq '.severity > 3'")
  (reduce "summarize" (llm :prompt "...")))</pre>
</div>
```

**How-it-works — add 4th step "04 compare and pick the best":**
Show the compare flow as the natural extension of gather/generate/verify.

**Playwright test updates:**
- Card count assertion: 6 -> 7
- How-it-works step count: 3 -> 4

### New Doc Pages (4)

All pages follow the existing pattern: stub in `docs/site/`, enriched content in `site/src/content/docs/`. Each page needs frontmatter with `title`, `order`, `description`.

#### 1. Compare Runs (`compare.md`, order: 6)
Replaces the thin `batch-comparison-runs.md` page with the real `(compare ...)` form.

Content outline:
- The compare form: `(compare ...)`, `(branch ...)`, `(review ...)`
- Real example: `compare-models.glitch` (verbatim from `examples/`)
- Real example: `compare-branches.glitch` (verbatim)
- CLI flags: `--variant`, `--compare`, `--review-criteria`
- How review scoring works (neutral local judge, criteria list)
- Saving and reading results

#### 2. DSL Reference (`dsl-reference.md`, order: 7)
The new forms shipped in the DSL improvements branch.

Content outline:
- Threading macro: `(->)` — pipe data between forms
- Collection forms: `(filter ...)`, `(reduce ...)`
- Conditionals: `(when ...)`, `(when-not ...)`
- Elasticsearch forms: `(search ...)`, `(index ...)`, `(delete ...)`
- Embedding: `(embed ...)`
- Data transforms: `(flatten)`, `assoc` template function, `pick` template function
- Each form gets a short example + one-line description

#### 3. Workspaces (`workspaces.md`, order: 8)
The workspace model: project-scoped config and plugin resolution.

Content outline:
- What a workspace is and why
- `workspace.glitch` manifest format
- `--workspace` CLI flag
- Workspace-aware plugin resolution
- Workspace Elasticsearch URL for knowledge forms
- Example workspace setup

#### 4. Phases & Gates (`phases-and-gates.md`, order: 9)
Already teased on the landing page, needs a dedicated reference.

Content outline:
- `(phase "name" :retries N ...)` form
- `(gate "name" (run "..."))` form
- How retries work (whole phase reruns)
- Real example: the site-create-page verification gates
- Composing with other control flow (retry, timeout)

### Cleanup

- Remove `batch-comparison-runs.md` from `docs/site/` and `site/src/content/docs/` — replaced by `compare.md`
- Update `workflow-syntax.md` "Next steps" links to include new pages
- Update `getting-started.md` "Next steps" to reference compare and DSL pages

### Playwright Tests for New Pages

Add to `site/tests/site.spec.ts`:

```typescript
// New doc pages load
test('compare page loads', async ({ page }) => {
  await page.goto('/docs/compare');
  await expect(page.locator('h2').first()).toContainText('Compare');
});

test('dsl-reference page loads', async ({ page }) => {
  await page.goto('/docs/dsl-reference');
  await expect(page.locator('h2').first()).toContainText('DSL Reference');
});

test('workspaces page loads', async ({ page }) => {
  await page.goto('/docs/workspaces');
  await expect(page.locator('h2').first()).toContainText('Workspaces');
});

test('phases-and-gates page loads', async ({ page }) => {
  await page.goto('/docs/phases-and-gates');
  await expect(page.locator('h2').first()).toContainText('Phases');
});

// Updated counts
// feature cards: 6 -> 7
// how-it-works steps: 3 -> 4

// New pages in content check
test('new doc pages have content', async ({ page }) => {
  const pages = ['compare', 'dsl-reference', 'workspaces', 'phases-and-gates'];
  for (const slug of pages) {
    await page.goto(`/docs/${slug}`);
    const content = page.locator('.doc-content');
    const text = await content.textContent();
    expect(text!.length, `${slug} has too little content`).toBeGreaterThan(300);
  }
});

// Banned terms check includes new pages
// Internal links check covers new pages (already handled by existing test)
```

## Implementation Strategy

### Parallelization: 5 independent subagents

1. **Agent: compare-stub** — Write `docs/site/compare.md` stub, create `site/src/content/docs/compare.md` with full content, remove old `batch-comparison-runs.md`
2. **Agent: dsl-reference-stub** — Write `docs/site/dsl-reference.md` stub, create `site/src/content/docs/dsl-reference.md` with full content
3. **Agent: workspaces-stub** — Write `docs/site/workspaces.md` stub, create `site/src/content/docs/workspaces.md` with full content
4. **Agent: phases-gates-stub** — Write `docs/site/phases-and-gates.md` stub, create `site/src/content/docs/phases-and-gates.md` with full content
5. **Agent: landing-page** — Update `index.astro` (cards 06/07, how-it-works step 4), update Playwright tests, update cross-references in existing pages

### Verification

After all agents complete:
- Run `npx astro build` in `site/` to catch broken templates
- Run `npx playwright test` for full test suite including new assertions

## Content Rules (from memory)

- "your" framing, never "the user"
- Examples before explanation
- No internals: no BubbleTea, tmux, SQLite, Go types
- Code examples must be real — from `examples/` or `.glitch/workflows/`
- No invented commands, flags, or features

## Dependencies

- Astro installed in `site/` (already present)
- Playwright installed (already present via `site/playwright.config.ts`)
- Example `.glitch` files exist for compare forms
- DSL forms are implemented in `internal/pipeline/`
