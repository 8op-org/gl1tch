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
    background: linear-gradient(145deg, rgba(17,24,32,0.95), rgba(26,34,48,0.85));
    border: 1px solid rgba(0,229,255,0.1);
    border-radius: 16px;
    box-shadow: 0 16px 64px rgba(0,0,0,0.5);
    backdrop-filter: blur(16px);
  }
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid rgba(30,42,58,0.5);
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
