<script>
  import { push } from 'svelte-spa-router'
  import { api } from '../lib/api.js'

  let { name, params = [], onclose } = $props()

  let values = $state(Object.fromEntries(params.map((p) => [p, ''])))
  let running = $state(false)

  async function run() {
    running = true
    try {
      const resp = await api.runWorkflow(name, values)
      onclose?.()
      if (resp.run_id) {
        push(`/run/${resp.run_id}`)
      }
    } catch (e) {
      alert(e.message)
    }
    running = false
  }

  function handleOverlayClick(e) {
    if (e.target === e.currentTarget) onclose?.()
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="overlay" onclick={handleOverlayClick}>
  <div class="dialog">
    <h3>Run {name}</h3>
    {#if params.length > 0}
      {#each params as param}
        <label>
          <span>{param}</span>
          <input bind:value={values[param]} placeholder={param} />
        </label>
      {/each}
    {:else}
      <p class="muted">No parameters required.</p>
    {/if}
    <div class="actions">
      <button onclick={() => onclose?.()}>Cancel</button>
      <button class="primary" onclick={run} disabled={running}>
        {running ? 'Starting...' : 'Start Run'}
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }
  .dialog {
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    min-width: 400px;
    max-width: 500px;
  }
  h3 { font-family: var(--font-mono); margin-bottom: 1rem; }
  label { display: block; margin-bottom: 0.75rem; }
  label span { display: block; color: var(--text-muted); font-size: 12px; margin-bottom: 4px; }
  input {
    width: 100%;
    background: var(--bg);
    border: 1px solid var(--border);
    color: var(--text);
    padding: 0.5rem;
    border-radius: 4px;
    font-family: var(--font-mono);
    font-size: 13px;
  }
  .actions { display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 1rem; }
  .muted { color: var(--text-muted); }
</style>
