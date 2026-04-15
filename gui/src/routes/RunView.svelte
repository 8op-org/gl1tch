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
  let steps = $state([]);
  let kibanaUrl = $state(null);
  let error = $state(null);
  let showTelemetry = $state(false);

  function deriveStatus(r) {
    if (!r.finished_at) return 'RUNNING';
    return r.exit_status === 0 ? 'PASS' : 'FAIL';
  }

  function stepStatus(step) {
    if (step.exit_status != null) return step.exit_status === 0 ? 'PASS' : 'FAIL';
    if (step.gate_passed != null) return step.gate_passed ? 'PASS' : 'FAIL';
    return 'PASS';
  }

  const status = $derived(run ? deriveStatus(run) : null);
  const displayName = $derived(run ? (run.workflow_file || run.name || '').replace(/\.glitch$/, '') : '');
  const breadcrumbs = $derived([{ label: 'Runs', href: '#/runs' }, { label: `#${id}${displayName ? ' ' + displayName : ''}` }]);

  function duration(started, finished) {
    if (!started) return '--';
    const sec = Math.round(((finished || Date.now() / 1000) - started));
    if (sec < 0) return '--';
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  function stepDuration(ms) {
    if (!ms) return '--';
    if (ms < 1000) return `${ms}ms`;
    const sec = Math.round(ms / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  function formatDate(unixTs) {
    if (!unixTs) return '--';
    return new Date(unixTs * 1000).toLocaleString();
  }

  function formatTokens(n) {
    if (!n) return null;
    return Number(n).toLocaleString();
  }

  function formatCost(n) {
    if (!n) return null;
    return `$${n.toFixed(4)}`;
  }

  onMount(async () => {
    try {
      const data = await getRun(id);
      run = data.run || data;
      steps = data.steps || [];
      try { kibanaUrl = (await getKibanaRun(id)).url; } catch (_) {}
    } catch (e) { error = e.message; }
  });
</script>

<div class="page-header">
  <Breadcrumb segments={breadcrumbs} onnavigate={(href) => push(href.replace('#', ''))} />
  <div class="flex items-center gap-sm">{#if status}<StatusBadge status={status} size="md" />{/if}</div>
</div>

<div class="page-content">
  {#if error}<p class="status-fail">{error}</p>
  {:else if !run}<p class="text-muted">Loading...</p>
  {:else}
    <div class="meta-row">
      <div class="meta-item"><span class="meta-label">Started</span><span class="mono">{formatDate(run.started_at)}</span></div>
      <div class="meta-item"><span class="meta-label">Duration</span><span class="mono">{duration(run.started_at, run.finished_at)}</span></div>
      {#if run.model}<div class="meta-item"><span class="meta-label">Model</span><span class="mono">{run.model}</span></div>{/if}
      {#if run.tokens_in || run.tokens_out}
        <div class="meta-item"><span class="meta-label">Tokens</span><span class="mono">{formatTokens(run.tokens_in)} in / {formatTokens(run.tokens_out)} out</span></div>
      {/if}
      {#if run.cost_usd}<div class="meta-item"><span class="meta-label">Cost</span><span class="mono">{formatCost(run.cost_usd)}</span></div>{/if}
    </div>

    {#if steps.length > 0}
      <section class="section"><h3>Steps</h3>
        <table class="steps-table"><thead><tr><th>#</th><th>Step</th><th>Kind</th><th>Model</th><th>Duration</th><th>Tokens</th><th>Status</th></tr></thead>
          <tbody>{#each steps as step, i}<tr>
            <td class="mono text-muted">{i + 1}</td>
            <td class="mono step-name">{step.step_id || ''}</td>
            <td class="mono text-muted">{step.kind || ''}</td>
            <td class="mono text-muted">{step.model || ''}</td>
            <td class="mono">{stepDuration(step.duration_ms)}</td>
            <td class="mono text-muted">{#if step.tokens_in || step.tokens_out}{formatTokens(step.tokens_in)}/{formatTokens(step.tokens_out)}{:else}--{/if}</td>
            <td><StatusBadge status={stepStatus(step)} /></td>
          </tr>{/each}</tbody>
        </table>
      </section>
    {/if}

    {#if run.output}
      <section class="section"><h3>Output</h3><div class="output-content">{@html renderMarkdown(run.output)}</div></section>
    {/if}

    {#if kibanaUrl}
      <section class="section">
        <button class="section-toggle" onclick={() => showTelemetry = !showTelemetry}>
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
  .step-name { color: var(--neon-cyan); }
  .output-content { line-height: 1.6; }
  .output-content :global(pre) { margin: 12px 0; }
  .kibana-frame { width: 100%; height: 400px; border: 1px solid var(--border); border-radius: 6px; background: var(--bg-deep); }
  .text-muted { color: var(--text-muted); }
</style>
