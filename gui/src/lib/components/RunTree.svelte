<script>
  import { getRunTree } from '../api.js';

  let { runId, initial = null } = $props();

  let node = $state(initial);
  let error = $state('');
  let expanded = $state(true);

  async function load() {
    if (node) return;
    try {
      node = await getRunTree(runId);
    } catch (e) {
      error = e.message;
    }
  }

  $effect(() => { if (runId && !node) load(); });
</script>

{#if error}
  <span class="error">{error}</span>
{:else if node}
  <div class="run-node">
    <div class="header">
      {#if node.children && node.children.length > 0}
        <button class="toggle" onclick={() => (expanded = !expanded)} aria-label={expanded ? 'Collapse' : 'Expand'}>
          {expanded ? '▾' : '▸'}
        </button>
      {:else}
        <span class="toggle-placeholder"></span>
      {/if}
      <a href={`#/run/${node.id}`} class="name">{node.name}</a>
      <span class="kind">{node.kind}</span>
      {#if node.exit_status !== null && node.exit_status !== undefined}
        <span class="status" class:ok={node.exit_status === 0} class:fail={node.exit_status !== 0}>
          exit {node.exit_status}
        </span>
      {:else}
        <span class="status running">running</span>
      {/if}
    </div>

    {#if expanded && node.children}
      <div class="children">
        {#each node.children as child (child.id)}
          <svelte:self runId={child.id} initial={child} />
        {/each}
      </div>
    {/if}
  </div>
{/if}

<style>
  .run-node { font-family: var(--font-mono, monospace); font-size: 13px; }
  .header { display: flex; gap: 8px; align-items: center; padding: 3px 0; }
  .toggle { background: none; border: none; cursor: pointer; padding: 0 4px; color: var(--text-muted, #888); font-size: 12px; width: 20px; }
  .toggle:hover { color: var(--text-primary, #fff); }
  .toggle-placeholder { width: 20px; display: inline-block; flex-shrink: 0; }
  .name { color: var(--neon-cyan, #4af); text-decoration: none; }
  .name:hover { text-decoration: underline; }
  .kind { color: var(--text-muted, #888); font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; }
  .status { font-size: 11px; padding: 1px 6px; border-radius: 3px; }
  .status.ok { color: var(--status-pass, #4c4); }
  .status.fail { color: var(--status-fail, #c44); }
  .status.running { color: var(--neon-amber, #fa0); }
  .children { margin-left: 20px; border-left: 1px solid var(--border, #333); padding-left: 8px; }
  .error { color: var(--status-fail, #c44); }
</style>
