<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { renderMarkdown } from '../lib/markdown.js'

  let { params = {} } = $props()

  let run = $state(null)
  let steps = $state([])
  let kibanaURL = $state(null)
  let error = $state(null)

  onMount(async () => {
    try {
      const data = await api.getRun(params.id)
      run = data.run
      steps = data.steps || []
      const kibana = await api.getKibanaRun(params.id)
      kibanaURL = kibana.url
    } catch (e) {
      error = e.message
    }
  })

  function formatTime(ms) {
    if (!ms) return '-'
    return new Date(ms).toLocaleString()
  }
</script>

<div class="run-view">
  {#if error}
    <p class="error">{error}</p>
  {:else if !run}
    <p>Loading...</p>
  {:else}
    <div class="header">
      <h2>{run.name}</h2>
      <span class="badge" class:success={run.exit_status === 0} class:fail={run.exit_status !== 0}>
        {run.exit_status === 0 ? 'PASS' : 'FAIL'}
      </span>
    </div>

    <div class="meta">
      <span>Started: {formatTime(run.started_at)}</span>
      <span>Finished: {formatTime(run.finished_at)}</span>
    </div>

    <h3>Steps</h3>
    <table>
      <thead><tr><th>Step</th><th>Model</th><th>Duration</th></tr></thead>
      <tbody>
        {#each steps as step}
          <tr>
            <td class="mono">{step.step_id}</td>
            <td>{step.model || '-'}</td>
            <td>{step.duration_ms ? `${(step.duration_ms / 1000).toFixed(1)}s` : '-'}</td>
          </tr>
        {/each}
      </tbody>
    </table>

    {#if run.output}
      <h3>Output</h3>
      <div class="output">{@html renderMarkdown(run.output)}</div>
    {/if}

    {#if kibanaURL}
      <h3>Telemetry</h3>
      <iframe src={kibanaURL} title="Kibana" class="kibana-frame"></iframe>
    {/if}
  {/if}
</div>

<style>
  .header { display: flex; align-items: center; gap: 1rem; margin-bottom: 0.5rem; }
  h2 { font-family: var(--font-mono); }
  .badge { padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
  .badge.success { background: var(--success); color: #000; }
  .badge.fail { background: var(--danger); color: #fff; }
  .meta { color: var(--text-muted); font-size: 12px; display: flex; gap: 1.5rem; margin-bottom: 1rem; }
  h3 { margin-top: 1.5rem; margin-bottom: 0.5rem; font-size: 14px; }
  table { width: 100%; border-collapse: collapse; }
  th, td { text-align: left; padding: 0.5rem; border-bottom: 1px solid var(--border); font-size: 13px; }
  th { color: var(--text-muted); font-weight: normal; }
  .mono { font-family: var(--font-mono); }
  .output { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 4px; padding: 1rem; }
  .kibana-frame { width: 100%; height: 400px; border: 1px solid var(--border); border-radius: 4px; margin-top: 0.5rem; }
  .error { color: var(--danger); }
</style>
