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
    border-radius: 20px;
    padding: 4px 12px;
    border: 1px solid transparent;
  }
  .badge.sm { font-size: 11px; }
  .badge.md { font-size: 13px; }
  .badge.lg { font-size: 15px; }
  .badge.status-pass {
    background: rgba(0,255,159,0.06);
    border-color: rgba(0,255,159,0.15);
  }
  .badge.status-fail {
    background: rgba(255,45,111,0.06);
    border-color: rgba(255,45,111,0.15);
  }
  .badge.status-running {
    background: rgba(255,184,0,0.06);
    border-color: rgba(255,184,0,0.15);
  }
  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: var(--badge-glow);
  }
</style>
