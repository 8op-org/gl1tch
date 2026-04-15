# Results Explorer Expansion + Bug Fixes

**Date:** 2026-04-15
**Branch:** gui-polish
**Worktree:** .worktrees/gui-polish

---

## Problem

1. The `/runs` page shows `500: SQL logic error: no such column: input` because the SQLite DB was created before the `input` column was added to the schema. `CREATE TABLE IF NOT EXISTS` silently skips recreation.

2. The results explorer has basic file browsing and editing but no way to trigger folder-level workflow actions like "review this folder's changes" or "create a PR from these results."

3. Previous session left Svelte 4 deprecated event handlers (`on:click`) and a backend inconsistency where `/api/results/{path}` behaves differently with/without trailing slashes.

## Solution

### 1. SQL Stale DB Auto-Recreate

In `store.OpenAt()`, after running `CREATE TABLE IF NOT EXISTS`, verify schema integrity:

```go
// After exec(createSchema):
var count int
err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('runs') WHERE name = 'input'`).Scan(&count)
if err == nil && count == 0 {
    db.Exec("DROP TABLE IF EXISTS runs")
    db.Exec("DROP TABLE IF EXISTS steps")
    db.Exec("DROP TABLE IF EXISTS research_events")
    db.ExecContext(ctx, createSchema)
}
```

This is pre-1.0 behavior — wipe and restart. No migration code.

### 2. Folder-Scoped ActionBar

**Current flow:**
- `ActionBar` receives `context="results"` (static)
- Calls `GET /api/workflows/actions/results`
- Shows matching workflows as buttons
- On click, opens `RunDialog` with workflow params

**New flow:**
- `ResultsBrowser` tracks `selectedFolder` (set when a folder node is clicked/expanded)
- `ActionBar` receives `context="results"` and `resultPath={selectedFolder}`
- When a folder is selected, ActionBar re-fetches actions for that context
- Workflow params auto-include `path` populated with the selected folder path
- RunDialog pre-fills the `path` param

**Backend:** No changes needed. `handleGetWorkflowActions` already supports prefix matching — `results:review` matches context `results`. Workflows declare granular actions like `actions: ["results:review"]` or broad `actions: ["results"]`.

**Frontend changes:**

`ResultsBrowser.svelte`:
- Add `selectedFolder` state, set on folder click
- Pass `selectedFolder` to ActionBar as `resultPath`
- ActionBar key updates when selectedFolder changes

`ActionBar.svelte`:
- Accept `resultPath` prop
- Re-fetch actions when `resultPath` changes (use `$effect`)
- Pass `resultPath` as auto-param when triggering workflow

`RunDialog.svelte`:
- Accept optional `autoParams` prop (map of param name to value)
- Pre-fill param inputs from autoParams
- User can still override

### 3. Review Workflow

File: `.glitch/workflows/review-results.yaml`

```yaml
name: review-results
description: Run a structured code review on a results folder
actions:
  - results:review
  - results
steps:
  - id: gather
    run: |
      DIR="{{.param.path}}"
      if [ -z "$DIR" ]; then echo "ERROR: no path provided" >&2; exit 1; fi
      echo "## Files in $DIR"
      find "$DIR" -type f | head -50
      echo "---"
      for f in $(find "$DIR" -type f \( -name '*.diff' -o -name '*.patch' -o -name '*.md' \) | head -20); do
        echo "### $f"
        head -200 "$f"
        echo "---"
      done

  - id: review
    llm:
      prompt: |
        Review the following results folder contents against this checklist:

        {{step "gather"}}

        Checklist — mark each PASS, FAIL, or N/A:
        - Functionality: do changes accomplish their stated goal?
        - Tests: adequate coverage?
        - Security: no secrets, no injection vectors?
        - Performance: no obvious bottlenecks?
        - Standards: follows repo conventions?
        - Breaking changes: documented?

        End with a summary and suggested actions.

  - id: save-review
    save: "{{.param.path}}/review.md"
    save_step: review
```

### 4. PR Creation Workflow

File: `.glitch/workflows/create-pr.yaml`

```yaml
name: create-pr
description: Create a draft GitHub PR from a results folder
actions:
  - results:pr
  - results
steps:
  - id: detect-repo
    run: |
      DIR="{{.param.path}}"
      ORG=$(echo "$DIR" | awk -F/ '{for(i=1;i<=NF;i++) if($(i+1) != "") {print $i; exit}}')
      REPO=$(echo "$DIR" | awk -F/ '{for(i=1;i<=NF;i++) if($(i+1) != "") {print $(i+1); exit}}')
      echo "${ORG}/${REPO}"

  - id: read-review
    run: |
      DIR="{{.param.path}}"
      if [ -f "$DIR/review.md" ]; then
        cat "$DIR/review.md"
      else
        echo "No review found. Run review-results first."
      fi

  - id: build-pr-body
    llm:
      prompt: |
        Create a pull request description from this review:

        {{step "read-review"}}

        Format as:
        ## Summary
        <2-3 bullet points>

        ## Review Notes
        <key findings from the review>

        ## Test Plan
        <checklist of testing steps>

  - id: create-draft
    run: |
      REPO="{{step "detect-repo"}}"
      TITLE="{{.param.title}}"
      BRANCH="{{.param.branch}}"
      if [ -z "$TITLE" ]; then TITLE="Draft PR from gl1tch results"; fi
      if [ -z "$BRANCH" ]; then echo "ERROR: branch param required" >&2; exit 1; fi
      BODY='{{step "build-pr-body"}}'
      gh pr create --repo "$REPO" --head "$BRANCH" --draft --title "$TITLE" --body "$BODY"

  - id: save-url
    save: "{{.param.path}}/pr-url.txt"
    save_step: create-draft
```

### 5. Previous Session Fixes

**Svelte 5 event handlers:** Replace `on:click` to `onclick`, `on:input` to `oninput`, etc. across all `.svelte` files in the worktree.

**Results trailing slash:** In `api_results.go` `handleGetResult()`, normalize: if the resolved path is a directory, always return directory JSON listing regardless of trailing slash.

### 6. Playwright Test Coverage

New tests added to `gui/e2e/gui.spec.js`:

**Folder actions:**
- Selecting a folder shows ActionBar (if workflows with results actions exist)
- Action buttons appear/disappear based on folder selection
- Clicking an action opens RunDialog with path pre-filled

**Runs page (SQL fix):**
- `/api/runs` returns valid JSON (not 500) after DB recreation
- Runs page loads without SQL errors

**Svelte 5:**
- No deprecation warnings in browser console across all page navigations

**Results normalization:**
- `GET /api/results/somedir` (no trailing slash) returns same as `GET /api/results/somedir/`

## Files Changed

| File | Change |
|------|--------|
| `internal/store/store.go` | Schema drift detection + auto-recreate |
| `internal/gui/api_results.go` | Trailing slash normalization |
| `gui/src/routes/ResultsBrowser.svelte` | selectedFolder state, ActionBar integration |
| `gui/src/lib/components/ActionBar.svelte` | Reactive re-fetch on path change, auto-params |
| `gui/src/routes/RunDialog.svelte` | autoParams pre-fill |
| `gui/src/routes/*.svelte` | Svelte 5 event handler migration |
| `gui/src/lib/components/*.svelte` | Svelte 5 event handler migration |
| `gui/e2e/gui.spec.js` | New folder action + SQL fix + Svelte 5 tests |
| `.glitch/workflows/review-results.yaml` | New review workflow |
| `.glitch/workflows/create-pr.yaml` | New PR creation workflow |
