<script>
  import { icon } from '../icons.js';

  let { entries = [], selectedPath = '', onselect, depth = 0 } = $props();
  let expanded = $state({});

  function toggle(name) { expanded[name] = !expanded[name]; expanded = expanded; }

  function fileIcon(name) {
    if (name.endsWith('.md')) return 'text-cyan';
    if (name.endsWith('.diff') || name.endsWith('.patch')) return 'status-pass';
    if (name.endsWith('.json')) return 'text-amber';
    if (name.endsWith('.yaml') || name.endsWith('.yml')) return 'text-magenta';
    if (name.endsWith('.glitch')) return 'text-cyan';
    return 'text-muted';
  }
</script>

<ul class="tree" style="--depth:{depth}">
  {#each entries as entry}
    {#if entry.isDir}
      <li>
        <button class="tree-item dir" on:click={() => { toggle(entry.name); if (entry.loadChildren && !entry.children) entry.loadChildren(); }}>
          <span class="tree-indent" style="width:{depth * 16}px"></span>
          <span class="tree-icon">{@html icon(expanded[entry.name] ? 'chevronDown' : 'chevronRight')}</span>
          <span class="tree-icon text-cyan">{@html icon(expanded[entry.name] ? 'folderOpen' : 'folder', 16)}</span>
          <span class="tree-name">{entry.name}</span>
        </button>
        {#if expanded[entry.name] && entry.children}
          <svelte:self entries={entry.children} {selectedPath} {onselect} depth={depth + 1} />
        {/if}
      </li>
    {:else}
      <li>
        <button class="tree-item file" class:selected={selectedPath === entry.path} on:click={() => onselect?.(entry.path)}>
          <span class="tree-indent" style="width:{depth * 16}px"></span>
          <span class="tree-icon {fileIcon(entry.name)}">{@html icon('file', 16)}</span>
          <span class="tree-name">{entry.name}</span>
        </button>
      </li>
    {/if}
  {/each}
</ul>

<style>
  .tree { list-style: none; padding: 0; margin: 0; }
  .tree-item { display: flex; align-items: center; gap: 4px; width: 100%; text-align: left; background: none; border: none; border-left: 3px solid transparent; color: var(--text-primary); font-size: 12px; font-family: var(--font-mono); padding: 4px 8px; cursor: pointer; white-space: nowrap; }
  .tree-item:hover { background: var(--bg-elevated); }
  .tree-item.selected { background: var(--bg-elevated); border-left-color: var(--neon-cyan); }
  .tree-indent { flex-shrink: 0; }
  .tree-icon { display: flex; align-items: center; flex-shrink: 0; }
  .tree-name { overflow: hidden; text-overflow: ellipsis; }
  .text-cyan { color: var(--neon-cyan); }
  .text-amber { color: var(--neon-amber); }
  .text-magenta { color: var(--neon-magenta); }
</style>
