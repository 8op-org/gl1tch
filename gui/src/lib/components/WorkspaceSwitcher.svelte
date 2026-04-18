<script>
  import { listWorkspaces, useWorkspace } from '../api.js';

  let workspaces = $state([]);
  let active = $state('');
  let error = $state(null);
  let open = $state(false);

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

  async function switchTo(name) {
    if (!name || name === active) { open = false; return; }
    try {
      await useWorkspace(name);
      location.reload();
    } catch (e) {
      error = e.message;
    }
    open = false;
  }

  function initial(name) {
    return name ? name.charAt(0).toUpperCase() : '?';
  }

  $effect(() => { load(); });
</script>

<div class="ws-switcher">
  <button
    class="ws-avatar"
    title={active || 'No workspace'}
    onclick={() => open = !open}
  >
    {initial(active)}
  </button>
  {#if open}
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="ws-backdrop" onclick={() => open = false}></div>
    <div class="ws-dropdown">
      <div class="ws-dropdown-header">Workspaces</div>
      {#each workspaces as w}
        <button
          class="ws-option"
          class:active={w.name === active}
          onclick={() => switchTo(w.name)}
        >
          <span class="ws-option-avatar">{initial(w.name)}</span>
          <span class="ws-option-name">{w.name}</span>
          {#if w.name === active}
            <span class="ws-check">&#10003;</span>
          {/if}
        </button>
      {/each}
      {#if !workspaces || workspaces.length === 0}
        <div class="ws-empty">No workspaces registered</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .ws-switcher {
    position: relative;
    display: flex;
    justify-content: center;
    padding: 10px 0;
  }

  .ws-avatar {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    background: linear-gradient(135deg, rgba(0, 229, 255, 0.15), rgba(255, 45, 111, 0.1));
    border: 1px solid var(--border);
    color: var(--neon-cyan);
    font-family: var(--font-mono);
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
    padding: 0;
  }
  .ws-avatar:hover {
    border-color: var(--neon-cyan);
    box-shadow: 0 0 8px rgba(0, 229, 255, 0.3);
    background: linear-gradient(135deg, rgba(0, 229, 255, 0.25), rgba(255, 45, 111, 0.15));
  }

  .ws-backdrop {
    position: fixed;
    inset: 0;
    z-index: 99;
  }

  .ws-dropdown {
    position: absolute;
    left: 52px;
    top: 4px;
    min-width: 200px;
    background: linear-gradient(145deg, rgba(17,24,32,0.95), rgba(26,34,48,0.85));
    border: 1px solid rgba(0,229,255,0.1);
    border-radius: 12px;
    box-shadow: 0 16px 64px rgba(0,0,0,0.5);
    backdrop-filter: blur(16px);
    z-index: 100;
    overflow: hidden;
  }

  .ws-dropdown-header {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-muted);
    padding: 10px 14px 6px;
  }

  .ws-option {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 8px 14px;
    background: none;
    border: none;
    border-radius: 0;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 13px;
    text-align: left;
    transition: background 0.1s;
  }
  .ws-option:hover {
    background: rgba(0, 229, 255, 0.04);
  }
  .ws-option.active {
    color: var(--neon-cyan);
  }

  .ws-option-avatar {
    width: 24px;
    height: 24px;
    border-radius: 6px;
    background: rgba(0, 229, 255, 0.1);
    border: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    color: var(--neon-cyan);
    flex-shrink: 0;
  }

  .ws-option-name {
    font-family: var(--font-mono);
    font-size: 13px;
    flex: 1;
  }

  .ws-check {
    color: var(--neon-green);
    font-size: 12px;
  }

  .ws-empty {
    padding: 12px 14px;
    font-size: 12px;
    color: var(--text-muted);
  }
</style>
