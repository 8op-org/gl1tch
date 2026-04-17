import { test, expect } from '@playwright/test'

async function runAndWait(request, workflow, params = {}, maxWaitMs = 15000) {
  const resp = await request.post(`/api/workflows/${workflow}/run`, {
    data: { params },
  })
  expect(resp.ok()).toBeTruthy()
  const { run_id } = await resp.json()

  const start = Date.now()
  let detail
  while (Date.now() - start < maxWaitMs) {
    const r = await request.get(`/api/runs/${run_id}`)
    detail = await r.json()
    if (detail.run?.finished_at) break
    await new Promise(r => setTimeout(r, 500))
  }
  return { runId: run_id, detail }
}

test.describe('Run Detail Page', () => {
  test('navigates to /run/:id and shows header', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.run-header')).toBeVisible()
    await expect(page.locator('.breadcrumb')).toContainText('Run')
  })

  test('shows metadata strip with duration and started', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.metadata-strip')).toBeVisible()
    await expect(page.locator('.meta-label').filter({ hasText: 'Duration' })).toBeVisible()
    await expect(page.locator('.meta-label').filter({ hasText: 'Started' })).toBeVisible()
  })

  test('renders pipeline graph with nodes', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    const nodeCount = await page.locator('.graph-node').count()
    expect(nodeCount).toBeGreaterThanOrEqual(2)
  })

  test('clicking a node opens slide-over panel', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.node-panel')).toBeVisible()
    await expect(page.locator('.panel-title')).toBeVisible()
  })

  test('panel shows Output tab by default', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.panel-tab.active')).toContainText('Output')
  })

  test('panel tabs switch content', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()
    await page.locator('.panel-tab').filter({ hasText: 'Metrics' }).click()
    await expect(page.locator('.metrics-grid')).toBeVisible()
    await page.locator('.panel-tab').filter({ hasText: 'Prompt' }).click()
    await expect(page.locator('.panel-content')).toBeVisible()
  })

  test('close panel via X button', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.node-panel')).toBeVisible()
    await page.locator('.close-btn').click()
    await expect(page.locator('.node-panel')).not.toBeVisible()
  })

  test('close panel via Escape key', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.node-panel')).toBeVisible()
    await page.keyboard.press('Escape')
    await expect(page.locator('.node-panel')).not.toBeVisible()
  })

  test('clicking different node swaps panel content', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    const nodes = page.locator('.graph-node')
    const nodeCount = await nodes.count()
    if (nodeCount < 2) return
    await nodes.first().click()
    const firstTitle = await page.locator('.panel-title').textContent()
    await nodes.nth(1).click()
    const secondTitle = await page.locator('.panel-title').textContent()
    expect(firstTitle).not.toEqual(secondTitle)
  })

  test('breadcrumb links navigate back', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.breadcrumb')).toBeVisible()
    await page.locator('.breadcrumb a').first().click()
    await expect(page).toHaveURL(/\/#\/$/)
  })

  test('graph edges are visible', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    const edges = page.locator('svg path[stroke-width="2"]')
    const count = await edges.count()
    expect(count).toBeGreaterThan(0)
  })

  test('zoom reset button is visible', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')
    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await expect(page.locator('.zoom-reset')).toBeVisible()
  })

  test('workflow detail run rows navigate to /run/:id', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/workflow/test-echo.glitch`)
    await expect(page.locator('.run-row').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.run-row').first().click()
    await expect(page).toHaveURL(/\/#\/run\/\d+/)
  })
})
