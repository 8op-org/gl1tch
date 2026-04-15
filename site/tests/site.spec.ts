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

test('feature grid has 6 cards', async ({ page }) => {
  await page.goto('/');
  const cards = page.locator('.feature-card');
  await expect(cards).toHaveCount(6);
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

test('doc pages have content (not empty shells)', async ({ page }) => {
  const pages = ['getting-started', 'workflow-syntax', 'plugins', 'local-models'];
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

test('no BubbleTea/SQLite/tmux on any page', async ({ page }) => {
  const pages = ['/', '/docs/getting-started', '/docs/workflow-syntax', '/docs/plugins', '/docs/local-models'];
  const banned = ['BubbleTea', 'bubbletea', 'tea.Model', 'SQLite', 'sqlite3', 'internal/tui'];
  for (const url of pages) {
    await page.goto(url);
    const text = await page.textContent('body');
    for (const term of banned) {
      expect(text, `"${term}" found on ${url}`).not.toContain(term);
    }
  }
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
