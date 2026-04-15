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

  const allTags = $derived([...new Set(workflows.flatMap(w => w.tags || []))].sort());

  const filtered = $derived.by(() => {
    let list = workflows;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      list = list.filter(w => w.name.toLowerCase().includes(q) || (w.description || '').toLowerCase().includes(q));
    }
    if (activeTags.length > 0) {
      list = list.filter(w => activeTags.some(t => (w.tags || []).includes(t)));
    }
    if (sortBy === 'name') list = [...list].sort((a, b) => a.name.localeCompare(b.name));
    return list;
  });

  const sortOptions = [{ value: 'name', label: 'Name' }];

  onMount(async () => {
    try { workflows = await listWorkflows(); } catch (e) { error = e.message; } finally { loading = false; }
  });
</script>

<div class="page-header"><h1>Workflows</h1></div>
<div class="page-content">
  {#if loading}
    <p class="text-muted">Loading workflows...</p>
  {:else if error}
    <p class="status-fail">{error}</p>
  {:else}
    <FilterBar tags={allTags} bind:activeTags bind:searchQuery bind:sortBy {sortOptions} />
    <div class="card-grid">
      {#each filtered as wf}
        <button class="card surface" on:click={() => push(`/workflow/${wf.name}`)}>
          <div class="card-name mono">{@html icon('zap', 14)} {wf.name}</div>
          {#if wf.description}<p class="card-desc text-muted">{wf.description}</p>{/if}
          {#if wf.tags?.length}
            <div class="card-tags">
              {#each wf.tags as tag}<span class="pill">{tag}</span>{/each}
              {#if wf.version}<span class="pill">v{wf.version}</span>{/if}
            </div>
          {/if}
          <div class="card-footer text-muted">
            {#if wf.last_run_status}<StatusBadge status={wf.last_run_status} />{:else}<span>Never run</span>{/if}
            {#if wf.author}<span>@{wf.author}</span>{/if}
          </div>
        </button>
      {/each}
    </div>
    {#if filtered.length === 0}<p class="text-muted" style="margin-top:24px">No workflows match your filters.</p>{/if}
  {/if}
</div>

<style>
  .card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; margin-top: 20px; }
  .card { text-align: left; padding: 16px; cursor: pointer; display: flex; flex-direction: column; gap: 10px; transition: border-color 0.15s; }
  .card:hover { border-color: rgba(0, 229, 255, 0.3); }
  .card-name { font-family: var(--font-mono); font-size: 14px; color: var(--neon-cyan); display: flex; align-items: center; gap: 6px; }
  .card-desc { font-size: 12px; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
  .card-tags { display: flex; gap: 6px; flex-wrap: wrap; }
  .card-footer { display: flex; align-items: center; gap: 12px; font-size: 12px; margin-top: auto; padding-top: 8px; border-top: 1px solid var(--border); }
</style>
