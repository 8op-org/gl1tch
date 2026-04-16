<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { listRuns } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import StatusBadge from '../lib/components/StatusBadge.svelte';
  import RunTree from '../lib/components/RunTree.svelte';

  let runs = $state([]);
  let error = $state(null);
  let loading = $state(true);
  let filterStatus = $state('all');
  let view = $state('flat'); // 'flat' | 'tree'

  // Derive status from API fields: exit_status + finished_at
  function deriveStatus(run) {
    if (!run.finished_at) return 'RUNNING';
    return run.exit_status === 0 ? 'PASS' : 'FAIL';
  }

  // Normalize API run entry to frontend shape
  function normalize(run) {
    return {
      ...run,
      status: deriveStatus(run),
      workflow: run.workflow_file || run.name || '',
      started: run.started_at ? new Date(run.started_at * 1000) : null,
      finished: run.finished_at ? new Date(run.finished_at * 1000) : null,
    };
  }

  const normalized = $derived(runs.map(normalize));
  const filtered = $derived(filterStatus === 'all' ? normalized : normalized.filter(r => r.status === filterStatus));
  // Top-level runs (no parent) for tree view
  const topLevel = $derived(runs.filter(r => !r.parent_run_id || r.parent_run_id === 0));

  function duration(started, finished) {
    if (!started) return '--';
    const end = finished || new Date();
    const sec = Math.round((end - started) / 1000);
    if (sec < 60) return `${sec}s`;
    const min = Math.floor(sec / 60);
    return `${min}m ${sec % 60}s`;
  }

  function relativeTime(ts) {
    if (!ts) return '';
    const diff = Date.now() - ts.getTime();
    const min = Math.round(diff / 60000);
    if (min < 1) return 'just now';
    if (min < 60) return `${min}m ago`;
    const hr = Math.round(min / 60);
    if (hr < 24) return `${hr}h ago`;
    return `${Math.round(hr / 24)}d ago`;
  }

  // Strip .glitch extension for display
  function displayName(wf) {
    return wf.replace(/\.glitch$/, '');
  }

  onMount(async () => {
    try { runs = await listRuns(); } catch (e) { error = e.message; } finally { loading = false; }
  });
</script>

<div class="page-header">
  <h1>{@html icon('terminal', 20)} Runs</h1>
  <div class="flex gap-sm">
    <div class="tabs">
      <button class:active={view === 'flat'} onclick={() => (view = 'flat')}>Flat</button>
      <button class:active={view === 'tree'} onclick={() => (view = 'tree')}>Tree</button>
    </div>
    {#if view === 'flat'}
      <select class="status-filter" bind:value={filterStatus}>
        <option value="all">All statuses</option>
        <option value="PASS">Pass</option>
        <option value="FAIL">Fail</option>
        <option value="RUNNING">Running</option>
      </select>
    {/if}
  </div>
</div>
<div class="page-content">
  {#if loading}<p class="text-muted">Loading runs...</p>
  {:else if error}<p class="status-fail">{error}</p>
  {:else if view === 'flat'}
    {#if filtered.length === 0}
      <div class="empty-state">
        <span class="empty-icon">{@html icon('terminal', 48)}</span>
        <p class="text-muted">No runs found.</p>
      </div>
    {:else}
      <div class="run-count text-muted">{filtered.length} run{filtered.length !== 1 ? 's' : ''}</div>
      <table class="runs-table">
        <thead><tr><th>ID</th><th>Workflow</th><th>Status</th><th>Duration</th><th>Started</th><th>Model</th></tr></thead>
        <tbody>
          {#each filtered as run}
            <tr class="run-row" onclick={() => push(`/run/${run.id}`)}>
              <td class="mono text-cyan run-id">#{run.id}</td>
              <td class="mono run-name">{displayName(run.workflow)}</td>
              <td><StatusBadge status={run.status} /></td>
              <td class="mono run-duration">{duration(run.started, run.finished)}</td>
              <td class="text-muted run-started">{relativeTime(run.started)}</td>
              <td class="mono text-muted run-model">{run.model || '--'}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  {:else}
    {#if topLevel.length === 0}
      <div class="empty-state">
        <span class="empty-icon">{@html icon('terminal', 48)}</span>
        <p class="text-muted">No runs found.</p>
      </div>
    {:else}
      <div class="run-count text-muted">{topLevel.length} top-level run{topLevel.length !== 1 ? 's' : ''}</div>
      <div class="tree-view">
        {#each topLevel as r (r.id)}
          <RunTree runId={r.id} />
        {/each}
      </div>
    {/if}
  {/if}
</div>

<style>
  h1 { display: flex; align-items: center; gap: 8px; }
  .run-row { cursor: pointer; transition: background 0.1s; }
  .run-row:hover td { background: var(--bg-elevated); }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .text-cyan { color: var(--neon-cyan); }
  .run-count { font-size: 12px; margin-bottom: 12px; }
  .status-filter { min-width: 140px; }
  .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; padding: 64px 0; opacity: 0.5; }
  .empty-icon { color: var(--text-muted); }
  .runs-table { margin-top: 0; }
  .tabs { display: inline-flex; gap: 0; border: 1px solid var(--border); border-radius: 4px; overflow: hidden; }
  .tabs button { background: none; border: none; padding: 6px 12px; color: var(--text-muted); cursor: pointer; font-size: 12px; font-family: var(--font-mono); }
  .tabs button:hover { color: var(--text-primary); background: var(--bg-elevated); }
  .tabs button.active { background: var(--bg-elevated); color: var(--neon-cyan); }
  .tree-view { display: flex; flex-direction: column; gap: 8px; padding: 8px 0; }
</style>
