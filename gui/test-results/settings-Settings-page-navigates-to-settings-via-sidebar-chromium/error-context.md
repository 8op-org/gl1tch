# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: settings.spec.js >> Settings page >> navigates to settings via sidebar
- Location: e2e/settings.spec.js:5:3

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: locator.hover: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('aside.sidebar')

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
      - heading "Workflows" [level=1] [ref=e18]:
        - img [ref=e19]
        - text: Workflows
      - generic [ref=e21]:
        - button "Cards" [ref=e22] [cursor=pointer]:
          - img [ref=e23]
        - button "Grouped" [ref=e25] [cursor=pointer]:
          - img [ref=e26]
        - button "List" [ref=e28] [cursor=pointer]:
          - img [ref=e29]
    - generic [ref=e32]:
      - generic [ref=e33]:
        - generic [ref=e34]:
          - img [ref=e36]
          - textbox "Search..." [ref=e39]
        - generic [ref=e40]:
          - button "architecture" [ref=e41] [cursor=pointer]
          - button "create" [ref=e42] [cursor=pointer]
          - button "dashboard" [ref=e43] [cursor=pointer]
          - button "issue" [ref=e44] [cursor=pointer]
          - button "knowledge" [ref=e45] [cursor=pointer]
          - button "morning" [ref=e46] [cursor=pointer]
          - button "other" [ref=e47] [cursor=pointer]
          - button "pr" [ref=e48] [cursor=pointer]
          - button "repo" [ref=e49] [cursor=pointer]
          - button "test" [ref=e50] [cursor=pointer]
          - button "work" [ref=e51] [cursor=pointer]
      - generic [ref=e52]:
        - button "architecture-analysis Analyze repo architecture, entry points, abstractions, and patterns Never run" [ref=e53] [cursor=pointer]:
          - generic [ref=e54]:
            - img [ref=e55]
            - text: architecture-analysis
          - paragraph [ref=e57]: Analyze repo architecture, entry points, abstractions, and patterns
          - generic [ref=e59]: Never run
        - button "clean Clear workflow results, logs, and temp files Never run" [ref=e60] [cursor=pointer]:
          - generic [ref=e61]:
            - img [ref=e62]
            - text: clean
          - paragraph [ref=e64]: Clear workflow results, logs, and temp files
          - generic [ref=e66]: Never run
        - button "create-pr Create a draft GitHub PR from a results folder Never run" [ref=e67] [cursor=pointer]:
          - generic [ref=e68]:
            - img [ref=e69]
            - text: create-pr
          - paragraph [ref=e71]: Create a draft GitHub PR from a results folder
          - generic [ref=e73]: Never run
        - button "dashboard-activity JSON feed — yesterday's authored PRs, reviews given, closed issues Never run" [ref=e74] [cursor=pointer]:
          - generic [ref=e75]:
            - img [ref=e76]
            - text: dashboard-activity
          - paragraph [ref=e78]: JSON feed — yesterday's authored PRs, reviews given, closed issues
          - generic [ref=e80]: Never run
        - button "dashboard-calendar JSON feed — today and tomorrow calendar events via gws elastic account Never run" [ref=e81] [cursor=pointer]:
          - generic [ref=e82]:
            - img [ref=e83]
            - text: dashboard-calendar
          - paragraph [ref=e85]: JSON feed — today and tomorrow calendar events via gws elastic account
          - generic [ref=e87]: Never run
        - button "dashboard-ci JSON feed — recent failed GitHub Actions runs across target repos Never run" [ref=e88] [cursor=pointer]:
          - generic [ref=e89]:
            - img [ref=e90]
            - text: dashboard-ci
          - paragraph [ref=e92]: JSON feed — recent failed GitHub Actions runs across target repos
          - generic [ref=e94]: Never run
        - button "dashboard-email JSON feed — attention-worthy unread emails via gws elastic account Never run" [ref=e95] [cursor=pointer]:
          - generic [ref=e96]:
            - img [ref=e97]
            - text: dashboard-email
          - paragraph [ref=e99]: JSON feed — attention-worthy unread emails via gws elastic account
          - generic [ref=e101]: Never run
        - button "dashboard-my-prs JSON feed — open PRs authored by adam-stokes with context and next steps Never run" [ref=e102] [cursor=pointer]:
          - generic [ref=e103]:
            - img [ref=e104]
            - text: dashboard-my-prs
          - paragraph [ref=e106]: JSON feed — open PRs authored by adam-stokes with context and next steps
          - generic [ref=e108]: Never run
        - button "dashboard-reviews JSON feed — PRs awaiting review with context and next steps Never run" [ref=e109] [cursor=pointer]:
          - generic [ref=e110]:
            - img [ref=e111]
            - text: dashboard-reviews
          - paragraph [ref=e113]: JSON feed — PRs awaiting review with context and next steps
          - generic [ref=e115]: Never run
        - button "dashboard-tasks JSON feed — prioritized tasks (HIGH/MEDIUM/LOW) with falling-behind detection Never run" [ref=e116] [cursor=pointer]:
          - generic [ref=e117]:
            - img [ref=e118]
            - text: dashboard-tasks
          - paragraph [ref=e120]: JSON feed — prioritized tasks (HIGH/MEDIUM/LOW) with falling-behind detection
          - generic [ref=e122]: Never run
        - button "issue-to-pr Analyze GitHub issue and produce PR-ready artifacts with tiered escalation Never run" [ref=e123] [cursor=pointer]:
          - generic [ref=e124]:
            - img [ref=e125]
            - text: issue-to-pr
          - paragraph [ref=e127]: Analyze GitHub issue and produce PR-ready artifacts with tiered escalation
          - generic [ref=e129]: Never run
        - button "knowledge-synthesis Synthesize all knowledge into materialized summaries Never run" [ref=e130] [cursor=pointer]:
          - generic [ref=e131]:
            - img [ref=e132]
            - text: knowledge-synthesis
          - paragraph [ref=e134]: Synthesize all knowledge into materialized summaries
          - generic [ref=e136]: Never run
        - button "morning-briefing Morning briefing — composite view of tasks, reviews, PRs, CI, email, and calendar Never run" [ref=e137] [cursor=pointer]:
          - generic [ref=e138]:
            - img [ref=e139]
            - text: morning-briefing
          - paragraph [ref=e141]: Morning briefing — composite view of tasks, reviews, PRs, CI, email, and calendar
          - generic [ref=e143]: Never run
        - button "pr-comments Summarize all comments on a GitHub pull request Never run" [ref=e144] [cursor=pointer]:
          - generic [ref=e145]:
            - img [ref=e146]
            - text: pr-comments
          - paragraph [ref=e148]: Summarize all comments on a GitHub pull request
          - generic [ref=e150]: Never run
        - button "pr-mining Mine merged PRs and closed issues for decisions, patterns, and gotchas Never run" [ref=e151] [cursor=pointer]:
          - generic [ref=e152]:
            - img [ref=e153]
            - text: pr-mining
          - paragraph [ref=e155]: Mine merged PRs and closed issues for decisions, patterns, and gotchas
          - generic [ref=e157]: Never run
        - button "pr-review Review a PR against its issue — code quality, conventions, API compat, and completeness Never run" [ref=e158] [cursor=pointer]:
          - generic [ref=e159]:
            - img [ref=e160]
            - text: pr-review
          - paragraph [ref=e162]: Review a PR against its issue — code quality, conventions, API compat, and completeness
          - generic [ref=e164]: Never run
        - button "pr-status Show PR state, reviewer feedback, and what's blocking Never run" [ref=e165] [cursor=pointer]:
          - generic [ref=e166]:
            - img [ref=e167]
            - text: pr-status
          - paragraph [ref=e169]: Show PR state, reviewer feedback, and what's blocking
          - generic [ref=e171]: Never run
        - button "repo-guide-gen Generate updated repo guide from accumulated ES knowledge Never run" [ref=e172] [cursor=pointer]:
          - generic [ref=e173]:
            - img [ref=e174]
            - text: repo-guide-gen
          - paragraph [ref=e176]: Generate updated repo guide from accumulated ES knowledge
          - generic [ref=e178]: Never run
        - button "repo-ingest Ingest repo documentation into ES knowledge index Never run" [ref=e179] [cursor=pointer]:
          - generic [ref=e180]:
            - img [ref=e181]
            - text: repo-ingest
          - paragraph [ref=e183]: Ingest repo documentation into ES knowledge index
          - generic [ref=e185]: Never run
        - button "test-echo Echo a message for testing Never run" [ref=e186] [cursor=pointer]:
          - generic [ref=e187]:
            - img [ref=e188]
            - text: test-echo
          - paragraph [ref=e190]: Echo a message for testing
          - generic [ref=e192]: Never run
        - button "test-fail Always fails for testing error states Never run" [ref=e193] [cursor=pointer]:
          - generic [ref=e194]:
            - img [ref=e195]
            - text: test-fail
          - paragraph [ref=e197]: Always fails for testing error states
          - generic [ref=e199]: Never run
        - button "test-multi Multi-step workflow for graph visualization testing Never run" [ref=e200] [cursor=pointer]:
          - generic [ref=e201]:
            - img [ref=e202]
            - text: test-multi
          - paragraph [ref=e204]: Multi-step workflow for graph visualization testing
          - generic [ref=e206]: Never run
        - button "test-params Parameterized test workflow Never run" [ref=e207] [cursor=pointer]:
          - generic [ref=e208]:
            - img [ref=e209]
            - text: test-params
          - paragraph [ref=e211]: Parameterized test workflow
          - generic [ref=e213]: Never run
        - button "test-save Saves a result file for testing file viewer Never run" [ref=e214] [cursor=pointer]:
          - generic [ref=e215]:
            - img [ref=e216]
            - text: test-save
          - paragraph [ref=e218]: Saves a result file for testing file viewer
          - generic [ref=e220]: Never run
        - button "work-on-issue Solve a GitHub issue end-to-end — fetch, analyze, implement, and prepare PR artifacts Never run" [ref=e221] [cursor=pointer]:
          - generic [ref=e222]:
            - img [ref=e223]
            - text: work-on-issue
          - paragraph [ref=e225]: Solve a GitHub issue end-to-end — fetch, analyze, implement, and prepare PR artifacts
          - generic [ref=e227]: Never run
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test'
  2   | 
  3   | // ── Settings page ──────────────────────────────────────────────────
  4   | test.describe('Settings page', () => {
  5   |   test('navigates to settings via sidebar', async ({ page }) => {
  6   |     await page.goto('/')
  7   |     const sidebar = page.locator('aside.sidebar')
> 8   |     await sidebar.hover()
      |                   ^ Error: locator.hover: Test timeout of 30000ms exceeded.
  9   |     await page.locator('.sidebar-footer .nav-item', { hasText: 'Settings' }).click()
  10  |     await expect(page).toHaveURL(/\/#\/settings/)
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
```