<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { listWorkflows } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import FilterBar from '../lib/components/FilterBar.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';

  function relativeTime(ms) {
    if (!ms) return null;
    const diff = Date.now() - ms;
    const secs = Math.floor(diff / 1000);
    if (secs < 60) return 'just now';
    const mins = Math.floor(secs / 60);
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  }

  let workflows = $state([]);
  let error = $state(null);
  let loading = $state(true);
  let searchQuery = $state('');
  let activeTags = $state([]);
  let sortBy = $state('name');
  let viewMode = $state('grid'); // 'grouped' | 'grid' | 'list'
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
  <h1>{@html icon('workflow', 20)} Workflows</h1>
  <div class="flex items-center gap-sm">
    <button class="view-btn" class:active={viewMode === 'grid'} onclick={() => viewMode = 'grid'} title="Cards">{@html icon('workflow', 16)}</button>
    <button class="view-btn" class:active={viewMode === 'grouped'} onclick={() => viewMode = 'grouped'} title="Grouped">{@html icon('folder', 16)}</button>
    <button class="view-btn" class:active={viewMode === 'list'} onclick={() => viewMode = 'list'} title="List">{@html icon('file', 16)}</button>
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
            <button class="group-header" onclick={() => toggleGroup(groupName)}>
              <span class="group-chevron">{@html icon(collapsedGroups[groupName] ? 'chevronRight' : 'chevronDown')}</span>
              <span class="group-name">{groupName}</span>
              <span class="group-count">{groupWorkflows.length}</span>
            </button>
            {#if !collapsedGroups[groupName]}
              <div class="group-items">
                {#each groupWorkflows as wf}
                  <button class="group-item" onclick={() => push(`/workflow/${wf.file}`)}>
                    <span class="group-item-icon">{@html icon('zap', 14)}</span>
                    <div class="group-item-info">
                      <span class="group-item-name">{wf.name}</span>
                      {#if wf.description}<span class="group-item-desc">{wf.description}</span>{/if}
                    </div>
                    <div class="group-item-status">
                      {#if wf.last_run_status}<StatusBadge status={wf.last_run_status} />{/if}
                      {#if relativeTime(wf.last_run_at)}<span class="group-item-time">{relativeTime(wf.last_run_at)}</span>{/if}
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
        <thead><tr><th>Name</th><th>Description</th><th>Group</th><th>Status</th><th>Last Run</th><th>Runs</th></tr></thead>
        <tbody>
          {#each filtered as wf}
            <tr class="clickable" onclick={() => push(`/workflow/${wf.file}`)}>
              <td class="mono text-cyan">{wf.name}</td>
              <td class="text-muted">{wf.description || ''}</td>
              <td><span class="pill">{getGroup(wf)}</span></td>
              <td>{#if wf.last_run_status}<StatusBadge status={wf.last_run_status} />{:else}<span class="text-muted">--</span>{/if}</td>
              <td class="text-muted">{relativeTime(wf.last_run_at) || '--'}</td>
              <td class="mono text-muted">{wf.run_count != null ? wf.run_count : '--'}</td>
            </tr>
          {/each}
        </tbody>
      </table>

    {:else}
      <div class="card-grid">
        {#each filtered as wf}
          <button class="card surface" onclick={() => push(`/workflow/${wf.file}`)}>
            <div class="card-name mono">{@html icon('zap', 14)} {wf.name}</div>
            {#if wf.description}<p class="card-desc text-muted">{wf.description}</p>{/if}
            {#if wf.tags?.length}
              <div class="card-tags">
                {#each wf.tags as tag}<span class="pill">{tag}</span>{/each}
              </div>
            {/if}
            <div class="card-footer text-muted">
              {#if wf.last_run_status}<StatusBadge status={wf.last_run_status} />{:else}<span>Never run</span>{/if}
              {#if relativeTime(wf.last_run_at)}<span>{relativeTime(wf.last_run_at)}</span>{/if}
              {#if wf.run_count != null}<span class="card-sep">|</span><span><span class="mono">{wf.run_count}</span> runs</span>{/if}
            </div>
          </button>
        {/each}
      </div>
    {/if}

    {#if filtered.length === 0}<p class="text-muted" style="margin-top:24px">No workflows match your filters.</p>{/if}
  {/if}
</div>

<style>
  h1 { display: flex; align-items: center; gap: 8px; }

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

  /* Grouped view — modern list style */
  .groups { margin-top: 16px; display: flex; flex-direction: column; gap: 12px; }
  .group { border: 1px solid var(--border); border-radius: 8px; overflow: hidden; background: var(--bg-surface); }
  .group-header {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    text-align: left;
    padding: 12px 20px;
    background: transparent;
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.02em;
    transition: background 0.15s;
  }
  .group-header:hover { background: var(--bg-elevated); }
  .group-chevron { display: flex; align-items: center; color: var(--neon-cyan); }
  .group-name { text-transform: capitalize; }
  .group-count {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--neon-cyan);
    background: rgba(0, 229, 255, 0.08);
    padding: 2px 10px;
    border-radius: 12px;
    border: 1px solid rgba(0, 229, 255, 0.15);
  }
  .group-items { display: flex; flex-direction: column; }
  .group-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 20px 10px 44px;
    background: none;
    border: none;
    border-top: 1px solid var(--border);
    color: var(--text-primary);
    cursor: pointer;
    text-align: left;
    transition: background 0.1s;
    width: 100%;
  }
  .group-item:hover { background: var(--bg-elevated); }
  .group-item-icon { color: var(--neon-cyan); display: flex; flex-shrink: 0; }
  .group-item-info { display: flex; flex-direction: column; gap: 2px; flex: 1; min-width: 0; }
  .group-item-name { font-family: var(--font-mono); font-size: 13px; color: var(--text-primary); }
  .group-item-desc { font-size: 12px; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .group-item-status { flex-shrink: 0; display: flex; align-items: center; gap: 8px; }
  .group-item-time { font-size: 11px; color: var(--text-muted); }

  /* Card grid */
  .card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; margin-top: 20px; }
  .card { text-align: left; padding: 16px; cursor: pointer; display: flex; flex-direction: column; gap: 10px; transition: border-color 0.15s, box-shadow 0.15s; }
  .card:hover { border-color: rgba(0, 229, 255, 0.3); box-shadow: 0 0 12px rgba(0, 229, 255, 0.06); }
  .card-name { font-family: var(--font-mono); font-size: 14px; color: var(--neon-cyan); display: flex; align-items: center; gap: 6px; }
  .card-desc { font-size: 12px; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
  .card-tags { display: flex; gap: 6px; flex-wrap: wrap; }
  .card-footer { display: flex; align-items: center; gap: 8px; font-size: 12px; margin-top: auto; padding-top: 8px; border-top: 1px solid var(--border); }
  .card-sep { color: var(--text-muted); opacity: 0.4; }

  /* List view */
  .wf-table { margin-top: 16px; }
  .clickable { cursor: pointer; }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .text-cyan { color: var(--neon-cyan); }
</style>
