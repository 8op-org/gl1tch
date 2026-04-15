<script>
  let { segments = [], onnavigate } = $props();
</script>

<nav class="breadcrumb">
  {#each segments as seg, i}
    {#if i > 0}<span class="sep">/</span>{/if}
    {#if seg.href && i < segments.length - 1}
      <a href={seg.href} on:click|preventDefault={() => onnavigate?.(seg.href)}>{seg.label}</a>
    {:else}
      <span class="current">{seg.label}</span>
    {/if}
  {/each}
</nav>

<style>
  .breadcrumb {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
  }
  .sep { color: var(--text-muted); }
  a { color: var(--text-muted); }
  a:hover { color: var(--neon-cyan); }
  .current { color: var(--text-primary); font-weight: 500; }
</style>
