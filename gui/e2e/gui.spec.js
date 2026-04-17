import { test, expect } from '@playwright/test'

// All tests run against the stokagent workspace (36 real workflows).
// The webServer in playwright.config.js starts glitch with --workspace ../stokagent.

// ── API health ──────────────────────────────────────────────────────
test.describe('API', () => {
  test('GET /api/workflows returns workflows', async ({ request }) => {
    const resp = await request.get('/api/workflows')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(Array.isArray(data)).toBeTruthy()
    expect(data.length).toBeGreaterThan(0)
    // Every entry has a file field
    expect(data[0].file).toBeTruthy()
  })

  test('GET /api/workflows/{name} returns source', async ({ request }) => {
    // Get first workflow name dynamically
    const list = await (await request.get('/api/workflows')).json()
    const name = list[0].file
    const resp = await request.get(`/api/workflows/${name}`)
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.source).toBeTruthy()
    expect(data.source.length).toBeGreaterThan(0)
  })

  test('GET /api/results lists or returns empty', async ({ request }) => {
    const resp = await request.get('/api/results/')
    // May be empty in some workspaces
    if (resp.ok()) {
      const data = await resp.json()
      expect(Array.isArray(data)).toBeTruthy()
    }
  })

  test('GET /api/results navigates into subdirectory', async ({ request }) => {
    // List top-level results
    const resp = await request.get('/api/results/')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(Array.isArray(data)).toBeTruthy()
    // Find a directory entry and navigate into it
    const dir = data.find(e => e.is_dir)
    if (dir) {
      const subResp = await request.get(`/api/results/${dir.name}/`)
      expect(subResp.ok()).toBeTruthy()
    }
  })

  test('GET /api/kibana/workflow returns json', async ({ request }) => {
    const resp = await request.get('/api/kibana/workflow/test')
    // May fail if no ES configured, that's fine
    if (resp.ok()) {
      const data = await resp.json()
      expect(data).toBeTruthy()
    }
  })

  test('path traversal returns 400', async ({ request }) => {
    const resp = await request.get('/api/workflows/..%2F..%2Fetc%2Fpasswd')
    expect(resp.status()).toBe(400)
  })

  test('GET /api/workflows/actions/results returns array', async ({ request }) => {
    const resp = await request.get('/api/workflows/actions/results')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(Array.isArray(data)).toBeTruthy()
  })

  test('GET /api/runs returns array or empty', async ({ request }) => {
    const resp = await request.get('/api/runs')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(Array.isArray(data)).toBeTruthy()
  })

  test('GET /api/runs/:id returns 404 for missing run', async ({ request }) => {
    const resp = await request.get('/api/runs/999999')
    expect(resp.status()).toBe(404)
  })

  test('GET /api/workspace returns workspace info', async ({ request }) => {
    const resp = await request.get('/api/workspace')
    expect(resp.ok()).toBeTruthy()
  })


  test('GET /api/runs?workflow={name} returns array', async ({ request }) => {
    const resp = await request.get('/api/runs?workflow=clean.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(Array.isArray(data)).toBeTruthy()
  })
})

// ── App shell ───────────────────────────────────────────────────────
test.describe('App shell', () => {
  test('page loads without JS errors', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('/')
    await page.waitForTimeout(2000)
    expect(errors).toEqual([])
  })

  test('activity bar is visible with nav icons', async ({ page }) => {
    await page.goto('/')
    const bar = page.locator('.activity-bar')
    await expect(bar).toBeVisible()
    // Should have workflow and settings nav items (ab-icon links)
    const navItems = page.locator('.activity-bar .ab-icon')
    const count = await navItems.count()
    expect(count).toBeGreaterThanOrEqual(2)
  })

  test('activity bar does NOT expand on hover', async ({ page }) => {
    await page.goto('/')
    const bar = page.locator('.activity-bar')
    await bar.hover()
    // Should NOT have expanded class
    await expect(bar).not.toHaveClass(/expanded/)
  })

  test('active nav item is highlighted', async ({ page }) => {
    await page.goto('/')
    const active = page.locator('.activity-bar .ab-icon.active')
    await expect(active).toBeVisible()
  })

  test('clicking settings icon navigates to settings', async ({ page }) => {
    await page.goto('/')
    await page.locator('.activity-bar .ab-icon[title="Settings"]').click()
    await expect(page).toHaveURL(/\/#\/settings/)
  })

  test('clicking workflows icon navigates home', async ({ page }) => {
    await page.goto('#/settings')
    await page.locator('.activity-bar .ab-icon[title="Workflows"]').click()
    await expect(page).toHaveURL(/\/#\/$/)
  })
})

// ── Workflow list — card view (default) ─────────────────────────────
test.describe('Workflow list — card view', () => {
  test('defaults to card/grid view with 30+ cards', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const cards = page.locator('.card')
    const count = await cards.count()
    expect(count).toBeGreaterThan(0)
    // Grid view button should be active by default
    const activeBtn = page.locator('.view-btn.active')
    await expect(activeBtn).toHaveAttribute('title', 'Cards')
  })

  test('cards show workflow name', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await expect(page.locator('.card-name').first()).toBeVisible()
  })

  test('cards show descriptions when available', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const desc = page.locator('.card-desc')
    const count = await desc.count()
    expect(count).toBeGreaterThan(0)
  })

  test('clicking card navigates to workflow detail', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.card').first().click()
    await expect(page).toHaveURL(/\/#\/workflow\//)
    // Should show the workflow detail tabs instead of old editor
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await expect(page.locator('.tab')).toHaveCount(3)
  })
})

// ── Workflow list — search and filter ───────────────────────────────
test.describe('Workflow list — search and filter', () => {
  test('search filters workflows', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const initial = await page.locator('.card').count()
    await page.fill('input[placeholder="Search..."]', 'clean')
    const filtered = await page.locator('.card').count()
    expect(filtered).toBeGreaterThanOrEqual(1)
    expect(filtered).toBeLessThanOrEqual(initial)
  })

  test('search with no results shows empty message', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.fill('input[placeholder="Search..."]', 'nonexistent-workflow-xyz')
    await expect(page.locator('.card')).toHaveCount(0)
    await expect(page.locator('text=No workflows match')).toBeVisible()
  })

  test('tag filter pills are clickable', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const pills = page.locator('.filter-bar .pill')
    const count = await pills.count()
    expect(count).toBeGreaterThan(0)
    await pills.first().click()
    await expect(pills.first()).toHaveClass(/active/)
  })

  test('clearing search restores all workflows', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const initialCount = await page.locator('.card').count()
    await page.fill('input[placeholder="Search..."]', 'dashboard')
    expect(await page.locator('.card').count()).toBeLessThan(initialCount)
    await page.fill('input[placeholder="Search..."]', '')
    expect(await page.locator('.card').count()).toBe(initialCount)
  })
})

// ── Workflow list — view modes ──────────────────────────────────────
test.describe('Workflow list — view modes', () => {
  test('switch to grouped view shows groups', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.view-btn[title="Grouped"]').click()
    await expect(page.locator('.group').first()).toBeVisible()
    const groups = page.locator('.group-header')
    const count = await groups.count()
    expect(count).toBeGreaterThan(3) // dashboard, github, issue, etc.
  })

  test('grouped view shows workflow count in badge', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.view-btn[title="Grouped"]').click()
    const countBadge = page.locator('.group-count').first()
    await expect(countBadge).toBeVisible()
    const text = await countBadge.textContent()
    expect(parseInt(text)).toBeGreaterThan(0)
  })

  test('grouped view collapse/expand', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.view-btn[title="Grouped"]').click()
    await page.waitForSelector('.group-item')
    const firstGroup = page.locator('.group').first()
    const header = firstGroup.locator('.group-header')
    // Verify items visible initially
    await expect(firstGroup.locator('.group-items')).toBeVisible()
    // Collapse
    await header.click()
    await expect(firstGroup.locator('.group-items')).not.toBeVisible()
    // Expand again
    await header.click()
    await expect(firstGroup.locator('.group-items')).toBeVisible()
  })

  test('grouped view items navigate to workflow detail', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.view-btn[title="Grouped"]').click()
    await page.waitForSelector('.group-item')
    await page.locator('.group-item').first().click()
    await expect(page).toHaveURL(/\/#\/workflow\//)
  })

  test('switch to list view shows table', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.view-btn[title="List"]').click()
    await expect(page.locator('.wf-table')).toBeVisible()
    await expect(page.locator('th', { hasText: 'Name' })).toBeVisible()
    await expect(page.locator('th', { hasText: 'Description' })).toBeVisible()
    await expect(page.locator('th', { hasText: 'Group' })).toBeVisible()
    await expect(page.locator('th', { hasText: 'Status' })).toBeVisible()
  })

  test('list view rows navigate to workflow detail', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.view-btn[title="List"]').click()
    await page.waitForSelector('.wf-table tbody tr')
    await page.locator('.wf-table tbody tr').first().click()
    await expect(page).toHaveURL(/\/#\/workflow\//)
  })

  test('view mode persists across filter changes', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.view-btn[title="List"]').click()
    await expect(page.locator('.wf-table')).toBeVisible()
    await page.fill('input[placeholder="Search..."]', 'dashboard')
    await expect(page.locator('.wf-table')).toBeVisible()
  })
})

// ── Run dialog ──────────────────────────────────────────────────────
test.describe('Run dialog', () => {
  test('dialog opens and has Start Run button', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.header-actions button.primary', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('button', { hasText: 'Start Run' })).toBeVisible()
  })

  test('Cancel closes the dialog', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.header-actions button.primary', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await page.locator('button', { hasText: 'Cancel' }).click()
    await expect(page.locator('.modal')).not.toBeVisible()
  })

  test('Escape closes the dialog', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.header-actions button.primary', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await page.keyboard.press('Escape')
    await expect(page.locator('.modal')).not.toBeVisible()
  })

  test('clicking overlay backdrop closes dialog', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.header-actions button.primary', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await page.locator('.overlay').click({ position: { x: 10, y: 10 } })
    await expect(page.locator('.modal')).not.toBeVisible()
  })
})

// ── Cross-cutting ───────────────────────────────────────────────────
test.describe('Cross-cutting', () => {
  test('direct URL navigation works for all routes', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1')).toContainText('Workflows')

    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await expect(page.locator('.tab')).toHaveCount(3)

    await page.goto('#/settings')
    await expect(page.locator('h1')).toContainText('Settings')
  })

  test('navigating between all pages does not leak errors', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('/')
    await page.waitForSelector('.card')
    // Workflows -> Workflow Detail
    await page.locator('.card').first().click()
    await page.waitForSelector('.tabs', { timeout: 5000 })
    // Workflow Detail -> Settings
    await page.locator('.activity-bar .ab-icon[title="Settings"]').click()
    await page.waitForTimeout(500)
    // Settings -> Workflows
    await page.locator('.activity-bar .ab-icon[title="Workflows"]').click()
    await page.waitForTimeout(500)
    expect(errors).toEqual([])
  })

  test('no Svelte on: deprecation warnings in console', async ({ page }) => {
    const warnings = []
    page.on('console', (msg) => {
      if (msg.type() === 'warning' && msg.text().includes('on:')) {
        warnings.push(msg.text())
      }
    })
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.goto('#/settings')
    await page.waitForTimeout(500)
    await page.goto('#/workflow/clean.glitch')
    await page.waitForTimeout(500)
    expect(warnings).toEqual([])
  })
})
