# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: stokagent.spec.js >> Stokagent workspace — API >> GET /api/workflows returns 30+ workflows
- Location: e2e/stokagent.spec.js:10:3

# Error details

```
Error: expect(received).toBeGreaterThan(expected)

Expected: > 30
Received:   25
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test'
  2   | 
  3   | // These tests run against the stokagent workspace (38 real workflows).
  4   | // Requires: glitch --workspace ~/Projects/stokagent workflow gui
  5   | // on port 8374 before running.
  6   | 
  7   | test.use({ baseURL: 'http://127.0.0.1:8374' })
  8   | 
  9   | test.describe('Stokagent workspace — API', () => {
  10  |   test('GET /api/workflows returns 30+ workflows', async ({ request }) => {
  11  |     const resp = await request.get('/api/workflows')
  12  |     expect(resp.ok()).toBeTruthy()
  13  |     const data = await resp.json()
> 14  |     expect(data.length).toBeGreaterThan(30)
      |                         ^ Error: expect(received).toBeGreaterThan(expected)
  15  |     // Verify converted .glitch files are present
  16  |     const names = data.map(w => w.file)
  17  |     expect(names).toContain('dashboard-reviews.glitch')
  18  |     expect(names).toContain('git-status.glitch')
  19  |     expect(names).toContain('work-on-issue.glitch')
  20  |   })
  21  | 
  22  |   test('GET /api/workflows/git-status.glitch returns source with sexpr', async ({ request }) => {
  23  |     const resp = await request.get('/api/workflows/git-status.glitch')
  24  |     expect(resp.ok()).toBeTruthy()
  25  |     const data = await resp.json()
  26  |     expect(data.source).toContain('(workflow')
  27  |     expect(data.source).toContain('git-status')
  28  |     expect(data.params).toEqual([])
  29  |   })
  30  | 
  31  |   test('GET /api/workflows/dashboard-reviews.glitch has description', async ({ request }) => {
  32  |     const resp = await request.get('/api/workflows/dashboard-reviews.glitch')
  33  |     expect(resp.ok()).toBeTruthy()
  34  |     const data = await resp.json()
  35  |     expect(data.source).toContain('(workflow')
  36  |     expect(data.source).toContain('dashboard-reviews')
  37  |   })
  38  | 
  39  |   test('parameterized workflow extracts params', async ({ request }) => {
  40  |     const resp = await request.get('/api/workflows/issue-to-pr-claude.glitch')
  41  |     expect(resp.ok()).toBeTruthy()
  42  |     const data = await resp.json()
  43  |     expect(data.params).toContain('repo')
  44  |     expect(data.params).toContain('issue')
  45  |   })
  46  | })
  47  | 
  48  | test.describe('Stokagent workspace — Workflow browsing', () => {
  49  |   test('workflow list loads with 30+ cards', async ({ page }) => {
  50  |     await page.goto('/')
  51  |     await page.waitForSelector('.card')
  52  |     const cards = page.locator('.card')
  53  |     const count = await cards.count()
  54  |     expect(count).toBeGreaterThan(30)
  55  |   })
  56  | 
  57  |   test('workflow cards show descriptions', async ({ page }) => {
  58  |     await page.goto('/')
  59  |     await page.waitForSelector('.card')
  60  |     // dashboard-reviews should have a description
  61  |     const desc = page.locator('.card-desc')
  62  |     const count = await desc.count()
  63  |     expect(count).toBeGreaterThan(0)
  64  |   })
  65  | 
  66  |   test('grouped view shows workflow groups', async ({ page }) => {
  67  |     await page.goto('/')
  68  |     await page.waitForSelector('.group-header')
  69  |     const groups = page.locator('.group-header')
  70  |     const count = await groups.count()
  71  |     expect(count).toBeGreaterThan(3) // dashboard, github, issue-to-pr, etc.
  72  |   })
  73  | 
  74  |   test('group collapse/expand works', async ({ page }) => {
  75  |     await page.goto('/')
  76  |     await page.waitForSelector('.group-header')
  77  |     const firstGroup = page.locator('.group').first()
  78  |     const header = firstGroup.locator('.group-header')
  79  |     // Verify grid is visible initially
  80  |     await expect(firstGroup.locator('.group-grid')).toBeVisible()
  81  |     // Click to collapse
  82  |     await header.click()
  83  |     // The group grid should be hidden
  84  |     await expect(firstGroup.locator('.group-grid')).not.toBeVisible()
  85  |     // Click again to expand
  86  |     await header.click()
  87  |     await expect(firstGroup.locator('.group-grid')).toBeVisible()
  88  |   })
  89  | 
  90  |   test('search filters to matching workflows', async ({ page }) => {
  91  |     await page.goto('/')
  92  |     await page.waitForSelector('.card')
  93  |     await page.fill('input[placeholder="Search..."]', 'dashboard')
  94  |     const cards = page.locator('.card')
  95  |     const count = await cards.count()
  96  |     expect(count).toBeGreaterThanOrEqual(8) // 8 dashboard workflows
  97  |     expect(count).toBeLessThan(38) // not all
  98  |   })
  99  | 
  100 |   test('search for nonexistent shows zero cards', async ({ page }) => {
  101 |     await page.goto('/')
  102 |     await page.waitForSelector('.card')
  103 |     await page.fill('input[placeholder="Search..."]', 'xyznonexistent999')
  104 |     await expect(page.locator('.card')).toHaveCount(0)
  105 |   })
  106 | 
  107 |   test('view mode switches work', async ({ page }) => {
  108 |     await page.goto('/')
  109 |     await page.waitForSelector('.card')
  110 |     // Switch to list view
  111 |     await page.locator('.view-btn', { hasText: /.*/ }).nth(2).click()
  112 |     await expect(page.locator('.wf-table')).toBeVisible()
  113 |     // Switch to grid view
  114 |     await page.locator('.view-btn', { hasText: /.*/ }).nth(1).click()
```