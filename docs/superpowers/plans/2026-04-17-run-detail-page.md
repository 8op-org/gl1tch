# Run Detail Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a dedicated `/run/:id` page with full-width pipeline graph, slide-over step detail panel, result file linking, and 2-second live polling.

**Architecture:** Graph-centric layout. RunDetail page owns the polling lifecycle and passes run+steps data down to PipelineGraph. PipelineGraph gets auto-fit via SVG viewBox and zoom/pan. Clicking a node opens a SlideOverPanel with Output/Prompt/Metrics/Files tabs. Backend adds `artifacts` column to steps table and exposes `run_id` in workflow templates.

**Tech Stack:** Svelte 5, elkjs, CodeMirror 6, Go (net/http + SQLite)

---

### Task 1: Backend — Add artifacts column and expose run_id in templates

**Files:**
- Modify: `internal/store/schema.go:27-41` — add `artifacts` column to steps table
- Modify: `internal/store/store.go:106-118` — add `Artifacts` field to `StepRecord`
- Modify: `internal/store/store.go:159-176` — update `RecordStep` INSERT to include artifacts
- Modify: `internal/pipeline/runner.go:142-152` — add `Artifacts` field to pipeline `StepRecord`
- Modify: `internal/pipeline/runner.go:154-167` — update `buildStepRecord` to pass artifacts through
- Modify: `internal/pipeline/runner.go:1640-1645` — add `run_id` to template `data` map
- Modify: `internal/pipeline/runner.go:1647-1666` — record artifact path after save
- Modify: `internal/pipeline/runner.go:1980-1994` — record artifact path after write_file
- Modify: `internal/gui/api_runs.go:125-136` — add `Artifacts` to stepEntry struct
- Modify: `internal/gui/api_runs.go:138-143` — add artifacts to SQL query
- Modify: `internal/gui/api_runs.go:150-169` — scan artifacts from DB
- Test: `internal/store/store_test.go`

- [ ] **Step 1: Write failing test for artifacts in store**

Add to `internal/store/store_test.go`:

```go
func TestRecordStepArtifacts(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	runID, err := s.RecordRun(RunRecord{Kind: "pipeline", Name: "artifact-test", Input: ""})
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}

	arts := []string{"results/report.md", "results/data.json"}
	if err := s.RecordStep(StepRecord{
		RunID: runID, StepID: "save-step", Output: "saved",
		Kind: "save", ExitStatus: intPtr(0), Artifacts: arts,
	}); err != nil {
		t.Fatalf("RecordStep: %v", err)
	}

	// Read back via raw query to verify
	var artsJSON string
	err = s.DB().QueryRow(`SELECT COALESCE(artifacts,'') FROM steps WHERE run_id = ? AND step_id = ?`, runID, "save-step").Scan(&artsJSON)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if artsJSON == "" {
		t.Fatal("artifacts should not be empty")
	}
	var got []string
	if err := json.Unmarshal([]byte(artsJSON), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 || got[0] != "results/report.md" {
		t.Fatalf("unexpected artifacts: %v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/store/ -run TestRecordStepArtifacts -v`
Expected: FAIL — `StepRecord` has no `Artifacts` field

- [ ] **Step 3: Update schema and store**

In `internal/store/schema.go`, add `artifacts TEXT` to the steps table:

```sql
CREATE TABLE IF NOT EXISTS steps (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id      INTEGER NOT NULL,
  step_id     TEXT NOT NULL,
  prompt      TEXT,
  output      TEXT,
  model       TEXT,
  duration_ms INTEGER,
  kind        TEXT,
  exit_status INTEGER,
  tokens_in   INTEGER,
  tokens_out  INTEGER,
  gate_passed INTEGER,
  artifacts   TEXT,
  UNIQUE(run_id, step_id)
);
```

In `internal/store/store.go`, add `Artifacts []string` to `StepRecord`:

```go
type StepRecord struct {
	RunID      int64
	StepID     string
	Prompt     string
	Output     string
	Model      string
	DurationMs int64
	Kind       string
	ExitStatus *int
	TokensIn   int64
	TokensOut  int64
	GatePassed *bool
	Artifacts  []string
}
```

Update `RecordStep` to serialize and insert artifacts:

```go
func (s *Store) RecordStep(rec StepRecord) error {
	var gatePassed *int
	if rec.GatePassed != nil {
		v := 0
		if *rec.GatePassed {
			v = 1
		}
		gatePassed = &v
	}
	var artifactsJSON *string
	if len(rec.Artifacts) > 0 {
		b, _ := json.Marshal(rec.Artifacts)
		s := string(b)
		artifactsJSON = &s
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO steps (run_id, step_id, prompt, output, model, duration_ms,
         kind, exit_status, tokens_in, tokens_out, gate_passed, artifacts)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.RunID, rec.StepID, rec.Prompt, rec.Output, rec.Model, rec.DurationMs,
		rec.Kind, rec.ExitStatus, rec.TokensIn, rec.TokensOut, gatePassed, artifactsJSON,
	)
	return err
}
```

Add `import "encoding/json"` to the imports if not already present.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/store/ -run TestRecordStepArtifacts -v`
Expected: PASS

- [ ] **Step 5: Update pipeline StepRecord and runner**

In `internal/pipeline/runner.go`, add `Artifacts []string` to the pipeline `StepRecord` (line 142):

```go
type StepRecord struct {
	StepID     string
	Prompt     string
	Output     string
	Model      string
	DurationMs int64
	Kind       string
	ExitStatus *int
	TokensIn   int64
	TokensOut  int64
	Artifacts  []string
}
```

In `buildStepRecord` (line 154), pass artifacts through:

```go
func buildStepRecord(step Step, outcome *stepOutcome, runErr error, dur time.Duration, defaultModel string) StepRecord {
	rec := StepRecord{
		StepID:     step.ID,
		Kind:       stepKind(step),
		DurationMs: dur.Milliseconds(),
	}
	exit := 0
	if runErr != nil {
		exit = 1
	}
	rec.ExitStatus = &exit
	if outcome != nil {
		rec.Output = outcome.output
		rec.TokensIn = int64(outcome.tokensIn)
		rec.TokensOut = int64(outcome.tokensOut)
		rec.Artifacts = outcome.artifacts
		// ... rest unchanged
```

Add `artifacts []string` to the `stepOutcome` struct (find it near the top of runner.go):

```go
type stepOutcome struct {
	output    string
	tokensIn  int
	tokensOut int
	model     string
	artifacts []string
}
```

Add `run_id` to the template data map in `runSingleStep` (line 1640):

```go
data := map[string]any{
	"input":     rctx.input,
	"param":     rctx.params,
	"workspace": rctx.workspace,
	"resource":  rctx.resources,
	"run_id":    fmt.Sprintf("%d", rctx.runID),
}
```

Update the save step handler (line 1647) to return artifact path:

```go
if step.Save != "" {
	rendered, err := renderInStep(step.Save, scopeFromData(data), stepsSnap, rctx.workflowObj, &step)
	if err != nil {
		return nil, fmt.Errorf("step %s: render: %w", step.ID, err)
	}
	ui.StepSave(step.ID, rendered)
	sourceStep := step.SaveStep
	if sourceStep == "" && rctx.prevStepID != "" {
		sourceStep = rctx.prevStepID
	}
	content := stepsSnap[sourceStep]
	if err := os.MkdirAll(filepath.Dir(rendered), 0o755); err != nil {
		return nil, fmt.Errorf("step %s: mkdir: %w", step.ID, err)
	}
	if err := os.WriteFile(rendered, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("step %s: write: %w", step.ID, err)
	}
	out := fmt.Sprintf("saved %s to %s", sourceStep, rendered)
	return &stepOutcome{output: out, artifacts: []string{rendered}}, nil
}
```

Update the write_file handler (line 1980) similarly:

```go
if step.WriteFile != nil {
	rendered, err := renderInStep(step.WriteFile.Path, scopeFromData(data), stepsSnap, rctx.workflowObj, &step)
	if err != nil {
		return nil, fmt.Errorf("step %s: render: %w", step.ID, err)
	}
	content := stepsSnap[step.WriteFile.From]
	if err := os.MkdirAll(filepath.Dir(rendered), 0o755); err != nil {
		return nil, fmt.Errorf("step %s: mkdir: %w", step.ID, err)
	}
	if err := os.WriteFile(rendered, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("step %s: write-file: %w", step.ID, err)
	}
	ui.StepSDK(step.ID, "write-file")
	return &stepOutcome{output: rendered, artifacts: []string{rendered}}, nil
}
```

- [ ] **Step 6: Update API to return artifacts**

In `internal/gui/api_runs.go`, update the `stepEntry` struct (line 125):

```go
type stepEntry struct {
	StepID     string   `json:"step_id"`
	Model      string   `json:"model"`
	DurationMs int64    `json:"duration_ms"`
	Kind       string   `json:"kind,omitempty"`
	ExitStatus *int     `json:"exit_status,omitempty"`
	TokensIn   int64    `json:"tokens_in"`
	TokensOut  int64    `json:"tokens_out"`
	GatePassed *bool    `json:"gate_passed,omitempty"`
	Output     string   `json:"output,omitempty"`
	Prompt     string   `json:"prompt,omitempty"`
	Artifacts  []string `json:"artifacts,omitempty"`
}
```

Update the SQL query (line 138) to include artifacts:

```go
stepRows, err := s.store.DB().Query(
	`SELECT step_id, COALESCE(model,''), COALESCE(duration_ms,0),
	        COALESCE(kind,''), exit_status, COALESCE(tokens_in,0),
	        COALESCE(tokens_out,0), gate_passed,
	        COALESCE(output,''), COALESCE(prompt,''),
	        COALESCE(artifacts,'')
	 FROM steps WHERE run_id = ?`, id)
```

Update the scan loop (line 150) to read and parse artifacts:

```go
for stepRows.Next() {
	var se stepEntry
	var exitStatus, gatePassed sql.NullInt64
	var artifactsJSON string
	if err := stepRows.Scan(&se.StepID, &se.Model, &se.DurationMs,
		&se.Kind, &exitStatus, &se.TokensIn, &se.TokensOut, &gatePassed,
		&se.Output, &se.Prompt, &artifactsJSON); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if exitStatus.Valid {
		v := int(exitStatus.Int64)
		se.ExitStatus = &v
	}
	if gatePassed.Valid {
		v := gatePassed.Int64 == 1
		se.GatePassed = &v
	}
	if artifactsJSON != "" {
		_ = json.Unmarshal([]byte(artifactsJSON), &se.Artifacts)
	}
	steps = append(steps, se)
}
```

- [ ] **Step 7: Run full Go test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/store/ ./internal/pipeline/ ./internal/gui/ -v -count=1`
Expected: All tests PASS

- [ ] **Step 8: Commit backend changes**

```bash
git add internal/store/schema.go internal/store/store.go internal/store/store_test.go internal/pipeline/runner.go internal/gui/api_runs.go
git commit -m "feat: add artifacts column to steps, expose run_id in templates"
```

---

### Task 2: Frontend — RunDetail page component with header and metadata

**Files:**
- Create: `gui/src/routes/RunDetail.svelte`
- Modify: `gui/src/App.svelte` — add `/run/:id` route
- Modify: `gui/src/routes/WorkflowDetail.svelte:241-276` — run rows link to `/run/:id`

- [ ] **Step 1: Create RunDetail.svelte**

```svelte
<script>
  import { onMount, onDestroy } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { getRun, runWorkflow } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';
  import PipelineGraph from '../lib/components/PipelineGraph.svelte';

  let { params } = $props();
  let runId = $derived(Number(params?.id));

  let run = $state(null);
  let steps = $state([]);
  let loading = $state(true);
  let error = $state(null);
  let pollTimer = $state(null);
  let pollFailures = $state(0);
  let pollPaused = $state(false);
  let elapsed = $state(0);
  let elapsedTimer = $state(null);

  function runStatus(r) {
    if (!r?.finished_at) return 'running';
    return r.exit_status === 0 ? 'pass' : 'fail';
  }

  function formatDuration(ms) {
    if (ms == null || ms <= 0) return '0s';
    const sec = Math.round(ms / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  function formatTokens(input, output) {
    if ((input == null || input === 0) && (output == null || output === 0)) return '--';
    return `${(input || 0).toLocaleString()} / ${(output || 0).toLocaleString()}`;
  }

  function formatCost(cost) {
    if (cost == null || cost === 0) return '--';
    return `$${cost.toFixed(4)}`;
  }

  function formatTime(ms) {
    if (!ms) return '--';
    return new Date(ms).toLocaleString();
  }

  const status = $derived(run ? runStatus(run) : 'running');
  const isRunning = $derived(status === 'running');

  const duration = $derived.by(() => {
    if (!run?.started_at) return '--';
    if (run.finished_at) return formatDuration(run.finished_at - run.started_at);
    return formatDuration(elapsed);
  });

  const breadcrumbSegments = $derived.by(() => {
    const segs = [{ label: 'Workflows', href: '/' }];
    if (run?.workflow_file) {
      segs.push({ label: run.workflow_file.replace('.glitch', ''), href: `/workflow/${encodeURIComponent(run.workflow_file)}` });
    }
    segs.push({ label: `Run #${runId}` });
    return segs;
  });

  async function fetchRun() {
    try {
      const data = await getRun(runId);
      run = data.run || data;
      steps = data.steps || [];
      pollFailures = 0;
      return true;
    } catch (e) {
      pollFailures++;
      if (!run) { error = e.message; }
      return false;
    }
  }

  function startPolling() {
    if (pollTimer) return;
    pollTimer = setInterval(async () => {
      if (pollPaused) return;
      const ok = await fetchRun();
      if (!ok && pollFailures >= 5) {
        pollPaused = true;
      }
      if (run?.finished_at) {
        stopPolling();
        stopElapsed();
      }
    }, 2000);
  }

  function stopPolling() {
    if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
  }

  function reconnect() {
    pollPaused = false;
    pollFailures = 0;
    startPolling();
  }

  function startElapsed() {
    if (elapsedTimer || !run?.started_at) return;
    elapsedTimer = setInterval(() => {
      elapsed = Date.now() - run.started_at;
    }, 1000);
  }

  function stopElapsed() {
    if (elapsedTimer) { clearInterval(elapsedTimer); elapsedTimer = null; }
  }

  async function handleRerun() {
    if (!run?.workflow_file) return;
    try {
      const result = await runWorkflow(run.workflow_file, {});
      if (result.run_id) push(`/run/${result.run_id}`);
    } catch (e) {
      console.error('Re-run failed:', e);
    }
  }

  $effect(() => {
    if (runId) {
      loading = true;
      error = null;
      stopPolling();
      stopElapsed();
      fetchRun().then(() => {
        loading = false;
        if (isRunning) {
          startPolling();
          startElapsed();
        }
      });
    }
  });

  onDestroy(() => {
    stopPolling();
    stopElapsed();
  });
</script>

<div class="run-detail-page">
  <div class="run-header">
    <Breadcrumb segments={breadcrumbSegments} onnavigate={(href) => push(href)} />
    <div class="header-actions">
      <StatusBadge {status} size="md" />
      <button class="primary" onclick={handleRerun}>
        {@html icon('play', 14)} Re-run
      </button>
    </div>
  </div>

  {#if pollPaused}
    <div class="poll-banner">
      <span>Connection lost.</span>
      <button onclick={reconnect}>Reconnect</button>
    </div>
  {/if}

  <div class="metadata-strip">
    <span class="meta-pill">
      <span class="meta-label">Duration</span>
      <span class="meta-val mono">{duration}</span>
    </span>
    <span class="meta-pill">
      <span class="meta-label">Model</span>
      <span class="meta-val mono">{run?.model || '--'}</span>
    </span>
    <span class="meta-pill">
      <span class="meta-label">Tokens</span>
      <span class="meta-val mono">{formatTokens(run?.tokens_in, run?.tokens_out)}</span>
    </span>
    <span class="meta-pill">
      <span class="meta-label">Cost</span>
      <span class="meta-val mono cost">{formatCost(run?.cost_usd)}</span>
    </span>
    <span class="meta-pill">
      <span class="meta-label">Started</span>
      <span class="meta-val">{formatTime(run?.started_at)}</span>
    </span>
  </div>

  <div class="graph-area">
    {#if loading}
      <div class="center-msg"><p class="text-muted">Loading run...</p></div>
    {:else if error}
      <div class="center-msg"><p class="status-fail">{error}</p></div>
    {:else}
      <PipelineGraph {runId} externalSteps={steps} />
    {/if}
  </div>
</div>

<style>
  .run-detail-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .run-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 24px;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .poll-banner {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 8px;
    background: rgba(255, 45, 111, 0.1);
    border-bottom: 1px solid rgba(255, 45, 111, 0.3);
    font-size: 12px;
    color: var(--neon-magenta);
    flex-shrink: 0;
  }
  .poll-banner button {
    font-size: 12px;
    padding: 4px 12px;
  }

  .metadata-strip {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 24px;
    border-bottom: 1px solid var(--border);
    background: var(--bg-surface);
    flex-shrink: 0;
    overflow-x: auto;
  }

  .meta-pill {
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
  }

  .meta-label {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .meta-val {
    font-size: 13px;
  }

  .meta-val.cost {
    color: var(--neon-amber);
  }

  .mono {
    font-family: var(--font-mono);
    font-size: 12px;
  }

  .graph-area {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .center-msg {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .text-muted { color: var(--text-muted); }
</style>
```

- [ ] **Step 2: Add route to App.svelte**

In `gui/src/App.svelte`, add the import and route:

```svelte
<script>
  import Router from 'svelte-spa-router';
  import ActivityBar from './lib/components/ActivityBar.svelte';
  import WorkflowList from './routes/WorkflowList.svelte';
  import WorkflowDetail from './routes/WorkflowDetail.svelte';
  import RunDetail from './routes/RunDetail.svelte';
  import Settings from './routes/Settings.svelte';

  const routes = {
    '/': WorkflowList,
    '/workflow/:name': WorkflowDetail,
    '/run/:id': RunDetail,
    '/settings': Settings,
  };
</script>

<ActivityBar />
<main class="main-area">
  <Router {routes} />
</main>
```

- [ ] **Step 3: Update WorkflowDetail run rows to link to /run/:id**

In `gui/src/routes/WorkflowDetail.svelte`, replace the inline expand behavior (lines 241-276) with navigation:

```svelte
{#each runs.toSorted((a, b) => (b.started_at || 0) - (a.started_at || 0)) as run}
  <button
    class="run-row"
    onclick={() => push(`/run/${run.id}`)}
  >
    <StatusBadge status={runStatus(run)} />
    <span class="run-id mono">#{run.id ?? '--'}</span>
    <span class="run-name">{run.name || run.workflow || '--'}</span>
    <span class="run-time text-muted">
      {run.started_at ? new Date(run.started_at).toLocaleString() : '--'}
    </span>
    <span class="run-duration mono">
      {formatDuration(run.started_at, run.finished_at)}
    </span>
    <span class="run-model text-muted">{run.model || '--'}</span>
    <span class="run-tokens mono text-muted">
      {formatTokens(run.tokens_in, run.tokens_out)}
    </span>
    <span class="run-cost mono">
      {formatCost(run.cost_usd)}
    </span>
    <span class="run-chevron">
      {@html icon('chevronRight')}
    </span>
  </button>
{/each}
```

Remove the `expandedRunId` state variable and the `.expanded` CSS class. Remove the `run-detail` div and PipelineGraph inline import (it's no longer used here). Remove the PipelineGraph import line.

- [ ] **Step 4: Verify the GUI builds**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx vite build`
Expected: Build succeeds with no errors

- [ ] **Step 5: Commit**

```bash
git add gui/src/routes/RunDetail.svelte gui/src/App.svelte gui/src/routes/WorkflowDetail.svelte
git commit -m "feat: add /run/:id page with header, metadata strip, and live polling"
```

---

### Task 3: Frontend — Upgrade PipelineGraph with auto-fit, zoom/pan, and edge glow

**Files:**
- Modify: `gui/src/lib/components/PipelineGraph.svelte`

- [ ] **Step 1: Refactor PipelineGraph to accept external data**

Add an `externalSteps` prop so RunDetail can pass pre-fetched data. When `externalSteps` is provided, skip the internal fetch. Keep backward compatibility for WorkflowDetail (if it still uses PipelineGraph elsewhere).

Replace the props and loadGraph sections at the top of `PipelineGraph.svelte`:

```svelte
<script>
  import { getRun } from '../api.js';
  import ELK from 'elkjs/lib/elk.bundled.js';
  import GraphNode from './GraphNode.svelte';
  import NodePanel from './NodePanel.svelte';

  let { runId, externalSteps = null } = $props();
  let run = $state(null);
  let steps = $state([]);
  let layoutNodes = $state([]);
  let layoutEdges = $state([]);
  let selectedStep = $state(null);
  let loading = $state(true);
  let error = $state(null);
  let svgWidth = $state(800);
  let svgHeight = $state(400);

  // Zoom/pan state
  let scale = $state(1);
  let panX = $state(0);
  let panY = $state(0);
  let isPanning = $state(false);
  let panStart = $state({ x: 0, y: 0 });
  let containerEl = $state(null);
  let naturalWidth = $state(800);
  let naturalHeight = $state(400);

  const elk = new ELK();
```

- [ ] **Step 2: Add auto-fit and zoom/pan logic**

Add these functions after the ELK instance:

```javascript
  function autoFit() {
    if (!containerEl || naturalWidth <= 0) return;
    const rect = containerEl.getBoundingClientRect();
    const scaleX = rect.width / (naturalWidth + 40);
    const scaleY = rect.height / (naturalHeight + 40);
    scale = Math.min(scaleX, scaleY, 1.5);
    panX = (rect.width - naturalWidth * scale) / 2;
    panY = (rect.height - naturalHeight * scale) / 2;
  }

  function handleWheel(e) {
    e.preventDefault();
    const delta = e.deltaY > 0 ? 0.9 : 1.1;
    const newScale = Math.max(0.2, Math.min(3, scale * delta));
    // Zoom toward cursor
    const rect = containerEl.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    panX = mx - (mx - panX) * (newScale / scale);
    panY = my - (my - panY) * (newScale / scale);
    scale = newScale;
  }

  function handleMouseDown(e) {
    if (e.target.closest('.graph-node')) return;
    isPanning = true;
    panStart = { x: e.clientX - panX, y: e.clientY - panY };
  }

  function handleMouseMove(e) {
    if (!isPanning) return;
    panX = e.clientX - panStart.x;
    panY = e.clientY - panStart.y;
  }

  function handleMouseUp() {
    isPanning = false;
  }

  function resetZoom() {
    autoFit();
  }
```

- [ ] **Step 3: Update doLayout to store natural dimensions**

Replace the SVG dimension calculation in `doLayout`:

```javascript
  async function doLayout(stepsData) {
    const graph = buildElkGraph(stepsData);
    if (!graph) {
      layoutNodes = [];
      layoutEdges = [];
      return;
    }

    try {
      const laid = await elk.layout(graph);
      const padding = 20;

      const stepMap = {};
      for (const s of stepsData) stepMap[s.step_id] = s;

      layoutNodes = (laid.children || []).map((n) => ({
        x: n.x + padding,
        y: n.y + padding,
        width: n.width,
        height: n.height,
        step: stepMap[n.id],
      }));

      layoutEdges = (laid.edges || []).map((e) => {
        const adjusted = {
          ...e,
          sections: (e.sections || []).map((s) => ({
            ...s,
            startPoint: { x: s.startPoint.x + padding, y: s.startPoint.y + padding },
            endPoint: { x: s.endPoint.x + padding, y: s.endPoint.y + padding },
            bendPoints: (s.bendPoints || []).map((bp) => ({
              x: bp.x + padding,
              y: bp.y + padding,
            })),
          })),
        };
        const sourceStep = stepMap[e.sources?.[0]];
        const st = sourceStep ? stepStatus(sourceStep) : 'default';
        const lastSection = adjusted.sections?.[adjusted.sections.length - 1];
        return {
          path: edgePath(adjusted),
          color: statusColor(st),
          running: st === 'running',
          endX: lastSection?.endPoint?.x || 0,
          endY: lastSection?.endPoint?.y || 0,
        };
      });

      naturalWidth = (laid.width || 800) + padding * 2;
      naturalHeight = (laid.height || 400) + padding * 2;
      svgWidth = naturalWidth;
      svgHeight = naturalHeight;

      // Auto-fit after layout
      requestAnimationFrame(() => autoFit());
    } catch (e) {
      console.error('ELK layout failed:', e);
    }
  }
```

- [ ] **Step 4: Update loadGraph to support external data**

```javascript
  async function loadGraph() {
    loading = true;
    error = null;
    selectedStep = null;
    try {
      if (externalSteps) {
        steps = externalSteps;
      } else {
        const data = await getRun(runId);
        run = data.run || data;
        steps = data.steps || [];
      }
      await doLayout(steps);
    } catch (e) {
      error = e.message;
      console.error('Failed to load run:', e);
    } finally {
      loading = false;
    }
  }

  // Reload when runId or externalSteps changes
  $effect(() => {
    if (runId || externalSteps) {
      loadGraph();
    }
  });
```

- [ ] **Step 5: Update the template with zoom/pan and edge glow**

Replace the template section:

```svelte
<div class="graph-container" class:panel-open={selectedStep}>
  {#if loading}
    <div class="graph-loading">
      <p class="text-muted">Loading graph...</p>
    </div>
  {:else if error}
    <div class="graph-loading">
      <p class="text-muted">Failed to load run: {error}</p>
    </div>
  {:else if steps.length === 0}
    <div class="graph-loading">
      <p class="text-muted">No steps in this run</p>
    </div>
  {:else}
    <div
      class="graph-canvas"
      bind:this={containerEl}
      onwheel={handleWheel}
      onmousedown={handleMouseDown}
      onmousemove={handleMouseMove}
      onmouseup={handleMouseUp}
      onmouseleave={handleMouseUp}
      role="img"
    >
      <div
        class="graph-transform"
        style="transform: translate({panX}px, {panY}px) scale({scale}); transform-origin: 0 0;"
      >
        <svg width={svgWidth} height={svgHeight} style="position: absolute; top: 0; left: 0;">
          <defs>
            <filter id="glow-cyan" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur"/>
              <feFlood flood-color="#00e5ff" flood-opacity="0.4"/>
              <feComposite in2="blur" operator="in"/>
              <feMerge><feMergeNode/><feMergeNode in="SourceGraphic"/></feMerge>
            </filter>
            <filter id="glow-magenta" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur"/>
              <feFlood flood-color="#ff2d6f" flood-opacity="0.4"/>
              <feComposite in2="blur" operator="in"/>
              <feMerge><feMergeNode/><feMergeNode in="SourceGraphic"/></feMerge>
            </filter>
            <filter id="glow-amber" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur"/>
              <feFlood flood-color="#ffb800" flood-opacity="0.4"/>
              <feComposite in2="blur" operator="in"/>
              <feMerge><feMergeNode/><feMergeNode in="SourceGraphic"/></feMerge>
            </filter>
          </defs>
          {#each layoutEdges as edge}
            <path
              d={edge.path}
              fill="none"
              stroke={edge.color}
              stroke-width="2"
              stroke-opacity="0.8"
              filter="url(#glow-{edge.color === '#00e5ff' ? 'cyan' : edge.color === '#ff2d6f' ? 'magenta' : 'amber'})"
              class:running={edge.running}
            />
            <circle cx={edge.endX} cy={edge.endY} r="4" fill={edge.color} opacity="0.9" />
          {/each}
        </svg>
        {#each layoutNodes as node}
          <div style="position: absolute; left: {node.x}px; top: {node.y}px; width: {node.width}px; height: {node.height}px;">
            <GraphNode
              step={node.step}
              selected={selectedStep?.step_id === node.step.step_id}
              onclick={() => selectedStep = selectedStep?.step_id === node.step.step_id ? null : node.step}
            />
          </div>
        {/each}
      </div>
      <button class="zoom-reset" onclick={resetZoom} title="Reset zoom">
        {@html icon('search', 14)}
      </button>
    </div>
  {/if}

  {#if selectedStep}
    <div class="slide-over" transition:slide>
      <NodePanel step={selectedStep} {runId} onclose={() => selectedStep = null} />
    </div>
  {/if}
</div>
```

- [ ] **Step 6: Update styles**

Replace the `<style>` block:

```css
<style>
  .graph-container {
    display: flex;
    flex-direction: row;
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  .graph-canvas {
    flex: 1;
    overflow: hidden;
    position: relative;
    cursor: grab;
    min-height: 200px;
  }
  .graph-canvas:active {
    cursor: grabbing;
  }

  .graph-transform {
    position: absolute;
    top: 0;
    left: 0;
    will-change: transform;
  }

  .graph-loading {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .text-muted {
    color: var(--text-muted);
    font-size: 13px;
  }

  .zoom-reset {
    position: absolute;
    bottom: 12px;
    right: 12px;
    padding: 6px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 6px;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 12px;
    display: flex;
    align-items: center;
    gap: 4px;
    z-index: 10;
  }
  .zoom-reset:hover {
    color: var(--text-primary);
    border-color: var(--neon-cyan);
  }

  .slide-over {
    width: 420px;
    flex-shrink: 0;
    border-left: 1px solid var(--border);
    background: var(--bg-surface);
    overflow-y: auto;
    height: 100%;
  }

  /* Running edge animation */
  @keyframes dash {
    to { stroke-dashoffset: -24; }
  }
  path.running {
    stroke-dasharray: 8 4;
    animation: dash 1s linear infinite;
  }
</style>
```

- [ ] **Step 7: Verify build**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx vite build`
Expected: Build succeeds

- [ ] **Step 8: Commit**

```bash
git add gui/src/lib/components/PipelineGraph.svelte
git commit -m "feat: pipeline graph auto-fit, zoom/pan, edge glow effects"
```

---

### Task 4: Frontend — Upgrade NodePanel to slide-over with Files tab

**Files:**
- Modify: `gui/src/lib/components/NodePanel.svelte`

- [ ] **Step 1: Update NodePanel with Files tab and modern styling**

Replace `gui/src/lib/components/NodePanel.svelte`:

```svelte
<script>
  import { renderMarkdown } from '../markdown.js';
  import { getResultText } from '../api.js';

  let { step, runId, onclose } = $props();
  let activeSection = $state('output');
  let fileContent = $state(null);
  let viewingFile = $state(null);
  let fileLoading = $state(false);

  const sections = ['output', 'prompt', 'metrics', 'files'];

  const status = $derived.by(() => {
    if (step.exit_status === undefined || step.exit_status === null) {
      if (step.gate_passed) return 'pass';
      return 'running';
    }
    return step.exit_status === 0 ? 'pass' : 'fail';
  });

  const statusColor = $derived(
    status === 'pass' ? 'var(--neon-cyan)' :
    status === 'fail' ? 'var(--neon-magenta)' :
    'var(--neon-amber)'
  );

  function formatDuration(ms) {
    if (ms == null) return '--';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${Math.floor(ms / 60000)}m ${((ms % 60000) / 1000).toFixed(0)}s`;
  }

  function formatTokens(n) {
    if (n == null) return '--';
    return n.toLocaleString();
  }

  function fileName(path) {
    return path.split('/').pop();
  }

  async function viewFile(path) {
    viewingFile = path;
    fileLoading = true;
    try {
      fileContent = await getResultText(path);
    } catch (e) {
      fileContent = `Error loading file: ${e.message}`;
    } finally {
      fileLoading = false;
    }
  }

  function backToFileList() {
    viewingFile = null;
    fileContent = null;
  }

  // Reset to output tab when step changes
  $effect(() => {
    if (step) {
      activeSection = 'output';
      viewingFile = null;
      fileContent = null;
    }
  });

  function handleKeydown(e) {
    if (e.key === 'Escape') onclose?.();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="node-panel">
  <div class="panel-header">
    <div class="panel-title-row">
      <span class="panel-title mono">{step.step_id || '--'}</span>
      {#if step.kind}
        <span class="kind-badge kind-{step.kind}">{step.kind}</span>
      {/if}
    </div>
    <div class="panel-header-right">
      <span class="panel-status" style="color: {statusColor}">{status}</span>
      <span class="panel-duration mono">{formatDuration(step.duration_ms)}</span>
      <button class="close-btn" onclick={onclose} type="button">&times;</button>
    </div>
  </div>

  <div class="panel-tabs">
    {#each sections as section}
      <button
        class="panel-tab"
        class:active={activeSection === section}
        onclick={() => { activeSection = section; viewingFile = null; fileContent = null; }}
        type="button"
      >
        {section.charAt(0).toUpperCase() + section.slice(1)}
        {#if section === 'files' && step.artifacts?.length}
          <span class="tab-count">{step.artifacts.length}</span>
        {/if}
      </button>
    {/each}
  </div>

  <div class="panel-content">
    {#if activeSection === 'output'}
      {#if step.output}
        {#if step.kind === 'run'}
          <pre class="output-code"><code>{step.output}</code></pre>
        {:else}
          <div class="rendered-content">
            {@html renderMarkdown(step.output)}
          </div>
        {/if}
      {:else if step.exit_status == null}
        <div class="placeholder-state">
          <div class="spinner"></div>
          <p>Step still running...</p>
        </div>
      {:else}
        <p class="text-muted placeholder">No output captured</p>
      {/if}

    {:else if activeSection === 'prompt'}
      {#if step.prompt}
        <pre class="prompt-block"><code>{step.prompt}</code></pre>
      {:else}
        <p class="text-muted placeholder">No prompt for this step</p>
      {/if}

    {:else if activeSection === 'metrics'}
      <div class="metrics-grid">
        <span class="metric-label">Kind</span>
        <span class="metric-value">{step.kind || '--'}</span>

        <span class="metric-label">Model</span>
        <span class="metric-value mono">{step.model || '--'}</span>

        <span class="metric-label">Duration</span>
        <span class="metric-value mono">{formatDuration(step.duration_ms)}</span>

        <span class="metric-label">Tokens In</span>
        <span class="metric-value mono">{formatTokens(step.tokens_in)}</span>

        <span class="metric-label">Tokens Out</span>
        <span class="metric-value mono">{formatTokens(step.tokens_out)}</span>

        {#if step.gate_passed != null}
          <span class="metric-label">Gate</span>
          <span class="metric-value">{step.gate_passed ? '\u2713 Passed' : '\u2717 Failed'}</span>
        {/if}

        <span class="metric-label">Exit Status</span>
        <span class="metric-value mono">
          {#if step.exit_status == null}
            <span class="text-muted">--</span>
          {:else if step.exit_status === 0}
            <span style="color: var(--neon-green);">0 (success)</span>
          {:else}
            <span style="color: var(--neon-magenta);">{step.exit_status}</span>
          {/if}
        </span>
      </div>

    {:else if activeSection === 'files'}
      {#if viewingFile}
        <div class="file-viewer">
          <button class="back-btn" onclick={backToFileList} type="button">&larr; Back</button>
          <span class="file-path mono">{viewingFile}</span>
          {#if fileLoading}
            <p class="text-muted">Loading...</p>
          {:else}
            <pre class="file-content"><code>{fileContent}</code></pre>
          {/if}
        </div>
      {:else if step.artifacts?.length}
        <div class="file-list">
          {#each step.artifacts as path}
            <button class="file-item" onclick={() => viewFile(path)} type="button">
              <span class="file-name mono">{fileName(path)}</span>
              <span class="file-path-hint text-muted">{path}</span>
            </button>
          {/each}
        </div>
      {:else}
        <p class="text-muted placeholder">No files produced by this step</p>
      {/if}
    {/if}
  </div>
</div>

<style>
  .node-panel {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    gap: 12px;
  }

  .panel-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .panel-title {
    font-size: 13px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .panel-header-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .panel-status {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .panel-duration {
    font-size: 11px;
    color: var(--text-muted);
  }

  .kind-badge {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 1px 6px;
    border-radius: 4px;
    flex-shrink: 0;
    font-weight: 600;
    line-height: 1.4;
  }
  .kind-llm { background: rgba(0, 229, 255, 0.2); color: #00e5ff; }
  .kind-run { background: rgba(0, 255, 159, 0.2); color: #00ff9f; }
  .kind-cond { background: rgba(255, 184, 0, 0.2); color: #ffb800; }
  .kind-map { background: rgba(255, 45, 111, 0.2); color: #ff2d6f; }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 18px;
    cursor: pointer;
    padding: 0 4px;
    line-height: 1;
    border-radius: 4px;
  }
  .close-btn:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.05);
  }

  .panel-tabs {
    display: flex;
    border-bottom: 1px solid var(--border);
  }

  .panel-tab {
    padding: 8px 12px;
    font-size: 12px;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.15s;
    border-radius: 0;
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .panel-tab:hover { color: var(--text-primary); }
  .panel-tab.active {
    color: var(--neon-cyan);
    border-bottom-color: var(--neon-cyan);
  }

  .tab-count {
    font-size: 10px;
    background: rgba(0, 229, 255, 0.15);
    color: var(--neon-cyan);
    padding: 0 5px;
    border-radius: 8px;
    font-weight: 600;
  }

  .panel-content {
    padding: 16px;
    flex: 1;
    overflow-y: auto;
  }

  .metrics-grid {
    display: grid;
    grid-template-columns: 100px 1fr;
    gap: 10px 12px;
  }
  .metric-label {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    padding-top: 1px;
  }
  .metric-value { font-size: 13px; }

  .mono {
    font-family: var(--font-mono, 'JetBrains Mono', monospace);
    font-size: 12px;
  }

  .text-muted { color: var(--text-muted); }

  .placeholder {
    font-size: 13px;
    font-style: italic;
    padding: 16px 0;
  }

  .output-code {
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    overflow-x: auto;
    font-size: 12px;
    line-height: 1.5;
    font-family: var(--font-mono);
    color: var(--neon-green);
    white-space: pre-wrap;
    word-break: break-word;
  }

  .prompt-block {
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    overflow-x: auto;
    font-size: 12px;
    line-height: 1.5;
  }
  .prompt-block code {
    font-family: var(--font-mono, 'JetBrains Mono', monospace);
    color: var(--text-primary);
  }

  .placeholder-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 32px 0;
    color: var(--text-muted);
    font-size: 13px;
  }

  .spinner {
    width: 20px;
    height: 20px;
    border: 2px solid var(--border);
    border-top-color: var(--neon-amber);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Files */
  .file-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .file-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 8px 12px;
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    cursor: pointer;
    text-align: left;
    color: var(--text-primary);
    transition: all 0.15s;
    width: 100%;
  }
  .file-item:hover {
    border-color: var(--neon-cyan);
    background: var(--bg-elevated);
  }

  .file-name { font-size: 13px; }
  .file-path-hint { font-size: 11px; }

  .file-viewer {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .back-btn {
    align-self: flex-start;
    font-size: 12px;
    padding: 4px 8px;
    background: none;
    border: 1px solid var(--border);
    border-radius: 4px;
    color: var(--text-muted);
    cursor: pointer;
  }
  .back-btn:hover {
    color: var(--text-primary);
    border-color: var(--neon-cyan);
  }

  .file-path {
    font-size: 11px;
    color: var(--text-muted);
  }

  .file-content {
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    overflow-x: auto;
    font-size: 12px;
    line-height: 1.5;
    font-family: var(--font-mono);
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 400px;
    overflow-y: auto;
  }

  .rendered-content :global(pre) {
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    overflow-x: auto;
    font-size: 12px;
  }
  .rendered-content :global(code) {
    font-family: var(--font-mono);
  }
  .rendered-content :global(p) {
    margin-bottom: 8px;
    line-height: 1.6;
  }
</style>
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx vite build`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add gui/src/lib/components/NodePanel.svelte
git commit -m "feat: upgrade NodePanel with slide-over layout, Files tab, and modern styling"
```

---

### Task 5: Frontend — Artifacts bar below graph

**Files:**
- Create: `gui/src/lib/components/ArtifactsBar.svelte`
- Modify: `gui/src/routes/RunDetail.svelte` — add ArtifactsBar below graph area

- [ ] **Step 1: Create ArtifactsBar.svelte**

```svelte
<script>
  import { getResultText } from '../api.js';

  let { steps = [] } = $props();
  let expanded = $state(false);
  let viewingFile = $state(null);
  let fileContent = $state(null);
  let fileLoading = $state(false);

  const artifactGroups = $derived.by(() => {
    const groups = [];
    for (const step of steps) {
      if (step.artifacts?.length) {
        groups.push({ stepId: step.step_id, files: step.artifacts });
      }
    }
    return groups;
  });

  const totalCount = $derived(
    artifactGroups.reduce((sum, g) => sum + g.files.length, 0)
  );

  function fileName(path) {
    return path.split('/').pop();
  }

  async function viewFile(path) {
    viewingFile = path;
    fileLoading = true;
    try {
      fileContent = await getResultText(path);
    } catch (e) {
      fileContent = `Error: ${e.message}`;
    } finally {
      fileLoading = false;
    }
  }

  function closeViewer() {
    viewingFile = null;
    fileContent = null;
  }
</script>

{#if totalCount > 0}
  <div class="artifacts-bar">
    <button class="artifacts-toggle" onclick={() => { expanded = !expanded; viewingFile = null; }} type="button">
      <span class="artifacts-label">Artifacts</span>
      <span class="artifacts-count">{totalCount}</span>
      <span class="artifacts-chevron">{expanded ? '\u25BE' : '\u25B8'}</span>
    </button>

    {#if expanded}
      <div class="artifacts-content">
        {#if viewingFile}
          <div class="artifact-viewer">
            <button class="back-btn" onclick={closeViewer} type="button">&larr; Back</button>
            <span class="file-path mono">{viewingFile}</span>
            {#if fileLoading}
              <p class="text-muted">Loading...</p>
            {:else}
              <pre class="file-content"><code>{fileContent}</code></pre>
            {/if}
          </div>
        {:else}
          {#each artifactGroups as group}
            <div class="artifact-group">
              <span class="group-label mono">{group.stepId}</span>
              <div class="group-files">
                {#each group.files as path}
                  <button class="artifact-file" onclick={() => viewFile(path)} type="button">
                    <span class="mono">{fileName(path)}</span>
                    <span class="file-hint text-muted">{path}</span>
                  </button>
                {/each}
              </div>
            </div>
          {/each}
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .artifacts-bar {
    border-top: 1px solid var(--border);
    background: var(--bg-surface);
    flex-shrink: 0;
  }

  .artifacts-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 24px;
    width: 100%;
    background: none;
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 12px;
    text-align: left;
  }
  .artifacts-toggle:hover {
    background: var(--bg-elevated);
  }

  .artifacts-label {
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 11px;
    color: var(--text-muted);
  }

  .artifacts-count {
    font-size: 10px;
    background: rgba(0, 229, 255, 0.15);
    color: var(--neon-cyan);
    padding: 0 6px;
    border-radius: 8px;
    font-weight: 600;
    font-family: var(--font-mono);
  }

  .artifacts-chevron {
    color: var(--text-muted);
    margin-left: auto;
  }

  .artifacts-content {
    padding: 0 24px 16px;
    max-height: 300px;
    overflow-y: auto;
  }

  .artifact-group {
    margin-bottom: 12px;
  }

  .group-label {
    font-size: 11px;
    color: var(--neon-cyan);
    display: block;
    margin-bottom: 4px;
  }

  .group-files {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .artifact-file {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 6px 10px;
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 4px;
    cursor: pointer;
    text-align: left;
    color: var(--text-primary);
    font-size: 12px;
    width: 100%;
  }
  .artifact-file:hover {
    border-color: var(--neon-cyan);
  }

  .file-hint {
    font-size: 10px;
  }

  .artifact-viewer {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .back-btn {
    align-self: flex-start;
    font-size: 12px;
    padding: 4px 8px;
    background: none;
    border: 1px solid var(--border);
    border-radius: 4px;
    color: var(--text-muted);
    cursor: pointer;
  }
  .back-btn:hover {
    color: var(--text-primary);
    border-color: var(--neon-cyan);
  }

  .file-path { font-size: 11px; color: var(--text-muted); }

  .file-content {
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    font-size: 12px;
    line-height: 1.5;
    font-family: var(--font-mono);
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 250px;
    overflow-y: auto;
  }

  .mono { font-family: var(--font-mono); }
  .text-muted { color: var(--text-muted); }
</style>
```

- [ ] **Step 2: Add ArtifactsBar to RunDetail**

In `gui/src/routes/RunDetail.svelte`, add the import and component:

Add to imports:
```javascript
import ArtifactsBar from '../lib/components/ArtifactsBar.svelte';
```

Add after the `.graph-area` div closing tag:
```svelte
  <ArtifactsBar {steps} />
</div>
```

So the bottom of the template looks like:
```svelte
  <div class="graph-area">
    {#if loading}
      <div class="center-msg"><p class="text-muted">Loading run...</p></div>
    {:else if error}
      <div class="center-msg"><p class="status-fail">{error}</p></div>
    {:else}
      <PipelineGraph {runId} externalSteps={steps} />
    {/if}
  </div>

  <ArtifactsBar {steps} />
</div>
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx vite build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add gui/src/lib/components/ArtifactsBar.svelte gui/src/routes/RunDetail.svelte
git commit -m "feat: add ArtifactsBar component below pipeline graph"
```

---

### Task 6: Add .gitignore entry and clean up test-results

**Files:**
- Modify: `gui/.gitignore` (or project root `.gitignore`)

- [ ] **Step 1: Add test-results to .gitignore**

Check which .gitignore exists and add `test-results/` to it:

```bash
echo "test-results/" >> /Users/stokes/Projects/gl1tch/gui/.gitignore
```

If `gui/.gitignore` doesn't exist, add to root `.gitignore`:

```bash
echo "gui/test-results/" >> /Users/stokes/Projects/gl1tch/.gitignore
```

- [ ] **Step 2: Remove tracked test-results if present**

```bash
git rm -r --cached gui/test-results/ 2>/dev/null || true
```

- [ ] **Step 3: Commit**

```bash
git add .gitignore gui/.gitignore 2>/dev/null; git add .
git commit -m "chore: gitignore test-results directory"
```

---

### Task 7: Playwright E2E tests for Run Detail page

**Files:**
- Create: `gui/e2e/run-detail.spec.js`

- [ ] **Step 1: Write E2E tests**

```javascript
import { test, expect } from '@playwright/test'

// Helper: run a workflow and wait for it to finish
async function runAndWait(request, workflow, params = {}, maxWaitMs = 15000) {
  const resp = await request.post(`/api/workflows/${workflow}/run`, {
    data: { params },
  })
  expect(resp.ok()).toBeTruthy()
  const { run_id } = await resp.json()

  const start = Date.now()
  let detail
  while (Date.now() - start < maxWaitMs) {
    const r = await request.get(`/api/runs/${run_id}`)
    detail = await r.json()
    if (detail.run?.finished_at) break
    await new Promise(r => setTimeout(r, 500))
  }
  return { runId: run_id, detail }
}

test.describe('Run Detail Page', () => {
  test('navigates to /run/:id and shows header', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.run-header')).toBeVisible()
    await expect(page.locator('.breadcrumb')).toContainText('Run')
    await expect(page.locator('.breadcrumb')).toContainText(`#${runId}`)
  })

  test('shows metadata strip with status, duration, model', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.metadata-strip')).toBeVisible()
    // Should show at least Duration and Started
    await expect(page.locator('.meta-label').filter({ hasText: 'Duration' })).toBeVisible()
    await expect(page.locator('.meta-label').filter({ hasText: 'Started' })).toBeVisible()
  })

  test('renders pipeline graph with nodes', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')

    await page.goto(`/#/run/${runId}`)
    // Wait for graph nodes to render
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    const nodeCount = await page.locator('.graph-node').count()
    expect(nodeCount).toBeGreaterThanOrEqual(2)
  })

  test('clicking a node opens slide-over panel', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })

    // Click first node
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.node-panel')).toBeVisible()
    await expect(page.locator('.panel-title')).toBeVisible()
  })

  test('panel shows Output tab by default', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()

    // Output tab should be active
    await expect(page.locator('.panel-tab.active')).toContainText('Output')
  })

  test('panel tabs switch content', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()

    // Switch to Metrics
    await page.locator('.panel-tab').filter({ hasText: 'Metrics' }).click()
    await expect(page.locator('.metrics-grid')).toBeVisible()

    // Switch to Prompt
    await page.locator('.panel-tab').filter({ hasText: 'Prompt' }).click()
    // Should show prompt or placeholder
    const panel = page.locator('.panel-content')
    await expect(panel).toBeVisible()
  })

  test('close panel via X button', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.node-panel')).toBeVisible()

    await page.locator('.close-btn').click()
    await expect(page.locator('.node-panel')).not.toBeVisible()
  })

  test('close panel via Escape key', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await page.locator('.graph-node').first().click()
    await expect(page.locator('.node-panel')).toBeVisible()

    await page.keyboard.press('Escape')
    await expect(page.locator('.node-panel')).not.toBeVisible()
  })

  test('clicking different node swaps panel content', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })

    const nodes = page.locator('.graph-node')
    const nodeCount = await nodes.count()
    if (nodeCount < 2) return // skip if only one node

    await nodes.first().click()
    const firstTitle = await page.locator('.panel-title').textContent()

    await nodes.nth(1).click()
    const secondTitle = await page.locator('.panel-title').textContent()

    expect(firstTitle).not.toEqual(secondTitle)
  })

  test('breadcrumb links navigate back', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.breadcrumb')).toBeVisible()

    // Click "Workflows" breadcrumb link
    await page.locator('.breadcrumb a').first().click()
    await expect(page).toHaveURL(/\/#\/$/)
  })

  test('graph edges are visible (2px stroke)', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })

    // Check that SVG path elements exist with stroke-width 2
    const edges = page.locator('svg path[stroke-width="2"]')
    const count = await edges.count()
    expect(count).toBeGreaterThan(0)
  })

  test('zoom reset button is visible', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-multi.glitch')

    await page.goto(`/#/run/${runId}`)
    await expect(page.locator('.graph-node').first()).toBeVisible({ timeout: 10000 })
    await expect(page.locator('.zoom-reset')).toBeVisible()
  })

  test('workflow detail run rows navigate to /run/:id', async ({ page, request }) => {
    const { runId } = await runAndWait(request, 'test-echo.glitch')

    await page.goto(`/#/workflow/test-echo.glitch`)
    await expect(page.locator('.run-row').first()).toBeVisible({ timeout: 10000 })

    await page.locator('.run-row').first().click()
    await expect(page).toHaveURL(/\/#\/run\/\d+/)
  })
})
```

- [ ] **Step 2: Run the E2E tests**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx playwright test e2e/run-detail.spec.js --reporter=list`
Expected: All tests PASS (requires the GUI server to be running)

- [ ] **Step 3: Commit**

```bash
git add gui/e2e/run-detail.spec.js
git commit -m "test: add Playwright E2E tests for Run Detail page"
```

---

### Task 8: Integration test — verify full flow

- [ ] **Step 1: Run Go tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/store/ ./internal/pipeline/ ./internal/gui/ -v -count=1`
Expected: All PASS

- [ ] **Step 2: Build GUI**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx vite build`
Expected: Build succeeds

- [ ] **Step 3: Run all Playwright tests**

Run: `cd /Users/stokes/Projects/gl1tch/gui && npx playwright test --reporter=list`
Expected: All existing + new tests PASS

- [ ] **Step 4: Final commit if any fixups needed**

Only commit if changes were made during fixups.
