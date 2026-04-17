# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: stokagent.spec.js >> Stokagent workspace — Workflow browsing >> grouped view shows workflow groups
- Location: e2e/stokagent.spec.js:66:3

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.waitForSelector: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('.group-header') to be visible

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
  3   | // These tests run against the stokagent workspace (38 real workflows).
  4   | // Requires: glitch --workspace ~/Projects/stokagent workflow gui
  5   | // on port 8374 before running.
  6   | 
  7   | test.use({ baseURL: 'http://127.0.0.1:8374' })
  8   | 
  9   | test.describe('Stokagent workspace — API', () => {
  10  |   test('GET /api/workflows returns 30+ workflows', async ({ request }) => {
  11  |     const resp = await request.get('/api/workflows')
  12  |     expect(resp.ok()).toBeTruthy()
  13  |     const data = await resp.json()
  14  |     expect(data.length).toBeGreaterThan(30)
  15  |     // Verify converted .glitch files are present
  16  |     const names = data.map(w => w.file)
  17  |     expect(names).toContain('dashboard-reviews.glitch')
  18  |     expect(names).toContain('git-status.glitch')
  19  |     expect(names).toContain('work-on-issue.glitch')
  20  |   })
  21  | 
  22  |   test('GET /api/workflows/git-status.glitch returns source with sexpr', async ({ request }) => {
  23  |     const resp = await request.get('/api/workflows/git-status.glitch')
  24  |     expect(resp.ok()).toBeTruthy()
  25  |     const data = await resp.json()
  26  |     expect(data.source).toContain('(workflow')
  27  |     expect(data.source).toContain('git-status')
  28  |     expect(data.params).toEqual([])
  29  |   })
  30  | 
  31  |   test('GET /api/workflows/dashboard-reviews.glitch has description', async ({ request }) => {
  32  |     const resp = await request.get('/api/workflows/dashboard-reviews.glitch')
  33  |     expect(resp.ok()).toBeTruthy()
  34  |     const data = await resp.json()
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
> 68  |     await page.waitForSelector('.group-header')
      |                ^ Error: page.waitForSelector: Test timeout of 30000ms exceeded.
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
  135 |     await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 5000 })
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
```