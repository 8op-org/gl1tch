import { test, expect } from '@playwright/test'

// ── Settings page ──────────────────────────────────────────────────
test.describe('Settings page', () => {
  test('navigates to settings via sidebar', async ({ page }) => {
    await page.goto('/')
    const sidebar = page.locator('aside.sidebar')
    await sidebar.hover()
    await page.locator('.sidebar-footer .nav-item', { hasText: 'Settings' }).click()
    await expect(page).toHaveURL(/\/#\/settings/)
  })

  test('sidebar settings link shows active state', async ({ page }) => {
    await page.goto('#/settings')
    await page.waitForTimeout(500)
    const settingsLink = page.locator('.sidebar-footer .nav-item')
    await expect(settingsLink).toHaveClass(/active/)
  })

  test('page loads without JS errors', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/settings')
    await page.waitForTimeout(2000)
    expect(errors).toEqual([])
  })

  test('shows Workflow Defaults section', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  })

  test('shows Workspace section', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workspace' })).toBeVisible({ timeout: 5000 })
  })

  test('displays current workspace name', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    const nameInput = page.locator('input[placeholder="workspace name"]')
    await expect(nameInput).toBeVisible()
    const val = await nameInput.inputValue()
    expect(val.length).toBeGreaterThan(0)
  })

  test('displays default model input', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await expect(page.locator('input[placeholder="e.g. qwen2.5:7b"]')).toBeVisible()
  })

  test('displays Kibana URL field', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workspace' })).toBeVisible({ timeout: 5000 })
    await expect(page.locator('input[placeholder="http://localhost:5601"]')).toBeVisible()
  })

  test('save button is disabled when no changes made', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await expect(page.locator('button', { hasText: 'Save' })).toBeDisabled()
  })

  test('editing a field enables save button', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    const modelInput = page.locator('input[placeholder="e.g. qwen2.5:7b"]')
    await modelInput.fill('test-model-change')
    await expect(page.locator('button', { hasText: 'Save' })).toBeEnabled()
  })

  test('saving workspace config persists and reloads', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })

    // Change model
    const modelInput = page.locator('input[placeholder="e.g. qwen2.5:7b"]')
    const original = await modelInput.inputValue()
    await modelInput.fill('test-persist-model')
    await page.locator('button', { hasText: 'Save' }).click()
    await expect(page.locator('text=Saved')).toBeVisible({ timeout: 3000 })

    // Reload and verify
    await page.reload()
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await expect(modelInput).toHaveValue('test-persist-model')

    // Restore original
    await modelInput.fill(original)
    await page.locator('button', { hasText: 'Save' }).click()
    await expect(page.locator('text=Saved')).toBeVisible({ timeout: 3000 })
  })

  test('adding a default parameter shows key-value row', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    await page.locator('input[placeholder="key"]').fill('test-param')
    await page.locator('input[placeholder="value"]').fill('test-value')
    await page.locator('.add-row button', { hasText: '+' }).first().click()
    await expect(page.locator('.param-key', { hasText: 'test-param' })).toBeVisible()
  })

  test('removing a default parameter removes the row', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    // Add a param first
    await page.locator('input[placeholder="key"]').fill('temp-param')
    await page.locator('input[placeholder="value"]').fill('temp-val')
    await page.locator('.add-row button', { hasText: '+' }).first().click()
    await expect(page.locator('.param-key', { hasText: 'temp-param' })).toBeVisible()
    // Remove it
    const row = page.locator('.param-row', { hasText: 'temp-param' })
    await row.locator('button.danger').click()
    await expect(page.locator('.param-key', { hasText: 'temp-param' })).not.toBeVisible()
  })

  test('page header shows Settings title with icon', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('.page-header h1')).toContainText('Settings')
    await expect(page.locator('.page-header h1 svg')).toBeVisible()
  })

  test('invalid Kibana URL shows validation feedback', async ({ page }) => {
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    const kibanaInput = page.locator('input[placeholder="http://localhost:5601"]')
    await kibanaInput.fill('not-a-url')
    await expect(page.locator('.url-hint')).toBeVisible()
    await kibanaInput.fill('http://valid:5601')
    await expect(page.locator('.url-hint')).not.toBeVisible()
  })

  test('no JS errors during interaction', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/settings')
    await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
    // Interact with various fields
    await page.locator('input[placeholder="e.g. qwen2.5:7b"]').fill('test')
    await page.locator('input[placeholder="http://localhost:5601"]').fill('http://test:5601')
    await page.locator('input[placeholder="key"]').fill('k')
    await page.locator('input[placeholder="value"]').fill('v')
    await page.locator('.add-row button', { hasText: '+' }).first().click()
    expect(errors).toEqual([])
  })
})
