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

  test('artifacts API serves files with results/ prefix fallback', async ({ request }) => {
    // Run demo-run-detail which has a save step producing results/demo/health-report.md
    const { runId, detail } = await runAndWait(request, 'demo-run-detail.glitch', {}, 90000)
    expect(detail.run.exit_status).toBe(0)

    // Find the save step with artifacts
    const saveStep = detail.steps.find(s => s.artifacts?.length > 0)
    expect(saveStep).toBeTruthy()
    const artifactPath = saveStep.artifacts[0]
    expect(artifactPath).toContain('health-report.md')

    // The API should serve the file via the fallback path
    const fileResp = await request.get(`/api/results/${artifactPath}`)
    expect(fileResp.ok()).toBeTruthy()
    const content = await fileResp.text()
    expect(content).toContain('Health')
  })

  test('files tab shows artifact and loads content', async ({ page, request }) => {
    const { runId, detail } = await runAndWait(request, 'demo-run-detail.glitch', {}, 90000)

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })

    // Click the save-report node (it has artifacts)
    const saveNode = page.locator('.graph-node').filter({ hasText: 'save-report' })
    if (await saveNode.count() > 0) {
      await saveNode.click()
      await expect(page.locator('.node-panel')).toBeVisible()

      // Switch to Files tab
      await page.locator('.panel-tab').filter({ hasText: 'Files' }).click()

      // Should show the file item
      await expect(page.locator('.file-item').first()).toBeVisible({ timeout: 5000 })

      // Click file to view content
      await page.locator('.file-item').first().click()
      await expect(page.locator('.file-content')).toBeVisible({ timeout: 5000 })
    }
  })

  test('metadata strip shows aggregated model and tokens', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'demo-run-detail.glitch', {}, 90000)

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.metadata-strip')).toBeVisible()

    // Model should show qwen2.5:7b (aggregated from the analyze step)
    await expect(page.locator('.meta-val').filter({ hasText: 'qwen2.5:7b' })).toBeVisible({ timeout: 5000 })

    // Tokens should show non-zero values
    const tokensVal = page.locator('.meta-pill').filter({ hasText: 'Tokens' }).locator('.meta-val')
    await expect(tokensVal).not.toHaveText('--')
  })

  test('live run shows running state then completes', async ({ page, request }) => {
    // Start a run and immediately navigate to it
    const resp = await request.post('/api/workflows/test-multi.glitch/run', {
      data: { params: {} },
    })
    const { run_id } = await resp.json()

    await page.goto(`/#/run/${run_id}`)
    await expect(page.locator('.run-detail-page')).toBeVisible()

    // Should eventually show graph nodes (run completes fast for test workflows)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 15000 })

    // Status should eventually show PASS
    await expect(page.locator('.badge').filter({ hasText: 'PASS' })).toBeVisible({ timeout: 15000 })
  })

  test('workflow detail run rows navigate to /run/:id', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')
    await page.goto(`/#/workflow/test-echo.glitch`)
    await expect(page.locator('.run-row').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.run-row').first().click()
    await expect(page).toHaveURL(/\/#\/run\/\d+/)
  })
})
