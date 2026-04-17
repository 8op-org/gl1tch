<script>
  import { getResultText } from '../api.js';

  let { steps = [] } = $props();
  let expanded = $state(false);
  let viewingFile = $state(null);
  let fileContent = $state(null);
  let fileLoading = $state(false);

  const artifactGroups = $derived.by(() => {
    const groups = [];
    for (const step of steps) {
      if (step.artifacts?.length) {
        groups.push({ stepId: step.step_id, files: step.artifacts });
      }
    }
    return groups;
  });

  const totalCount = $derived(
    artifactGroups.reduce((sum, g) => sum + g.files.length, 0)
  );

  function fileName(path) {
    return path.split('/').pop();
  }

  async function viewFile(path) {
    viewingFile = path;
    fileLoading = true;
    try {
      const stripped = path.replace(/^results\//, '');
      try {
        fileContent = await getResultText(stripped);
      } catch {
        fileContent = await getResultText(path);
      }
    } catch (e) {
      fileContent = `Error: ${e.message}`;
    } finally {
      fileLoading = false;
    }
  }

  function closeViewer() {
    viewingFile = null;
    fileContent = null;
  }
</script>

{#if totalCount > 0}
  <div class="artifacts-bar">
    <button class="artifacts-toggle" onclick={() => { expanded = !expanded; viewingFile = null; }} type="button">
      <span class="artifacts-label">Artifacts</span>
      <span class="artifacts-count">{totalCount}</span>
      <span class="artifacts-chevron">{expanded ? '\u25BE' : '\u25B8'}</span>
    </button>

    {#if expanded}
      <div class="artifacts-content">
        {#if viewingFile}
          <div class="artifact-viewer">
            <button class="back-btn" onclick={closeViewer} type="button">&larr; Back</button>
            <span class="file-path mono">{viewingFile}</span>
            {#if fileLoading}
              <p class="text-muted">Loading...</p>
            {:else}
              <pre class="file-content"><code>{fileContent}</code></pre>
            {/if}
          </div>
        {:else}
          {#each artifactGroups as group}
            <div class="artifact-group">
              <span class="group-label mono">{group.stepId}</span>
              <div class="group-files">
                {#each group.files as path}
                  <button class="artifact-file" onclick={() => viewFile(path)} type="button">
                    <span class="mono">{fileName(path)}</span>
                    <span class="file-hint text-muted">{path}</span>
                  </button>
                {/each}
              </div>
            </div>
          {/each}
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .artifacts-bar { border-top: 1px solid var(--border); background: var(--bg-surface); flex-shrink: 0; }
  .artifacts-toggle { display: flex; align-items: center; gap: 8px; padding: 8px 24px; width: 100%; background: none; border: none; color: var(--text-primary); cursor: pointer; font-size: 12px; text-align: left; }
  .artifacts-toggle:hover { background: var(--bg-elevated); }
  .artifacts-label { font-weight: 500; text-transform: uppercase; letter-spacing: 0.05em; font-size: 11px; color: var(--text-muted); }
  .artifacts-count { font-size: 10px; background: rgba(0, 229, 255, 0.15); color: var(--neon-cyan); padding: 0 6px; border-radius: 8px; font-weight: 600; font-family: var(--font-mono); }
  .artifacts-chevron { color: var(--text-muted); margin-left: auto; }
  .artifacts-content { padding: 0 24px 16px; max-height: 300px; overflow-y: auto; }
  .artifact-group { margin-bottom: 12px; }
  .group-label { font-size: 11px; color: var(--neon-cyan); display: block; margin-bottom: 4px; }
  .group-files { display: flex; flex-direction: column; gap: 4px; }
  .artifact-file { display: flex; align-items: center; gap: 12px; padding: 6px 10px; background: var(--bg-deep); border: 1px solid var(--border); border-radius: 4px; cursor: pointer; text-align: left; color: var(--text-primary); font-size: 12px; width: 100%; }
  .artifact-file:hover { border-color: var(--neon-cyan); }
  .file-hint { font-size: 10px; }
  .artifact-viewer { display: flex; flex-direction: column; gap: 8px; }
  .back-btn { align-self: flex-start; font-size: 12px; padding: 4px 8px; background: none; border: 1px solid var(--border); border-radius: 4px; color: var(--text-muted); cursor: pointer; }
  .back-btn:hover { color: var(--text-primary); border-color: var(--neon-cyan); }
  .file-path { font-size: 11px; color: var(--text-muted); }
  .file-content { background: var(--bg-deep); border: 1px solid var(--border); border-radius: 6px; padding: 12px; font-size: 12px; line-height: 1.5; font-family: var(--font-mono); color: var(--text-primary); white-space: pre-wrap; word-break: break-word; max-height: 250px; overflow-y: auto; }
  .mono { font-family: var(--font-mono); }
  .text-muted { color: var(--text-muted); }
</style>
