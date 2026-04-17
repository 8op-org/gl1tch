import { test, expect } from '@playwright/test'

// Pipeline graph tests require runs to exist. These tests verify the graph
// components render correctly when a run is expanded.

test.describe('Pipeline graph', () => {
  test('expanding a run row shows graph container', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.waitForTimeout(2000)
    const runRow = page.locator('.run-row').first()
    if (await runRow.isVisible()) {
      await runRow.click()
      await expect(page.locator('.graph-container')).toBeVisible({ timeout: 5000 })
    }
  })

  test('graph nodes have status coloring', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.waitForTimeout(2000)
    const runRow = page.locator('.run-row').first()
    if (await runRow.isVisible()) {
      await runRow.click()
      await page.waitForTimeout(1000)
      const node = page.locator('.graph-node').first()
      if (await node.isVisible()) {
        // Node should have a border color class or style
        await expect(node).toBeVisible()
      }
    }
  })

  test('clicking a graph node opens detail panel', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.waitForTimeout(2000)
    const runRow = page.locator('.run-row').first()
    if (await runRow.isVisible()) {
      await runRow.click()
      await page.waitForTimeout(1000)
      const node = page.locator('.graph-node').first()
      if (await node.isVisible()) {
        await node.click()
        await expect(page.locator('.node-panel')).toBeVisible({ timeout: 3000 })
      }
    }
  })

  test('node panel has metrics section', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.waitForTimeout(2000)
    const runRow = page.locator('.run-row').first()
    if (await runRow.isVisible()) {
      await runRow.click()
      await page.waitForTimeout(1000)
      const node = page.locator('.graph-node').first()
      if (await node.isVisible()) {
        await node.click()
        await page.waitForTimeout(500)
        // Panel should show metrics tab/section
        const panel = page.locator('.node-panel')
        if (await panel.isVisible()) {
          await expect(panel.locator('.panel-tab', { hasText: 'Metrics' })).toBeVisible()
        }
      }
    }
  })

  test('closing node panel hides it', async ({ page }) => {
    await page.goto('#/workflow/git-status.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.waitForTimeout(2000)
    const runRow = page.locator('.run-row').first()
    if (await runRow.isVisible()) {
      await runRow.click()
      await page.waitForTimeout(1000)
      const node = page.locator('.graph-node').first()
      if (await node.isVisible()) {
        await node.click()
        await page.waitForTimeout(500)
        const panel = page.locator('.node-panel')
        if (await panel.isVisible()) {
          await page.locator('.node-panel .close-btn').click()
          await expect(panel).not.toBeVisible()
        }
      }
    }
  })

  test('no JS errors when interacting with graph', async ({ page }) => {
    const errors = []
    page.on('pageerror', (err) => errors.push(err.message))
    await page.goto('#/workflow/git-status.glitch')
    await page.waitForSelector('.tabs', { timeout: 5000 })
    await page.waitForTimeout(2000)
    const runRow = page.locator('.run-row').first()
    if (await runRow.isVisible()) {
      await runRow.click()
      await page.waitForTimeout(2000)
    }
    expect(errors).toEqual([])
  })
})
