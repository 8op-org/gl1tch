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

  // Snapshot of the original workspace JSON for dirty detection
  let originalJson = $state('');

  // Default param editing
  let newParamKey = $state('');
  let newParamVal = $state('');

  // New repo input
  let newRepo = $state('');

  $effect(() => {
    loadData();
  });

  async function loadData() {
    try {
      const [ws, prov] = await Promise.all([getWorkspace(), getProviders()]);
      workspace = ws;
      providers = prov;
      originalJson = JSON.stringify(ws);
      dirty = false;
      error = null;
    } catch (e) {
      error = e.message;
    }
  }

  function markDirty() {
    if (workspace) {
      dirty = JSON.stringify(workspace) !== originalJson;
    }
  }

  async function handleSave() {
    saving = true;
    saveStatus = null;
    try {
      await updateWorkspace(workspace);
      originalJson = JSON.stringify(workspace);
      dirty = false;
      saveStatus = 'saved';
      setTimeout(() => saveStatus = null, 2000);
    } catch (e) {
      saveStatus = 'error';
      error = e.message;
    } finally {
      saving = false;
    }
  }

  function addParam() {
    if (!newParamKey.trim()) return;
    workspace.defaults.params[newParamKey.trim()] = newParamVal;
    newParamKey = '';
    newParamVal = '';
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
      <span class="status-pass" style="font-size:12px">Saved</span>
    {/if}
    {#if saveStatus === 'error'}
      <span class="status-fail" style="font-size:12px">Error saving</span>
    {/if}
    <button class="primary" disabled={!dirty || saving} onclick={handleSave}>
      {#if saving}Saving...{:else}{@html icon('save', 14)} Save{/if}
    </button>
  </div>
</div>

<div class="page-content settings-content">
  {#if error && !workspace}
    <p class="status-fail">{error}</p>
  {:else if !workspace}
    <p class="text-muted">Loading...</p>
  {:else}
    <!-- Workflow Defaults -->
    <section class="settings-section">
      <h2>Workflow Defaults</h2>

      <label class="settings-field">
        <span class="settings-label">Default Model</span>
        <input
          type="text"
          bind:value={workspace.defaults.model}
          oninput={markDirty}
          placeholder="e.g. qwen2.5:7b"
        />
      </label>

      <label class="settings-field">
        <span class="settings-label">Default Provider</span>
        {#if providers.length > 0}
          <select bind:value={workspace.defaults.provider} onchange={markDirty}>
            <option value="">— select —</option>
            {#each providers as p}
              <option value={p}>{p}</option>
            {/each}
          </select>
        {:else}
          <input
            type="text"
            bind:value={workspace.defaults.provider}
            oninput={markDirty}
            placeholder="e.g. ollama"
          />
        {/if}
      </label>

      <div class="settings-field">
        <span class="settings-label">Default Parameters</span>
        <div class="params-list">
          {#each Object.entries(workspace.defaults.params) as [key, val]}
            <div class="param-row">
              <span class="param-key">{key}</span>
              <input
                type="text"
                value={val}
                oninput={(e) => { workspace.defaults.params[key] = e.target.value; markDirty(); }}
              />
              <button class="danger" onclick={() => removeParam(key)} title="Remove">&times;</button>
            </div>
          {/each}
          <div class="param-row add-row">
            <input type="text" bind:value={newParamKey} placeholder="key" />
            <input type="text" bind:value={newParamVal} placeholder="value" />
            <button onclick={addParam} disabled={!newParamKey.trim()}>+</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Workspace Config -->
    <section class="settings-section">
      <h2>Workspace</h2>

      <label class="settings-field">
        <span class="settings-label">Workspace Name</span>
        <input
          type="text"
          bind:value={workspace.name}
          oninput={markDirty}
          placeholder="workspace name"
        />
      </label>

      <label class="settings-field">
        <span class="settings-label">Kibana URL</span>
        <input
          type="text"
          bind:value={workspace.defaults.elasticsearch}
          oninput={markDirty}
          placeholder="http://localhost:5601"
        />
        {#if workspace.defaults.elasticsearch && !workspace.defaults.elasticsearch.startsWith('http://') && !workspace.defaults.elasticsearch.startsWith('https://')}
          <span class="status-fail url-hint" style="font-size:11px">URL should start with http:// or https://</span>
        {/if}
      </label>

      <div class="settings-field">
        <span class="settings-label">Repositories</span>
        <div class="repos-list">
          {#each workspace.repos as repo, i}
            <div class="repo-row">
              <span class="mono">{repo}</span>
              <button class="danger" onclick={() => removeRepo(i)} title="Remove">&times;</button>
            </div>
          {/each}
          <div class="repo-row add-row">
            <input type="text" bind:value={newRepo} placeholder="owner/repo" />
            <button onclick={addRepo} disabled={!newRepo.trim()}>+</button>
          </div>
        </div>
      </div>
    </section>

    <!-- Workspace Resources -->
    <section class="settings-section">
      <ResourcesPanel />
    </section>
  {/if}
</div>

<style>
  .settings-content {
    max-width: 640px;
  }

  .settings-section {
    margin-bottom: 32px;
  }
  .settings-section h2 {
    color: var(--neon-cyan);
    margin-bottom: 16px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
  }

  .settings-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 16px;
  }
  .settings-label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .params-list, .repos-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .param-row, .repo-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .param-row input, .repo-row input {
    flex: 1;
  }
  .param-key {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--neon-cyan);
    min-width: 120px;
  }
  .param-row button, .repo-row button {
    padding: 4px 10px;
    font-size: 14px;
  }
  .add-row {
    opacity: 0.7;
  }
  .add-row:focus-within {
    opacity: 1;
  }
</style>
