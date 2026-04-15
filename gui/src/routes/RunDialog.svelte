<script>
  import { push } from 'svelte-spa-router';
  import { runWorkflow } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Modal from '../lib/components/Modal.svelte';

  let { name, params = [], onclose } = $props();
  let values = $state({});
  let running = $state(false);
  let error = $state(null);

  async function handleSubmit() {
    running = true; error = null;
    try { const result = await runWorkflow(name, values); onclose?.(); if (result.run_id) push(`/run/${result.run_id}`); } catch (e) { error = e.message; running = false; }
  }
</script>

<Modal title="Run {name}" {onclose}>
  {#if params.length > 0}
    <form on:submit|preventDefault={handleSubmit} class="flex flex-col gap-md">
      {#each params as param}<label class="field"><span class="field-label">{param}</span><input type="text" bind:value={values[param]} placeholder={param} /></label>{/each}
      {#if error}<p class="status-fail" style="font-size:12px">{error}</p>{/if}
      <div class="flex justify-between" style="margin-top:8px">
        <button type="button" on:click={onclose}>Cancel</button>
        <button type="submit" class="primary" disabled={running}>{#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}</button>
      </div>
    </form>
  {:else}
    <div class="flex flex-col gap-md">
      <p class="text-muted">No parameters required.</p>
      {#if error}<p class="status-fail" style="font-size:12px">{error}</p>{/if}
      <div class="flex justify-between">
        <button on:click={onclose}>Cancel</button>
        <button class="primary" disabled={running} on:click={handleSubmit}>{#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}</button>
      </div>
    </div>
  {/if}
</Modal>

<style>
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field-label { font-family: var(--font-mono); font-size: 12px; color: var(--text-muted); }
</style>
