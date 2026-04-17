<script>
  import { getWorkspace, updateWorkspace, getProviders } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import ResourcesPanel from '../lib/components/ResourcesPanel.svelte';

  let workspace = $state(null);
  let providers = $state([]);
  let saving = $state(false);
  let saveStatus = $state(null);
  let dirty = $state(false);
  let error = $state(null);

  let originalJson = $state('');
  let newParamKey = $state('');
  let newParamVal = $state('');
  let newRepo = $state('');

  $effect(() => { loadData(); });

  async function loadData() {
    try {
      const [ws, prov] = await Promise.all([getWorkspace(), getProviders()]);
      workspace = ws;
      providers = prov;
      originalJson = JSON.stringify(ws);
      dirty = false;
      error = null;
    } catch (e) { error = e.message; }
  }

  function markDirty() {
    if (workspace) dirty = JSON.stringify(workspace) !== originalJson;
  }

  async function handleSave() {
    saving = true; saveStatus = null;
    try {
      await updateWorkspace(workspace);
      originalJson = JSON.stringify(workspace);
      dirty = false;
      saveStatus = 'saved';
      setTimeout(() => saveStatus = null, 2000);
    } catch (e) { saveStatus = 'error'; error = e.message; }
    finally { saving = false; }
  }

  function addParam() {
    if (!newParamKey.trim()) return;
    workspace.defaults.params[newParamKey.trim()] = newParamVal;
    newParamKey = ''; newParamVal = '';
    markDirty();
  }

  function removeParam(key) {
    delete workspace.defaults.params[key];
    workspace.defaults = { ...workspace.defaults, params: { ...workspace.defaults.params } };
    markDirty();
  }

  function addRepo() {
    if (!newRepo.trim()) return;
    workspace.repos = [...workspace.repos, newRepo.trim()];
    newRepo = '';
    markDirty();
  }

  function removeRepo(index) {
    workspace.repos = workspace.repos.filter((_, i) => i !== index);
    markDirty();
  }
</script>

<div class="page-header">
  <h1>{@html icon('settings', 20)} Settings</h1>
  <div class="flex gap-sm items-center">
    {#if saveStatus === 'saved'}
      <span class="save-indicator saved">Saved</span>
    {:else if saveStatus === 'error'}
      <span class="save-indicator error">Error</span>
    {/if}
    <button class="primary" disabled={!dirty || saving} onclick={handleSave}>
      {#if saving}Saving...{:else}{@html icon('save', 14)} Save{/if}
    </button>
  </div>
</div>

<div class="page-content settings-layout">
  {#if error && !workspace}
    <p class="status-fail">{error}</p>
  {:else if !workspace}
    <p class="text-muted">Loading...</p>
  {:else}
    <div class="settings-grid">
      <!-- Left column -->
      <div class="settings-col">
        <!-- Workflow Defaults card -->
        <div class="settings-card">
          <div class="card-header">
            <h2>{@html icon('zap', 16)} Workflow Defaults</h2>
          </div>
          <div class="card-body">
            <div class="field-group">
              <label class="field">
                <span class="field-label">Model</span>
                <input type="text" bind:value={workspace.defaults.model} oninput={markDirty} placeholder="e.g. qwen2.5:7b" />
              </label>

              <label class="field">
                <span class="field-label">Provider</span>
                {#if providers.length > 0}
                  <select bind:value={workspace.defaults.provider} onchange={markDirty}>
                    <option value="">— select —</option>
                    {#each providers as p}<option value={p}>{p}</option>{/each}
                  </select>
                {:else}
                  <input type="text" bind:value={workspace.defaults.provider} oninput={markDirty} placeholder="e.g. ollama" />
                {/if}
              </label>
            </div>

            <div class="field">
              <span class="field-label">Parameters</span>
              <div class="kv-list">
                {#each Object.entries(workspace.defaults.params) as [key, val]}
                  <div class="kv-row">
                    <span class="kv-key">{key}</span>
                    <input type="text" value={val} oninput={(e) => { workspace.defaults.params[key] = e.target.value; markDirty(); }} />
                    <button class="icon-btn danger" onclick={() => removeParam(key)} title="Remove">&times;</button>
                  </div>
                {/each}
                <div class="kv-row add-row">
                  <input type="text" bind:value={newParamKey} placeholder="key" />
                  <input type="text" bind:value={newParamVal} placeholder="value" />
                  <button class="icon-btn" onclick={addParam} disabled={!newParamKey.trim()}>+</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Workspace card -->
        <div class="settings-card">
          <div class="card-header">
            <h2>{@html icon('folder', 16)} Workspace</h2>
          </div>
          <div class="card-body">
            <div class="field-group">
              <label class="field">
                <span class="field-label">Name</span>
                <input type="text" bind:value={workspace.name} oninput={markDirty} placeholder="workspace name" />
              </label>

              <label class="field">
                <span class="field-label">Elasticsearch URL</span>
                <input type="text" bind:value={workspace.defaults.elasticsearch} oninput={markDirty} placeholder="http://localhost:9200" />
                {#if workspace.defaults.elasticsearch && !workspace.defaults.elasticsearch.startsWith('http://') && !workspace.defaults.elasticsearch.startsWith('https://')}
                  <span class="field-hint error">URL should start with http:// or https://</span>
                {/if}
              </label>
            </div>

            <div class="field">
              <span class="field-label">Repositories</span>
              <div class="item-list">
                {#each workspace.repos as repo, i}
                  <div class="item-row">
                    <span class="item-value">{repo}</span>
                    <button class="icon-btn danger" onclick={() => removeRepo(i)} title="Remove">&times;</button>
                  </div>
                {/each}
                <div class="item-row add-row">
                  <input type="text" bind:value={newRepo} placeholder="owner/repo or URL" />
                  <button class="icon-btn" onclick={addRepo} disabled={!newRepo.trim()}>+</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right column -->
      <div class="settings-col">
        <div class="settings-card full-height">
          <ResourcesPanel />
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .settings-layout {
    max-width: 1200px;
  }

  .settings-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
    align-items: start;
  }

  @media (max-width: 900px) {
    .settings-grid { grid-template-columns: 1fr; }
  }

  .settings-col {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* Card */
  .settings-card {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 10px;
    overflow: hidden;
  }
  .settings-card.full-height {
    min-height: 400px;
  }

  .card-header {
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

  .card-body {
    padding: 20px;
  }

  /* Fields */
  .field-group {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    margin-bottom: 20px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-label {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .field input, .field select {
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 8px 12px;
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 13px;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .field input:focus, .field select:focus {
    outline: none;
    border-color: var(--neon-cyan);
    box-shadow: 0 0 0 2px rgba(0, 229, 255, 0.1);
  }

  .field-hint {
    font-size: 11px;
    color: var(--text-muted);
  }
  .field-hint.error {
    color: var(--neon-magenta);
  }

  /* Key-value list */
  .kv-list, .item-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 4px;
  }

  .kv-row, .item-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: var(--bg-deep);
    border: 1px solid var(--border);
    border-radius: 6px;
    transition: border-color 0.15s;
  }
  .kv-row:hover, .item-row:hover {
    border-color: var(--text-muted);
  }
  .kv-row.add-row, .item-row.add-row {
    background: transparent;
    border-style: dashed;
    border-color: var(--border);
    opacity: 0.6;
  }
  .kv-row.add-row:focus-within, .item-row.add-row:focus-within {
    opacity: 1;
    border-color: var(--neon-cyan);
  }

  .kv-key {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--neon-cyan);
    min-width: 100px;
    flex-shrink: 0;
  }

  .kv-row input, .item-row input {
    flex: 1;
    background: transparent;
    border: none;
    padding: 2px 4px;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-primary);
    outline: none;
  }

  .item-value {
    flex: 1;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-primary);
  }

  .icon-btn {
    width: 28px;
    height: 28px;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: 1px solid transparent;
    border-radius: 6px;
    color: var(--text-muted);
    font-size: 16px;
    cursor: pointer;
    flex-shrink: 0;
    transition: all 0.15s;
  }
  .icon-btn:hover {
    background: var(--bg-elevated);
    border-color: var(--border);
    color: var(--text-primary);
  }
  .icon-btn.danger:hover {
    color: var(--neon-magenta);
    border-color: rgba(255, 45, 111, 0.3);
    background: rgba(255, 45, 111, 0.06);
  }
  .icon-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Save indicator */
  .save-indicator {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 3px 10px;
    border-radius: 4px;
  }
  .save-indicator.saved {
    color: var(--neon-green);
    background: rgba(0, 255, 159, 0.08);
    border: 1px solid rgba(0, 255, 159, 0.2);
  }
  .save-indicator.error {
    color: var(--neon-magenta);
    background: rgba(255, 45, 111, 0.08);
    border: 1px solid rgba(255, 45, 111, 0.2);
  }
</style>
