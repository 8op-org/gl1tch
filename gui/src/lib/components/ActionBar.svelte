<script>
  import { getWorkflowActions } from '../api.js';
  import { icon } from '../icons.js';

  let { context = '', resultPath = '', onrun } = $props();
  let actions = $state([]);

  async function fetchActions() {
    try { actions = await getWorkflowActions(context) || []; } catch (_) { actions = []; }
  }

  $effect(() => {
    context;
    fetchActions();
  });

  // Short display name for the folder path
  const folderLabel = $derived(resultPath ? resultPath.split('/').filter(Boolean).slice(-1)[0] || resultPath : '');
</script>

{#if actions.length > 0 && resultPath}
  <div class="action-bar">
    <span class="action-context">{@html icon('folder', 14)} {folderLabel}</span>
    {#each actions as wf}
      <button class="action-btn" onclick={() => onrun?.({ ...wf, name: wf.file, autoParams: { path: resultPath } })}>
        {@html icon('zap', 12)} {wf.description || wf.name}
      </button>
    {/each}
  </div>
{/if}

<style>
  .action-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: rgba(0, 229, 255, 0.04);
    border-bottom: 1px solid var(--border);
    flex-wrap: wrap;
  }
  .action-context {
    display: flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--neon-cyan);
    padding-right: 8px;
    border-right: 1px solid var(--border);
    margin-right: 4px;
  }
  .action-btn {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    padding: 4px 10px;
    border: 1px solid var(--border);
    border-radius: 4px;
    background: var(--bg-surface);
    color: var(--text-primary);
    cursor: pointer;
    transition: all 0.15s;
  }
  .action-btn:hover {
    border-color: var(--neon-cyan);
    color: var(--neon-cyan);
    background: rgba(0, 229, 255, 0.06);
  }
</style>
