<script>
  let { leftWidth = 250, minLeft = 200, maxLeftPercent = 50 } = $props();
  let width = $state(leftWidth);
  let dragging = $state(false);
  let containerEl = $state(null);

  function onMouseDown(e) {
    e.preventDefault();
    dragging = true;
    const onMove = (e) => {
      if (!containerEl) return;
      const rect = containerEl.getBoundingClientRect();
      const maxPx = rect.width * maxLeftPercent / 100;
      width = Math.min(maxPx, Math.max(minLeft, e.clientX - rect.left));
    };
    const onUp = () => {
      dragging = false;
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }
</script>

<div class="split-pane" bind:this={containerEl} class:dragging>
  <div class="split-left" style="width:{width}px"><slot name="left" /></div>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="split-handle" onmousedown={onMouseDown}></div>
  <div class="split-right"><slot name="right" /></div>
</div>

<style>
  .split-pane { display: flex; flex: 1; overflow: hidden; }
  .split-pane.dragging { cursor: col-resize; user-select: none; }
  .split-left { flex-shrink: 0; overflow-y: auto; border-right: 1px solid var(--border); }
  .split-handle { width: 4px; cursor: col-resize; background: transparent; transition: background 0.15s; flex-shrink: 0; }
  .split-handle:hover, .dragging .split-handle { background: var(--neon-cyan); opacity: 0.3; }
  .split-right { flex: 1; overflow-y: auto; display: flex; flex-direction: column; }
</style>
