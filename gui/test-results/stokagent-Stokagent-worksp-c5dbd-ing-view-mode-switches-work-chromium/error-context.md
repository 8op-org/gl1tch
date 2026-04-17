# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: stokagent.spec.js >> Stokagent workspace — Workflow browsing >> view mode switches work
- Location: e2e/stokagent.spec.js:107:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.card-grid')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.card-grid')

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
        - button "Grouped" [active] [ref=e25] [cursor=pointer]:
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
        - generic [ref=e53]:
          - button "architecture 1" [ref=e54] [cursor=pointer]:
            - img [ref=e56]
            - generic [ref=e58]: architecture
            - generic [ref=e59]: "1"
          - button "architecture-analysis Analyze repo architecture, entry points, abstractions, and patterns" [ref=e61] [cursor=pointer]:
            - img [ref=e63]
            - generic [ref=e65]:
              - generic [ref=e66]: architecture-analysis
              - generic [ref=e67]: Analyze repo architecture, entry points, abstractions, and patterns
        - generic [ref=e68]:
          - button "create 1" [ref=e69] [cursor=pointer]:
            - img [ref=e71]
            - generic [ref=e73]: create
            - generic [ref=e74]: "1"
          - button "create-pr Create a draft GitHub PR from a results folder" [ref=e76] [cursor=pointer]:
            - img [ref=e78]
            - generic [ref=e80]:
              - generic [ref=e81]: create-pr
              - generic [ref=e82]: Create a draft GitHub PR from a results folder
        - generic [ref=e83]:
          - button "dashboard 7" [ref=e84] [cursor=pointer]:
            - img [ref=e86]
            - generic [ref=e88]: dashboard
            - generic [ref=e89]: "7"
          - generic [ref=e90]:
            - button "dashboard-activity JSON feed — yesterday's authored PRs, reviews given, closed issues" [ref=e91] [cursor=pointer]:
              - img [ref=e93]
              - generic [ref=e95]:
                - generic [ref=e96]: dashboard-activity
                - generic [ref=e97]: JSON feed — yesterday's authored PRs, reviews given, closed issues
            - button "dashboard-calendar JSON feed — today and tomorrow calendar events via gws elastic account" [ref=e98] [cursor=pointer]:
              - img [ref=e100]
              - generic [ref=e102]:
                - generic [ref=e103]: dashboard-calendar
                - generic [ref=e104]: JSON feed — today and tomorrow calendar events via gws elastic account
            - button "dashboard-ci JSON feed — recent failed GitHub Actions runs across target repos" [ref=e105] [cursor=pointer]:
              - img [ref=e107]
              - generic [ref=e109]:
                - generic [ref=e110]: dashboard-ci
                - generic [ref=e111]: JSON feed — recent failed GitHub Actions runs across target repos
            - button "dashboard-email JSON feed — attention-worthy unread emails via gws elastic account" [ref=e112] [cursor=pointer]:
              - img [ref=e114]
              - generic [ref=e116]:
                - generic [ref=e117]: dashboard-email
                - generic [ref=e118]: JSON feed — attention-worthy unread emails via gws elastic account
            - button "dashboard-my-prs JSON feed — open PRs authored by adam-stokes with context and next steps" [ref=e119] [cursor=pointer]:
              - img [ref=e121]
              - generic [ref=e123]:
                - generic [ref=e124]: dashboard-my-prs
                - generic [ref=e125]: JSON feed — open PRs authored by adam-stokes with context and next steps
            - button "dashboard-reviews JSON feed — PRs awaiting review with context and next steps" [ref=e126] [cursor=pointer]:
              - img [ref=e128]
              - generic [ref=e130]:
                - generic [ref=e131]: dashboard-reviews
                - generic [ref=e132]: JSON feed — PRs awaiting review with context and next steps
            - button "dashboard-tasks JSON feed — prioritized tasks (HIGH/MEDIUM/LOW) with falling-behind detection" [ref=e133] [cursor=pointer]:
              - img [ref=e135]
              - generic [ref=e137]:
                - generic [ref=e138]: dashboard-tasks
                - generic [ref=e139]: JSON feed — prioritized tasks (HIGH/MEDIUM/LOW) with falling-behind detection
        - generic [ref=e140]:
          - button "issue 1" [ref=e141] [cursor=pointer]:
            - img [ref=e143]
            - generic [ref=e145]: issue
            - generic [ref=e146]: "1"
          - button "issue-to-pr Analyze GitHub issue and produce PR-ready artifacts with tiered escalation" [ref=e148] [cursor=pointer]:
            - img [ref=e150]
            - generic [ref=e152]:
              - generic [ref=e153]: issue-to-pr
              - generic [ref=e154]: Analyze GitHub issue and produce PR-ready artifacts with tiered escalation
        - generic [ref=e155]:
          - button "knowledge 1" [ref=e156] [cursor=pointer]:
            - img [ref=e158]
            - generic [ref=e160]: knowledge
            - generic [ref=e161]: "1"
          - button "knowledge-synthesis Synthesize all knowledge into materialized summaries" [ref=e163] [cursor=pointer]:
            - img [ref=e165]
            - generic [ref=e167]:
              - generic [ref=e168]: knowledge-synthesis
              - generic [ref=e169]: Synthesize all knowledge into materialized summaries
        - generic [ref=e170]:
          - button "morning 1" [ref=e171] [cursor=pointer]:
            - img [ref=e173]
            - generic [ref=e175]: morning
            - generic [ref=e176]: "1"
          - button "morning-briefing Morning briefing — composite view of tasks, reviews, PRs, CI, email, and calendar" [ref=e178] [cursor=pointer]:
            - img [ref=e180]
            - generic [ref=e182]:
              - generic [ref=e183]: morning-briefing
              - generic [ref=e184]: Morning briefing — composite view of tasks, reviews, PRs, CI, email, and calendar
        - generic [ref=e185]:
          - button "other 1" [ref=e186] [cursor=pointer]:
            - img [ref=e188]
            - generic [ref=e190]: other
            - generic [ref=e191]: "1"
          - button "clean Clear workflow results, logs, and temp files" [ref=e193] [cursor=pointer]:
            - img [ref=e195]
            - generic [ref=e197]:
              - generic [ref=e198]: clean
              - generic [ref=e199]: Clear workflow results, logs, and temp files
        - generic [ref=e200]:
          - button "pr 4" [ref=e201] [cursor=pointer]:
            - img [ref=e203]
            - generic [ref=e205]: pr
            - generic [ref=e206]: "4"
          - generic [ref=e207]:
            - button "pr-comments Summarize all comments on a GitHub pull request" [ref=e208] [cursor=pointer]:
              - img [ref=e210]
              - generic [ref=e212]:
                - generic [ref=e213]: pr-comments
                - generic [ref=e214]: Summarize all comments on a GitHub pull request
            - button "pr-mining Mine merged PRs and closed issues for decisions, patterns, and gotchas" [ref=e215] [cursor=pointer]:
              - img [ref=e217]
              - generic [ref=e219]:
                - generic [ref=e220]: pr-mining
                - generic [ref=e221]: Mine merged PRs and closed issues for decisions, patterns, and gotchas
            - button "pr-review Review a PR against its issue — code quality, conventions, API compat, and completeness" [ref=e222] [cursor=pointer]:
              - img [ref=e224]
              - generic [ref=e226]:
                - generic [ref=e227]: pr-review
                - generic [ref=e228]: Review a PR against its issue — code quality, conventions, API compat, and completeness
            - button "pr-status Show PR state, reviewer feedback, and what's blocking" [ref=e229] [cursor=pointer]:
              - img [ref=e231]
              - generic [ref=e233]:
                - generic [ref=e234]: pr-status
                - generic [ref=e235]: Show PR state, reviewer feedback, and what's blocking
        - generic [ref=e236]:
          - button "repo 2" [ref=e237] [cursor=pointer]:
            - img [ref=e239]
            - generic [ref=e241]: repo
            - generic [ref=e242]: "2"
          - generic [ref=e243]:
            - button "repo-guide-gen Generate updated repo guide from accumulated ES knowledge" [ref=e244] [cursor=pointer]:
              - img [ref=e246]
              - generic [ref=e248]:
                - generic [ref=e249]: repo-guide-gen
                - generic [ref=e250]: Generate updated repo guide from accumulated ES knowledge
            - button "repo-ingest Ingest repo documentation into ES knowledge index" [ref=e251] [cursor=pointer]:
              - img [ref=e253]
              - generic [ref=e255]:
                - generic [ref=e256]: repo-ingest
                - generic [ref=e257]: Ingest repo documentation into ES knowledge index
        - generic [ref=e258]:
          - button "test 5" [ref=e259] [cursor=pointer]:
            - img [ref=e261]
            - generic [ref=e263]: test
            - generic [ref=e264]: "5"
          - generic [ref=e265]:
            - button "test-echo Echo a message for testing" [ref=e266] [cursor=pointer]:
              - img [ref=e268]
              - generic [ref=e270]:
                - generic [ref=e271]: test-echo
                - generic [ref=e272]: Echo a message for testing
            - button "test-fail Always fails for testing error states" [ref=e273] [cursor=pointer]:
              - img [ref=e275]
              - generic [ref=e277]:
                - generic [ref=e278]: test-fail
                - generic [ref=e279]: Always fails for testing error states
            - button "test-multi Multi-step workflow for graph visualization testing" [ref=e280] [cursor=pointer]:
              - img [ref=e282]
              - generic [ref=e284]:
                - generic [ref=e285]: test-multi
                - generic [ref=e286]: Multi-step workflow for graph visualization testing
            - button "test-params Parameterized test workflow" [ref=e287] [cursor=pointer]:
              - img [ref=e289]
              - generic [ref=e291]:
                - generic [ref=e292]: test-params
                - generic [ref=e293]: Parameterized test workflow
            - button "test-save Saves a result file for testing file viewer" [ref=e294] [cursor=pointer]:
              - img [ref=e296]
              - generic [ref=e298]:
                - generic [ref=e299]: test-save
                - generic [ref=e300]: Saves a result file for testing file viewer
        - generic [ref=e301]:
          - button "work 1" [ref=e302] [cursor=pointer]:
            - img [ref=e304]
            - generic [ref=e306]: work
            - generic [ref=e307]: "1"
          - button "work-on-issue Solve a GitHub issue end-to-end — fetch, analyze, implement, and prepare PR artifacts" [ref=e309] [cursor=pointer]:
            - img [ref=e311]
            - generic [ref=e313]:
              - generic [ref=e314]: work-on-issue
              - generic [ref=e315]: Solve a GitHub issue end-to-end — fetch, analyze, implement, and prepare PR artifacts
```

# Test source

```ts
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
> 115 |     await expect(page.locator('.card-grid')).toBeVisible()
      |                                              ^ Error: expect(locator).toBeVisible() failed
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