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
    <button class="save-btn" disabled={!dirty || saving} onclick={handleSave}>
      {#if saving}
        <span class="save-spinner"></span> Saving
      {:else}
        {@html icon('save', 14)} Save changes
      {/if}
    </button>
  </div>
</div>

<div class="page-content settings-layout">
  {#if error && !workspace}
    <p class="status-fail">{error}</p>
  {:else if !workspace}
    <div class="loading-state">
      <div class="loading-shimmer"></div>
      <div class="loading-shimmer short"></div>
    </div>
  {:else}
    <div class="settings-grid loaded">
      <!-- Left column -->
      <div class="settings-col">
        <!-- Workflow Defaults card -->
        <section class="glass-card" style="animation-delay: 0ms">
          <div class="card-header">
            <div class="card-icon">{@html icon('zap', 18)}</div>
            <div>
              <h2>Workflow Defaults</h2>
              <p class="card-subtitle">Model, provider, and default parameters</p>
            </div>
          </div>
          <div class="card-body">
            <div class="field-group">
              <label class="field">
                <span class="field-label">Model</span>
                <div class="input-wrap">
                  <input type="text" bind:value={workspace.defaults.model} oninput={markDirty} placeholder="e.g. qwen2.5:7b" />
                </div>
              </label>

              <label class="field">
                <span class="field-label">Provider</span>
                <div class="input-wrap">
                  {#if providers.length > 0}
                    <select bind:value={workspace.defaults.provider} onchange={markDirty}>
                      <option value="">-- select --</option>
                      {#each providers as p}<option value={p}>{p}</option>{/each}
                    </select>
                  {:else}
                    <input type="text" bind:value={workspace.defaults.provider} oninput={markDirty} placeholder="e.g. ollama" />
                  {/if}
                </div>
              </label>
            </div>

            <div class="field">
              <span class="field-label">Parameters</span>
              <div class="kv-list">
                {#each Object.entries(workspace.defaults.params) as [key, val]}
                  <div class="kv-row">
                    <span class="kv-key">{key}</span>
                    <span class="kv-sep">=</span>
                    <input type="text" value={val} oninput={(e) => { workspace.defaults.params[key] = e.target.value; markDirty(); }} />
                    <button class="remove-btn" onclick={() => removeParam(key)} title="Remove">&times;</button>
                  </div>
                {/each}
                <div class="kv-row add-row">
                  <input type="text" bind:value={newParamKey} placeholder="key" />
                  <span class="kv-sep">=</span>
                  <input type="text" bind:value={newParamVal} placeholder="value" />
                  <button class="add-btn" onclick={addParam} disabled={!newParamKey.trim()}>+</button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Workspace card -->
        <section class="glass-card" style="animation-delay: 80ms">
          <div class="card-header">
            <div class="card-icon">{@html icon('folder', 18)}</div>
            <div>
              <h2>Workspace</h2>
              <p class="card-subtitle">Environment and repository configuration</p>
            </div>
          </div>
          <div class="card-body">
            <div class="field-group">
              <label class="field">
                <span class="field-label">Name</span>
                <div class="input-wrap">
                  <input type="text" bind:value={workspace.name} oninput={markDirty} placeholder="workspace name" />
                </div>
              </label>

              <label class="field">
                <span class="field-label">Elasticsearch URL</span>
                <div class="input-wrap" class:input-error={workspace.defaults.elasticsearch && !workspace.defaults.elasticsearch.startsWith('http://') && !workspace.defaults.elasticsearch.startsWith('https://')}>
                  <input type="text" bind:value={workspace.defaults.elasticsearch} oninput={markDirty} placeholder="http://localhost:9200" />
                </div>
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
                    <button class="remove-btn" onclick={() => removeRepo(i)} title="Remove">&times;</button>
                  </div>
                {/each}
                <div class="item-row add-row">
                  <input type="text" bind:value={newRepo} placeholder="owner/repo or URL" />
                  <button class="add-btn" onclick={addRepo} disabled={!newRepo.trim()}>+</button>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- Right column -->
      <div class="settings-col">
        <section class="glass-card full-height" style="animation-delay: 160ms">
          <ResourcesPanel />
        </section>
      </div>
    </div>
  {/if}
</div>

<style>
  .settings-layout {
    max-width: 1280px;
  }

  .settings-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    align-items: start;
  }

  .settings-grid > * {
    opacity: 0;
    transform: translateY(12px);
  }
  .settings-grid.loaded > * {
    animation: card-enter 0.4s ease-out forwards;
  }
  .settings-grid.loaded > :nth-child(2) {
    animation-delay: 80ms;
  }

  @keyframes card-enter {
    to { opacity: 1; transform: translateY(0); }
  }

  @media (max-width: 900px) {
    .settings-grid { grid-template-columns: 1fr; }
  }

  .settings-col {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  /* ── Glass card ──────────────────────────────────────── */
  .glass-card {
    background: linear-gradient(
      145deg,
      rgba(17, 24, 32, 0.85),
      rgba(26, 34, 48, 0.6)
    );
    border: 1px solid rgba(0, 229, 255, 0.08);
    border-radius: 16px;
    overflow: hidden;
    backdrop-filter: blur(12px);
    transition: border-color 0.3s, box-shadow 0.3s;
  }
  .glass-card:hover {
    border-color: rgba(0, 229, 255, 0.15);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3),
                0 0 0 1px rgba(0, 229, 255, 0.05);
  }
  .glass-card.full-height {
    min-height: 400px;
  }

  .card-header {
    padding: 20px 24px;
    display: flex;
    align-items: center;
    gap: 14px;
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

  .card-body {
    padding: 24px;
  }

  /* ── Fields ──────────────────────────────────────────── */
  .field-group {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
    margin-bottom: 24px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .field-label {
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--text-muted);
  }

  .input-wrap {
    position: relative;
    border-radius: 10px;
    background: var(--bg-deep);
    border: 1px solid rgba(30, 42, 58, 0.8);
    transition: border-color 0.2s, box-shadow 0.2s;
  }
  .input-wrap:focus-within {
    border-color: rgba(0, 229, 255, 0.4);
    box-shadow: 0 0 0 3px rgba(0, 229, 255, 0.06),
                0 2px 8px rgba(0, 0, 0, 0.2);
  }
  .input-wrap.input-error {
    border-color: rgba(255, 45, 111, 0.4);
  }
  .input-wrap.input-error:focus-within {
    box-shadow: 0 0 0 3px rgba(255, 45, 111, 0.06);
  }

  .input-wrap input, .input-wrap select {
    width: 100%;
    background: transparent;
    border: none;
    border-radius: 10px;
    padding: 10px 14px;
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 13px;
    outline: none;
  }
  .input-wrap select {
    appearance: none;
    cursor: pointer;
  }

  .field-hint {
    font-size: 11px;
    color: var(--text-muted);
  }
  .field-hint.error {
    color: var(--neon-magenta);
  }

  /* ── Key-value list ──────────────────────────────────── */
  .kv-list, .item-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: 6px;
  }

  .kv-row, .item-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 14px;
    background: rgba(10, 14, 20, 0.5);
    border: 1px solid rgba(30, 42, 58, 0.6);
    border-radius: 10px;
    transition: border-color 0.2s, background 0.2s;
  }
  .kv-row:hover, .item-row:hover {
    border-color: rgba(0, 229, 255, 0.15);
    background: rgba(10, 14, 20, 0.7);
  }
  .kv-row.add-row, .item-row.add-row {
    background: transparent;
    border-style: dashed;
    border-color: rgba(30, 42, 58, 0.4);
    opacity: 0.5;
    transition: opacity 0.2s, border-color 0.2s;
  }
  .kv-row.add-row:focus-within, .item-row.add-row:focus-within {
    opacity: 1;
    border-color: rgba(0, 229, 255, 0.3);
    border-style: solid;
  }

  .kv-key {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--neon-cyan);
    min-width: 80px;
    flex-shrink: 0;
    font-weight: 500;
  }

  .kv-sep {
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 12px;
    opacity: 0.4;
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

  .remove-btn {
    width: 26px;
    height: 26px;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: none;
    border-radius: 8px;
    color: var(--text-muted);
    font-size: 16px;
    cursor: pointer;
    flex-shrink: 0;
    transition: all 0.2s;
    opacity: 0;
  }
  .kv-row:hover .remove-btn,
  .item-row:hover .remove-btn {
    opacity: 1;
  }
  .remove-btn:hover {
    color: var(--neon-magenta);
    background: rgba(255, 45, 111, 0.1);
  }

  .add-btn {
    width: 28px;
    height: 28px;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 229, 255, 0.06);
    border: 1px solid rgba(0, 229, 255, 0.15);
    border-radius: 8px;
    color: var(--neon-cyan);
    font-size: 16px;
    cursor: pointer;
    flex-shrink: 0;
    transition: all 0.2s;
  }
  .add-btn:hover {
    background: rgba(0, 229, 255, 0.12);
    border-color: rgba(0, 229, 255, 0.3);
    box-shadow: 0 0 8px rgba(0, 229, 255, 0.15);
  }
  .add-btn:disabled {
    opacity: 0.2;
    cursor: not-allowed;
  }

  /* ── Save button ─────────────────────────────────────── */
  .save-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 20px;
    border-radius: 10px;
    border: 1px solid rgba(0, 229, 255, 0.3);
    background: linear-gradient(
      135deg,
      rgba(0, 229, 255, 0.1),
      rgba(0, 229, 255, 0.04)
    );
    color: var(--neon-cyan);
    font-family: var(--font-sans);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.25s;
  }
  .save-btn:hover:not(:disabled) {
    background: linear-gradient(
      135deg,
      rgba(0, 229, 255, 0.18),
      rgba(0, 229, 255, 0.08)
    );
    border-color: rgba(0, 229, 255, 0.5);
    box-shadow: 0 0 20px rgba(0, 229, 255, 0.12);
    transform: translateY(-1px);
  }
  .save-btn:active:not(:disabled) {
    transform: translateY(0);
  }
  .save-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }
  .save-btn :global(svg) {
    opacity: 0.8;
  }

  .save-spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(0, 229, 255, 0.2);
    border-top-color: var(--neon-cyan);
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* ── Save indicator ──────────────────────────────────── */
  .save-indicator {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 4px 12px;
    border-radius: 20px;
    animation: indicator-in 0.3s ease-out;
  }
  @keyframes indicator-in {
    from { opacity: 0; transform: scale(0.9); }
    to { opacity: 1; transform: scale(1); }
  }
  .save-indicator.saved {
    color: var(--neon-green);
    background: rgba(0, 255, 159, 0.08);
    border: 1px solid rgba(0, 255, 159, 0.15);
  }
  .save-indicator.error {
    color: var(--neon-magenta);
    background: rgba(255, 45, 111, 0.08);
    border: 1px solid rgba(255, 45, 111, 0.15);
  }

  /* ── Loading state ───────────────────────────────────── */
  .loading-state {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 20px 0;
  }
  .loading-shimmer {
    height: 200px;
    border-radius: 16px;
    background: linear-gradient(
      90deg,
      rgba(17, 24, 32, 0.6) 25%,
      rgba(26, 34, 48, 0.4) 50%,
      rgba(17, 24, 32, 0.6) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s ease-in-out infinite;
  }
  .loading-shimmer.short { height: 140px; }
  @keyframes shimmer {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }
</style>
