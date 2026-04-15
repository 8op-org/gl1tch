<script>
  import { onMount } from 'svelte'
  import { EditorView, basicSetup } from 'codemirror'
  import { EditorState } from '@codemirror/state'
  import { oneDark } from '@codemirror/theme-one-dark'
  import { api } from '../lib/api.js'
  import RunDialog from './RunDialog.svelte'

  let { params = {} } = $props()

  let editorEl
  let view
  let saving = $state(false)
  let saved = $state(false)
  let showRunDialog = $state(false)
  let workflowParams = $state([])

  onMount(async () => {
    const data = await api.getWorkflow(params.name)
    workflowParams = data.params || []

    view = new EditorView({
      state: EditorState.create({
        doc: data.source,
        extensions: [basicSetup, oneDark],
      }),
      parent: editorEl,
    })
  })

  async function save() {
    saving = true
    const source = view.state.doc.toString()
    await api.saveWorkflow(params.name, source)
    saving = false
    saved = true
    setTimeout(() => (saved = false), 2000)
  }
</script>

<div class="editor-page">
  <div class="toolbar">
    <h2>{params.name}</h2>
    <div class="actions">
      <button onclick={save} disabled={saving}>
        {saving ? 'Saving...' : saved ? 'Saved' : 'Save'}
      </button>
      <button class="primary" onclick={() => (showRunDialog = true)}>
        Run
      </button>
    </div>
  </div>
  <div class="editor" bind:this={editorEl}></div>
</div>

{#if showRunDialog}
  <RunDialog
    name={params.name}
    params={workflowParams}
    onclose={() => (showRunDialog = false)}
  />
{/if}

<style>
  .editor-page { display: flex; flex-direction: column; height: calc(100vh - 60px); }
  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border);
    margin-bottom: 0.5rem;
  }
  .toolbar h2 { font-family: var(--font-mono); font-size: 14px; }
  .actions { display: flex; gap: 0.5rem; }
  .editor { flex: 1; overflow: auto; }
  .editor :global(.cm-editor) { height: 100%; }
</style>
