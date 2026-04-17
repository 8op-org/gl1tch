# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: settings.spec.js >> Settings page >> removing a default parameter removes the row
- Location: e2e/settings.spec.js:104:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.param-key').filter({ hasText: 'temp-param' })
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.param-key').filter({ hasText: 'temp-param' })

```

# Page snapshot

```yaml
- generic [ref=e2]:
  - complementary [ref=e3]:
    - button "S" [ref=e6] [cursor=pointer]
    - navigation [ref=e7]:
      - link "Workflows" [ref=e8] [cursor=pointer]:
        - /url: "#/"
        - img [ref=e9]
    - link "Settings" [ref=e12] [cursor=pointer]:
      - /url: "#/settings"
      - img [ref=e13]
  - main [ref=e16]:
    - generic [ref=e17]:
      - heading "Settings" [level=1] [ref=e18]:
        - img [ref=e19]
        - text: Settings
      - button "Save" [ref=e23] [cursor=pointer]:
        - img [ref=e24]
        - text: Save
    - generic [ref=e29]:
      - generic [ref=e30]:
        - generic [ref=e31]:
          - heading "Workflow Defaults" [level=2] [ref=e33]:
            - img [ref=e34]
            - text: Workflow Defaults
          - generic [ref=e36]:
            - generic [ref=e37]:
              - generic [ref=e38]:
                - generic [ref=e39]: Model
                - textbox "Model" [ref=e40]:
                  - /placeholder: e.g. qwen2.5:7b
                  - text: qwen3-8b
              - generic [ref=e41]:
                - generic [ref=e42]: Provider
                - combobox "Provider" [ref=e43]:
                  - option "— select —"
                  - option "claude"
                  - option "codex"
                  - option "copilot" [selected]
                  - option "gemini"
                  - option "opencode"
            - generic [ref=e44]:
              - generic [ref=e45]: Parameters
              - generic [ref=e46]:
                - generic [ref=e47]:
                  - generic [ref=e48]: temp-param
                  - textbox [ref=e49]: temp-val
                  - button "×" [ref=e50] [cursor=pointer]
                - generic [ref=e51]:
                  - textbox "key" [ref=e52]
                  - textbox "value" [ref=e53]
                  - button "+" [disabled] [ref=e54]
        - generic [ref=e55]:
          - heading "Workspace" [level=2] [ref=e57]:
            - img [ref=e58]
            - text: Workspace
          - generic [ref=e60]:
            - generic [ref=e61]:
              - generic [ref=e62]:
                - generic [ref=e63]: Name
                - textbox "Name" [ref=e64]:
                  - /placeholder: workspace name
                  - text: stokagent
              - generic [ref=e65]:
                - generic [ref=e66]: Elasticsearch URL
                - textbox "Elasticsearch URL" [ref=e67]:
                  - /placeholder: http://localhost:9200
                  - text: http://localhost:9200
            - generic [ref=e68]:
              - generic [ref=e69]: Repositories
              - generic [ref=e71]:
                - textbox "owner/repo or URL" [ref=e72]
                - button "+" [disabled] [ref=e73]
      - generic [ref=e76]:
        - generic [ref=e77]:
          - heading "Resources" [level=2] [ref=e78]:
            - img [ref=e79]
            - text: Resources
          - button "+ Add" [ref=e81] [cursor=pointer]
        - paragraph [ref=e82]: No resources declared. Click “Add resource” to add one.
```

# Test source

```ts
  11  |   })
  12  | 
  13  |   test('sidebar settings link shows active state', async ({ page }) => {
  14  |     await page.goto('#/settings')
  15  |     await page.waitForTimeout(500)
  16  |     const settingsLink = page.locator('.sidebar-footer .nav-item')
  17  |     await expect(settingsLink).toHaveClass(/active/)
  18  |   })
  19  | 
  20  |   test('page loads without JS errors', async ({ page }) => {
  21  |     const errors = []
  22  |     page.on('pageerror', (err) => errors.push(err.message))
  23  |     await page.goto('#/settings')
  24  |     await page.waitForTimeout(2000)
  25  |     expect(errors).toEqual([])
  26  |   })
  27  | 
  28  |   test('shows Workflow Defaults section', async ({ page }) => {
  29  |     await page.goto('#/settings')
  30  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  31  |   })
  32  | 
  33  |   test('shows Workspace section', async ({ page }) => {
  34  |     await page.goto('#/settings')
  35  |     await expect(page.locator('h2', { hasText: 'Workspace' })).toBeVisible({ timeout: 5000 })
  36  |   })
  37  | 
  38  |   test('displays current workspace name', async ({ page }) => {
  39  |     await page.goto('#/settings')
  40  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  41  |     const nameInput = page.locator('input[placeholder="workspace name"]')
  42  |     await expect(nameInput).toBeVisible()
  43  |     const val = await nameInput.inputValue()
  44  |     expect(val.length).toBeGreaterThan(0)
  45  |   })
  46  | 
  47  |   test('displays default model input', async ({ page }) => {
  48  |     await page.goto('#/settings')
  49  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  50  |     await expect(page.locator('input[placeholder="e.g. qwen2.5:7b"]')).toBeVisible()
  51  |   })
  52  | 
  53  |   test('displays Kibana URL field', async ({ page }) => {
  54  |     await page.goto('#/settings')
  55  |     await expect(page.locator('h2', { hasText: 'Workspace' })).toBeVisible({ timeout: 5000 })
  56  |     await expect(page.locator('input[placeholder="http://localhost:5601"]')).toBeVisible()
  57  |   })
  58  | 
  59  |   test('save button is disabled when no changes made', async ({ page }) => {
  60  |     await page.goto('#/settings')
  61  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  62  |     await expect(page.locator('button', { hasText: 'Save' })).toBeDisabled()
  63  |   })
  64  | 
  65  |   test('editing a field enables save button', async ({ page }) => {
  66  |     await page.goto('#/settings')
  67  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  68  |     const modelInput = page.locator('input[placeholder="e.g. qwen2.5:7b"]')
  69  |     await modelInput.fill('test-model-change')
  70  |     await expect(page.locator('button', { hasText: 'Save' })).toBeEnabled()
  71  |   })
  72  | 
  73  |   test('saving workspace config persists and reloads', async ({ page }) => {
  74  |     await page.goto('#/settings')
  75  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  76  | 
  77  |     // Change model
  78  |     const modelInput = page.locator('input[placeholder="e.g. qwen2.5:7b"]')
  79  |     const original = await modelInput.inputValue()
  80  |     await modelInput.fill('test-persist-model')
  81  |     await page.locator('button', { hasText: 'Save' }).click()
  82  |     await expect(page.locator('text=Saved')).toBeVisible({ timeout: 3000 })
  83  | 
  84  |     // Reload and verify
  85  |     await page.reload()
  86  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  87  |     await expect(modelInput).toHaveValue('test-persist-model')
  88  | 
  89  |     // Restore original
  90  |     await modelInput.fill(original)
  91  |     await page.locator('button', { hasText: 'Save' }).click()
  92  |     await expect(page.locator('text=Saved')).toBeVisible({ timeout: 3000 })
  93  |   })
  94  | 
  95  |   test('adding a default parameter shows key-value row', async ({ page }) => {
  96  |     await page.goto('#/settings')
  97  |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  98  |     await page.locator('input[placeholder="key"]').fill('test-param')
  99  |     await page.locator('input[placeholder="value"]').fill('test-value')
  100 |     await page.locator('.add-row button', { hasText: '+' }).first().click()
  101 |     await expect(page.locator('.param-key', { hasText: 'test-param' })).toBeVisible()
  102 |   })
  103 | 
  104 |   test('removing a default parameter removes the row', async ({ page }) => {
  105 |     await page.goto('#/settings')
  106 |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  107 |     // Add a param first
  108 |     await page.locator('input[placeholder="key"]').fill('temp-param')
  109 |     await page.locator('input[placeholder="value"]').fill('temp-val')
  110 |     await page.locator('.add-row button', { hasText: '+' }).first().click()
> 111 |     await expect(page.locator('.param-key', { hasText: 'temp-param' })).toBeVisible()
      |                                                                         ^ Error: expect(locator).toBeVisible() failed
  112 |     // Remove it
  113 |     const row = page.locator('.param-row', { hasText: 'temp-param' })
  114 |     await row.locator('button.danger').click()
  115 |     await expect(page.locator('.param-key', { hasText: 'temp-param' })).not.toBeVisible()
  116 |   })
  117 | 
  118 |   test('page header shows Settings title with icon', async ({ page }) => {
  119 |     await page.goto('#/settings')
  120 |     await expect(page.locator('.page-header h1')).toContainText('Settings')
  121 |     await expect(page.locator('.page-header h1 svg')).toBeVisible()
  122 |   })
  123 | 
  124 |   test('invalid Kibana URL shows validation feedback', async ({ page }) => {
  125 |     await page.goto('#/settings')
  126 |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  127 |     const kibanaInput = page.locator('input[placeholder="http://localhost:5601"]')
  128 |     await kibanaInput.fill('not-a-url')
  129 |     await expect(page.locator('.url-hint')).toBeVisible()
  130 |     await kibanaInput.fill('http://valid:5601')
  131 |     await expect(page.locator('.url-hint')).not.toBeVisible()
  132 |   })
  133 | 
  134 |   test('no JS errors during interaction', async ({ page }) => {
  135 |     const errors = []
  136 |     page.on('pageerror', (err) => errors.push(err.message))
  137 |     await page.goto('#/settings')
  138 |     await expect(page.locator('h2', { hasText: 'Workflow Defaults' })).toBeVisible({ timeout: 5000 })
  139 |     // Interact with various fields
  140 |     await page.locator('input[placeholder="e.g. qwen2.5:7b"]').fill('test')
  141 |     await page.locator('input[placeholder="http://localhost:5601"]').fill('http://test:5601')
  142 |     await page.locator('input[placeholder="key"]').fill('k')
  143 |     await page.locator('input[placeholder="value"]').fill('v')
  144 |     await page.locator('.add-row button', { hasText: '+' }).first().click()
  145 |     expect(errors).toEqual([])
  146 |   })
  147 | })
  148 | 
```