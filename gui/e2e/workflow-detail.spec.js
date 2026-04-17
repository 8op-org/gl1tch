import { test, expect } from '@playwright/test'

test.describe('Workflow Detail — tabs', () => {
  test('loads with Runs tab active by default', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tab.active', { timeout: 5000 })
    const activeTab = page.locator('.tab.active')
    await expect(activeTab).toContainText('Runs')
  })

  test('has three tabs: Runs, Source, Metadata', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await expect(page.locator('.tab')).toHaveCount(3)
    await expect(page.locator('.tab').nth(0)).toContainText('Runs')
    await expect(page.locator('.tab').nth(1)).toContainText('Source')
    await expect(page.locator('.tab').nth(2)).toContainText('Metadata')
  })

  test('switching to Source tab shows editor', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Source' }).click()
    // Monaco editor or textarea should appear
    await expect(page.locator('.editor-container')).toBeVisible({ timeout: 5000 })
  })

  test('switching to Metadata tab shows workflow info', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Metadata' }).click()
    await expect(page.locator('.meta-grid')).toBeVisible({ timeout: 3000 })
  })

  test('breadcrumb shows Workflows link and workflow name', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await expect(page.locator('main a', { hasText: 'Workflows' })).toBeVisible()
  })

  test('breadcrumb Workflows link navigates home', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('main a', { hasText: 'Workflows' }).click()
    await expect(page).toHaveURL(/\/#\/$/)
  })

  test('Run button is always visible', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await expect(page.locator('.header-actions button.primary', { hasText: 'Run' })).toBeVisible()
  })

  test('Save button only visible on Source tab', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    // On Runs tab — Save should not be visible
    await expect(page.locator('button', { hasText: 'Save' })).not.toBeVisible()
    // Switch to Source
    await page.locator('.tab', { hasText: 'Source' }).click()
    await expect(page.locator('button', { hasText: 'Save' })).toBeVisible({ timeout: 5000 })
  })

  test('Run button opens run dialog', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.header-actions button.primary', { hasText: 'Run' }).click()
    await expect(page.locator('.modal')).toBeVisible()
  })

  test('no JS errors across tab switches', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Source' }).click()
    await page.waitForTimeout(1000)
    await page.locator('.tab', { hasText: 'Metadata' }).click()
    await page.waitForTimeout(500)
    await page.locator('.tab', { hasText: 'Runs' }).click()
    await page.waitForTimeout(500)
    expect(errors).toEqual([])
  })
})

test.describe('Workflow Detail — runs tab', () => {
  test('shows run list or empty state', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    // Either run rows exist or empty state message
    await page.waitForTimeout(2000)
    const hasRuns = await page.locator('.run-row').count()
    const hasEmpty = await page.locator('.text-muted', { hasText: /no runs/i }).isVisible().catch(() => false)
    expect(hasRuns > 0 || hasEmpty).toBeTruthy()
  })
})

test.describe('Workflow Detail — source tab', () => {
  test('Save button disabled when clean', async ({ page }) => {
    await page.goto('#/workflow/clean.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.locator('.tab', { hasText: 'Source' }).click()
    await page.waitForTimeout(1000)
    await expect(page.locator('button', { hasText: 'Save' })).toBeDisabled()
  })
})
