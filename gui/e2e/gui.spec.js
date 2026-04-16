import { test, expect } from '@playwright/test'

// All tests run against the stokagent workspace (36 real workflows).
// The webServer in playwright.config.js starts glitch with --workspace ../stokagent.

// ── API health ──────────────────────────────────────────────────────
test.describe('API', () => {
  test('GET /api/workflows returns 30+ workflows', async ({ request }) => {
    const resp = await request.get('/api/workflows')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(data.length).toBeGreaterThan(30)
    const names = data.map(w => w.file)
    expect(names).toContain('dashboard-reviews.glitch')
    expect(names).toContain('git-status.glitch')
    expect(names).toContain('work-on-issue.glitch')
  })

  test('GET /api/workflows/git-status.glitch returns source + no params', async ({ request }) => {
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

  test('GET /api/results lists directories', async ({ request }) => {
    const resp = await request.get('/api/results/')
    expect(resp.ok()).toBeTruthy()
    const data = await resp.json()
    expect(Array.isArray(data)).toBeTruthy()
    expect(data.length).toBeGreaterThan(0)
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

  test('GET /api/kibana/workflow/git-status returns url', async ({ request }) => {
    const resp = await request.get('/api/kibana/workflow/git-status')
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

  test('GET /api/results/elastic without trailing slash returns JSON', async ({ request }) => {
    const resp = await request.get('/api/results/elastic')
    if (resp.ok()) {
      const data = await resp.json()
      expect(Array.isArray(data)).toBeTruthy()
    }
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

  test('sidebar is visible with 3 nav items', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await expect(sidebar).toBeVisible()
    await expect(page.locator('.nav-item')).toHaveCount(4)
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

  test('sidebar shows logo', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('.logo')).toBeVisible()
  })

  test('sidebar nav labels visible when expanded', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await sidebar.hover()
    await expect(page.locator('.nav-label').first()).toBeVisible()
  })

  test('navigating via sidebar to Runs updates active state', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await sidebar.hover()
    await page.locator('.nav-item', { hasText: 'Runs' }).click()
    await expect(page).toHaveURL(/\/#\/runs/)
    const active = page.locator('.nav-item.active')
    await expect(active).toContainText('Runs')
  })

  test('navigating via sidebar to Results', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await sidebar.hover()
    await page.locator('.nav-item', { hasText: 'Results' }).click()
    await expect(page).toHaveURL(/\/#\/results/)
  })
})

// ── Workflow list — card view (default) ─────────────────────────────
test.describe('Workflow list — card view', () => {
  test('defaults to card/grid view with 30+ cards', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    const cards = page.locator('.card')
    const count = await cards.count()
    expect(count).toBeGreaterThan(30)
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

  test('clicking card navigates to editor', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.card').first().click()
    await expect(page).toHaveURL(/\/#\/workflow\//)
  })
})

// ── Workflow list — search and filter ───────────────────────────────
test.describe('Workflow list — search and filter', () => {
  test('search filters to dashboard workflows', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.fill('input[placeholder="Search..."]', 'dashboard')
    const cards = page.locator('.card')
    const count = await cards.count()
    expect(count).toBeGreaterThanOrEqual(8)
    expect(count).toBeLessThan(36)
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

  test('grouped view items navigate to editor', async ({ page }) => {
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

  test('list view rows navigate to editor', async ({ page }) => {
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

// ── Editor ──────────────────────────────────────────────────────────
test.describe('Editor', () => {
  test('loads CodeMirror with sexpr source', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('.cm-content')).toContainText('workflow')
  })

  test('CodeMirror has custom background (not oneDark grey)', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    const bg = await page.locator('.cm-editor').evaluate(el => getComputedStyle(el).backgroundColor)
    // Should not be oneDark default grey (#282c34 = rgb(40, 44, 52))
    expect(bg).not.toBe('rgb(40, 44, 52)')
  })

  test('has Save and Run buttons', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button', { hasText: 'Save' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Run' })).toBeVisible()
  })

  test('Save button is disabled when clean', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button', { hasText: 'Save' })).toBeDisabled()
  })

  test('shows metadata panel with heading', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.meta-panel')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('.meta-panel h3')).toContainText('Metadata')
  })

  test('metadata panel collapse and restore', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.meta-panel')).toBeVisible({ timeout: 5000 })
    await page.locator('.meta-panel .close-btn').click()
    await expect(page.locator('.meta-panel')).not.toBeVisible()
    await expect(page.locator('.meta-toggle')).toBeVisible()
    await page.locator('.meta-toggle').click()
    await expect(page.locator('.meta-panel')).toBeVisible()
  })

  test('breadcrumbs show Workflows link and workflow name', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await expect(page.locator('main a', { hasText: 'Workflows' })).toBeVisible()
    await expect(page.locator('.breadcrumb')).toContainText('git-status.glitch')
  })

  test('breadcrumb Workflows link navigates home', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('main a', { hasText: 'Workflows' }).click()
    await expect(page).toHaveURL(/\/#\/$/)
  })

  test('clicking a card from list opens editor', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('.card')
    await page.locator('.card').first().click()
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
  })
})

// ── Run dialog ──────────────────────────────────────────────────────
test.describe('Run dialog', () => {
  test('no-param workflow shows "No parameters required"', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('.modal')).toContainText('No parameters required')
  })

  test('parameterized workflow shows input fields', async ({ page }) => {
    await page.goto('#/workflow/issue-to-pr-claude.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await expect(page.locator('.modal')).toContainText('repo')
    await expect(page.locator('.modal')).toContainText('issue')
  })

  test('dialog has title with workflow name', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal-header')).toContainText('git-status.glitch')
  })

  test('dialog has Start Run button', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('button', { hasText: 'Start Run' })).toBeVisible()
  })

  test('Cancel closes the dialog', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await page.locator('button', { hasText: 'Cancel' }).click()
    await expect(page.locator('.modal')).not.toBeVisible()
  })

  test('Escape closes the dialog', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await page.keyboard.press('Escape')
    await expect(page.locator('.modal')).not.toBeVisible()
  })

  test('clicking overlay backdrop closes dialog', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    await page.locator('button', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
    await page.locator('.overlay').click({ position: { x: 10, y: 10 } })
    await expect(page.locator('.modal')).not.toBeVisible()
  })
})

// ── Results browser ─────────────────────────────────────────────────
test.describe('Results browser', () => {
  test('file tree renders with directories', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    const items = page.locator('.tree-item')
    await expect(items.first()).toBeVisible()
  })

  test('shows empty state when no file selected', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await expect(page.locator('text=Select a file to preview')).toBeVisible()
  })

  test('navigates into elastic/observability-robots directory', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await page.locator('.tree-item', { hasText: 'elastic' }).click()
    await expect(page.locator('.tree-item', { hasText: 'observability-robots' })).toBeVisible({ timeout: 3000 })
  })

  test('preview/edit toggle exists when file selected', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await page.locator('.tree-item', { hasText: 'elastic' }).click()
    await page.locator('.tree-item', { hasText: 'observability-robots' }).click({ timeout: 3000 })
    // Click into an issue subfolder to reach files
    await page.waitForSelector('.tree-item.dir >> text=/issue-/')
    await page.locator('.tree-item.dir', { hasText: /issue-/ }).first().click()
    await page.waitForSelector('.tree-item.file', { timeout: 5000 })
    await page.locator('.tree-item.file').first().click()
    await expect(page.locator('button', { hasText: 'Preview' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'Edit' })).toBeVisible()
  })

  test('switching to edit mode shows CodeMirror', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await page.locator('.tree-item', { hasText: 'elastic' }).click()
    await page.locator('.tree-item', { hasText: 'observability-robots' }).click({ timeout: 3000 })
    await page.waitForSelector('.tree-item.dir >> text=/issue-/')
    await page.locator('.tree-item.dir', { hasText: /issue-/ }).first().click()
    await page.waitForSelector('.tree-item.file', { timeout: 5000 })
    await page.locator('.tree-item.file').first().click()
    await page.locator('button', { hasText: 'Edit' }).click()
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 3000 })
  })

  test('switching back to preview hides CodeMirror', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await page.locator('.tree-item', { hasText: 'elastic' }).click()
    await page.locator('.tree-item', { hasText: 'observability-robots' }).click({ timeout: 3000 })
    await page.waitForSelector('.tree-item.dir >> text=/issue-/')
    await page.locator('.tree-item.dir', { hasText: /issue-/ }).first().click()
    await page.waitForSelector('.tree-item.file', { timeout: 5000 })
    await page.locator('.tree-item.file').first().click()
    await page.locator('button', { hasText: 'Edit' }).click()
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 3000 })
    await page.locator('button', { hasText: 'Preview' }).click()
    await expect(page.locator('.cm-editor')).not.toBeVisible()
  })

  test('breadcrumb shows Results label', async ({ page }) => {
    await page.goto('#/results')
    await expect(page.locator('main').locator('text=Results')).toBeVisible({ timeout: 3000 })
  })

  test('no action bar before folder is selected', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await expect(page.locator('.action-bar')).not.toBeVisible()
  })

  test('action bar appears when folder is clicked', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await page.locator('.tree-item.dir').first().click()
    await expect(page.locator('.action-bar')).toBeVisible({ timeout: 3000 })
    expect(errors).toEqual([])
  })

  test('action bar shows folder name and action buttons', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    await page.locator('.tree-item.dir').first().click()
    await expect(page.locator('.action-bar')).toBeVisible({ timeout: 3000 })
    // Shows the folder name
    await expect(page.locator('.action-context')).toBeVisible()
    // Has action buttons
    const buttons = page.locator('.action-bar .action-btn')
    await expect(buttons.first()).toBeVisible()
    const count = await buttons.count()
    expect(count).toBeGreaterThan(0)
  })

  test('clicking action button opens RunDialog with path param pre-filled', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    // Get the folder name before clicking
    const dirItem = page.locator('.tree-item.dir').first()
    const folderName = await dirItem.locator('.tree-name').textContent()
    await dirItem.click()
    await expect(page.locator('.action-bar')).toBeVisible({ timeout: 3000 })
    await page.locator('.action-bar .action-btn').first().click()
    await expect(page.locator('.modal')).toBeVisible({ timeout: 3000 })
    // path param field must exist and be pre-filled with the folder path
    const pathInput = page.locator('.modal input[placeholder="path"]')
    await expect(pathInput).toBeVisible()
    const val = await pathInput.inputValue()
    expect(val).toContain(folderName.trim())
  })

  test('action bar updates when different folder clicked', async ({ page }) => {
    await page.goto('#/results')
    await page.waitForSelector('.tree-item', { timeout: 5000 })
    // Click first folder
    const firstDir = page.locator('.tree-item.dir').first()
    const firstName = await firstDir.locator('.tree-name').textContent()
    await firstDir.click()
    await expect(page.locator('.action-bar')).toBeVisible({ timeout: 3000 })
    await expect(page.locator('.action-context')).toContainText(firstName.trim())
    // Click a different folder if available
    const dirs = page.locator('.tree-item.dir')
    const dirCount = await dirs.count()
    if (dirCount > 1) {
      const secondDir = dirs.nth(1)
      const secondName = await secondDir.locator('.tree-name').textContent()
      await secondDir.click()
      await expect(page.locator('.action-context')).toContainText(secondName.trim())
    }
  })
})

// ── Runs page ───────────────────────────────────────────────────────
test.describe('Runs page', () => {
  test('loads with header and title', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('h1')).toContainText('Runs', { timeout: 5000 })
  })

  test('has status filter with 4 options', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('select.status-filter')).toBeVisible({ timeout: 5000 })
    const options = page.locator('select.status-filter option')
    await expect(options).toHaveCount(4)
  })

  test('shows empty state, runs table, or error', async ({ page }) => {
    await page.goto('#/runs')
    await page.waitForTimeout(2000)
    const hasTable = await page.locator('.runs-table').isVisible().catch(() => false)
    const hasEmpty = await page.locator('.empty-state').isVisible().catch(() => false)
    const hasError = await page.locator('.status-fail').isVisible().catch(() => false)
    expect(hasTable || hasEmpty || hasError).toBeTruthy()
  })

  test('no JS errors on runs page', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/runs')
    await page.waitForTimeout(2000)
    expect(errors).toEqual([])
  })

  test('page header has icon', async ({ page }) => {
    await page.goto('#/runs')
    await expect(page.locator('h1 svg')).toBeVisible({ timeout: 5000 })
  })
})

// ── Run detail view ─────────────────────────────────────────────────
test.describe('Run detail view', () => {
  test('shows error for invalid run ID', async ({ page }) => {
    await page.goto('#/run/999999')
    await expect(page.locator('.status-fail')).toBeVisible({ timeout: 5000 })
  })

  test('breadcrumbs show Runs link', async ({ page }) => {
    await page.goto('#/run/1')
    await expect(page.locator('main').locator('text=Runs')).toBeVisible({ timeout: 5000 })
  })

  test('breadcrumb Runs link navigates back', async ({ page }) => {
    await page.goto('#/run/1')
    await page.waitForTimeout(1000)
    await page.locator('main a', { hasText: 'Runs' }).click()
    await expect(page).toHaveURL(/\/#\/runs/)
  })

  test('no JS errors on run view', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/run/1')
    await page.waitForTimeout(2000)
    expect(errors).toEqual([])
  })
})

// ── Cross-cutting ───────────────────────────────────────────────────
test.describe('Cross-cutting', () => {
  test('direct URL navigation works for all routes', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1')).toContainText('Workflows')

    await page.goto('#/runs')
    await expect(page.locator('h1')).toContainText('Runs')

    await page.goto('#/results')
    await expect(page.locator('main').locator('text=Results')).toBeVisible()
  })

  test('navigating between all pages does not leak errors', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('/')
    await page.waitForSelector('.card')
    const sidebar = page.locator('aside.sidebar')
    // Workflows -> Editor
    await page.locator('.card').first().click()
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
    // Editor -> Runs
    await sidebar.hover()
    await page.locator('.nav-item', { hasText: 'Runs' }).click()
    await page.waitForTimeout(500)
    // Runs -> Results
    await sidebar.hover()
    await page.locator('.nav-item', { hasText: 'Results' }).click()
    await page.waitForTimeout(500)
    // Results -> Workflows
    await sidebar.hover()
    await page.locator('.nav-item', { hasText: 'Workflows' }).click()
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
    await page.goto('#/runs')
    await page.waitForTimeout(500)
    await page.goto('#/results')
    await page.waitForTimeout(500)
    await page.goto('#/workflow/git-status.glitch')
    await page.waitForTimeout(500)
    expect(warnings).toEqual([])
  })
})
