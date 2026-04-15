<script>
  import { onMount } from 'svelte'
  import { link } from 'svelte-spa-router'
  import { api } from '../lib/api.js'

  let workflows = $state([])
  let error = $state(null)

  onMount(async () => {
    try {
      workflows = await api.listWorkflows()
    } catch (e) {
      error = e.message
    }
  })
</script>

<div class="workflow-list">
  <h2>Workflows</h2>
  {#if error}
    <p class="error">{error}</p>
  {:else if workflows.length === 0}
    <p class="muted">No workflows found in workspace.</p>
  {:else}
    <ul>
      {#each workflows as wf}
        <li>
          <a href="/workflow/{wf.file}" use:link>
            <span class="name">{wf.name}</span>
            {#if wf.description}
              <span class="desc">{wf.description}</span>
            {/if}
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .workflow-list ul { list-style: none; }
  .workflow-list li {
    border: 1px solid var(--border);
    border-radius: 4px;
    margin-bottom: 0.5rem;
  }
  .workflow-list a {
    display: block;
    padding: 0.75rem 1rem;
    text-decoration: none;
    color: var(--text);
  }
  .workflow-list a:hover { background: var(--bg-hover); }
  .name { font-family: var(--font-mono); color: var(--accent); }
  .desc { display: block; color: var(--text-muted); font-size: 12px; margin-top: 2px; }
  .error { color: var(--danger); }
  .muted { color: var(--text-muted); }
</style>
