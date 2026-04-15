import { test, expect } from '@playwright/test'

// ── API health ──────────────────────────────────────────────────────
test.describe('API', () => {
  test('GET /api/workflows returns workflow list', async ({ request }) => {
    const resp = await request.get('/api/workflows')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(Array.isArray(data)).toBeTruthy()
    expect(data.length).toBeGreaterThan(0)
    expect(data[0]).toHaveProperty('name')
    expect(data[0]).toHaveProperty('file')
  })

  test('GET /api/workflows/hello.glitch returns source + params', async ({ request }) => {
    const resp = await request.get('/api/workflows/hello.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.source).toContain('workflow')
    expect(data.params).toContain('name')
  })

  test('GET /api/results/elastic lists directory', async ({ request }) => {
    const resp = await request.get('/api/results/elastic')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data[0]).toHaveProperty('is_dir')
  })

  test('GET /api/kibana/workflow/hello returns url', async ({ request }) => {
    const resp = await request.get('/api/kibana/workflow/hello')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.url).toContain('localhost:5601')
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

  test('sidebar is visible with nav items', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await expect(sidebar).toBeVisible()
    await expect(page.locator('.nav-item')).toHaveCount(3)
  })

  test('sidebar expands on hover', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await sidebar.hover()
    await expect(sidebar).toHaveClass(/expanded/)
  })

  test('active nav item is highlighted', async ({ page }) => {
    await page.goto('/')
    const active = page.locator('.nav-item.active')
    await expect(active).toBeVisible()
  })
})

// ── Workflow list ───────────────────────────────────────────────────
test.describe('Workflow list', () => {
  test('renders workflow cards', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const cards = page.locator('.card')
    await expect(cards.first()).toBeVisible()
  })

  test('shows workflow description', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('text=A simple test workflow')).toBeVisible({ timeout: 5000 })
  })

  test('search filters workflows', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.fill('input[placeholder="Search..."]', 'nonexistent-workflow-xyz')
    await expect(page.locator('.card')).toHaveCount(0)
  })

  test('clicking card navigates to editor', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.card').first().click()
    await expect(page).toHaveURL(/\/#\/workflow\//)
  })
})

// ── Editor ──────────────────────────────────────────────────────────
test.describe('Editor', () => {
  test('loads CodeMirror with workflow source', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('.cm-content')).toContainText('workflow')
  })

  test('has Save and Run buttons', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button', { hasText: 'Save' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Run' })).toBeVisible()
  })

  test('shows metadata panel', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.meta-panel')).toBeVisible({ timeout: 5000 })
  })

  test('metadata panel can be collapsed', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.meta-panel')).toBeVisible({ timeout: 5000 })
    await page.locator('.meta-panel .close-btn').click()
    await expect(page.locator('.meta-panel')).not.toBeVisible()
  })

  test('Run opens dialog with extracted params', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('.modal')).toContainText('name')
    await expect(page.locator('.modal input')).toBeVisible()
  })

  test('Cancel closes the run dialog', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await page.locator('button', { hasText: 'Cancel' }).click()
    await expect(page.locator('.modal')).not.toBeVisible()
  })
})

// ── Results browser ─────────────────────────────────────────────────
test.describe('Results browser', () => {
  test('file tree renders directories', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item')
    const items = page.locator('.tree-item')
    await expect(items.first()).toBeVisible()
  })

  test('navigates into nested directories', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item')
    await page.locator('.tree-item', { hasText: 'elastic' }).click()
    await expect(page.locator('.tree-item', { hasText: 'observability-robots' })).toBeVisible({ timeout: 3000 })
  })

  test('preview/edit toggle exists when file selected', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item')
    // Navigate to a file
    await page.locator('.tree-item', { hasText: 'elastic' }).click()
    await page.locator('.tree-item', { hasText: 'observability-robots' }).click()
    await page.locator('.tree-item', { hasText: 'issue-3916' }).click()
    await page.waitForSelector('.tree-item.file')
    await page.locator('.tree-item.file').first().click()
    await expect(page.locator('button', { hasText: 'Preview' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Edit' })).toBeVisible()
  })
})

// ── Runs page ───────────────────────────────────────────────────────
test.describe('Runs page', () => {
  test('loads without errors', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('h1', { hasText: 'Runs' })).toBeVisible({ timeout: 5000 })
  })

  test('has status filter', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('select')).toBeVisible({ timeout: 5000 })
  })
})
