<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { getRun, getKibanaRun } from '../lib/api.js';
  import { renderMarkdown } from '../lib/markdown.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';

  let { params } = $props();
  let id = $derived(params?.id || '');
  let run = $state(null);
  let kibanaUrl = $state(null);
  let error = $state(null);
  let showTelemetry = $state(false);

  const breadcrumbs = $derived([{ label: 'Runs', href: '#/runs' }, { label: `#${id}${run?.workflow ? ' ' + run.workflow : ''}` }]);

  function duration(started, finished) {
    if (!started) return '--';
    const sec = Math.round((new Date(finished || Date.now()) - new Date(started)) / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  onMount(async () => {
    try { run = await getRun(id); try { kibanaUrl = (await getKibanaRun(id)).url; } catch (_) {} } catch (e) { error = e.message; }
  });
</script>

<div class="page-header">
  <Breadcrumb segments={breadcrumbs} onnavigate={(href) => push(href.replace('#', ''))} />
  <div class="flex items-center gap-sm">{#if run}<StatusBadge status={run.status} size="md" />{/if}</div>
</div>

<div class="page-content">
  {#if error}<p class="status-fail">{error}</p>
  {:else if !run}<p class="text-muted">Loading...</p>
  {:else}
    <div class="meta-row">
      <div class="meta-item"><span class="meta-label">Started</span><span class="mono">{run.started ? new Date(run.started).toLocaleString() : '--'}</span></div>
      <div class="meta-item"><span class="meta-label">Duration</span><span class="mono">{duration(run.started, run.finished)}</span></div>
      {#if run.model}<div class="meta-item"><span class="meta-label">Model</span><span class="mono">{run.model}</span></div>{/if}
      {#if run.tokens}<div class="meta-item"><span class="meta-label">Tokens</span><span class="mono">{Number(run.tokens).toLocaleString()}</span></div>{/if}
    </div>

    {#if run.steps?.length}
      <section class="section"><h3>Steps</h3>
        <table><thead><tr><th>#</th><th>Step</th><th>Model</th><th>Duration</th><th>Status</th></tr></thead>
          <tbody>{#each run.steps as step, i}<tr>
            <td class="mono text-muted">{i + 1}</td><td class="mono">{step.step_id || step.name || ''}</td>
            <td class="mono text-muted">{step.model || ''}</td><td class="mono">{duration(step.started, step.finished)}</td>
            <td><StatusBadge status={step.status || 'pass'} /></td>
          </tr>{/each}</tbody>
        </table>
      </section>
    {/if}

    {#if run.output}<section class="section"><h3>Output</h3><div class="output-content">{@html renderMarkdown(run.output)}</div></section>{/if}

    {#if kibanaUrl}
      <section class="section">
        <button class="section-toggle" on:click={() => showTelemetry = !showTelemetry}>
          <h3>{@html icon(showTelemetry ? 'chevronDown' : 'chevronRight')} Telemetry</h3>
        </button>
        {#if showTelemetry}<iframe src={kibanaUrl} title="Kibana telemetry" class="kibana-frame"></iframe>{/if}
      </section>
    {/if}
  {/if}
</div>

<style>
  .meta-row { display: flex; gap: 32px; padding: 16px 0; border-bottom: 1px solid var(--border); margin-bottom: 24px; flex-wrap: wrap; }
  .meta-item { display: flex; flex-direction: column; gap: 4px; }
  .meta-label { font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .section { margin-bottom: 24px; }
  .section h3 { margin-bottom: 12px; display: flex; align-items: center; gap: 6px; }
  .section-toggle { background: none; border: none; color: var(--text-primary); cursor: pointer; padding: 0; }
  .output-content { line-height: 1.6; }
  .output-content :global(pre) { margin: 12px 0; }
  .kibana-frame { width: 100%; height: 400px; border: 1px solid var(--border); border-radius: 6px; background: var(--bg-deep); }
  .text-muted { color: var(--text-muted); }
</style>
