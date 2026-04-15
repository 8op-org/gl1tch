<script>
  let { title = '', onclose } = $props();

  function handleKeydown(e) {
    if (e.key === 'Escape' && onclose) onclose();
  }

  function handleBackdrop(e) {
    if (e.target === e.currentTarget && onclose) onclose();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div class="overlay" onclick={handleBackdrop}>
  <div class="modal surface">
    <div class="modal-header">
      <h2>{title}</h2>
      {#if onclose}
        <button class="close-btn" onclick={onclose}>&times;</button>
      {/if}
    </div>
    <div class="modal-body">
      <slot />
    </div>
  </div>
</div>

<style>
  .modal {
    width: 90%;
    max-width: 480px;
    border-color: var(--neon-cyan);
    box-shadow: 0 0 20px rgba(0, 229, 255, 0.15);
  }
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border);
  }
  .modal-header h2 { font-size: 15px; font-weight: 500; }
  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 20px;
    cursor: pointer;
    padding: 0 4px;
  }
  .close-btn:hover { color: var(--text-primary); }
  .modal-body { padding: 20px; }
</style>
