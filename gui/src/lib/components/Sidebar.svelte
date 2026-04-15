<script>
  import { icon } from '../icons.js';
  import { location } from 'svelte-spa-router';

  let expanded = $state(false);

  const navItems = [
    { path: '/', icon: 'workflow', label: 'Workflows' },
    { path: '/runs', icon: 'terminal', label: 'Runs' },
    { path: '/results', icon: 'folder', label: 'Results' },
  ];

  function isActive(itemPath, currentPath) {
    if (itemPath === '/') return currentPath === '/' || currentPath === '';
    return currentPath.startsWith(itemPath);
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<aside
  class="sidebar"
  class:expanded
  onmouseenter={() => expanded = true}
  onmouseleave={() => expanded = false}
>
  <div class="logo">
    <a href="#/">
      <span class="logo-short">1</span><span class="logo-full"><span class="logo-text">gl</span><span class="logo-accent">1</span><span class="logo-text">tch</span></span>
    </a>
  </div>

  <nav class="nav-items">
    {#each navItems as item}
      <a
        href="#{item.path}"
        class="nav-item"
        class:active={isActive(item.path, $location)}
      >
        <span class="nav-icon">{@html icon(item.icon)}</span>
        <span class="nav-label">{item.label}</span>
      </a>
    {/each}
  </nav>

  <div class="sidebar-footer">
    <span class="nav-icon">{@html icon('settings')}</span>
    <span class="nav-label text-muted">Settings</span>
  </div>
</aside>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    background: var(--bg-surface);
    border-right: 1px solid var(--border);
    width: var(--sidebar-width-collapsed);
    overflow: hidden;
    transition: width 0.2s ease;
    z-index: 10;
  }
  .sidebar.expanded { width: var(--sidebar-width-expanded); }

  .logo {
    padding: 0 16px;
    height: 52px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-bottom: 1px solid var(--border);
    overflow: hidden;
  }
  .logo a {
    font-family: var(--font-mono);
    font-size: 16px;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
    display: flex;
    align-items: center;
  }
  .logo-short {
    color: var(--neon-cyan);
    font-size: 22px;
    font-weight: 700;
  }
  .logo-full { display: none; }
  .sidebar.expanded .logo-short { display: none; }
  .sidebar.expanded .logo-full { display: inline; }
  .logo-text { color: var(--text-primary); }
  .logo-accent { color: var(--neon-cyan); }
  .logo a:hover .logo-text { color: var(--neon-cyan); }

  .nav-items {
    flex: 1;
    padding: 8px 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 16px;
    color: var(--text-muted);
    text-decoration: none;
    white-space: nowrap;
    transition: all 0.15s;
    border-left: 3px solid transparent;
  }
  .nav-item:hover {
    color: var(--text-primary);
    background: var(--bg-elevated);
  }
  .nav-item.active {
    color: var(--neon-cyan);
    border-left-color: var(--neon-cyan);
  }
  .nav-item.active .nav-icon { filter: drop-shadow(0 0 4px rgba(0, 229, 255, 0.5)); }

  .nav-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    flex-shrink: 0;
  }

  .nav-label {
    font-size: 13px;
    opacity: 0;
    transition: opacity 0.15s;
  }
  .sidebar.expanded .nav-label { opacity: 1; }

  .sidebar-footer {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-top: 1px solid var(--border);
    white-space: nowrap;
  }
</style>
