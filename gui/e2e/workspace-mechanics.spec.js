import { test, expect } from '@playwright/test';
import { resolve } from 'path';
import { mkdirSync, writeFileSync, rmSync } from 'fs';
import { tmpdir, homedir } from 'os';
import { execFileSync } from 'child_process';

// The webServer already starts glitch with GLITCH_WORKSPACE pointing at a real
// workspace. We register that primary workspace (if needed) plus one additional
// workspace via the glitch CLI so the switcher has at least two entries.

const API = 'http://127.0.0.1:8374';
const ROOT = resolve(import.meta.dirname, '..', '..');
const GLITCH = resolve(ROOT, 'glitch');
const PRIMARY_WS = process.env.GLITCH_TEST_WORKSPACE || resolve(homedir(), 'Projects/stokagent');

const tmpWS = resolve(tmpdir(), `gl1tch-e2e-${Date.now()}`);

function safeGlitch(args) {
  try {
    return execFileSync(GLITCH, args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
  } catch (e) {
    return (e.stdout || '') + (e.stderr || '');
  }
}

test.describe('workspace mechanics', () => {
  test.beforeAll(async () => {
    // Create a second workspace on disk
    mkdirSync(tmpWS, { recursive: true });
    writeFileSync(resolve(tmpWS, 'workspace.glitch'), '(workspace "e2e-alt")\n');

    // Register both — ignore "already registered" errors
    safeGlitch(['workspace', 'register', PRIMARY_WS]);
    safeGlitch(['workspace', 'register', tmpWS]);
  });

  test.afterAll(() => {
    // Clean up the alt workspace from the registry (best effort)
    safeGlitch(['workspace', 'unregister', 'e2e-alt']);
    try { rmSync(tmpWS, { recursive: true, force: true }); } catch {}
  });

  test('API lists registered workspaces', async ({ request }) => {
    const res = await request.get(`${API}/api/workspaces`);
    expect(res.ok()).toBeTruthy();
    const entries = await res.json();
    expect(Array.isArray(entries)).toBe(true);
    expect(entries.length).toBeGreaterThanOrEqual(1);
    const names = entries.map(e => e.name);
    expect(names).toContain('e2e-alt');
  });

  test('workspace switcher is visible in top bar', async ({ page }) => {
    await page.goto('/');
    const select = page.locator('.workspace-switcher select');
    await expect(select).toBeVisible();
    // Expect at least one option in the switcher (the registered workspace).
    const optionCount = await select.locator('option').count();
    expect(optionCount).toBeGreaterThanOrEqual(1);
  });

  test('resources panel visible in settings', async ({ page }) => {
    await page.goto('/#/settings');
    await expect(page.locator('.resources-panel')).toBeVisible();
    await expect(page.getByRole('button', { name: /add resource/i })).toBeVisible();
  });

  test('add a local-path resource and sync it', async ({ page, request }) => {
    // Create a fresh tiny directory we can point a resource at
    const resPath = resolve(tmpdir(), `gl1tch-e2e-res-${Date.now()}`);
    mkdirSync(resPath, { recursive: true });
    writeFileSync(resolve(resPath, 'README.md'), '# e2e\n');

    await page.goto('/#/settings');
    await expect(page.locator('.resources-panel')).toBeVisible();

    // Open the Add resource modal
    await page.getByRole('button', { name: /add resource/i }).click();

    // Modal input: fill URL/path field (placeholder includes "/abs/path")
    const input = page.locator('input[placeholder*="/abs/path"]').first();
    await expect(input).toBeVisible();
    await input.fill(resPath);

    // Submit the modal
    await page.getByRole('button', { name: /^add$/i }).click();

    // Wait for the modal to close and the table row to appear
    const expectedName = resPath.split('/').pop();
    const row = page.locator('.resources-table tbody tr', { hasText: expectedName });
    await expect(row).toBeVisible({ timeout: 10000 });

    // Click Sync on the row
    await row.getByRole('button', { name: /^sync$/i }).click();

    // Verify that the fetched column populated (non-empty) — poll via API
    await expect.poll(async () => {
      const res = await request.get(`${API}/api/workspace/resources`);
      if (!res.ok()) return null;
      const list = await res.json();
      const entry = (list || []).find(r => r.name === expectedName);
      return entry && entry.fetched ? entry.fetched : null;
    }, { timeout: 10000 }).not.toBeNull();

    // Cleanup: remove resource via API (DELETE)
    await request.delete(`${API}/api/workspace/resources/${encodeURIComponent(expectedName)}`);
    try { rmSync(resPath, { recursive: true, force: true }); } catch {}
  });

  test('runs tree view renders', async ({ page }) => {
    await page.goto('/#/runs');
    // Tab bar with Flat / Tree
    const treeTab = page.getByRole('button', { name: /^tree$/i });
    await expect(treeTab).toBeVisible();
    await treeTab.click();
    // Either the tree-view container, an empty state, or a run-node is present.
    // All three are acceptable — we just want the view to render without error.
    const anyView = page.locator('.tree-view, .empty-state, .run-node');
    await expect(anyView.first()).toBeVisible();
  });
});
