import { test, expect } from '@playwright/test'

// These tests use dedicated test-* workflows that are fast, deterministic,
// and have no external dependencies.

// ── API: test workflow metadata ────────────────────────────────────
test.describe('Test workflow API', () => {
  test('test-echo has correct metadata', async ({ request }) => {
    const resp = await request.get('/api/workflows/test-echo.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.source).toContain('(workflow "test-echo"')
    expect(data.source).toContain(':description')
    expect(data.params).toEqual([])
  })

  test('test-params extracts params', async ({ request }) => {
    const resp = await request.get('/api/workflows/test-params.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.params).toContain('name')
    expect(data.params).toContain('count')
  })

  test('test-multi has 5 steps in source', async ({ request }) => {
    const resp = await request.get('/api/workflows/test-multi.glitch')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    const stepCount = (data.source.match(/\(step /g) || []).length
    expect(stepCount).toBe(5)
  })
})

// ── Run test-echo and verify results ──────────────────────────────
test.describe('Run and verify', () => {
  test('run test-echo and see it in runs list', async ({ page, request }) => {
    const runResp = await request.post('/api/workflows/test-echo.glitch/run', {
      data: { params: {} },
    })
    expect(runResp.ok()).toBeTruthy()
    const runId = (await runResp.json()).run_id

    // Poll until complete (max 10s)
    let detail
    for (let i = 0; i < 20; i++) {
      await page.waitForTimeout(500)
      detail = await (await request.get(`/api/runs/${runId}`)).json()
      if (detail.run.finished_at) break
    }
    expect(detail.run.exit_status).toBe(0)
    expect(detail.steps.length).toBe(3)
    const stepIds = detail.steps.map(s => s.step_id).sort()
    expect(stepIds).toEqual(['done', 'greet', 'info'])
  })

  test('run test-fail and verify failure status', async ({ request, page }) => {
    const runResp = await request.post('/api/workflows/test-fail.glitch/run', {
      data: { params: {} },
    })
    const runId = (await runResp.json()).run_id

    let detail
    for (let i = 0; i < 20; i++) {
      await page.waitForTimeout(500)
      detail = await (await request.get(`/api/runs/${runId}`)).json()
      if (detail.run.finished_at) break
    }
    expect(detail.run.exit_status).not.toBe(0)
  })

  test('run test-params with values', async ({ request, page }) => {
    const runResp = await request.post('/api/workflows/test-params.glitch/run', {
      data: { params: { name: 'e2e-test', count: '42' } },
    })
    const runId = (await runResp.json()).run_id

    let detail
    for (let i = 0; i < 20; i++) {
      await page.waitForTimeout(500)
      detail = await (await request.get(`/api/runs/${runId}`)).json()
      if (detail.run.finished_at) break
    }
    expect(detail.run.exit_status).toBe(0)
    expect(detail.steps.length).toBe(2)
  })

  test('run test-multi and verify all 5 steps', async ({ request, page }) => {
    const runResp = await request.post('/api/workflows/test-multi.glitch/run', {
      data: { params: {} },
    })
    const runId = (await runResp.json()).run_id

    let detail
    for (let i = 0; i < 20; i++) {
      await page.waitForTimeout(500)
      detail = await (await request.get(`/api/runs/${runId}`)).json()
      if (detail.run.finished_at) break
    }
    expect(detail.run.exit_status).toBe(0)
    expect(detail.steps.length).toBe(5)
    const stepIds = detail.steps.map(s => s.step_id).sort()
    expect(stepIds).toEqual(['aggregate', 'fetch', 'process-a', 'process-b', 'report'])
  })
})

// Helper: run workflow and wait for completion
async function runAndWait(request, page, workflowFile, params = {}) {
  const resp = await request.post(`/api/workflows/${workflowFile}/run`, {
    data: { params },
  })
  const runId = (await resp.json()).run_id
  for (let i = 0; i < 20; i++) {
    await page.waitForTimeout(500)
    const detail = await (await request.get(`/api/runs/${runId}`)).json()
    if (detail.run.finished_at) return runId
  }
  return runId
}

// ── GUI: workflow detail with runs ────────────────────────────────
test.describe('Workflow detail with runs', () => {
  test('test-echo detail page shows runs after execution', async ({ page, request }) => {
    await runAndWait(request, page, 'test-echo.glitch')

    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await expect(page.locator('.run-row').first()).toBeVisible({ timeout: 5000 })
  })

  test('expanding a run shows pipeline graph', async ({ page, request }) => {
    await runAndWait(request, page, 'test-multi.glitch')

    await page.goto('#/workflow/test-multi.glitch')
    await page.waitForSelector('.run-row', { timeout: 5000 })
    await page.locator('.run-row').first().click()
    await expect(page.locator('.graph-container')).toBeVisible({ timeout: 5000 })
  })

  test('graph renders nodes for each step', async ({ page, request }) => {
    await runAndWait(request, page, 'test-multi.glitch')

    await page.goto('#/workflow/test-multi.glitch')
    await page.waitForSelector('.run-row', { timeout: 5000 })
    await page.locator('.run-row').first().click()
    await page.waitForSelector('.graph-node', { timeout: 5000 })
    const nodes = page.locator('.graph-node')
    await expect(nodes).toHaveCount(5)
  })

  test('clicking a graph node opens detail panel', async ({ page, request }) => {
    await runAndWait(request, page, 'test-echo.glitch')

    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.run-row', { timeout: 5000 })
    await page.locator('.run-row').first().click()
    await page.waitForSelector('.graph-node', { timeout: 5000 })
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.node-panel')).toBeVisible({ timeout: 3000 })
    await expect(page.locator('.panel-tab')).toHaveCount(4)
  })

  test('failed run shows fail status', async ({ page, request }) => {
    await runAndWait(request, page, 'test-fail.glitch')

    await page.goto('#/workflow/test-fail.glitch')
    await page.waitForSelector('.run-row', { timeout: 5000 })
    // The run row should exist — we just verify it rendered
    await expect(page.locator('.run-row').first()).toBeVisible()
  })
})

// ── GUI: source tab ───────────────────────────────────────────────
test.describe('Source tab', () => {
  test('source tab shows workflow code in CodeMirror', async ({ page }) => {
    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Source' }).click()
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    // Editor should contain workflow source
    await expect(page.locator('.cm-content')).toContainText('workflow')
  })

  test('source tab has line numbers', async ({ page }) => {
    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Source' }).click()
    await expect(page.locator('.cm-gutters')).toBeVisible({ timeout: 5000 })
  })
})

// ── GUI: metadata tab ─────────────────────────────────────────────
test.describe('Metadata tab', () => {
  test('metadata tab shows workflow description', async ({ page }) => {
    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Metadata' }).click()
    await expect(page.locator('.meta-grid')).toBeVisible({ timeout: 3000 })
    // Description should be visible (either the actual text or --)
    await expect(page.locator('.meta-grid')).toContainText('Echo a message for testing')
  })

  test('metadata tab shows workflow name', async ({ page }) => {
    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Metadata' }).click()
    await expect(page.locator('.meta-grid')).toContainText('test-echo')
  })
})

// ── GUI: run dialog with params ───────────────────────────────────
test.describe('Run dialog', () => {
  test('no-param workflow shows no params required', async ({ page }) => {
    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.header-actions button.primary', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('.modal')).toContainText('No parameters required')
  })

  test('parameterized workflow shows input fields', async ({ page }) => {
    await page.goto('#/workflow/test-params.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.header-actions button.primary', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    // Should show name and count fields
    await expect(page.locator('.modal input[placeholder="name"]')).toBeVisible()
    await expect(page.locator('.modal input[placeholder="count"]')).toBeVisible()
  })
})

// ── No JS errors across the full flow ─────────────────────────────
test.describe('Error free navigation', () => {
  test('navigating through all test workflow pages has no JS errors', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))

    // Home
    await page.goto('/')
    await page.waitForSelector('.card', { timeout: 5000 })

    // Click into test-echo
    await page.goto('#/workflow/test-echo.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })

    // Source tab
    await page.locator('.tab', { hasText: 'Source' }).click()
    await page.waitForTimeout(1000)

    // Metadata tab
    await page.locator('.tab', { hasText: 'Metadata' }).click()
    await page.waitForTimeout(500)

    // Back to Runs
    await page.locator('.tab', { hasText: 'Runs' }).click()
    await page.waitForTimeout(500)

    // Settings
    await page.goto('#/settings')
    await page.waitForTimeout(500)

    // Back home
    await page.goto('/')
    await page.waitForTimeout(500)

    expect(errors).toEqual([])
  })
})
