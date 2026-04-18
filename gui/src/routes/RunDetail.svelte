<script>
  import { onDestroy } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { getRun, runWorkflow } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';
  import PipelineGraph from '../lib/components/PipelineGraph.svelte';
  import ArtifactsBar from '../lib/components/ArtifactsBar.svelte';

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

  // Aggregate model/tokens/cost from steps when run-level is empty
  const aggModel = $derived.by(() => {
    if (run?.model) return run.model;
    const models = [...new Set(steps.filter(s => s.model).map(s => s.model))];
    return models.length === 1 ? models[0] : models.length > 1 ? models.join(', ') : '--';
  });

  const aggTokensIn = $derived(
    (run?.tokens_in || 0) || steps.reduce((sum, s) => sum + (s.tokens_in || 0), 0)
  );

  const aggTokensOut = $derived(
    (run?.tokens_out || 0) || steps.reduce((sum, s) => sum + (s.tokens_out || 0), 0)
  );

  const aggCost = $derived(
    steps.reduce((sum, s) => sum + (s.cost_usd || 0), 0)
  );

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
      <span class="meta-val mono">{aggModel}</span>
    </span>
    <span class="meta-pill">
      <span class="meta-label">Tokens</span>
      <span class="meta-val mono">{formatTokens(aggTokensIn, aggTokensOut)}</span>
    </span>
    <span class="meta-pill">
      <span class="meta-label">Cost</span>
      <span class="meta-val mono cost">{formatCost(run?.cost_usd || aggCost)}</span>
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
  <ArtifactsBar {steps} />
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
    border-bottom: 1px solid rgba(0, 229, 255, 0.08);
    flex-shrink: 0;
    transition: border-color 0.2s ease;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-actions .primary {
    border: 1px solid rgba(0, 229, 255, 0.3);
    background: linear-gradient(135deg, rgba(0, 229, 255, 0.1), rgba(0, 229, 255, 0.04));
    border-radius: 10px;
    color: var(--neon-cyan);
    padding: 6px 16px;
    font-size: 13px;
    cursor: pointer;
    transition: background 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .header-actions .primary:hover {
    background: linear-gradient(135deg, rgba(0, 229, 255, 0.18), rgba(0, 229, 255, 0.08));
    border-color: rgba(0, 229, 255, 0.5);
    box-shadow: 0 0 12px rgba(0, 229, 255, 0.1);
  }

  .poll-banner {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 10px 16px;
    background: rgba(255, 45, 111, 0.08);
    border-bottom: 1px solid rgba(255, 45, 111, 0.15);
    border-radius: 0 0 12px 12px;
    font-size: 12px;
    color: var(--neon-magenta);
    flex-shrink: 0;
    transition: background 0.2s ease;
  }
  .poll-banner button {
    font-size: 12px;
    padding: 4px 14px;
    border-radius: 8px;
    border: 1px solid rgba(255, 45, 111, 0.3);
    background: rgba(255, 45, 111, 0.1);
    color: var(--neon-magenta);
    cursor: pointer;
    transition: background 0.2s ease, border-color 0.2s ease;
  }
  .poll-banner button:hover {
    background: rgba(255, 45, 111, 0.18);
    border-color: rgba(255, 45, 111, 0.5);
  }

  .metadata-strip {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 24px;
    border-bottom: 1px solid rgba(0, 229, 255, 0.08);
    background: linear-gradient(145deg, rgba(17, 24, 32, 0.85), rgba(26, 34, 48, 0.6));
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    flex-shrink: 0;
    overflow-x: auto;
    border-radius: 0;
    transition: background 0.3s ease;
  }

  .meta-pill {
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    background: rgba(10, 14, 20, 0.5);
    border: 1px solid rgba(30, 42, 58, 0.6);
    border-radius: 10px;
    padding: 6px 14px;
    transition: border-color 0.2s ease, background 0.2s ease;
  }

  .meta-pill:hover {
    border-color: rgba(0, 229, 255, 0.15);
    background: rgba(10, 14, 20, 0.65);
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
