<script>
  import { icon } from '../icons.js';
  import { location } from 'svelte-spa-router';
  import WorkspaceSwitcher from './WorkspaceSwitcher.svelte';

  function isActive(itemPath, currentPath) {
    if (itemPath === '/') return currentPath === '/' || currentPath === '';
    return currentPath.startsWith(itemPath);
  }
</script>

<aside class="activity-bar">
  <div class="ab-top">
    <WorkspaceSwitcher />
  </div>

  <nav class="ab-nav">
    <a
      href="#/"
      class="ab-icon"
      class:active={isActive('/', $location)}
      title="Workflows"
    >
      {@html icon('workflow')}
    </a>
  </nav>

  <div class="ab-bottom">
    <a
      href="#/settings"
      class="ab-icon"
      class:active={isActive('/settings', $location)}
      title="Settings"
    >
      {@html icon('settings')}
    </a>
  </div>
</aside>

<style>
  .activity-bar {
    display: flex;
    flex-direction: column;
    width: 48px;
    background: linear-gradient(180deg, rgba(17,24,32,1), rgba(10,14,20,1));
    border-right: 1px solid rgba(0,229,255,0.06);
    z-index: 10;
  }

  .ab-top {
    border-bottom: 1px solid rgba(30,42,58,0.5);
  }

  .ab-nav {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 8px 0;
    gap: 4px;
  }

  .ab-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 40px;
    color: var(--text-muted);
    text-decoration: none;
    border-left: 3px solid transparent;
    transition: color 0.15s, background 0.15s;
  }
  .ab-icon:hover {
    color: var(--text-primary);
    opacity: 1;
  }
  .ab-icon.active {
    color: var(--neon-cyan);
    border-left: 3px solid var(--neon-cyan);
    background: rgba(0,229,255,0.04);
  }

  .ab-bottom {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 8px 0;
    border-top: 1px solid rgba(30,42,58,0.5);
  }
</style>
