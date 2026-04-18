<script>
  import { icon } from '../icons.js';

  let { tags = [], activeTags = $bindable([]), searchQuery = $bindable(''), sortBy = $bindable('name'), sortOptions = [] } = $props();

  function toggleTag(tag) {
    if (activeTags.includes(tag)) {
      activeTags = activeTags.filter(t => t !== tag);
    } else {
      activeTags = [...activeTags, tag];
    }
  }
</script>

<div class="filter-bar">
  <div class="search-wrap">
    <span class="search-icon">{@html icon('search', 16)}</span>
    <input type="text" placeholder="Search..." bind:value={searchQuery} />
  </div>
  {#if tags.length > 0}
    <div class="tags">
      {#each tags as tag}
        <button class="pill" class:active={activeTags.includes(tag)} onclick={() => toggleTag(tag)}>{tag}</button>
      {/each}
    </div>
  {/if}
  {#if sortOptions.length > 0}
    <select bind:value={sortBy}>
      {#each sortOptions as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>
  {/if}
</div>

<style>
  .filter-bar { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
  .search-wrap { position: relative; flex: 1; min-width: 200px; max-width: 360px; }
  .search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: var(--text-muted); display: flex; z-index: 1; }
  .search-wrap input {
    width: 100%;
    padding-left: 36px;
    border-radius: 10px;
    background: rgba(10,14,20,0.5);
    border: 1px solid rgba(30,42,58,0.6);
    color: var(--text-primary);
    transition: border-color 0.2s, box-shadow 0.2s;
  }
  .search-wrap input:focus {
    outline: none;
    border-color: rgba(0,229,255,0.4);
    box-shadow: 0 0 0 3px rgba(0,229,255,0.06);
  }
  .tags { display: flex; gap: 6px; flex-wrap: wrap; }
  .pill {
    border-radius: 20px;
    background: rgba(30,42,58,0.4);
    border: 1px solid rgba(30,42,58,0.6);
    color: var(--text-muted);
    padding: 4px 14px;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
  }
  .pill:hover {
    background: rgba(0,229,255,0.06);
    border-color: rgba(0,229,255,0.2);
    color: var(--text-primary);
  }
  .pill.active {
    background: rgba(0,229,255,0.1);
    border-color: rgba(0,229,255,0.3);
    color: var(--neon-cyan);
  }
  select { min-width: 140px; }
</style>
