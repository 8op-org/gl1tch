<script>
  import { renderMarkdown } from '../markdown.js';

  let { step, runId, onclose } = $props();
  let activeSection = $state('metrics');

  const sections = ['metrics', 'output', 'prompt', 'files'];

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
</script>

<div class="node-panel">
  <div class="panel-header">
    <span class="panel-title mono">{step.step_id || '--'}</span>
    <button class="close-btn" onclick={onclose} type="button">&times;</button>
  </div>

  <div class="panel-tabs">
    {#each sections as section}
      <button
        class="panel-tab"
        class:active={activeSection === section}
        onclick={() => activeSection = section}
        type="button"
      >
        {section.charAt(0).toUpperCase() + section.slice(1)}
      </button>
    {/each}
  </div>

  <div class="panel-content">
    {#if activeSection === 'metrics'}
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
            <span style="color: #00ff9f;">0 (success)</span>
          {:else}
            <span style="color: #ff2d6f;">{step.exit_status}</span>
          {/if}
        </span>
      </div>

    {:else if activeSection === 'output'}
      {#if step.output}
        <div class="rendered-content">
          {@html renderMarkdown(step.output)}
        </div>
      {:else}
        <p class="text-muted placeholder">Step output not available in summary view</p>
      {/if}

    {:else if activeSection === 'prompt'}
      {#if step.prompt}
        <pre class="prompt-block"><code>{step.prompt}</code></pre>
      {:else}
        <p class="text-muted placeholder">Step prompt not available in summary view</p>
      {/if}

    {:else if activeSection === 'files'}
      <p class="text-muted placeholder">No result files linked</p>
    {/if}
  </div>
</div>

<style>
  .node-panel {
    width: 400px;
    border-left: 1px solid var(--border);
    background: var(--bg-surface);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    flex-shrink: 0;
    height: 100%;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
  }

  .panel-title {
    font-size: 13px;
    font-weight: 600;
  }

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
  }
  .panel-tab:hover {
    color: var(--text-primary);
  }
  .panel-tab.active {
    color: var(--neon-cyan);
    border-bottom-color: var(--neon-cyan);
  }

  .panel-content {
    padding: 16px;
    flex: 1;
    overflow-y: auto;
  }

  /* Metrics grid */
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
  .metric-value {
    font-size: 13px;
  }

  .mono {
    font-family: var(--font-mono, 'JetBrains Mono', monospace);
    font-size: 12px;
  }

  .text-muted {
    color: var(--text-muted);
  }

  .placeholder {
    font-size: 13px;
    font-style: italic;
    padding: 16px 0;
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
</style>
