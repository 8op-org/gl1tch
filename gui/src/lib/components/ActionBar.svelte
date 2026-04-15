<script>
  import { onMount } from 'svelte';
  import { getWorkflowActions } from '../api.js';
  import { icon } from '../icons.js';

  let { context = '', resultPath = '', onrun } = $props();
  let actions = $state([]);

  onMount(async () => {
    try { actions = await getWorkflowActions(context) || []; } catch (_) { actions = []; }
  });
</script>

{#if actions.length > 0}
  <div class="action-bar">
    <span class="action-label text-muted">Actions:</span>
    {#each actions as wf}
      <button class="primary" on:click={() => onrun?.(wf)}>
        {@html icon('zap', 14)} {wf.name}
      </button>
    {/each}
  </div>
{/if}

<style>
  .action-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border);
    flex-wrap: wrap;
  }
  .action-label {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
</style>
