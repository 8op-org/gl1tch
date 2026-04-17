<script>
  import { onMount, tick } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { getWorkflow, saveWorkflow, listWorkflowRuns } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';
  import RunDialog from './RunDialog.svelte';
  import PipelineGraph from '../lib/components/PipelineGraph.svelte';
  import { EditorView, basicSetup } from 'codemirror';
  import { EditorState } from '@codemirror/state';
  import { keymap } from '@codemirror/view';
  import { StreamLanguage } from '@codemirror/language';
  import { scheme } from '@codemirror/legacy-modes/mode/scheme';
  import { oneDark } from '@codemirror/theme-one-dark';

  let { params } = $props();
  let name = $derived(params?.name || '');

  // Tab state
  let activeTab = $state('runs');

  // Workflow data
  let source = $state('');
  let workflowParams = $state([]);
  let metadata = $state({});
  let runs = $state([]);
  let dirty = $state(false);
  let saveStatus = $state('');
  let showRunDialog = $state(false);
  let error = $state(null);
  let loading = $state(true);

  // Editor
  let editorEl = $state(null);
  let editorView = $state(null);

  // Expanded run (for pipeline graph)
  let expandedRunId = $state(null);

  // CodeMirror cyberpunk theme
  const cyberpunkTheme = EditorView.theme({
    '&': { backgroundColor: '#0a0e14', color: '#e0e6ed', fontSize: '13px' },
    '.cm-content': { fontFamily: "'JetBrains Mono', monospace", lineHeight: '1.7', padding: '16px 0' },
    '.cm-gutters': { backgroundColor: '#111820', color: '#5a6a7a', borderRight: '1px solid #1e2a3a', paddingLeft: '8px' },
    '.cm-activeLineGutter': { backgroundColor: '#1a2230' },
    '.cm-activeLine': { backgroundColor: 'rgba(0, 229, 255, 0.03)' },
    '.cm-cursor': { borderLeftColor: '#00e5ff' },
    '.cm-selectionBackground, ::selection': { backgroundColor: 'rgba(0, 229, 255, 0.15) !important' },
    '.cm-matchingBracket': { backgroundColor: 'rgba(0, 229, 255, 0.12)', outline: '1px solid rgba(0, 229, 255, 0.3)' },
    '.cm-foldGutter .cm-gutterElement': { color: '#5a6a7a', cursor: 'pointer' },
    '.cm-scroller': { overflow: 'auto' },
  }, { dark: true });

  function initEditor(content) {
    if (editorView) editorView.destroy();
    editorView = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          basicSetup,
          StreamLanguage.define(scheme),
          oneDark,
          cyberpunkTheme,
          EditorView.updateListener.of(update => {
            if (update.docChanged) dirty = true;
          }),
        ],
      }),
      parent: editorEl,
    });
  }

  function destroyEditor() {
    if (editorView) { editorView.destroy(); editorView = null; }
  }

  // Data loading
  async function load() {
    loading = true;
    error = null;
    try {
      const data = await getWorkflow(name);
      source = data.source || '';
      workflowParams = data.params || [];
      metadata = {
        name: data.name,
        description: data.description,
        tags: data.tags,
        author: data.author,
        version: data.version,
        created: data.created,
      };
      // Load runs
      try { runs = await listWorkflowRuns(name); } catch (_) { runs = []; }
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  // Save handler
  async function handleSave() {
    if (!dirty) return;
    const content = editorView ? editorView.state.doc.toString() : source;
    saveStatus = 'saving';
    try {
      await saveWorkflow(name, content);
      source = content;
      dirty = false;
      saveStatus = 'saved';
      setTimeout(() => { saveStatus = ''; }, 2000);
    } catch (e) {
      saveStatus = 'error';
      console.error('Save failed:', e);
    }
  }

  // Keyboard shortcut: Cmd/Ctrl+S
  function handleKeydown(e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 's') {
      e.preventDefault();
      if (activeTab === 'source') handleSave();
    }
  }

  // Helpers
  function formatDuration(startMs, endMs) {
    if (!endMs) return '...';
    const sec = Math.round((endMs - startMs) / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  function runStatus(run) {
    if (!run.finished_at) return 'running';
    return run.exit_status === 0 ? 'pass' : 'fail';
  }

  function formatTokens(input, output) {
    if ((input == null || input === 0) && (output == null || output === 0)) return '--';
    return `${(input || 0).toLocaleString()} / ${(output || 0).toLocaleString()}`;
  }

  function formatCost(cost) {
    if (cost == null || cost === 0) return '--';
    return `$${cost.toFixed(4)}`;
  }

  // Breadcrumb segments
  const breadcrumbSegments = $derived([
    { label: 'Workflows', href: '/' },
    { label: decodeURIComponent(name) },
  ]);

  // Use action to init CodeMirror when editor container mounts
  function mountEditor(node) {
    editorEl = node;
    if (!editorView && source != null) initEditor(source);
    return { destroy() { destroyEditor(); } };
  }

  // Reload when name changes (via navigation between workflows)
  let prevName = '';
  $effect(() => {
    if (name && name !== prevName) {
      const isInitial = prevName === '';
      prevName = name;
      if (!isInitial) {
        destroyEditor();
        dirty = false;
        activeTab = 'runs';
        load();
      }
    }
  });

  onMount(() => {
    load();
    return () => destroyEditor();
  });
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="page-header">
  <Breadcrumb segments={breadcrumbSegments} onnavigate={(href) => push(href)} />
  <div class="header-actions">
    {#if activeTab === 'source'}
      <button
        class="primary"
        disabled={!dirty || saveStatus === 'saving'}
        onclick={handleSave}
      >
        {#if saveStatus === 'saving'}
          Saving...
        {:else if saveStatus === 'saved'}
          Saved
        {:else}
          {@html icon('save', 14)} Save
        {/if}
      </button>
    {/if}
    <button class="primary" onclick={() => { showRunDialog = true; }}>
      {@html icon('play', 14)} Run
    </button>
  </div>
</div>

<div class="tabs">
  <button class="tab" class:active={activeTab === 'runs'} onclick={() => activeTab = 'runs'}>
    {@html icon('terminal', 16)} Runs
  </button>
  <button class="tab" class:active={activeTab === 'source'} onclick={() => activeTab = 'source'}>
    {@html icon('file', 16)} Source
  </button>
  <button class="tab" class:active={activeTab === 'metadata'} onclick={() => activeTab = 'metadata'}>
    {@html icon('settings', 16)} Metadata
  </button>
</div>

<div class="tab-content">
  {#if loading}
    <div class="page-content">
      <p class="text-muted">Loading workflow...</p>
    </div>
  {:else if error}
    <div class="page-content">
      <p class="status-fail">{error}</p>
    </div>

  {:else if activeTab === 'runs'}
    <div class="page-content">
      {#if runs.length === 0}
        <div class="empty-state">
          <p class="text-muted">No runs yet. Click Run to execute this workflow.</p>
        </div>
      {:else}
        <div class="runs-list">
          {#each runs.toSorted((a, b) => (b.started_at || 0) - (a.started_at || 0)) as run}
            <button
              class="run-row"
              class:expanded={expandedRunId === run.id}
              onclick={() => { expandedRunId = expandedRunId === run.id ? null : run.id; }}
            >
              <StatusBadge status={runStatus(run)} />
              <span class="run-id mono">#{run.id ?? '--'}</span>
              <span class="run-name">{run.name || run.workflow || '--'}</span>
              <span class="run-time text-muted">
                {run.started_at ? new Date(run.started_at).toLocaleString() : '--'}
              </span>
              <span class="run-duration mono">
                {formatDuration(run.started_at, run.finished_at)}
              </span>
              <span class="run-model text-muted">{run.model || '--'}</span>
              <span class="run-tokens mono text-muted">
                {formatTokens(run.tokens_in, run.tokens_out)}
              </span>
              <span class="run-cost mono">
                {formatCost(run.cost_usd)}
              </span>
              <span class="run-chevron">
                {#if expandedRunId === run.id}
                  {@html icon('chevronDown')}
                {:else}
                  {@html icon('chevronRight')}
                {/if}
              </span>
            </button>
            {#if expandedRunId === run.id}
              <div class="run-detail">
                <PipelineGraph runId={run.id} />
              </div>
            {/if}
          {/each}
        </div>
      {/if}
    </div>

  {:else if activeTab === 'source'}
    <div class="editor-wrap">
      <div class="editor-container" use:mountEditor></div>
    </div>

  {:else if activeTab === 'metadata'}
    <div class="page-content">
      <div class="meta-grid">
        <span class="meta-label">Name</span>
        <span class="meta-value mono">{metadata.name || '--'}</span>

        <span class="meta-label">Description</span>
        <span class="meta-value">{metadata.description || '--'}</span>

        <span class="meta-label">Tags</span>
        <span class="meta-value">
          {#if metadata.tags?.length}
            <span class="tag-list">
              {#each metadata.tags as tag}
                <span class="pill">{tag}</span>
              {/each}
            </span>
          {:else}
            <span class="text-muted">--</span>
          {/if}
        </span>

        <span class="meta-label">Author</span>
        <span class="meta-value">{metadata.author || '--'}</span>

        <span class="meta-label">Version</span>
        <span class="meta-value mono">{metadata.version || '--'}</span>

        <span class="meta-label">Created</span>
        <span class="meta-value">
          {metadata.created ? new Date(metadata.created).toLocaleString() : '--'}
        </span>

        <span class="meta-label">Parameters</span>
        <span class="meta-value">
          {#if workflowParams.length}
            <div class="params-list">
              {#each workflowParams as param}
                <div class="param-item">
                  <span class="mono">{typeof param === 'string' ? param : param.name || param}</span>
                  {#if typeof param === 'object' && param.type}
                    <span class="pill">{param.type}</span>
                  {/if}
                </div>
              {/each}
            </div>
          {:else}
            <span class="text-muted">None</span>
          {/if}
        </span>
      </div>
    </div>
  {/if}
</div>

{#if showRunDialog}
  <RunDialog {name} params={workflowParams} autoParams={{}} onclose={() => { showRunDialog = false; load(); }} />
{/if}

<style>
  /* Header */
  .header-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* Tabs */
  .tabs {
    display: flex;
    gap: 0;
    border-bottom: 1px solid var(--border);
    padding: 0 24px;
  }
  .tab {
    padding: 10px 16px;
    font-size: 13px;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: var(--text-muted);
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 6px;
    transition: all 0.15s;
    border-radius: 0;
  }
  .tab:hover {
    color: var(--text-primary);
    background: none;
  }
  .tab.active {
    color: var(--neon-cyan);
    border-bottom-color: var(--neon-cyan);
  }

  /* Tab content area */
  .tab-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Runs */
  .runs-list {
    display: flex;
    flex-direction: column;
  }
  .run-row {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 16px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg-surface);
    cursor: pointer;
    transition: all 0.15s;
    margin-bottom: 8px;
    text-align: left;
    width: 100%;
    font-size: 13px;
    color: var(--text-primary);
  }
  .run-row:hover {
    border-color: rgba(0, 229, 255, 0.3);
    background: var(--bg-elevated);
  }
  .run-row.expanded {
    border-radius: 6px 6px 0 0;
    margin-bottom: 0;
    border-color: rgba(0, 229, 255, 0.3);
  }
  .run-id {
    font-size: 11px;
    color: var(--text-muted);
    flex-shrink: 0;
  }
  .run-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .run-time {
    font-size: 12px;
    flex-shrink: 0;
  }
  .run-duration {
    font-size: 12px;
    color: var(--neon-cyan);
    flex-shrink: 0;
    min-width: 48px;
  }
  .run-model {
    font-size: 12px;
    flex-shrink: 0;
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .run-tokens {
    font-size: 11px;
    flex-shrink: 0;
  }
  .run-cost {
    font-size: 11px;
    color: var(--neon-amber);
    flex-shrink: 0;
  }
  .run-chevron {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    color: var(--text-muted);
  }

  .run-detail {
    padding: 0;
    border: 1px solid rgba(0, 229, 255, 0.3);
    border-top: none;
    border-radius: 0 0 6px 6px;
    background: var(--bg-elevated);
    margin-bottom: 8px;
    min-height: 200px;
    overflow-x: auto;
    overflow: hidden;
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 48px 24px;
  }

  /* Editor */
  .editor-wrap {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .editor-container {
    flex: 1;
    overflow: hidden;
  }
  .editor-container :global(.cm-editor) {
    height: 100%;
  }

  /* Metadata */
  .meta-grid {
    display: grid;
    grid-template-columns: 140px 1fr;
    gap: 12px 16px;
    max-width: 600px;
  }
  .meta-label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    padding-top: 2px;
  }
  .meta-value {
    font-size: 13px;
  }
  .tag-list {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }
  .params-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .param-item {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .mono {
    font-family: var(--font-mono);
    font-size: 12px;
  }
</style>
