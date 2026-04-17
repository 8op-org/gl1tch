# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: stokagent.spec.js >> Stokagent workspace — Editor >> breadcrumb shows workflow name
- Location: e2e/stokagent.spec.js:133:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.cm-editor')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.cm-editor')

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
      - navigation [ref=e18]:
        - link "Workflows" [ref=e19] [cursor=pointer]:
          - /url: /
        - generic [ref=e20]: /
        - generic [ref=e21]: git-status.glitch
      - button "Run" [ref=e23] [cursor=pointer]:
        - img [ref=e24]
        - text: Run
    - generic [ref=e26]:
      - button "Runs" [ref=e27] [cursor=pointer]:
        - img [ref=e28]
        - text: Runs
      - button "Source" [ref=e30] [cursor=pointer]:
        - img [ref=e31]
        - text: Source
      - button "Metadata" [ref=e34] [cursor=pointer]:
        - img [ref=e35]
        - text: Metadata
    - paragraph [ref=e40]: "404: not found"
```

# Test source

```ts
  35  |     expect(data.source).toContain('(workflow')
  36  |     expect(data.source).toContain('dashboard-reviews')
  37  |   })
  38  | 
  39  |   test('parameterized workflow extracts params', async ({ request }) => {
  40  |     const resp = await request.get('/api/workflows/issue-to-pr-claude.glitch')
  41  |     expect(resp.ok()).toBeTruthy()
  42  |     const data = await resp.json()
  43  |     expect(data.params).toContain('repo')
  44  |     expect(data.params).toContain('issue')
  45  |   })
  46  | })
  47  | 
  48  | test.describe('Stokagent workspace — Workflow browsing', () => {
  49  |   test('workflow list loads with 30+ cards', async ({ page }) => {
  50  |     await page.goto('/')
  51  |     await page.waitForSelector('.card')
  52  |     const cards = page.locator('.card')
  53  |     const count = await cards.count()
  54  |     expect(count).toBeGreaterThan(30)
  55  |   })
  56  | 
  57  |   test('workflow cards show descriptions', async ({ page }) => {
  58  |     await page.goto('/')
  59  |     await page.waitForSelector('.card')
  60  |     // dashboard-reviews should have a description
  61  |     const desc = page.locator('.card-desc')
  62  |     const count = await desc.count()
  63  |     expect(count).toBeGreaterThan(0)
  64  |   })
  65  | 
  66  |   test('grouped view shows workflow groups', async ({ page }) => {
  67  |     await page.goto('/')
  68  |     await page.waitForSelector('.group-header')
  69  |     const groups = page.locator('.group-header')
  70  |     const count = await groups.count()
  71  |     expect(count).toBeGreaterThan(3) // dashboard, github, issue-to-pr, etc.
  72  |   })
  73  | 
  74  |   test('group collapse/expand works', async ({ page }) => {
  75  |     await page.goto('/')
  76  |     await page.waitForSelector('.group-header')
  77  |     const firstGroup = page.locator('.group').first()
  78  |     const header = firstGroup.locator('.group-header')
  79  |     // Verify grid is visible initially
  80  |     await expect(firstGroup.locator('.group-grid')).toBeVisible()
  81  |     // Click to collapse
  82  |     await header.click()
  83  |     // The group grid should be hidden
  84  |     await expect(firstGroup.locator('.group-grid')).not.toBeVisible()
  85  |     // Click again to expand
  86  |     await header.click()
  87  |     await expect(firstGroup.locator('.group-grid')).toBeVisible()
  88  |   })
  89  | 
  90  |   test('search filters to matching workflows', async ({ page }) => {
  91  |     await page.goto('/')
  92  |     await page.waitForSelector('.card')
  93  |     await page.fill('input[placeholder="Search..."]', 'dashboard')
  94  |     const cards = page.locator('.card')
  95  |     const count = await cards.count()
  96  |     expect(count).toBeGreaterThanOrEqual(8) // 8 dashboard workflows
  97  |     expect(count).toBeLessThan(38) // not all
  98  |   })
  99  | 
  100 |   test('search for nonexistent shows zero cards', async ({ page }) => {
  101 |     await page.goto('/')
  102 |     await page.waitForSelector('.card')
  103 |     await page.fill('input[placeholder="Search..."]', 'xyznonexistent999')
  104 |     await expect(page.locator('.card')).toHaveCount(0)
  105 |   })
  106 | 
  107 |   test('view mode switches work', async ({ page }) => {
  108 |     await page.goto('/')
  109 |     await page.waitForSelector('.card')
  110 |     // Switch to list view
  111 |     await page.locator('.view-btn', { hasText: /.*/ }).nth(2).click()
  112 |     await expect(page.locator('.wf-table')).toBeVisible()
  113 |     // Switch to grid view
  114 |     await page.locator('.view-btn', { hasText: /.*/ }).nth(1).click()
  115 |     await expect(page.locator('.card-grid')).toBeVisible()
  116 |   })
  117 | })
  118 | 
  119 | test.describe('Stokagent workspace — Editor', () => {
  120 |   test('clicking a workflow card opens the editor', async ({ page }) => {
  121 |     await page.goto('/')
  122 |     await page.waitForSelector('.card')
  123 |     await page.locator('.card').first().click()
  124 |     await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
  125 |   })
  126 | 
  127 |   test('editor loads sexpr source for git-status', async ({ page }) => {
  128 |     await page.goto('#/workflow/git-status.glitch')
  129 |     await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
  130 |     await expect(page.locator('.cm-content')).toContainText('workflow')
  131 |   })
  132 | 
  133 |   test('breadcrumb shows workflow name', async ({ page }) => {
  134 |     await page.goto('#/workflow/git-status.glitch')
> 135 |     await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
      |                                              ^ Error: expect(locator).toBeVisible() failed
  136 |     await expect(page.locator('.breadcrumb')).toContainText('git-status.glitch')
  137 |   })
  138 | 
  139 |   test('metadata panel is visible', async ({ page }) => {
  140 |     await page.goto('#/workflow/git-status.glitch')
  141 |     await expect(page.locator('.meta-panel')).toBeVisible({ timeout: 5000 })
  142 |   })
  143 | 
  144 |   test('Save and Run buttons exist', async ({ page }) => {
  145 |     await page.goto('#/workflow/git-status.glitch')
  146 |     await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
  147 |     await expect(page.locator('button', { hasText: 'Save' })).toBeVisible()
  148 |     await expect(page.locator('button', { hasText: 'Run' })).toBeVisible()
  149 |   })
  150 | 
  151 |   test('Run button opens dialog for parameterized workflow', async ({ page }) => {
  152 |     await page.goto('#/workflow/issue-to-pr-claude.glitch')
  153 |     await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
  154 |     await page.locator('button', { hasText: 'Run' }).click()
  155 |     await expect(page.locator('.modal')).toBeVisible()
  156 |     // Should show repo and issue params
  157 |     await expect(page.locator('.modal')).toContainText('repo')
  158 |     await expect(page.locator('.modal')).toContainText('issue')
  159 |     // Cancel closes it
  160 |     await page.locator('button', { hasText: 'Cancel' }).click()
  161 |     await expect(page.locator('.modal')).not.toBeVisible()
  162 |   })
  163 | 
  164 |   test('Run button opens dialog for no-param workflow', async ({ page }) => {
  165 |     await page.goto('#/workflow/git-status.glitch')
  166 |     await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
  167 |     await page.locator('button', { hasText: 'Run' }).click()
  168 |     await expect(page.locator('.modal')).toBeVisible()
  169 |     await expect(page.locator('.modal')).toContainText('No parameters required')
  170 |   })
  171 | })
  172 | 
  173 | test.describe('Stokagent workspace — Results browser', () => {
  174 |   test('results tree loads with directories', async ({ page }) => {
  175 |     await page.goto('#/results')
  176 |     await page.waitForSelector('.tree-item', { timeout: 5000 })
  177 |     const items = page.locator('.tree-item')
  178 |     await expect(items.first()).toBeVisible()
  179 |   })
  180 | })
  181 | 
  182 | test.describe('Stokagent workspace — Runs page', () => {
  183 |   test('runs page loads', async ({ page }) => {
  184 |     await page.goto('#/runs')
  185 |     await expect(page.locator('h1', { hasText: 'Runs' })).toBeVisible({ timeout: 5000 })
  186 |   })
  187 | 
  188 |   test('runs page has status filter', async ({ page }) => {
  189 |     await page.goto('#/runs')
  190 |     await expect(page.locator('select')).toBeVisible({ timeout: 5000 })
  191 |   })
  192 | })
  193 | 
```