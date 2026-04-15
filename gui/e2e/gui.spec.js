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

  test('nav bar is visible with all links', async ({ page }) => {
    await page.goto('/')
    const nav = page.locator('nav')
    await expect(nav).toBeVisible()
    await expect(nav).toContainText('gl1tch')
    await expect(nav).toContainText('Workflows')
    await expect(nav).toContainText('Runs')
    await expect(nav).toContainText('Results')
  })
})

// ── Workflow list ───────────────────────────────────────────────────
test.describe('Workflow list', () => {
  test('shows workflows from API', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('text=hello')).toBeVisible({ timeout: 5000 })
  })

  test('shows workflow description', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('text=A simple test workflow')).toBeVisible({ timeout: 5000 })
  })

  test('clicking workflow navigates to editor', async ({ page }) => {
    await page.goto('/')
    await page.locator('text=hello').first().click()
    await expect(page).toHaveURL(/#\/workflow\/hello\.glitch/)
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

  test('Run opens dialog with extracted params', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.dialog')).toBeVisible()
    await expect(page.locator('.dialog')).toContainText('name')
    await expect(page.locator('.dialog input')).toBeVisible()
  })

  test('Cancel closes the run dialog', async ({ page }) => {
    await page.goto('#/workflow/hello.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.dialog')).toBeVisible()
    await page.locator('button', { hasText: 'Cancel' }).click()
    await expect(page.locator('.dialog')).not.toBeVisible()
  })
})

// ── Results browser ─────────────────────────────────────────────────
test.describe('Results browser', () => {
  test('shows top-level results directory', async ({ page }) => {
    await page.goto('#/results')
    await expect(page.locator('text=elastic')).toBeVisible({ timeout: 5000 })
  })

  test('navigates into nested directories', async ({ page }) => {
    await page.goto('#/results')
    await page.locator('text=elastic').click()
    await expect(page.locator('text=observability-robots')).toBeVisible({ timeout: 3000 })
    await page.locator('text=observability-robots').click()
    await expect(page.locator('text=issue-3916')).toBeVisible({ timeout: 3000 })
  })

  test('renders markdown file in preview', async ({ page }) => {
    await page.goto('#/results')
    await page.locator('text=elastic').click()
    await page.locator('text=observability-robots').click()
    await page.locator('text=issue-3916').click()
    await page.locator('text=plan.md').click()
    await expect(page.locator('.preview')).toContainText('Implementation Plan', { timeout: 3000 })
  })

  test('shows JSON file content', async ({ page }) => {
    await page.goto('#/results')
    await page.locator('text=elastic').click()
    await page.locator('text=observability-robots').click()
    await page.locator('text=issue-3916').click()
    await page.locator('text=classification.json').click()
    await expect(page.locator('.preview')).toContainText('documentation', { timeout: 3000 })
  })

  test('breadcrumb .. goes up one level', async ({ page }) => {
    await page.goto('#/results')
    await page.locator('text=elastic').click()
    await expect(page.locator('.breadcrumb')).toContainText('elastic')
    await page.locator('button.link', { hasText: '..' }).click()
    await expect(page.locator('text=elastic')).toBeVisible()
  })
})

// ── Runs page ───────────────────────────────────────────────────────
test.describe('Runs page', () => {
  test('loads without errors', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('h2', { hasText: 'Runs' })).toBeVisible({ timeout: 5000 })
  })
})
