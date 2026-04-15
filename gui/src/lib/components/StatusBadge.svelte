<script>
  let { status = 'unknown', size = 'sm' } = $props();

  const config = {
    PASS: { cls: 'status-pass', glow: 'var(--glow-green)' },
    FAIL: { cls: 'status-fail', glow: 'var(--glow-magenta)' },
    RUNNING: { cls: 'status-running', glow: 'var(--glow-amber)' },
  };

  const upper = $derived((status || '').toUpperCase());
  const c = $derived(config[upper] || { cls: 'text-muted', glow: 'none' });
  const isRunning = $derived(upper === 'RUNNING');
</script>

<span class="badge {c.cls} {size}" class:pulse={isRunning} style="--badge-glow: {c.glow}">
  <span class="dot"></span>
  {upper || 'UNKNOWN'}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-weight: 500;
  }
  .badge.sm { font-size: 11px; }
  .badge.md { font-size: 13px; }
  .badge.lg { font-size: 15px; }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: var(--badge-glow);
  }
</style>
