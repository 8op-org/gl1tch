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
  .search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); color: var(--text-muted); display: flex; }
  .search-wrap input { width: 100%; padding-left: 32px; }
  .tags { display: flex; gap: 6px; flex-wrap: wrap; }
  select { min-width: 140px; }
</style>
