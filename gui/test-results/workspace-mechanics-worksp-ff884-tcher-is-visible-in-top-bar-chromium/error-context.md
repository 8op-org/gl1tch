# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace-mechanics.spec.js >> workspace mechanics >> workspace switcher is visible in top bar
- Location: e2e/workspace-mechanics.spec.js:53:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.workspace-switcher select')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.workspace-switcher select')

```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | import { resolve } from 'path';
  3   | import { mkdirSync, writeFileSync, rmSync } from 'fs';
  4   | import { tmpdir, homedir } from 'os';
  5   | import { execFileSync } from 'child_process';
  6   | 
  7   | // The webServer already starts glitch with GLITCH_WORKSPACE pointing at a real
  8   | // workspace. We register that primary workspace (if needed) plus one additional
  9   | // workspace via the glitch CLI so the switcher has at least two entries.
  10  | 
  11  | const API = 'http://127.0.0.1:8374';
  12  | const ROOT = resolve(import.meta.dirname, '..', '..');
  13  | const GLITCH = resolve(ROOT, 'glitch');
  14  | const PRIMARY_WS = process.env.GLITCH_TEST_WORKSPACE || resolve(homedir(), 'Projects/stokagent');
  15  | 
  16  | const tmpWS = resolve(tmpdir(), `gl1tch-e2e-${Date.now()}`);
  17  | 
  18  | function safeGlitch(args) {
  19  |   try {
  20  |     return execFileSync(GLITCH, args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
  21  |   } catch (e) {
  22  |     return (e.stdout || '') + (e.stderr || '');
  23  |   }
  24  | }
  25  | 
  26  | test.describe('workspace mechanics', () => {
  27  |   test.beforeAll(async () => {
  28  |     // Create a second workspace on disk
  29  |     mkdirSync(tmpWS, { recursive: true });
  30  |     writeFileSync(resolve(tmpWS, 'workspace.glitch'), '(workspace "e2e-alt")\n');
  31  | 
  32  |     // Register both — ignore "already registered" errors
  33  |     safeGlitch(['workspace', 'register', PRIMARY_WS]);
  34  |     safeGlitch(['workspace', 'register', tmpWS]);
  35  |   });
  36  | 
  37  |   test.afterAll(() => {
  38  |     // Clean up the alt workspace from the registry (best effort)
  39  |     safeGlitch(['workspace', 'unregister', 'e2e-alt']);
  40  |     try { rmSync(tmpWS, { recursive: true, force: true }); } catch {}
  41  |   });
  42  | 
  43  |   test('API lists registered workspaces', async ({ request }) => {
  44  |     const res = await request.get(`${API}/api/workspaces`);
  45  |     expect(res.ok()).toBeTruthy();
  46  |     const entries = await res.json();
  47  |     expect(Array.isArray(entries)).toBe(true);
  48  |     expect(entries.length).toBeGreaterThanOrEqual(1);
  49  |     const names = entries.map(e => e.name);
  50  |     expect(names).toContain('e2e-alt');
  51  |   });
  52  | 
  53  |   test('workspace switcher is visible in top bar', async ({ page }) => {
  54  |     await page.goto('/');
  55  |     const select = page.locator('.workspace-switcher select');
> 56  |     await expect(select).toBeVisible();
      |                          ^ Error: expect(locator).toBeVisible() failed
  57  |     // Expect at least one option in the switcher (the registered workspace).
  58  |     const optionCount = await select.locator('option').count();
  59  |     expect(optionCount).toBeGreaterThanOrEqual(1);
  60  |   });
  61  | 
  62  |   test('resources panel visible in settings', async ({ page }) => {
  63  |     await page.goto('/#/settings');
  64  |     await expect(page.locator('.resources-panel')).toBeVisible();
  65  |     await expect(page.getByRole('button', { name: /add resource/i })).toBeVisible();
  66  |   });
  67  | 
  68  |   test('add a local-path resource and sync it', async ({ page, request }) => {
  69  |     // Create a fresh tiny directory we can point a resource at
  70  |     const resPath = resolve(tmpdir(), `gl1tch-e2e-res-${Date.now()}`);
  71  |     mkdirSync(resPath, { recursive: true });
  72  |     writeFileSync(resolve(resPath, 'README.md'), '# e2e\n');
  73  | 
  74  |     await page.goto('/#/settings');
  75  |     await expect(page.locator('.resources-panel')).toBeVisible();
  76  | 
  77  |     // Open the Add resource modal
  78  |     await page.getByRole('button', { name: /add resource/i }).click();
  79  | 
  80  |     // Modal input: fill URL/path field (placeholder includes "/abs/path")
  81  |     const input = page.locator('input[placeholder*="/abs/path"]').first();
  82  |     await expect(input).toBeVisible();
  83  |     await input.fill(resPath);
  84  | 
  85  |     // Submit the modal
  86  |     await page.getByRole('button', { name: /^add$/i }).click();
  87  | 
  88  |     // Wait for the modal to close and the table row to appear
  89  |     const expectedName = resPath.split('/').pop();
  90  |     const row = page.locator('.resources-table tbody tr', { hasText: expectedName });
  91  |     await expect(row).toBeVisible({ timeout: 10000 });
  92  | 
  93  |     // Click Sync on the row
  94  |     await row.getByRole('button', { name: /^sync$/i }).click();
  95  | 
  96  |     // Verify that the fetched column populated (non-empty) — poll via API
  97  |     await expect.poll(async () => {
  98  |       const res = await request.get(`${API}/api/workspace/resources`);
  99  |       if (!res.ok()) return null;
  100 |       const list = await res.json();
  101 |       const entry = (list || []).find(r => r.name === expectedName);
  102 |       return entry && entry.fetched ? entry.fetched : null;
  103 |     }, { timeout: 10000 }).not.toBeNull();
  104 | 
  105 |     // Cleanup: remove resource via API (DELETE)
  106 |     await request.delete(`${API}/api/workspace/resources/${encodeURIComponent(expectedName)}`);
  107 |     try { rmSync(resPath, { recursive: true, force: true }); } catch {}
  108 |   });
  109 | 
  110 |   test('runs tree view renders', async ({ page }) => {
  111 |     await page.goto('/#/runs');
  112 |     // Tab bar with Flat / Tree
  113 |     const treeTab = page.getByRole('button', { name: /^tree$/i });
  114 |     await expect(treeTab).toBeVisible();
  115 |     await treeTab.click();
  116 |     // Either the tree-view container, an empty state, or a run-node is present.
  117 |     // All three are acceptable — we just want the view to render without error.
  118 |     const anyView = page.locator('.tree-view, .empty-state, .run-node');
  119 |     await expect(anyView.first()).toBeVisible();
  120 |   });
  121 | });
  122 | 
```