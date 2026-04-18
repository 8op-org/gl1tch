<script>
  import { listResources, addResource, removeResource, syncWorkspace, pinResource } from '../api.js';
  import { icon } from '../icons.js';
  import Modal from './Modal.svelte';

  let resources = $state([]);
  let error = $state(null);
  let loading = $state(false);
  let showAdd = $state(false);
  let busyName = $state('');

  // Add form state
  let addInput = $state('');
  let addName = $state('');
  let addPin = $state('');
  let adding = $state(false);

  async function load() {
    loading = true;
    error = null;
    try {
      const data = await listResources();
      resources = Array.isArray(data) ? data : [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function onAdd() {
    if (!addInput.trim()) return;
    adding = true;
    error = null;
    try {
      await addResource({
        input: addInput.trim(),
        name: addName.trim(),
        pin: addPin.trim(),
      });
      addInput = '';
      addName = '';
      addPin = '';
      showAdd = false;
      await load();
    } catch (e) {
      error = e.message;
    } finally {
      adding = false;
    }
  }

  async function onSync(name) {
    busyName = name;
    error = null;
    try {
      await syncWorkspace(name);
      await load();
    } catch (e) {
      error = e.message;
    } finally {
      busyName = '';
    }
  }

  async function onRemove(name) {
    if (!confirm(`Remove resource "${name}"?`)) return;
    busyName = name;
    error = null;
    try {
      await removeResource(name);
      await load();
    } catch (e) {
      error = e.message;
    } finally {
      busyName = '';
    }
  }

  async function onPin(name) {
    const ref = prompt(`Pin "${name}" to ref, branch, or sha:`);
    if (!ref) return;
    busyName = name;
    error = null;
    try {
      await pinResource(name, ref.trim());
      await load();
    } catch (e) {
      error = e.message;
    } finally {
      busyName = '';
    }
  }

  function closeAdd() {
    showAdd = false;
    addInput = '';
    addName = '';
    addPin = '';
  }

  function shortPin(p) {
    if (!p) return '';
    return p.length > 8 ? p.slice(0, 8) : p;
  }

  function formatFetched(t) {
    if (!t) return '';
    try {
      const d = new Date(t);
      if (isNaN(d.getTime())) return t;
      return d.toLocaleString();
    } catch {
      return t;
    }
  }

  $effect(() => {
    load();
  });
</script>

<section class="resources-panel">
  <div class="card-header">
    <div class="card-icon">{@html icon('folder', 18)}</div>
    <div>
      <h2>Resources</h2>
      <p class="card-subtitle">Synced repositories and data sources</p>
    </div>
    <button class="add-resource-btn" onclick={() => showAdd = true}>
      <span class="add-resource-icon">+</span> Add
    </button>
  </div>

  <div class="resources-body">
    {#if error}
      <p class="status-fail" style="font-size:12px; padding: 12px 20px;">{error}</p>
    {/if}

    {#if loading && resources.length === 0}
      <div class="empty-state">
        <div class="loading-dots">
          <span></span><span></span><span></span>
        </div>
      </div>
    {:else if resources.length === 0}
      <div class="empty-state">
        <div class="empty-icon">{@html icon('folder', 32)}</div>
        <p>No resources yet</p>
        <span class="text-muted">Add a repository or data source to get started</span>
      </div>
    {:else}
      <div class="resource-list">
        {#each resources as r, i}
          <div class="resource-card" class:busy={busyName === r.name} style="animation-delay: {i * 40}ms">
            <div class="resource-main">
              <div class="resource-info">
                <span class="resource-name">{r.name}</span>
                <div class="resource-meta">
                  <span class="resource-type">{r.type}</span>
                  {#if r.ref}
                    <span class="resource-ref">{r.ref}</span>
                  {/if}
                  {#if r.pin}
                    <span class="resource-pin" title={r.pin}>{shortPin(r.pin)}</span>
                  {/if}
                </div>
              </div>
              <div class="resource-status">
                {#if busyName === r.name}
                  <span class="status-badge busy">
                    <span class="status-dot busy-dot"></span> syncing
                  </span>
                {:else if r.fetched}
                  <span class="status-badge synced">
                    <span class="status-dot synced"></span> synced
                  </span>
                {:else}
                  <span class="status-badge stale">
                    <span class="status-dot stale"></span> stale
                  </span>
                {/if}
                {#if r.fetched && busyName !== r.name}
                  <span class="resource-fetched">{formatFetched(r.fetched)}</span>
                {/if}
              </div>
            </div>
            <div class="resource-actions">
              <button
                class="action-btn"
                disabled={busyName === r.name}
                onclick={() => onSync(r.name)}
                title="Re-fetch resource"
              >Sync</button>
              <button
                class="action-btn"
                disabled={busyName === r.name}
                onclick={() => onPin(r.name)}
                title="Pin to a specific ref/sha"
              >Pin</button>
              <button
                class="action-btn danger"
                disabled={busyName === r.name}
                onclick={() => onRemove(r.name)}
                title="Remove resource"
              >Remove</button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>

{#if showAdd}
  <Modal title="Add resource" onclose={closeAdd}>
    <form onsubmit={(e) => { e.preventDefault(); onAdd(); }} class="flex flex-col gap-md">
      <label class="field">
        <span class="field-label">URL or path</span>
        <div class="input-wrap">
          <!-- svelte-ignore a11y_autofocus -->
          <input
            type="text"
            bind:value={addInput}
            placeholder="https://github.com/org/repo or /abs/path or org/repo"
            autofocus
          />
        </div>
      </label>
      <label class="field">
        <span class="field-label">Name (optional)</span>
        <div class="input-wrap">
          <input
            type="text"
            bind:value={addName}
            placeholder="inferred from input"
          />
        </div>
      </label>
      <label class="field">
        <span class="field-label">Pin / ref (optional)</span>
        <div class="input-wrap">
          <input
            type="text"
            bind:value={addPin}
            placeholder="main, tag, sha"
          />
        </div>
      </label>
      {#if error}
        <p class="status-fail" style="font-size:12px">{error}</p>
      {/if}
      <div class="flex justify-between" style="margin-top:12px">
        <button type="button" onclick={closeAdd}>Cancel</button>
        <button type="submit" class="primary" disabled={adding || !addInput.trim()}>
          {#if adding}Adding...{:else}Add{/if}
        </button>
      </div>
    </form>
  </Modal>
{/if}

<style>
  .resources-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 20px 24px;
    border-bottom: 1px solid rgba(0, 229, 255, 0.06);
    background: linear-gradient(
      180deg,
      rgba(0, 229, 255, 0.03) 0%,
      transparent 100%
    );
  }
  .card-icon {
    width: 38px;
    height: 38px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    background: rgba(0, 229, 255, 0.08);
    border: 1px solid rgba(0, 229, 255, 0.12);
    flex-shrink: 0;
  }
  .card-icon :global(svg) {
    color: var(--neon-cyan);
  }
  .card-header h2 {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
    letter-spacing: -0.01em;
  }
  .card-subtitle {
    font-size: 12px;
    color: var(--text-muted);
    margin: 2px 0 0;
    font-weight: 400;
  }

  .add-resource-btn {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border-radius: 8px;
    border: 1px solid rgba(0, 229, 255, 0.2);
    background: rgba(0, 229, 255, 0.06);
    color: var(--neon-cyan);
    font-family: var(--font-sans);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  .add-resource-btn:hover {
    background: rgba(0, 229, 255, 0.12);
    border-color: rgba(0, 229, 255, 0.35);
    box-shadow: 0 0 12px rgba(0, 229, 255, 0.1);
  }
  .add-resource-icon {
    font-size: 15px;
    font-weight: 400;
    line-height: 1;
  }

  .resources-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
  }

  /* ── Empty state ─────────────────────────────────────── */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 20px;
    text-align: center;
    gap: 8px;
  }
  .empty-icon {
    opacity: 0.15;
    margin-bottom: 8px;
  }
  .empty-icon :global(svg) {
    color: var(--text-primary);
  }
  .empty-state p {
    font-size: 14px;
    color: var(--text-primary);
    margin: 0;
  }
  .empty-state span {
    font-size: 12px;
  }

  .loading-dots {
    display: flex;
    gap: 6px;
  }
  .loading-dots span {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--neon-cyan);
    animation: dot-pulse 1.2s ease-in-out infinite;
  }
  .loading-dots span:nth-child(2) { animation-delay: 0.2s; }
  .loading-dots span:nth-child(3) { animation-delay: 0.4s; }
  @keyframes dot-pulse {
    0%, 80%, 100% { opacity: 0.2; transform: scale(0.8); }
    40% { opacity: 1; transform: scale(1); }
  }

  /* ── Resource cards ──────────────────────────────────── */
  .resource-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .resource-card {
    padding: 14px 16px;
    background: rgba(10, 14, 20, 0.4);
    border: 1px solid rgba(30, 42, 58, 0.5);
    border-radius: 12px;
    transition: border-color 0.2s, background 0.2s, box-shadow 0.2s;
    animation: resource-in 0.3s ease-out both;
  }
  @keyframes resource-in {
    from { opacity: 0; transform: translateX(-8px); }
    to { opacity: 1; transform: translateX(0); }
  }
  .resource-card:hover {
    border-color: rgba(0, 229, 255, 0.12);
    background: rgba(10, 14, 20, 0.6);
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  }
  .resource-card.busy {
    opacity: 0.6;
  }

  .resource-main {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .resource-info {
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-width: 0;
  }

  .resource-name {
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .resource-meta {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .resource-type,
  .resource-ref,
  .resource-pin {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 2px 8px;
    border-radius: 4px;
    background: rgba(26, 34, 48, 0.6);
    color: var(--text-muted);
    border: 1px solid rgba(30, 42, 58, 0.5);
  }

  .resource-status {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 4px;
    flex-shrink: 0;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 3px 10px;
    border-radius: 20px;
  }
  .status-badge.synced {
    color: var(--neon-green);
    background: rgba(0, 255, 159, 0.06);
  }
  .status-badge.stale {
    color: var(--neon-amber);
    background: rgba(255, 184, 0, 0.06);
  }
  .status-badge.busy {
    color: var(--neon-cyan);
    background: rgba(0, 229, 255, 0.06);
  }

  .status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    display: inline-block;
    flex-shrink: 0;
  }
  .status-dot.synced {
    background: var(--neon-green);
    box-shadow: 0 0 6px rgba(0, 255, 159, 0.5);
  }
  .status-dot.stale {
    background: var(--neon-amber);
    box-shadow: 0 0 6px rgba(255, 184, 0, 0.5);
  }
  .status-dot.busy-dot {
    background: var(--neon-cyan);
    box-shadow: 0 0 6px rgba(0, 255, 255, 0.5);
    animation: pulse 1s ease-in-out infinite;
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }

  .resource-fetched {
    font-size: 10px;
    color: var(--text-muted);
    font-family: var(--font-mono);
  }

  .resource-actions {
    display: flex;
    gap: 6px;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid rgba(30, 42, 58, 0.3);
    opacity: 0;
    transition: opacity 0.2s;
  }
  .resource-card:hover .resource-actions {
    opacity: 1;
  }

  .action-btn {
    padding: 4px 12px;
    font-size: 11px;
    font-family: var(--font-mono);
    border-radius: 6px;
    border: 1px solid rgba(30, 42, 58, 0.6);
    background: rgba(26, 34, 48, 0.4);
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.15s;
  }
  .action-btn:hover {
    color: var(--text-primary);
    border-color: rgba(0, 229, 255, 0.2);
    background: rgba(0, 229, 255, 0.05);
  }
  .action-btn.danger:hover {
    color: var(--neon-magenta);
    border-color: rgba(255, 45, 111, 0.25);
    background: rgba(255, 45, 111, 0.06);
  }
  .action-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* ── Modal fields ────────────────────────────────────── */
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .field-label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }
  .input-wrap {
    border-radius: 10px;
    background: var(--bg-deep);
    border: 1px solid rgba(30, 42, 58, 0.8);
    transition: border-color 0.2s, box-shadow 0.2s;
  }
  .input-wrap:focus-within {
    border-color: rgba(0, 229, 255, 0.4);
    box-shadow: 0 0 0 3px rgba(0, 229, 255, 0.06);
  }
  .input-wrap input {
    width: 100%;
    background: transparent;
    color: var(--text-primary);
    border: none;
    border-radius: 10px;
    padding: 10px 14px;
    font-family: var(--font-mono);
    font-size: 12px;
    outline: none;
  }
</style>
