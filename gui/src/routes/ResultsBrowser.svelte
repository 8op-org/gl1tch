<script>
  import { onMount } from 'svelte'
  import { api } from '../lib/api.js'
  import { renderMarkdown } from '../lib/markdown.js'

  let tree = $state([])
  let currentPath = $state('')
  let fileContent = $state(null)
  let fileType = $state(null)
  let error = $state(null)

  onMount(() => loadDir(''))

  async function loadDir(path) {
    currentPath = path
    fileContent = null
    try {
      tree = await api.getResult(path || '.')
    } catch (e) {
      error = e.message
      tree = []
    }
  }

  async function openFile(name) {
    const path = currentPath ? `${currentPath}/${name}` : name
    try {
      const content = await api.getResult(path)
      fileContent = typeof content === 'string' ? content : JSON.stringify(content, null, 2)
      fileType = name.split('.').pop()
    } catch (e) {
      error = e.message
    }
  }

  function navigate(name, isDir) {
    if (isDir) {
      loadDir(currentPath ? `${currentPath}/${name}` : name)
    } else {
      openFile(name)
    }
  }

  function goUp() {
    const parts = currentPath.split('/')
    parts.pop()
    loadDir(parts.join('/'))
  }

  function renderContent(content, type) {
    if (type === 'md') return renderMarkdown(content)
    return `<pre><code>${content.replace(/</g, '&lt;')}</code></pre>`
  }
</script>

<div class="results">
  <div class="breadcrumb">
    <span>results/</span>
    {#if currentPath}
      <button class="link" onclick={goUp}>..</button>
      <span>{currentPath}</span>
    {/if}
  </div>

  <div class="split">
    <div class="file-tree">
      {#each tree as entry}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="entry" class:dir={entry.is_dir} onclick={() => navigate(entry.name, entry.is_dir)}>
          <span class="icon">{entry.is_dir ? '/' : ''}</span>
          {entry.name}
        </div>
      {/each}
    </div>

    <div class="preview">
      {#if fileContent !== null}
        <div class="rendered">{@html renderContent(fileContent, fileType)}</div>
      {:else}
        <p class="muted">Select a file to preview.</p>
      {/if}
    </div>
  </div>
</div>

<style>
  .results { height: calc(100vh - 60px); display: flex; flex-direction: column; }
  .breadcrumb {
    padding: 0.5rem 0;
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 13px;
    border-bottom: 1px solid var(--border);
    margin-bottom: 0.5rem;
  }
  .link { background: none; border: none; color: var(--accent); cursor: pointer; font-family: var(--font-mono); padding: 0; }
  .split { display: flex; flex: 1; gap: 1rem; overflow: hidden; }
  .file-tree { width: 250px; overflow-y: auto; border-right: 1px solid var(--border); padding-right: 1rem; }
  .entry {
    padding: 0.35rem 0.5rem;
    cursor: pointer;
    font-family: var(--font-mono);
    font-size: 13px;
    border-radius: 4px;
  }
  .entry:hover { background: var(--bg-hover); }
  .entry.dir { color: var(--accent); }
  .icon { display: inline-block; width: 1em; }
  .preview { flex: 1; overflow-y: auto; }
  .rendered { padding: 1rem; background: var(--bg-surface); border-radius: 4px; }
  .rendered :global(pre) { overflow-x: auto; }
  .muted { color: var(--text-muted); }
</style>
