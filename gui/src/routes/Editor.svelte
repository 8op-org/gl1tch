<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { EditorView, basicSetup } from 'codemirror';
  import { EditorState } from '@codemirror/state';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { getWorkflow, saveWorkflow, listRuns } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';
  import RunDialog from './RunDialog.svelte';

  let { params } = $props();
  let name = $derived(params?.name || '');
  let source = $state('');
  let workflowParams = $state([]);
  let metadata = $state({});
  let recentRuns = $state([]);
  let dirty = $state(false);
  let saveStatus = $state('');
  let showRunDialog = $state(false);
  let showMeta = $state(true);
  let editorEl = $state(null);
  let editorView = $state(null);
  let error = $state(null);

  const breadcrumbs = $derived([{ label: 'Workflows', href: '#/' }, { label: name }]);

  async function load() {
    try {
      const data = await getWorkflow(name);
      source = data.source || '';
      workflowParams = data.params || [];
      metadata = { tags: data.tags, author: data.author, version: data.version, created: data.created };
      initEditor(source);
      try { const runs = await listRuns(); recentRuns = runs.filter(r => r.workflow === name).slice(0, 5); } catch (_) {}
    } catch (e) { error = e.message; }
  }

  function initEditor(content) {
    if (editorView) editorView.destroy();
    editorView = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [basicSetup, oneDark, EditorView.updateListener.of(update => { if (update.docChanged) dirty = true; })],
      }),
      parent: editorEl,
    });
  }

  async function handleSave() {
    if (!editorView) return;
    const content = editorView.state.doc.toString();
    saveStatus = 'saving';
    try { await saveWorkflow(name, content); dirty = false; saveStatus = 'saved'; setTimeout(() => { saveStatus = ''; }, 2000); } catch (e) { saveStatus = 'error'; }
  }

  function handleKeydown(e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 's') { e.preventDefault(); handleSave(); }
  }

  onMount(() => { load(); });
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="page-header">
  <Breadcrumb segments={breadcrumbs} onnavigate={(href) => push(href.replace('#', ''))} />
  <div class="flex items-center gap-sm">
    {#if saveStatus === 'saved'}<span class="status-pass" style="font-size:12px">Saved</span>{:else if saveStatus === 'saving'}<span class="text-muted" style="font-size:12px">Saving...</span>{/if}
    <button class:primary={dirty} disabled={!dirty} on:click={handleSave}>{@html icon('save', 14)} Save</button>
    <button class="primary" on:click={() => showRunDialog = true}>{@html icon('play', 14)} Run</button>
  </div>
</div>

{#if error}
  <div class="page-content"><p class="status-fail">{error}</p></div>
{:else}
  <div class="editor-layout">
    <div class="editor-pane" bind:this={editorEl}></div>
    {#if showMeta}
      <aside class="meta-panel">
        <div class="meta-header"><h3>Metadata</h3><button class="close-btn" on:click={() => showMeta = false}>&times;</button></div>
        <div class="meta-body">
          {#if metadata.tags?.length}<div class="meta-section"><span class="meta-label">Tags</span><div class="flex gap-sm" style="flex-wrap:wrap">{#each metadata.tags as tag}<span class="pill">{tag}</span>{/each}</div></div>{/if}
          {#if metadata.author}<div class="meta-section"><span class="meta-label">Author</span><span class="mono">@{metadata.author}</span></div>{/if}
          {#if metadata.version}<div class="meta-section"><span class="meta-label">Version</span><span class="mono">{metadata.version}</span></div>{/if}
          {#if recentRuns.length > 0}
            <div class="meta-section"><span class="meta-label">Recent Runs</span>
              <div class="run-list">{#each recentRuns as run}<a href="#/run/{run.id}" class="run-item"><StatusBadge status={run.status} /><span class="text-muted" style="font-size:11px">{run.started ? new Date(run.started).toLocaleString() : ''}</span></a>{/each}</div>
            </div>
          {/if}
        </div>
      </aside>
    {:else}
      <button class="meta-toggle" on:click={() => showMeta = true} title="Show metadata">{@html icon('chevronRight')}</button>
    {/if}
  </div>
{/if}

{#if showRunDialog}<RunDialog {name} params={workflowParams} onclose={() => showRunDialog = false} />{/if}

<style>
  .editor-layout { display: flex; flex: 1; overflow: hidden; }
  .editor-pane { flex: 1; overflow: auto; }
  .editor-pane :global(.cm-editor) { height: 100%; }
  .editor-pane :global(.cm-editor .cm-scroller) { font-family: var(--font-mono); font-size: 13px; }
  .meta-panel { width: 250px; border-left: 1px solid var(--border); background: var(--bg-surface); overflow-y: auto; flex-shrink: 0; }
  .meta-header { display: flex; align-items: center; justify-content: space-between; padding: 12px 16px; border-bottom: 1px solid var(--border); }
  .close-btn { background: none; border: none; color: var(--text-muted); font-size: 18px; cursor: pointer; padding: 0; }
  .meta-body { padding: 16px; }
  .meta-section { margin-bottom: 16px; }
  .meta-label { display: block; font-size: 11px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin-bottom: 6px; }
  .run-list { display: flex; flex-direction: column; gap: 6px; }
  .run-item { display: flex; align-items: center; gap: 8px; padding: 4px 0; text-decoration: none; }
  .run-item:hover { background: var(--bg-elevated); border-radius: 4px; padding: 4px 6px; margin: 0 -6px; }
  .meta-toggle { position: absolute; right: 0; top: 50%; transform: translateY(-50%); background: var(--bg-surface); border: 1px solid var(--border); border-right: none; border-radius: 4px 0 0 4px; padding: 8px 4px; cursor: pointer; color: var(--text-muted); }
</style>
