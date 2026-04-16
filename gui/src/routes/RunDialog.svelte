<script>
  import { untrack } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { runWorkflow, getWorkspace } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Modal from '../lib/components/Modal.svelte';

  let { name, params = [], autoParams = {}, onclose } = $props();
  let values = $state({});
  let running = $state(false);
  let error = $state(null);
  let workspaceDefaults = $state({});

  // Fetch workspace defaults on mount
  $effect(() => {
    getWorkspace().then(ws => {
      workspaceDefaults = ws.defaults?.params || {};
    }).catch(() => {});
  });

  // Merge autoParams keys into params so fields appear, and pre-fill values
  const allParams = $derived(() => {
    const set = new Set(params);
    for (const k of Object.keys(autoParams)) set.add(k);
    for (const k of Object.keys(workspaceDefaults)) set.add(k);
    return [...set];
  });

  // Pre-fill from workspace defaults first, then autoParams override
  // Priority: autoParams > workspace defaults > empty
  $effect(() => {
    const wd = workspaceDefaults;
    const ap = autoParams;
    untrack(() => {
      for (const [k, v] of Object.entries(wd)) {
        if (v && !values[k]) values[k] = v;
      }
      for (const [k, v] of Object.entries(ap)) {
        if (v) values[k] = v;
      }
    });
  });

  async function handleSubmit() {
    running = true; error = null;
    try { const result = await runWorkflow(name, values); onclose?.(); if (result.run_id) push(`/run/${result.run_id}`); } catch (e) { error = e.message; running = false; }
  }
</script>

<Modal title="Run {name}" {onclose}>
  {#if allParams().length > 0}
    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="flex flex-col gap-md">
      {#each allParams() as param}<label class="field"><span class="field-label">{param}</span><input type="text" bind:value={values[param]} placeholder={param} /></label>{/each}
      {#if error}<p class="status-fail" style="font-size:12px">{error}</p>{/if}
      <div class="flex justify-between" style="margin-top:8px">
        <button type="button" onclick={onclose}>Cancel</button>
        <button type="submit" class="primary" disabled={running}>{#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}</button>
      </div>
    </form>
  {:else}
    <div class="flex flex-col gap-md">
      <p class="text-muted">No parameters required.</p>
      {#if error}<p class="status-fail" style="font-size:12px">{error}</p>{/if}
      <div class="flex justify-between">
        <button onclick={onclose}>Cancel</button>
        <button class="primary" disabled={running} onclick={handleSubmit}>{#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}</button>
      </div>
    </div>
  {/if}
</Modal>

<style>
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field-label { font-family: var(--font-mono); font-size: 12px; color: var(--text-muted); }
</style>
