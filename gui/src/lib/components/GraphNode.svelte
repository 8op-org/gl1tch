<script>
  let { step, selected = false, onclick } = $props();

  const status = $derived.by(() => {
    if (step.exit_status === undefined || step.exit_status === null) {
      if (step.gate_passed) return 'pass';
      return 'running';
    }
    return step.exit_status === 0 ? 'pass' : 'fail';
  });

  function formatDuration(ms) {
    if (ms == null) return '...';
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  function formatTokens(n) {
    if (n == null) return '--';
    if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
    return `${n}`;
  }

  const kindColors = {
    llm: 'kind-llm',
    run: 'kind-run',
    cond: 'kind-cond',
    map: 'kind-map',
  };
</script>

<button
  class="graph-node status-{status}"
  class:selected
  {onclick}
  type="button"
>
  <div class="node-top">
    <span class="node-id">{step.step_id || '--'}</span>
    {#if step.kind}
      <span class="kind-badge {kindColors[step.kind] || 'kind-default'}">{step.kind}</span>
    {/if}
  </div>
  <div class="node-bottom">
    <span class="node-duration">{formatDuration(step.duration_ms)}</span>
    {#if step.tokens_in != null || step.tokens_out != null}
      <span class="node-tokens">{formatTokens(step.tokens_in)} in / {formatTokens(step.tokens_out)} out</span>
    {/if}
  </div>
</button>

<style>
  .graph-node {
    width: 100%;
    height: 100%;
    border-radius: 8px;
    background: #1a2230;
    border: 1px solid var(--border);
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    cursor: pointer;
    transition: all 0.15s;
    box-sizing: border-box;
    text-align: left;
    color: var(--text-primary);
    font-family: inherit;
    font-size: inherit;
  }

  .graph-node:hover {
    background: #1e2a3a;
  }

  /* Status border colors */
  .status-pass { border-color: #00e5ff; }
  .status-fail { border-color: #ff2d6f; }
  .status-running { border-color: #ffb800; }

  /* Selected state */
  .graph-node.selected {
    border-width: 2px;
    padding: 9px 11px; /* compensate for thicker border */
  }
  .status-pass.selected { box-shadow: 0 0 12px rgba(0, 229, 255, 0.3); }
  .status-fail.selected { box-shadow: 0 0 12px rgba(255, 45, 111, 0.3); }
  .status-running.selected { box-shadow: 0 0 12px rgba(255, 184, 0, 0.3); }

  /* Running pulse animation */
  .status-running {
    animation: pulse-border 2s ease-in-out infinite;
  }
  @keyframes pulse-border {
    0%, 100% { border-color: #ffb800; }
    50% { border-color: rgba(255, 184, 0, 0.3); }
  }

  /* Top row */
  .node-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .node-id {
    font-family: var(--font-mono, 'JetBrains Mono', monospace);
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
    flex: 1;
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
  .kind-default { background: rgba(90, 106, 122, 0.2); color: #5a6a7a; }

  /* Bottom row */
  .node-bottom {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 4px;
  }
  .node-duration {
    font-family: var(--font-mono, 'JetBrains Mono', monospace);
    font-size: 11px;
    color: var(--text-muted, #5a6a7a);
  }
  .node-tokens {
    font-family: var(--font-mono, 'JetBrains Mono', monospace);
    font-size: 10px;
    color: var(--text-muted, #5a6a7a);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
