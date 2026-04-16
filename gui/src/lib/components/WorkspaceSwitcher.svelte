<script>
  import { listWorkspaces, useWorkspace } from '../api.js';

  let workspaces = $state([]);
  let active = $state('');
  let error = $state(null);

  async function load() {
    try {
      workspaces = await listWorkspaces();
      const a = (workspaces || []).find(w => w.active);
      active = a ? a.name : '';
      error = null;
    } catch (e) {
      error = e.message;
    }
  }

  async function onChange(ev) {
    const next = ev.target.value;
    if (!next || next === active) return;
    try {
      await useWorkspace(next);
      location.reload();
    } catch (e) {
      error = e.message;
      ev.target.value = active;
    }
  }

  $effect(() => {
    load();
  });
</script>

<div class="workspace-switcher">
  <span class="ws-label">workspace</span>
  {#if error}
    <span class="ws-error" title={error}>error</span>
  {:else if !workspaces || workspaces.length === 0}
    <span class="ws-empty">none registered</span>
  {:else}
    <select value={active} onchange={onChange}>
      {#each workspaces as w}
        <option value={w.name}>{w.name}</option>
      {/each}
    </select>
  {/if}
</div>

<style>
  .workspace-switcher {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    height: 44px;
    border-bottom: 1px solid var(--border);
    background: var(--bg-surface);
  }
  .ws-label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }
  .workspace-switcher select {
    background: var(--bg-deep);
    color: var(--text-primary);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 4px 8px;
    font-family: var(--font-mono);
    font-size: 12px;
    min-width: 160px;
  }
  .workspace-switcher select:focus {
    outline: none;
    border-color: var(--neon-cyan);
    box-shadow: var(--glow-cyan);
  }
  .ws-empty, .ws-error {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-muted);
  }
  .ws-error { color: var(--neon-magenta); }
</style>
