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
    <h2>{@html icon('folder', 16)} Resources</h2>
    <button class="primary" onclick={() => showAdd = true}>+ Add</button>
  </div>

  {#if error}
    <p class="status-fail" style="font-size:12px; margin-bottom: 8px;">{error}</p>
  {/if}

  {#if loading && resources.length === 0}
    <p class="text-muted">Loading...</p>
  {:else if resources.length === 0}
    <p class="text-muted">No resources declared. Click &ldquo;Add resource&rdquo; to add one.</p>
  {:else}
    <table class="resources-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Type</th>
          <th>Ref</th>
          <th>Pin</th>
          <th>Status</th>
          <th>Fetched</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each resources as r}
          <tr class:busy={busyName === r.name}>
            <td class="mono">{r.name}</td>
            <td class="mono">{r.type}</td>
            <td class="mono">{r.ref ?? ''}</td>
            <td class="mono" title={r.pin ?? ''}>{shortPin(r.pin)}</td>
            <td class="status-cell">
              {#if busyName === r.name}
                <span class="status-dot busy-dot"></span><span class="text-muted">busy</span>
              {:else if r.fetched}
                <span class="status-dot synced"></span><span class="status-synced">synced</span>
              {:else}
                <span class="status-dot stale"></span><span class="status-stale">stale</span>
              {/if}
            </td>
            <td class="text-muted">{formatFetched(r.fetched)}</td>
            <td class="actions">
              <button
                disabled={busyName === r.name}
                onclick={() => onSync(r.name)}
                title="Re-fetch resource"
              >Sync</button>
              <button
                disabled={busyName === r.name}
                onclick={() => onPin(r.name)}
                title="Pin to a specific ref/sha"
              >Pin</button>
              <button
                class="danger"
                disabled={busyName === r.name}
                onclick={() => onRemove(r.name)}
                title="Remove resource"
              >Remove</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</section>

{#if showAdd}
  <Modal title="Add resource" onclose={closeAdd}>
    <form onsubmit={(e) => { e.preventDefault(); onAdd(); }} class="flex flex-col gap-md">
      <label class="field">
        <span class="field-label">URL or path</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          type="text"
          bind:value={addInput}
          placeholder="https://github.com/org/repo or /abs/path or org/repo"
          autofocus
        />
      </label>
      <label class="field">
        <span class="field-label">Name (optional)</span>
        <input
          type="text"
          bind:value={addName}
          placeholder="inferred from input"
        />
      </label>
      <label class="field">
        <span class="field-label">Pin / ref (optional)</span>
        <input
          type="text"
          bind:value={addPin}
          placeholder="main, tag, sha"
        />
      </label>
      {#if error}
        <p class="status-fail" style="font-size:12px">{error}</p>
      {/if}
      <div class="flex justify-between" style="margin-top:8px">
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
    justify-content: space-between;
    align-items: center;
    padding: 14px 20px;
    border-bottom: 1px solid var(--border);
    background: rgba(0, 229, 255, 0.02);
  }
  .card-header h2 {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
  }
  .card-header h2 :global(svg) {
    color: var(--neon-cyan);
  }
  .card-header .primary {
    font-size: 12px;
    padding: 4px 12px;
  }

  .resources-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }
  .resources-table th {
    text-align: left;
    padding: 8px 10px;
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    border-bottom: 1px solid var(--border);
  }
  .resources-table td {
    padding: 8px 10px;
    border-bottom: 1px solid var(--border);
    vertical-align: middle;
  }
  .resources-table tr.busy { opacity: 0.5; }
  .status-cell {
    white-space: nowrap;
    font-size: 11px;
  }
  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
    margin-right: 6px;
    vertical-align: middle;
  }
  .status-dot.synced {
    background: var(--neon-green);
    box-shadow: 0 0 4px rgba(0, 255, 159, 0.4);
  }
  .status-dot.stale {
    background: var(--neon-amber);
    box-shadow: 0 0 4px rgba(255, 184, 0, 0.4);
  }
  .status-dot.busy-dot {
    background: var(--neon-cyan);
    box-shadow: 0 0 4px rgba(0, 255, 255, 0.4);
    animation: pulse 1s ease-in-out infinite;
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  .status-synced { color: var(--neon-green); }
  .status-stale { color: var(--neon-amber); }
  .resources-table .actions {
    display: flex;
    gap: 6px;
    justify-content: flex-end;
  }
  .resources-table .actions button {
    padding: 3px 10px;
    font-size: 12px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .field-label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }
  .field input {
    background: var(--bg-deep);
    color: var(--text-primary);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 6px 10px;
    font-family: var(--font-mono);
    font-size: 12px;
  }
  .field input:focus {
    outline: none;
    border-color: var(--neon-cyan);
  }
</style>
