<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { listRuns } from '../lib/api.js';
  import StatusBadge from '../lib/components/StatusBadge.svelte';

  let runs = $state([]);
  let error = $state(null);
  let loading = $state(true);
  let filterStatus = $state('all');

  const filtered = $derived(filterStatus === 'all' ? runs : runs.filter(r => r.status?.toUpperCase() === filterStatus));

  function duration(started, finished) {
    if (!started) return '--';
    const start = new Date(started);
    const end = finished ? new Date(finished) : new Date();
    const sec = Math.round((end - start) / 1000);
    if (sec < 60) return `${sec}s`;
    const min = Math.floor(sec / 60);
    return `${min}m ${sec % 60}s`;
  }

  function relativeTime(ts) {
    if (!ts) return '';
    const diff = Date.now() - new Date(ts).getTime();
    const min = Math.round(diff / 60000);
    if (min < 1) return 'just now';
    if (min < 60) return `${min}m ago`;
    const hr = Math.round(min / 60);
    if (hr < 24) return `${hr}h ago`;
    return `${Math.round(hr / 24)}d ago`;
  }

  onMount(async () => {
    try { runs = await listRuns(); } catch (e) { error = e.message; } finally { loading = false; }
  });
</script>

<div class="page-header">
  <h1>Runs</h1>
  <div class="flex gap-sm">
    <select bind:value={filterStatus}>
      <option value="all">All</option>
      <option value="PASS">Pass</option>
      <option value="FAIL">Fail</option>
      <option value="RUNNING">Running</option>
    </select>
  </div>
</div>
<div class="page-content">
  {#if loading}<p class="text-muted">Loading runs...</p>
  {:else if error}<p class="status-fail">{error}</p>
  {:else if filtered.length === 0}<p class="text-muted">No runs found.</p>
  {:else}
    <table>
      <thead><tr><th>ID</th><th>Workflow</th><th>Status</th><th>Duration</th><th>Started</th></tr></thead>
      <tbody>
        {#each filtered as run}
          <tr class="clickable" on:click={() => push(`/run/${run.id}`)}>
            <td class="mono text-cyan">#{run.id}</td>
            <td class="mono">{run.workflow || run.name || ''}</td>
            <td><StatusBadge status={run.status} /></td>
            <td class="mono">{duration(run.started, run.finished)}</td>
            <td class="text-muted">{relativeTime(run.started)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .clickable { cursor: pointer; }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .text-cyan { color: var(--neon-cyan); }
</style>
