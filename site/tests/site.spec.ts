import { test, expect } from '@playwright/test';

// ── Homepage ────────────────────────────────────

test('homepage loads with hero', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('.hero-title')).toHaveText('gl1tch');
  await expect(page.locator('.hero-tagline')).toBeVisible();
});

test('homepage has all sections', async ({ page }) => {
  await page.goto('/');
  for (const id of ['features', 'how', 'meta', 'agents']) {
    await expect(page.locator(`#${id}`)).toBeVisible();
  }
});

test('feature grid has 7 cards', async ({ page }) => {
  await page.goto('/');
  const cards = page.locator('.feature-card');
  await expect(cards).toHaveCount(7);
});

test('how-it-works has 4 steps', async ({ page }) => {
  await page.goto('/');
  const steps = page.locator('.how-step');
  await expect(steps).toHaveCount(4);
});

test('meta section shows both gate phases', async ({ page }) => {
  await page.goto('/');
  const trace = page.locator('.meta-trace');
  await expect(trace).toContainText('content-verify');
  await expect(trace).toContainText('page-tests');
  await expect(trace).toContainText('playwright');
});

test('workflow suite has 4 commands', async ({ page }) => {
  await page.goto('/');
  const cmds = page.locator('.meta-cmd');
  await expect(cmds).toHaveCount(4);
});

test('install command is copyable', async ({ page }) => {
  await page.goto('/');
  const install = page.locator('.hero-cmd');
  await expect(install).toContainText('brew install 8op-org/tap/glitch');
});

test('no broken internal links', async ({ page }) => {
  await page.goto('/');
  const links = page.locator('a[href^="/docs/"]');
  const count = await links.count();
  expect(count).toBeGreaterThan(0);
  for (let i = 0; i < count; i++) {
    const href = await links.nth(i).getAttribute('href');
    const resp = await page.request.get(href!);
    expect(resp.status(), `broken link: ${href}`).toBe(200);
  }
});

// ── Doc pages ───────────────────────────────────

test('getting-started page loads', async ({ page }) => {
  await page.goto('/docs/getting-started');
  await expect(page.locator('h2').first()).toContainText('Getting Started');
  await expect(page.locator('.doc-content')).toBeVisible();
});

test('workflow-syntax page loads', async ({ page }) => {
  await page.goto('/docs/workflow-syntax');
  await expect(page.locator('h2').first()).toContainText('Workflow Syntax');
});

test('plugins page loads', async ({ page }) => {
  await page.goto('/docs/plugins');
  await expect(page.locator('h2').first()).toContainText('Plugins');
});

test('local-models page loads', async ({ page }) => {
  await page.goto('/docs/local-models');
  await expect(page.locator('h2').first()).toContainText('Local Models');
});

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

test('doc pages have content (not empty shells)', async ({ page }) => {
  const pages = ['getting-started', 'workflow-syntax', 'plugins', 'local-models', 'compare', 'dsl-reference', 'workspaces', 'phases-and-gates'];
  for (const slug of pages) {
    await page.goto(`/docs/${slug}`);
    const content = page.locator('.doc-content');
    const text = await content.textContent();
    expect(text!.length, `${slug} has too little content`).toBeGreaterThan(200);
  }
});

// ── Changelog ───────────────────────────────────

test('changelog page loads', async ({ page }) => {
  await page.goto('/changelog');
  await expect(page.locator('h2').first()).toContainText('Changelog');
});

// ── No leaked internals ─────────────────────────

test('no BubbleTea/SQLite/tmux on any page', async ({ page, }, testInfo) => {
  testInfo.setTimeout(30000);
  const pages = ['/', '/docs/getting-started', '/docs/workflow-syntax', '/docs/plugins', '/docs/local-models', '/docs/compare', '/docs/dsl-reference', '/docs/workspaces', '/docs/phases-and-gates'];
  const banned = ['BubbleTea', 'bubbletea', 'tea.Model', 'SQLite', 'sqlite3', 'internal/tui'];
  for (const url of pages) {
    await page.goto(url);
    const text = await page.textContent('body');
    for (const term of banned) {
      expect(text, `"${term}" found on ${url}`).not.toContain(term);
    }
  }
});

// ── Navigation ─────────────────────────────────────

test('nav bar exists with home link', async ({ page }) => {
  await page.goto('/');
  const nav = page.locator('.site-nav');
  await expect(nav).toBeVisible();
  const logo = page.locator('.nav-logo');
  await expect(logo).toHaveAttribute('href', '/');
  await expect(logo).toHaveText('gl1tch');
});

test('can navigate from doc page back to home', async ({ page }) => {
  await page.goto('/docs/getting-started');
  await page.click('.nav-logo');
  await expect(page).toHaveURL('/');
  await expect(page.locator('.hero-title')).toBeVisible();
});

test('nav has docs, labs, and changelog links', async ({ page }) => {
  await page.goto('/');
  const navLinks = page.locator('.nav-links a');
  const texts = await navLinks.allTextContents();
  expect(texts).toContain('docs');
  expect(texts).toContain('labs');
  expect(texts).toContain('changelog');
});

// ── Labs ──────────────────────────────────────────

const labSlugs = [
  'issue-triage-kubernetes',
  'pr-review-prometheus',
  'model-showdown-containerd',
  'bug-triage-kibana',
];

test('labs index lists all 4 labs', async ({ page }) => {
  await page.goto('/labs');
  const cards = page.locator('.lab-card');
  await expect(cards).toHaveCount(4);
});

test('labs index cards link to detail pages', async ({ page }) => {
  await page.goto('/labs');
  for (const slug of labSlugs) {
    const link = page.locator(`a.lab-card[href="/labs/${slug}"]`);
    await expect(link, `missing card link for ${slug}`).toBeVisible();
  }
});

test('labs index cards have title, description, and date', async ({ page }) => {
  await page.goto('/labs');
  const cards = page.locator('.lab-card');
  const count = await cards.count();
  for (let i = 0; i < count; i++) {
    const card = cards.nth(i);
    await expect(card.locator('h2')).not.toBeEmpty();
    await expect(card.locator('p')).not.toBeEmpty();
    await expect(card.locator('time')).not.toBeEmpty();
  }
});

test('labs index renders without JS errors', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (err) => errors.push(err.message));
  await page.goto('/labs');
  await page.waitForTimeout(500);
  expect(errors).toEqual([]);
});

for (const slug of labSlugs) {
  test(`lab detail page loads: ${slug}`, async ({ page }) => {
    await page.goto(`/labs/${slug}`);
    await expect(page.locator('.lab-header h1')).not.toBeEmpty();
    await expect(page.locator('.lab-content')).toBeVisible();
    await expect(page.locator('.lab-back')).toHaveAttribute('href', '/labs');
  });

  test(`lab has required sections: ${slug}`, async ({ page }) => {
    await page.goto(`/labs/${slug}`);
    const content = page.locator('.lab-content');
    const h2s = content.locator('h2');
    const count = await h2s.count();
    expect(count, `${slug} should have at least 3 h2 headings`).toBeGreaterThanOrEqual(3);

    // Every lab must have "The Scenario" and "The Workflow"
    const text = await content.textContent();
    expect(text, `${slug} missing "The Scenario"`).toContain('The Scenario');
    expect(text, `${slug} missing "The Workflow"`).toContain('The Workflow');
  });

  test(`lab has code blocks: ${slug}`, async ({ page }) => {
    await page.goto(`/labs/${slug}`);
    const codeBlocks = page.locator('.lab-content pre');
    const count = await codeBlocks.count();
    expect(count, `${slug} should have at least 1 code block`).toBeGreaterThanOrEqual(1);
  });

  test(`lab has comparison table: ${slug}`, async ({ page }) => {
    await page.goto(`/labs/${slug}`);
    const tables = page.locator('.lab-content table');
    const count = await tables.count();
    expect(count, `${slug} should have at least 1 comparison table`).toBeGreaterThanOrEqual(1);

    // Table should have header row and data rows
    const firstTable = tables.first();
    const headerCells = firstTable.locator('th');
    const dataCells = firstTable.locator('td');
    expect(await headerCells.count(), `${slug} table needs headers`).toBeGreaterThanOrEqual(2);
    expect(await dataCells.count(), `${slug} table needs data`).toBeGreaterThanOrEqual(2);
  });

  test(`lab has no raw markdown artifacts: ${slug}`, async ({ page }) => {
    await page.goto(`/labs/${slug}`);
    // Check text OUTSIDE code blocks for unrendered markdown
    const nonCodeText = await page.locator('.lab-content').evaluate((el) => {
      const clone = el.cloneNode(true) as HTMLElement;
      clone.querySelectorAll('pre, code').forEach((n) => n.remove());
      return clone.textContent || '';
    });
    expect(nonCodeText, `${slug} has raw triple backticks outside code blocks`).not.toMatch(/```/);
    expect(nonCodeText, `${slug} has raw ~(step ...) outside code blocks`).not.toMatch(/~\(step /);
    expect(nonCodeText, `${slug} has raw {{step ...}} outside code blocks`).not.toMatch(/\{\{step /);
  });

  test(`lab content is substantial: ${slug}`, async ({ page }) => {
    await page.goto(`/labs/${slug}`);
    const text = await page.locator('.lab-content').textContent();
    expect(text!.length, `${slug} content too short`).toBeGreaterThan(2000);
  });

  test(`lab has no JS errors: ${slug}`, async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', (err) => errors.push(err.message));
    await page.goto(`/labs/${slug}`);
    await page.waitForTimeout(500);
    expect(errors).toEqual([]);
  });
}

test('lab detail pages have correct layout (no sidebar)', async ({ page }) => {
  await page.goto(`/labs/${labSlugs[0]}`);
  // Lab layout should NOT have doc-sidebar
  await expect(page.locator('.doc-sidebar-col')).toHaveCount(0);
  // Should have lab-specific layout
  await expect(page.locator('.lab-page')).toBeVisible();
});

test('lab code blocks are visually distinct', async ({ page }) => {
  await page.goto(`/labs/${labSlugs[0]}`);
  const pre = page.locator('.lab-content pre').first();
  const bg = await pre.evaluate((el) => getComputedStyle(el).backgroundColor);
  expect(bg).not.toBe('rgba(0, 0, 0, 0)');
  expect(bg).not.toBe('transparent');
});

test('lab tables have styled headers', async ({ page }) => {
  await page.goto(`/labs/${labSlugs[0]}`);
  const th = page.locator('.lab-content th').first();
  const color = await th.evaluate((el) => getComputedStyle(el).color);
  // Headers should be teal-ish, not default white
  expect(color).not.toBe('rgb(255, 255, 255)');
});

test('lab content width is wider than doc pages', async ({ page }) => {
  await page.goto(`/labs/${labSlugs[0]}`);
  const labBox = await page.locator('.lab-page').boundingBox();

  await page.goto('/docs/getting-started');
  const docBox = await page.locator('.doc-main').boundingBox();

  expect(labBox).not.toBeNull();
  expect(docBox).not.toBeNull();
  // Lab page max-width (960px) should be >= doc main width
  expect(labBox!.width).toBeGreaterThanOrEqual(docBox!.width * 0.9);
});

test('no broken internal links on labs pages', async ({ page }) => {
  await page.goto('/labs');
  const links = page.locator('a.lab-card');
  const count = await links.count();
  expect(count).toBeGreaterThan(0);
  for (let i = 0; i < count; i++) {
    const href = await links.nth(i).getAttribute('href');
    const resp = await page.request.get(href!);
    expect(resp.status(), `broken lab link: ${href}`).toBe(200);
  }
});

// ── Syntax highlighting ────────────────────────────

test('doc code blocks have syntax highlighting', async ({ page }) => {
  await page.goto('/docs/getting-started');
  // Shiki wraps tokens in <span style="color:..."> elements
  const highlightedSpans = page.locator('.doc-content pre code span[style*="color"]');
  const count = await highlightedSpans.count();
  expect(count, 'code blocks should have colored spans from Shiki').toBeGreaterThan(0);
});

test('glitch code blocks have keyword + string colors', async ({ page }) => {
  await page.goto('/docs/workflow-syntax');
  // Shiki should produce blocks with data-language="glitch"
  const glitchBlocks = page.locator('pre[data-language="glitch"]');
  const blockCount = await glitchBlocks.count();
  expect(blockCount, 'should have glitch-language code blocks').toBeGreaterThan(0);
  // Check the first non-trivial block has multiple distinct token colors
  const spans = glitchBlocks.nth(1).locator('code span[style*="color"]');
  const colors = new Set<string>();
  for (let i = 0; i < await spans.count(); i++) {
    const style = await spans.nth(i).getAttribute('style');
    const match = style?.match(/color:(#[A-Fa-f0-9]+)/);
    if (match) colors.add(match[1]);
  }
  expect(colors.size, 'glitch blocks need at least 3 distinct token colors (keyword, string, punctuation)').toBeGreaterThanOrEqual(3);
});

// ── Card styling ───────────────────────────────────

test('feature cards have distinct background from page', async ({ page }) => {
  await page.goto('/');
  const card = page.locator('.feature-card').first();
  const cardBg = await card.evaluate((el) => getComputedStyle(el).backgroundColor);
  // Should not be fully transparent or same as page bg
  expect(cardBg).not.toBe('rgba(0, 0, 0, 0)');
  expect(cardBg).not.toBe('transparent');
});

test('feature cards have hover border accent', async ({ page }) => {
  await page.goto('/');
  const card = page.locator('.feature-card').first();
  await card.hover();
  const borderLeft = await card.evaluate((el) => getComputedStyle(el).borderLeftColor);
  // After hover, border-left should be teal (not transparent)
  expect(borderLeft).not.toBe('rgba(0, 0, 0, 0)');
});

// ── Brew install card ──────────────────────────────

test('brew install text does not clip under copy button', async ({ page }) => {
  await page.goto('/');
  const cmd = page.locator('.hero-cmd');
  const cmdBox = await cmd.boundingBox();
  // Wait for JS to inject the copy button
  await page.waitForSelector('.hero-cmd .copy-btn');
  const btn = page.locator('.hero-cmd .copy-btn');
  const btnBox = await btn.boundingBox();
  // The text area (cmd width minus padding-right) should not overlap with button
  expect(cmdBox).not.toBeNull();
  expect(btnBox).not.toBeNull();
  // Verify the cmd block is wide enough for text + button
  expect(cmdBox!.width).toBeGreaterThan(btnBox!.width + 200);
});

// ── Visual regression guard ─────────────────────

test('homepage renders without JS errors', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (err) => errors.push(err.message));
  await page.goto('/');
  await page.waitForTimeout(1000);
  expect(errors).toEqual([]);
});

test('no 404 resources on homepage', async ({ page }) => {
  const failures: string[] = [];
  page.on('response', (resp) => {
    if (resp.status() === 404 && !resp.url().includes('favicon')) {
      failures.push(`${resp.status()} ${resp.url()}`);
    }
  });
  await page.goto('/');
  await page.waitForTimeout(1000);
  expect(failures).toEqual([]);
});
