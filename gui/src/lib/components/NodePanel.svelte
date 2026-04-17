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
