<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { listWorkflows } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import FilterBar from '../lib/components/FilterBar.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';

  let workflows = $state([]);
  let error = $state(null);
  let loading = $state(true);
  let searchQuery = $state('');
  let activeTags = $state([]);
  let sortBy = $state('name');
  let viewMode = $state('grouped'); // 'grouped' | 'grid' | 'list'
  let collapsedGroups = $state({});

  const allTags = $derived([...new Set(workflows.flatMap(w => w.tags || []))].sort());

  // Derive group from tags or name prefix
  function getGroup(wf) {
    if (wf.tags?.length) return wf.tags[0];
    const name = wf.name;
    const dash = name.indexOf('-');
    if (dash > 0) return name.substring(0, dash);
    return 'other';
  }

  const filtered = $derived.by(() => {
    let list = workflows;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      list = list.filter(w => w.name.toLowerCase().includes(q) || (w.description || '').toLowerCase().includes(q));
    }
    if (activeTags.length > 0) {
      list = list.filter(w => activeTags.some(t => (w.tags || []).includes(t) || getGroup(w) === t));
    }
    list = [...list].sort((a, b) => a.name.localeCompare(b.name));
    return list;
  });

  const grouped = $derived.by(() => {
    const groups = {};
    for (const wf of filtered) {
      const g = getGroup(wf);
      if (!groups[g]) groups[g] = [];
      groups[g].push(wf);
    }
    return Object.entries(groups).sort(([a], [b]) => a.localeCompare(b));
  });

  const allGroups = $derived([...new Set(workflows.map(getGroup))].sort());

  function toggleGroup(name) {
    collapsedGroups = { ...collapsedGroups, [name]: !collapsedGroups[name] };
  }

  onMount(async () => {
    try { workflows = await listWorkflows(); } catch (e) { error = e.message; } finally { loading = false; }
  });
</script>

<div class="page-header">
  <h1>Workflows</h1>
  <div class="flex items-center gap-sm">
    <button class="view-btn" class:active={viewMode === 'grouped'} on:click={() => viewMode = 'grouped'} title="Grouped">{@html icon('folder', 16)}</button>
    <button class="view-btn" class:active={viewMode === 'grid'} on:click={() => viewMode = 'grid'} title="Grid">{@html icon('workflow', 16)}</button>
    <button class="view-btn" class:active={viewMode === 'list'} on:click={() => viewMode = 'list'} title="List">{@html icon('file', 16)}</button>
  </div>
</div>
<div class="page-content">
  {#if loading}
    <p class="text-muted">Loading workflows...</p>
  {:else if error}
    <p class="status-fail">{error}</p>
  {:else}
    <FilterBar tags={allGroups} bind:activeTags bind:searchQuery sortOptions={[]} />

    {#if viewMode === 'grouped'}
      <div class="groups">
        {#each grouped as [groupName, groupWorkflows]}
          <div class="group">
            <button class="group-header" on:click={() => toggleGroup(groupName)}>
              <span class="group-chevron">{@html icon(collapsedGroups[groupName] ? 'chevronRight' : 'chevronDown')}</span>
              <span class="group-name">{groupName}</span>
              <span class="group-count">{groupWorkflows.length}</span>
            </button>
            {#if !collapsedGroups[groupName]}
              <div class="group-grid">
                {#each groupWorkflows as wf}
                  <button class="card surface" on:click={() => push(`/workflow/${wf.file}`)}>
                    <div class="card-name mono">{@html icon('zap', 14)} {wf.name}</div>
                    {#if wf.description}<p class="card-desc text-muted">{wf.description}</p>{/if}
                    <div class="card-footer text-muted">
                      {#if wf.last_run_status}<StatusBadge status={wf.last_run_status} />{:else}<span>Never run</span>{/if}
                    </div>
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        {/each}
      </div>

    {:else if viewMode === 'list'}
      <table class="wf-table">
        <thead><tr><th>Name</th><th>Description</th><th>Group</th><th>Status</th></tr></thead>
        <tbody>
          {#each filtered as wf}
            <tr class="clickable" on:click={() => push(`/workflow/${wf.file}`)}>
              <td class="mono text-cyan">{wf.name}</td>
              <td class="text-muted">{wf.description || ''}</td>
              <td><span class="pill">{getGroup(wf)}</span></td>
              <td>{#if wf.last_run_status}<StatusBadge status={wf.last_run_status} />{:else}<span class="text-muted">--</span>{/if}</td>
            </tr>
          {/each}
        </tbody>
      </table>

    {:else}
      <div class="card-grid">
        {#each filtered as wf}
          <button class="card surface" on:click={() => push(`/workflow/${wf.file}`)}>
            <div class="card-name mono">{@html icon('zap', 14)} {wf.name}</div>
            {#if wf.description}<p class="card-desc text-muted">{wf.description}</p>{/if}
            {#if wf.tags?.length}
              <div class="card-tags">
                {#each wf.tags as tag}<span class="pill">{tag}</span>{/each}
              </div>
            {/if}
            <div class="card-footer text-muted">
              {#if wf.last_run_status}<StatusBadge status={wf.last_run_status} />{:else}<span>Never run</span>{/if}
              {#if wf.author}<span>@{wf.author}</span>{/if}
            </div>
          </button>
        {/each}
      </div>
    {/if}

    {#if filtered.length === 0}<p class="text-muted" style="margin-top:24px">No workflows match your filters.</p>{/if}
  {/if}
</div>

<style>
  /* View toggle */
  .view-btn {
    background: none;
    border: 1px solid transparent;
    color: var(--text-muted);
    padding: 4px 8px;
    border-radius: 4px;
    display: flex;
    align-items: center;
  }
  .view-btn.active {
    color: var(--neon-cyan);
    border-color: var(--neon-cyan);
    background: rgba(0, 229, 255, 0.08);
  }

  /* Grouped view */
  .groups { margin-top: 16px; display: flex; flex-direction: column; gap: 8px; }
  .group { border: 1px solid var(--border); border-radius: 6px; overflow: hidden; }
  .group-header {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    text-align: left;
    padding: 10px 16px;
    background: var(--bg-surface);
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
  }
  .group-header:hover { background: var(--bg-elevated); }
  .group-chevron { display: flex; align-items: center; color: var(--text-muted); }
  .group-name { text-transform: capitalize; }
  .group-count {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-muted);
    background: var(--bg-elevated);
    padding: 1px 8px;
    border-radius: 10px;
  }
  .group-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 12px;
    padding: 12px 16px;
  }

  /* Card grid */
  .card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; margin-top: 20px; }
  .card { text-align: left; padding: 16px; cursor: pointer; display: flex; flex-direction: column; gap: 10px; transition: border-color 0.15s; }
  .card:hover { border-color: rgba(0, 229, 255, 0.3); }
  .card-name { font-family: var(--font-mono); font-size: 14px; color: var(--neon-cyan); display: flex; align-items: center; gap: 6px; }
  .card-desc { font-size: 12px; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
  .card-tags { display: flex; gap: 6px; flex-wrap: wrap; }
  .card-footer { display: flex; align-items: center; gap: 12px; font-size: 12px; margin-top: auto; padding-top: 8px; border-top: 1px solid var(--border); }

  /* List view */
  .wf-table { margin-top: 16px; }
  .clickable { cursor: pointer; }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .text-cyan { color: var(--neon-cyan); }
</style>
