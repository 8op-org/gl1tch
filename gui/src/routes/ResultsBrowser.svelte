<script>
  import { onMount } from 'svelte';
  import { EditorView, basicSetup } from 'codemirror';
  import { EditorState } from '@codemirror/state';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { getResult, getResultText, saveResult } from '../lib/api.js';
  import { renderMarkdown } from '../lib/markdown.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import SplitPane from '../lib/components/SplitPane.svelte';
  import FileTree from '../lib/components/FileTree.svelte';

  let tree = $state([]);
  let selectedPath = $state('');
  let fileContent = $state('');
  let mode = $state('preview');
  let dirty = $state(false);
  let loading = $state(true);
  let error = $state(null);
  let editorEl = $state(null);
  let editorView = $state(null);

  const breadcrumbSegments = $derived(() => {
    const parts = selectedPath.split('/').filter(Boolean);
    return [{ label: 'Results', href: '#/results' }, ...parts.map(p => ({ label: p }))];
  });

  const isMarkdown = $derived(selectedPath.endsWith('.md'));
  const isDiff = $derived(selectedPath.endsWith('.diff') || selectedPath.endsWith('.patch'));
  const isJson = $derived(selectedPath.endsWith('.json'));

  async function loadTree(path) {
    try {
      const data = await getResult(path || '');
      if (Array.isArray(data)) return buildTree(data, path || '');
    } catch (e) { error = e.message; }
    return [];
  }

  function buildTree(entries, basePath) {
    return entries.map(entry => {
      const fullPath = basePath ? `${basePath}/${entry.name}` : entry.name;
      const node = { name: entry.name, path: fullPath, isDir: entry.is_dir };
      if (entry.is_dir) {
        node.children = null;
        node.loadChildren = async () => { node.children = await loadTree(fullPath); tree = [...tree]; };
      }
      return node;
    });
  }

  async function selectFile(path) {
    if (editorView) { editorView.destroy(); editorView = null; }
    selectedPath = path; mode = 'preview'; dirty = false;
    try { fileContent = await getResultText(path); } catch (e) { fileContent = `Error loading file: ${e.message}`; }
  }

  function switchToEdit() {
    mode = 'edit';
    requestAnimationFrame(() => {
      if (!editorEl) return;
      if (editorView) editorView.destroy();
      editorView = new EditorView({
        state: EditorState.create({
          doc: fileContent,
          extensions: [basicSetup, oneDark, EditorView.updateListener.of(update => { if (update.docChanged) dirty = true; })],
        }),
        parent: editorEl,
      });
    });
  }

  async function handleSave() {
    if (!editorView) return;
    const content = editorView.state.doc.toString();
    try { await saveResult(selectedPath, content); fileContent = content; dirty = false; } catch (e) { error = e.message; }
  }

  async function handleSelect(path) {
    const entry = findEntry(tree, path);
    if (entry?.isDir && entry.loadChildren && !entry.children) { await entry.loadChildren(); }
    else if (entry && !entry.isDir) { await selectFile(path); }
  }

  function findEntry(entries, path) {
    for (const e of entries) {
      if (e.path === path) return e;
      if (e.children) { const found = findEntry(e.children, path); if (found) return found; }
    }
    return null;
  }

  onMount(async () => { tree = await loadTree(''); loading = false; });
</script>

<div class="page-header">
  <Breadcrumb segments={breadcrumbSegments()} />
  {#if selectedPath}
    <div class="flex items-center gap-sm">
      <button class:primary={mode === 'preview'} on:click={() => { mode = 'preview'; if (editorView) { editorView.destroy(); editorView = null; } }}>{@html icon('eye', 14)} Preview</button>
      <button class:primary={mode === 'edit'} on:click={switchToEdit}>{@html icon('edit', 14)} Edit</button>
      {#if mode === 'edit' && dirty}<button class="primary" on:click={handleSave}>{@html icon('save', 14)} Save</button>{/if}
    </div>
  {/if}
</div>

{#if loading}<div class="page-content"><p class="text-muted">Loading...</p></div>
{:else if error}<div class="page-content"><p class="status-fail">{error}</p></div>
{:else}
  <SplitPane leftWidth={250}>
    <div slot="left" class="tree-pane"><FileTree entries={tree} {selectedPath} onselect={handleSelect} /></div>
    <div slot="right" class="preview-pane">
      {#if !selectedPath}
        <div class="empty-state"><span class="text-muted">{@html icon('folder', 48)}</span><p class="text-muted">Select a file to preview</p></div>
      {:else if mode === 'edit'}
        <div class="editor-wrap" bind:this={editorEl}></div>
      {:else if isMarkdown}
        <div class="rendered-content">{@html renderMarkdown(fileContent)}</div>
      {:else if isJson}
        <pre><code>{(() => { try { return JSON.stringify(JSON.parse(fileContent), null, 2); } catch { return fileContent; } })()}</code></pre>
      {:else if isDiff}
        <pre class="diff-content">{fileContent}</pre>
      {:else}
        <pre><code>{fileContent}</code></pre>
      {/if}
    </div>
  </SplitPane>
{/if}

<style>
  .tree-pane { padding: 8px 0; height: 100%; }
  .preview-pane { flex: 1; display: flex; flex-direction: column; }
  .empty-state { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; opacity: 0.5; }
  .rendered-content { padding: 24px; line-height: 1.6; }
  .rendered-content :global(pre) { margin: 12px 0; }
  .editor-wrap { flex: 1; }
  .editor-wrap :global(.cm-editor) { height: 100%; }
  pre { padding: 24px; margin: 0; white-space: pre-wrap; word-break: break-word; font-family: var(--font-mono); font-size: 12px; }
</style>
