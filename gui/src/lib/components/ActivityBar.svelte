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
    background: var(--bg-surface);
    border-right: 1px solid var(--border);
    z-index: 10;
  }

  .ab-top {
    border-bottom: 1px solid var(--border);
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
    transition: color 0.15s;
  }
  .ab-icon:hover {
    color: var(--text-primary);
    opacity: 1;
  }
  .ab-icon.active {
    color: var(--neon-cyan);
    border-left-color: var(--neon-cyan);
    filter: drop-shadow(0 0 4px rgba(0, 229, 255, 0.5));
  }

  .ab-bottom {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 8px 0;
    border-top: 1px solid var(--border);
  }
</style>
