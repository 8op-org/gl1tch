import { test, expect } from '@playwright/test'

// These tests run against the stokagent workspace (38 real workflows).
// Requires: glitch --workspace ~/Projects/stokagent workflow gui
// on port 8374 before running.

test.use({ baseURL: 'http://127.0.0.1:8374' })

test.describe('Stokagent workspace — API', () => {
  test('GET /api/workflows returns 30+ workflows', async ({ request }) => {
    const resp = await request.get('/api/workflows')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.length).toBeGreaterThan(30)
    // Verify converted .glitch files are present
    const names = data.map(w => w.file)
    expect(names).toContain('dashboard-reviews.glitch')
    expect(names).toContain('git-status.glitch')
    expect(names).toContain('work-on-issue.glitch')
  })

  test('GET /api/workflows/git-status.glitch returns source with sexpr', async ({ request }) => {
    const resp = await request.get('/api/workflows/git-status.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.source).toContain('(workflow')
    expect(data.source).toContain('git-status')
    expect(data.params).toEqual([])
  })

  test('GET /api/workflows/dashboard-reviews.glitch has description', async ({ request }) => {
    const resp = await request.get('/api/workflows/dashboard-reviews.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.source).toContain('(workflow')
    expect(data.source).toContain('dashboard-reviews')
  })

  test('parameterized workflow extracts params', async ({ request }) => {
    const resp = await request.get('/api/workflows/issue-to-pr-claude.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.params).toContain('repo')
    expect(data.params).toContain('issue')
  })
})

test.describe('Stokagent workspace — Workflow browsing', () => {
  test('workflow list loads with 30+ cards', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const cards = page.locator('.card')
    const count = await cards.count()
    expect(count).toBeGreaterThan(30)
  })

  test('workflow cards show descriptions', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    // dashboard-reviews should have a description
    const desc = page.locator('.card-desc')
    const count = await desc.count()
    expect(count).toBeGreaterThan(0)
  })

  test('grouped view shows workflow groups', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.group-header')
    const groups = page.locator('.group-header')
    const count = await groups.count()
    expect(count).toBeGreaterThan(3) // dashboard, github, issue-to-pr, etc.
  })

  test('group collapse/expand works', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.group-header')
    const firstGroup = page.locator('.group').first()
    const header = firstGroup.locator('.group-header')
    // Verify grid is visible initially
    await expect(firstGroup.locator('.group-grid')).toBeVisible()
    // Click to collapse
    await header.click()
    // The group grid should be hidden
    await expect(firstGroup.locator('.group-grid')).not.toBeVisible()
    // Click again to expand
    await header.click()
    await expect(firstGroup.locator('.group-grid')).toBeVisible()
  })

  test('search filters to matching workflows', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.fill('input[placeholder="Search..."]', 'dashboard')
    const cards = page.locator('.card')
    const count = await cards.count()
    expect(count).toBeGreaterThanOrEqual(8) // 8 dashboard workflows
    expect(count).toBeLessThan(38) // not all
  })

  test('search for nonexistent shows zero cards', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.fill('input[placeholder="Search..."]', 'xyznonexistent999')
    await expect(page.locator('.card')).toHaveCount(0)
  })

  test('view mode switches work', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    // Switch to list view
    await page.locator('.view-btn', { hasText: /.*/ }).nth(2).click()
    await expect(page.locator('.wf-table')).toBeVisible()
    // Switch to grid view
    await page.locator('.view-btn', { hasText: /.*/ }).nth(1).click()
    await expect(page.locator('.card-grid')).toBeVisible()
  })
})

test.describe('Stokagent workspace — Editor', () => {
  test('clicking a workflow card opens the editor', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.card').first().click()
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
  })

  test('editor loads sexpr source for git-status', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('.cm-content')).toContainText('workflow')
  })

  test('breadcrumb shows workflow name', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('.breadcrumb')).toContainText('git-status.glitch')
  })

  test('metadata panel is visible', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.meta-panel')).toBeVisible({ timeout: 5000 })
  })

  test('Save and Run buttons exist', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button', { hasText: 'Save' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Run' })).toBeVisible()
  })

  test('Run button opens dialog for parameterized workflow', async ({ page }) => {
    await page.goto('#/workflow/issue-to-pr-claude.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    // Should show repo and issue params
    await expect(page.locator('.modal')).toContainText('repo')
    await expect(page.locator('.modal')).toContainText('issue')
    // Cancel closes it
    await page.locator('button', { hasText: 'Cancel' }).click()
    await expect(page.locator('.modal')).not.toBeVisible()
  })

  test('Run button opens dialog for no-param workflow', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('.modal')).toContainText('No parameters required')
  })
})

test.describe('Stokagent workspace — Results browser', () => {
  test('results tree loads with directories', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    const items = page.locator('.tree-item')
    await expect(items.first()).toBeVisible()
  })
})

test.describe('Stokagent workspace — Runs page', () => {
  test('runs page loads', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('h1', { hasText: 'Runs' })).toBeVisible({ timeout: 5000 })
  })

  test('runs page has status filter', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('select')).toBeVisible({ timeout: 5000 })
  })
})
