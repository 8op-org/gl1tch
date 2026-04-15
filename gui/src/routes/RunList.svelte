<script>
  import { onMount } from 'svelte'
  import { link } from 'svelte-spa-router'
  import { api } from '../lib/api.js'

  let runs = $state([])
  let error = $state(null)

  onMount(async () => {
    try {
      runs = await api.listRuns()
    } catch (e) {
      error = e.message
    }
  })

  function formatTime(ms) {
    if (!ms) return '-'
    return new Date(ms).toLocaleString()
  }

  function duration(run) {
    if (!run.finished_at || !run.started_at) return '-'
    const secs = (run.finished_at - run.started_at) / 1000
    return `${secs.toFixed(1)}s`
  }
</script>

<div class="run-list">
  <h2>Runs</h2>
  {#if error}
    <p class="error">{error}</p>
  {:else if runs.length === 0}
    <p class="muted">No runs yet.</p>
  {:else}
    <table>
      <thead>
        <tr><th>ID</th><th>Workflow</th><th>Status</th><th>Duration</th><th>Started</th></tr>
      </thead>
      <tbody>
        {#each runs as run}
          <tr>
            <td><a href="/run/{run.id}" use:link>#{run.id}</a></td>
            <td class="mono">{run.name}</td>
            <td>
              <span class="badge" class:success={run.exit_status === 0 && run.finished_at} class:fail={run.exit_status !== 0 && run.finished_at} class:pending={!run.finished_at}>
                {#if !run.finished_at}RUNNING{:else if run.exit_status === 0}PASS{:else}FAIL{/if}
              </span>
            </td>
            <td>{duration(run)}</td>
            <td class="muted">{formatTime(run.started_at)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  h2 { margin-bottom: 1rem; }
  table { width: 100%; border-collapse: collapse; }
  th, td { text-align: left; padding: 0.5rem; border-bottom: 1px solid var(--border); font-size: 13px; }
  th { color: var(--text-muted); font-weight: normal; }
  .mono { font-family: var(--font-mono); }
  a { color: var(--accent); text-decoration: none; }
  a:hover { text-decoration: underline; }
  .badge { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; }
  .badge.success { background: var(--success); color: #000; }
  .badge.fail { background: var(--danger); color: #fff; }
  .badge.pending { background: var(--border); color: var(--text); }
  .error { color: var(--danger); }
  .muted { color: var(--text-muted); }
</style>
