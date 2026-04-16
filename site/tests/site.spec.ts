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
    expect(text!.length, `${slug} has too little content`).toBeGreaterThan(500);
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

test('nav has docs and changelog links', async ({ page }) => {
  await page.goto('/');
  const navLinks = page.locator('.nav-links a');
  const texts = await navLinks.allTextContents();
  expect(texts).toContain('docs');
  expect(texts).toContain('changelog');
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
