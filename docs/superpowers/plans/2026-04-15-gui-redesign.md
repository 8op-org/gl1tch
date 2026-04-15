# GUI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the gl1tch GUI from bare-functional to cyberpunk dev tool with sidebar nav, card grid, full file browser, and dynamic workflow action system.

**Architecture:** Custom CSS design system (no libraries), evolve existing Svelte 5 components, add `(action ...)` sexpr keyword for dynamic GUI actions, 2 new backend endpoints.

**Tech Stack:** Svelte 5 (compat mode), Vite 6, CodeMirror 6, marked + highlight.js, Go net/http, Playwright

**Spec:** `docs/superpowers/specs/2026-04-15-gui-redesign-design.md`

---

## File Structure

### New files
- `gui/src/lib/components/StatusBadge.svelte` — reusable status dot + text
- `gui/src/lib/components/Modal.svelte` — overlay + backdrop blur + cyan border
- `gui/src/lib/components/Breadcrumb.svelte` — clickable path segments
- `gui/src/lib/components/FilterBar.svelte` — search + tag pills + sort
- `gui/src/lib/components/SplitPane.svelte` — resizable two-panel layout
- `gui/src/lib/components/FileTree.svelte` — recursive directory tree
- `gui/src/lib/components/ActionBar.svelte` — dynamic workflow action strip
- `gui/src/lib/components/Sidebar.svelte` — collapsible sidebar nav
- `gui/src/lib/icons.js` — inline SVG icon functions

### Modified files
- `gui/src/app.css` — complete replacement with cyberpunk design system
- `gui/src/App.svelte` — replace top nav with sidebar layout shell
- `gui/src/main.js` — add Inter + JetBrains Mono font imports
- `gui/src/lib/api.js` — add `getWorkflowActions(context)` function
- `gui/src/routes/WorkflowList.svelte` — card grid with filtering
- `gui/src/routes/Editor.svelte` — add metadata panel, restyle
- `gui/src/routes/RunDialog.svelte` — restyle with new design system
- `gui/src/routes/RunList.svelte` — restyle table with neon status dots
- `gui/src/routes/RunView.svelte` — restyle with steps, output, telemetry sections
- `gui/src/routes/ResultsBrowser.svelte` — full file browser with edit mode + action bar
- `gui/src/lib/markdown.js` — import highlight.js CSS
- `gui/e2e/gui.spec.js` — add tests for new features
- `internal/pipeline/sexpr.go` — add `action` keyword parsing
- `internal/pipeline/pipeline.go` or workflow struct file — add Action field
- `internal/gui/api_workflows.go` — add actions to response, add actions endpoint
- `internal/gui/server.go` — register new routes

---

## Task 1: CSS Design System

**Files:**
- Replace: `gui/src/app.css`
- Modify: `gui/src/main.js`

This is the foundation. Every other task depends on these CSS custom properties existing.

- [ ] **Step 1: Replace app.css with cyberpunk design system**

Replace the entire contents of `gui/src/app.css` with:

```css
/* gl1tch GUI — Cyberpunk Dev Tool Design System */

/* --- Fonts --- */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap');

/* --- Tokens --- */
:root {
  --bg-deep: #0a0e14;
  --bg-surface: #111820;
  --bg-elevated: #1a2230;
  --neon-cyan: #00e5ff;
  --neon-magenta: #ff2d6f;
  --neon-amber: #ffb800;
  --neon-green: #00ff9f;
  --text-primary: #e0e6ed;
  --text-muted: #5a6a7a;
  --border: #1e2a3a;
  --font-sans: 'Inter', system-ui, -apple-system, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', monospace;
  --glow-cyan: 0 0 8px rgba(0, 229, 255, 0.4);
  --glow-green: 0 0 8px rgba(0, 255, 159, 0.4);
  --glow-magenta: 0 0 8px rgba(255, 45, 111, 0.4);
  --glow-amber: 0 0 8px rgba(255, 184, 0, 0.4);
  --sidebar-width-collapsed: 56px;
  --sidebar-width-expanded: 200px;
}

/* --- Reset & Base --- */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

html, body {
  height: 100%;
  background: var(--bg-deep);
  color: var(--text-primary);
  font-family: var(--font-sans);
  font-size: 13px;
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

#app {
  display: grid;
  grid-template-columns: var(--sidebar-width-collapsed) 1fr;
  height: 100vh;
  overflow: hidden;
}

/* --- Typography --- */
h1 { font-size: 20px; font-weight: 600; }
h2 { font-size: 16px; font-weight: 500; }
h3 { font-size: 14px; font-weight: 500; }
code, pre, .mono { font-family: var(--font-mono); font-size: 12px; }

a {
  color: var(--neon-cyan);
  text-decoration: none;
  transition: opacity 0.15s;
}
a:hover { opacity: 0.8; }

/* --- Surfaces --- */
.surface {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 6px;
}

.elevated {
  background: var(--bg-elevated);
}

/* --- Buttons --- */
button {
  font-family: var(--font-sans);
  font-size: 13px;
  padding: 6px 14px;
  border-radius: 4px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.15s;
}
button:hover {
  background: var(--bg-elevated);
  border-color: var(--text-muted);
}
button.primary {
  background: transparent;
  border-color: var(--neon-cyan);
  color: var(--neon-cyan);
}
button.primary:hover {
  background: rgba(0, 229, 255, 0.1);
  box-shadow: var(--glow-cyan);
}
button.danger {
  border-color: var(--neon-magenta);
  color: var(--neon-magenta);
}
button.danger:hover {
  background: rgba(255, 45, 111, 0.1);
  box-shadow: var(--glow-magenta);
}
button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* --- Inputs --- */
input, select, textarea {
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 6px 10px;
  border-radius: 4px;
  border: 1px solid var(--border);
  background: var(--bg-deep);
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}
input:focus, select:focus, textarea:focus {
  border-color: var(--neon-cyan);
  box-shadow: var(--glow-cyan);
}

/* --- Tables --- */
table { width: 100%; border-collapse: collapse; }
th {
  text-align: left;
  font-size: 11px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border);
}
td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border);
}
tr:hover td {
  background: var(--bg-elevated);
}

/* --- Scrollbar --- */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover { background: var(--text-muted); }

/* --- Tags / Pills --- */
.pill {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 3px;
  background: var(--bg-elevated);
  color: var(--text-muted);
  border: 1px solid var(--border);
}
.pill.active {
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
  background: rgba(0, 229, 255, 0.08);
}

/* --- Page Layout --- */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
  min-height: 52px;
}
.page-header h1, .page-header h2 { margin: 0; }
.page-content {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}
.main-area {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

/* --- Status Colors --- */
.status-pass { color: var(--neon-green); }
.status-fail { color: var(--neon-magenta); }
.status-running { color: var(--neon-amber); }

/* --- Code Blocks (highlight.js) --- */
pre code.hljs {
  background: var(--bg-deep);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 12px 16px;
  font-size: 12px;
  overflow-x: auto;
}

/* --- Animations --- */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
.pulse { animation: pulse 2s ease-in-out infinite; }

@keyframes glow-fade {
  0% { box-shadow: var(--glow-green); }
  100% { box-shadow: none; }
}

/* --- Overlay --- */
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

/* --- Utility --- */
.flex { display: flex; }
.flex-col { flex-direction: column; }
.gap-sm { gap: 8px; }
.gap-md { gap: 16px; }
.gap-lg { gap: 24px; }
.items-center { align-items: center; }
.justify-between { justify-content: space-between; }
.text-muted { color: var(--text-muted); }
.text-cyan { color: var(--neon-cyan); }
.truncate { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
```

- [ ] **Step 2: Verify build still works**

Run: `cd gui && npm run build`
Expected: Build succeeds (CSS is just replaced, no Svelte changes yet)

- [ ] **Step 3: Commit**

```bash
git add gui/src/app.css
git commit -m "feat(gui): replace CSS with cyberpunk design system"
```

---

## Task 2: SVG Icons Module

**Files:**
- Create: `gui/src/lib/icons.js`

Small utility — inline SVG strings for the ~8 icons we need. No dependencies.

- [ ] **Step 1: Create icons module**

Create `gui/src/lib/icons.js`:

```javascript
// Inline SVG icons — no library dependency
// Each returns an SVG string. Size controlled by parent via width/height or font-size.

export const icons = {
  workflow: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>`,

  play: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>`,

  folder: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2z"/></svg>`,

  folderOpen: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 14 1.5-2.9A2 2 0 0 1 9.24 10H20a2 2 0 0 1 1.94 2.5l-1.54 6a2 2 0 0 1-1.95 1.5H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.93a2 2 0 0 1 1.66.9l.82 1.2a2 2 0 0 0 1.66.9H18a2 2 0 0 1 2 2v2"/></svg>`,

  file: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/></svg>`,

  settings: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`,

  search: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>`,

  chevronRight: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6"/></svg>`,

  chevronDown: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>`,

  terminal: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"/><line x1="12" x2="20" y1="19" y2="19"/></svg>`,

  edit: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/></svg>`,

  eye: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>`,

  save: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>`,

  zap: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>`,
};

/** Render an icon inline. Usage: {@html icon('workflow')} */
export function icon(name, size) {
  const svg = icons[name];
  if (!svg) return '';
  if (size) {
    return svg.replace(/width="\d+"/, `width="${size}"`).replace(/height="\d+"/, `height="${size}"`);
  }
  return svg;
}
```

- [ ] **Step 2: Commit**

```bash
git add gui/src/lib/icons.js
git commit -m "feat(gui): add inline SVG icons module"
```

---

## Task 3: Shared Components

**Files:**
- Create: `gui/src/lib/components/StatusBadge.svelte`
- Create: `gui/src/lib/components/Modal.svelte`
- Create: `gui/src/lib/components/Breadcrumb.svelte`
- Create: `gui/src/lib/components/Sidebar.svelte`

These are reusable building blocks used by multiple route components.

- [ ] **Step 1: Create StatusBadge component**

Create `gui/src/lib/components/StatusBadge.svelte`:

```svelte
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
```

- [ ] **Step 2: Create Modal component**

Create `gui/src/lib/components/Modal.svelte`:

```svelte
<script>
  let { title = '', onclose } = $props();

  function handleKeydown(e) {
    if (e.key === 'Escape' && onclose) onclose();
  }

  function handleBackdrop(e) {
    if (e.target === e.currentTarget && onclose) onclose();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div class="overlay" on:click={handleBackdrop}>
  <div class="modal surface">
    <div class="modal-header">
      <h2>{title}</h2>
      {#if onclose}
        <button class="close-btn" on:click={onclose}>&times;</button>
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
```

- [ ] **Step 3: Create Breadcrumb component**

Create `gui/src/lib/components/Breadcrumb.svelte`:

```svelte
<script>
  let { segments = [], onnavigate } = $props();
  // segments: [{label: "Workflows", href: "#/"}, {label: "cross-review.glitch"}]
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
```

- [ ] **Step 4: Create Sidebar component**

Create `gui/src/lib/components/Sidebar.svelte`:

```svelte
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
  on:mouseenter={() => expanded = true}
  on:mouseleave={() => expanded = false}
>
  <div class="logo">
    <a href="#/">
      <span class="logo-text">gl</span><span class="logo-accent">1</span><span class="logo-text">tch</span>
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
  .sidebar.expanded {
    width: var(--sidebar-width-expanded);
  }

  .logo {
    padding: 16px;
    height: 52px;
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--border);
  }
  .logo a {
    font-family: var(--font-mono);
    font-size: 16px;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
  }
  .logo-text { color: var(--text-primary); }
  .logo-accent { color: var(--neon-cyan); }
  .logo a:hover .logo-text { color: var(--neon-cyan); }
  .logo a:hover { box-shadow: var(--glow-cyan); }

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
```

- [ ] **Step 5: Verify build**

Run: `cd gui && npm run build`
Expected: Build succeeds

- [ ] **Step 6: Commit**

```bash
git add gui/src/lib/components/
git commit -m "feat(gui): add shared components — StatusBadge, Modal, Breadcrumb, Sidebar"
```

---

## Task 4: Layout Shell (App.svelte)

**Files:**
- Modify: `gui/src/App.svelte`

Replace the top nav with sidebar layout. The `#app` grid is already defined in app.css.

- [ ] **Step 1: Rewrite App.svelte**

Replace the entire contents of `gui/src/App.svelte` with:

```svelte
<script>
  import Router from 'svelte-spa-router';
  import Sidebar from './lib/components/Sidebar.svelte';
  import WorkflowList from './routes/WorkflowList.svelte';
  import Editor from './routes/Editor.svelte';
  import RunList from './routes/RunList.svelte';
  import RunView from './routes/RunView.svelte';
  import ResultsBrowser from './routes/ResultsBrowser.svelte';

  const routes = {
    '/': WorkflowList,
    '/workflow/:name': Editor,
    '/runs': RunList,
    '/run/:id': RunView,
    '/results': ResultsBrowser,
  };
</script>

<Sidebar />
<main class="main-area">
  <Router {routes} />
</main>
```

- [ ] **Step 2: Verify build**

Run: `cd gui && npm run build`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add gui/src/App.svelte
git commit -m "feat(gui): replace top nav with sidebar layout shell"
```

---

## Task 5: Workflow List (Card Grid)

**Files:**
- Create: `gui/src/lib/components/FilterBar.svelte`
- Modify: `gui/src/routes/WorkflowList.svelte`

- [ ] **Step 1: Create FilterBar component**

Create `gui/src/lib/components/FilterBar.svelte`:

```svelte
<script>
  import { icon } from '../icons.js';

  let { tags = [], activeTags = $bindable([]), searchQuery = $bindable(''), sortBy = $bindable('name'), sortOptions = [] } = $props();

  function toggleTag(tag) {
    if (activeTags.includes(tag)) {
      activeTags = activeTags.filter(t => t !== tag);
    } else {
      activeTags = [...activeTags, tag];
    }
  }
</script>

<div class="filter-bar">
  <div class="search-wrap">
    <span class="search-icon">{@html icon('search', 16)}</span>
    <input
      type="text"
      placeholder="Search..."
      bind:value={searchQuery}
    />
  </div>

  {#if tags.length > 0}
    <div class="tags">
      {#each tags as tag}
        <button
          class="pill"
          class:active={activeTags.includes(tag)}
          on:click={() => toggleTag(tag)}
        >{tag}</button>
      {/each}
    </div>
  {/if}

  {#if sortOptions.length > 0}
    <select bind:value={sortBy}>
      {#each sortOptions as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>
  {/if}
</div>

<style>
  .filter-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }
  .search-wrap {
    position: relative;
    flex: 1;
    min-width: 200px;
    max-width: 360px;
  }
  .search-icon {
    position: absolute;
    left: 10px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-muted);
    display: flex;
  }
  .search-wrap input {
    width: 100%;
    padding-left: 32px;
  }
  .tags { display: flex; gap: 6px; flex-wrap: wrap; }
  select { min-width: 140px; }
</style>
```

- [ ] **Step 2: Rewrite WorkflowList.svelte as card grid**

Replace the entire contents of `gui/src/routes/WorkflowList.svelte` with:

```svelte
<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { listWorkflows } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import FilterBar from '../lib/components/FilterBar.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';

  let workflows = $state([]);
  let error = $state(null);
  let loading = $state(true);
  let searchQuery = $state('');
  let activeTags = $state([]);
  let sortBy = $state('name');

  const allTags = $derived([...new Set(workflows.flatMap(w => w.tags || []))].sort());

  const filtered = $derived.by(() => {
    let list = workflows;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      list = list.filter(w => w.name.toLowerCase().includes(q) || (w.description || '').toLowerCase().includes(q));
    }
    if (activeTags.length > 0) {
      list = list.filter(w => activeTags.some(t => (w.tags || []).includes(t)));
    }
    if (sortBy === 'name') list = [...list].sort((a, b) => a.name.localeCompare(b.name));
    return list;
  });

  const sortOptions = [
    { value: 'name', label: 'Name' },
  ];

  onMount(async () => {
    try {
      workflows = await listWorkflows();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<div class="page-header">
  <h1>Workflows</h1>
</div>
<div class="page-content">
  {#if loading}
    <p class="text-muted">Loading workflows...</p>
  {:else if error}
    <p class="status-fail">{error}</p>
  {:else}
    <FilterBar
      tags={allTags}
      bind:activeTags
      bind:searchQuery
      bind:sortBy
      {sortOptions}
    />

    <div class="card-grid">
      {#each filtered as wf}
        <button class="card surface" on:click={() => push(`/workflow/${wf.name}`)}>
          <div class="card-name mono">{@html icon('zap', 14)} {wf.name}</div>
          {#if wf.description}
            <p class="card-desc text-muted">{wf.description}</p>
          {/if}
          {#if wf.tags?.length}
            <div class="card-tags">
              {#each wf.tags as tag}<span class="pill">{tag}</span>{/each}
              {#if wf.version}<span class="pill">v{wf.version}</span>{/if}
            </div>
          {/if}
          <div class="card-footer text-muted">
            {#if wf.last_run_status}
              <StatusBadge status={wf.last_run_status} />
            {:else}
              <span>Never run</span>
            {/if}
            {#if wf.author}<span>@{wf.author}</span>{/if}
          </div>
        </button>
      {/each}
    </div>

    {#if filtered.length === 0}
      <p class="text-muted" style="margin-top:24px">No workflows match your filters.</p>
    {/if}
  {/if}
</div>

<style>
  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
    margin-top: 20px;
  }
  .card {
    text-align: left;
    padding: 16px;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    gap: 10px;
    transition: border-color 0.15s;
  }
  .card:hover {
    border-color: rgba(0, 229, 255, 0.3);
  }
  .card-name {
    font-family: var(--font-mono);
    font-size: 14px;
    color: var(--neon-cyan);
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .card-desc {
    font-size: 12px;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .card-tags { display: flex; gap: 6px; flex-wrap: wrap; }
  .card-footer {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 12px;
    margin-top: auto;
    padding-top: 8px;
    border-top: 1px solid var(--border);
  }
</style>
```

- [ ] **Step 3: Verify build**

Run: `cd gui && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add gui/src/lib/components/FilterBar.svelte gui/src/routes/WorkflowList.svelte
git commit -m "feat(gui): workflow list as card grid with search and tag filtering"
```

---

## Task 6: Workflow Editor Redesign

**Files:**
- Modify: `gui/src/routes/Editor.svelte`
- Modify: `gui/src/routes/RunDialog.svelte`

- [ ] **Step 1: Rewrite Editor.svelte with metadata panel**

Replace the entire contents of `gui/src/routes/Editor.svelte` with:

```svelte
<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { EditorView, basicSetup } from 'codemirror';
  import { EditorState } from '@codemirror/state';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { getWorkflow, saveWorkflow, listRuns } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';
  import RunDialog from './RunDialog.svelte';

  let { params } = $props();
  let name = $derived(params?.name || '');

  let source = $state('');
  let workflowParams = $state([]);
  let metadata = $state({});
  let recentRuns = $state([]);
  let dirty = $state(false);
  let saveStatus = $state('');
  let showRunDialog = $state(false);
  let showMeta = $state(true);
  let editorEl = $state(null);
  let editorView = $state(null);
  let error = $state(null);

  const breadcrumbs = $derived([
    { label: 'Workflows', href: '#/' },
    { label: name },
  ]);

  async function load() {
    try {
      const data = await getWorkflow(name);
      source = data.source || '';
      workflowParams = data.params || [];
      metadata = { tags: data.tags, author: data.author, version: data.version, created: data.created };
      initEditor(source);

      try {
        const runs = await listRuns();
        recentRuns = runs.filter(r => r.workflow === name).slice(0, 5);
      } catch (_) { /* runs optional */ }
    } catch (e) {
      error = e.message;
    }
  }

  function initEditor(content) {
    if (editorView) editorView.destroy();
    editorView = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          basicSetup,
          oneDark,
          EditorView.updateListener.of(update => {
            if (update.docChanged) dirty = true;
          }),
        ],
      }),
      parent: editorEl,
    });
  }

  async function handleSave() {
    if (!editorView) return;
    const content = editorView.state.doc.toString();
    saveStatus = 'saving';
    try {
      await saveWorkflow(name, content);
      dirty = false;
      saveStatus = 'saved';
      setTimeout(() => { saveStatus = ''; }, 2000);
    } catch (e) {
      saveStatus = 'error';
    }
  }

  function handleKeydown(e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 's') {
      e.preventDefault();
      handleSave();
    }
  }

  onMount(() => { load(); });
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="page-header">
  <Breadcrumb segments={breadcrumbs} onnavigate={(href) => push(href.replace('#', ''))} />
  <div class="flex items-center gap-sm">
    {#if saveStatus === 'saved'}
      <span class="status-pass" style="font-size:12px">Saved</span>
    {:else if saveStatus === 'saving'}
      <span class="text-muted" style="font-size:12px">Saving...</span>
    {/if}
    <button class:primary={dirty} disabled={!dirty} on:click={handleSave}>
      {@html icon('save', 14)} Save
    </button>
    <button class="primary" on:click={() => showRunDialog = true}>
      {@html icon('play', 14)} Run
    </button>
  </div>
</div>

{#if error}
  <div class="page-content"><p class="status-fail">{error}</p></div>
{:else}
  <div class="editor-layout">
    <div class="editor-pane" bind:this={editorEl}></div>

    {#if showMeta}
      <aside class="meta-panel">
        <div class="meta-header">
          <h3>Metadata</h3>
          <button class="close-btn" on:click={() => showMeta = false}>&times;</button>
        </div>
        <div class="meta-body">
          {#if metadata.tags?.length}
            <div class="meta-section">
              <span class="meta-label">Tags</span>
              <div class="flex gap-sm" style="flex-wrap:wrap">{#each metadata.tags as tag}<span class="pill">{tag}</span>{/each}</div>
            </div>
          {/if}
          {#if metadata.author}
            <div class="meta-section">
              <span class="meta-label">Author</span>
              <span class="mono">@{metadata.author}</span>
            </div>
          {/if}
          {#if metadata.version}
            <div class="meta-section">
              <span class="meta-label">Version</span>
              <span class="mono">{metadata.version}</span>
            </div>
          {/if}

          {#if recentRuns.length > 0}
            <div class="meta-section">
              <span class="meta-label">Recent Runs</span>
              <div class="run-list">
                {#each recentRuns as run}
                  <a href="#/run/{run.id}" class="run-item">
                    <StatusBadge status={run.status} />
                    <span class="text-muted" style="font-size:11px">{run.started ? new Date(run.started).toLocaleString() : ''}</span>
                  </a>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </aside>
    {:else}
      <button class="meta-toggle" on:click={() => showMeta = true} title="Show metadata">
        {@html icon('chevronRight')}
      </button>
    {/if}
  </div>
{/if}

{#if showRunDialog}
  <RunDialog {name} params={workflowParams} onclose={() => showRunDialog = false} />
{/if}

<style>
  .editor-layout {
    display: flex;
    flex: 1;
    overflow: hidden;
  }
  .editor-pane {
    flex: 1;
    overflow: auto;
  }
  .editor-pane :global(.cm-editor) {
    height: 100%;
  }
  .editor-pane :global(.cm-editor .cm-scroller) {
    font-family: var(--font-mono);
    font-size: 13px;
  }

  .meta-panel {
    width: 250px;
    border-left: 1px solid var(--border);
    background: var(--bg-surface);
    overflow-y: auto;
    flex-shrink: 0;
  }
  .meta-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
  }
  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 18px;
    cursor: pointer;
    padding: 0;
  }
  .meta-body { padding: 16px; }
  .meta-section {
    margin-bottom: 16px;
  }
  .meta-label {
    display: block;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    margin-bottom: 6px;
  }
  .run-list { display: flex; flex-direction: column; gap: 6px; }
  .run-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 0;
    text-decoration: none;
  }
  .run-item:hover { background: var(--bg-elevated); border-radius: 4px; padding: 4px 6px; margin: 0 -6px; }

  .meta-toggle {
    position: absolute;
    right: 0;
    top: 50%;
    transform: translateY(-50%);
    background: var(--bg-surface);
    border: 1px solid var(--border);
    border-right: none;
    border-radius: 4px 0 0 4px;
    padding: 8px 4px;
    cursor: pointer;
    color: var(--text-muted);
  }
</style>
```

- [ ] **Step 2: Rewrite RunDialog.svelte with new design**

Replace the entire contents of `gui/src/routes/RunDialog.svelte` with:

```svelte
<script>
  import { push } from 'svelte-spa-router';
  import { runWorkflow } from '../lib/api.js';
  import { icon } from '../lib/icons.js';
  import Modal from '../lib/components/Modal.svelte';

  let { name, params = [], onclose } = $props();
  let values = $state({});
  let running = $state(false);
  let error = $state(null);

  async function handleSubmit() {
    running = true;
    error = null;
    try {
      const result = await runWorkflow(name, values);
      onclose?.();
      if (result.run_id) push(`/run/${result.run_id}`);
    } catch (e) {
      error = e.message;
      running = false;
    }
  }
</script>

<Modal title="Run {name}" {onclose}>
  {#if params.length > 0}
    <form on:submit|preventDefault={handleSubmit} class="flex flex-col gap-md">
      {#each params as param}
        <label class="field">
          <span class="field-label">{param}</span>
          <input type="text" bind:value={values[param]} placeholder={param} />
        </label>
      {/each}

      {#if error}
        <p class="status-fail" style="font-size:12px">{error}</p>
      {/if}

      <div class="flex justify-between" style="margin-top:8px">
        <button type="button" on:click={onclose}>Cancel</button>
        <button type="submit" class="primary" disabled={running}>
          {#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}
        </button>
      </div>
    </form>
  {:else}
    <div class="flex flex-col gap-md">
      <p class="text-muted">No parameters required.</p>
      {#if error}
        <p class="status-fail" style="font-size:12px">{error}</p>
      {/if}
      <div class="flex justify-between">
        <button on:click={onclose}>Cancel</button>
        <button class="primary" disabled={running} on:click={handleSubmit}>
          {#if running}Running...{:else}{@html icon('play', 14)} Start Run{/if}
        </button>
      </div>
    </div>
  {/if}
</Modal>

<style>
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field-label {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--text-muted);
  }
</style>
```

- [ ] **Step 3: Verify build**

Run: `cd gui && npm run build`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add gui/src/routes/Editor.svelte gui/src/routes/RunDialog.svelte
git commit -m "feat(gui): editor with metadata panel and restyled run dialog"
```

---

## Task 7: Runs List & Run Detail Redesign

**Files:**
- Modify: `gui/src/routes/RunList.svelte`
- Modify: `gui/src/routes/RunView.svelte`
- Modify: `gui/src/lib/markdown.js`

- [ ] **Step 1: Fix highlight.js CSS import in markdown.js**

Add a highlight.js theme import at the top of `gui/src/lib/markdown.js`. Add this as the first import:

```javascript
import 'highlight.js/styles/github-dark.min.css';
```

- [ ] **Step 2: Rewrite RunList.svelte**

Replace the entire contents of `gui/src/routes/RunList.svelte` with:

```svelte
<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { listRuns } from '../lib/api.js';
  import StatusBadge from '../lib/components/StatusBadge.svelte';

  let runs = $state([]);
  let error = $state(null);
  let loading = $state(true);
  let filterStatus = $state('all');

  const filtered = $derived(
    filterStatus === 'all' ? runs : runs.filter(r => r.status?.toUpperCase() === filterStatus)
  );

  function duration(started, finished) {
    if (!started) return '--';
    const start = new Date(started);
    const end = finished ? new Date(finished) : new Date();
    const sec = Math.round((end - start) / 1000);
    if (sec < 60) return `${sec}s`;
    const min = Math.floor(sec / 60);
    const rem = sec % 60;
    return `${min}m ${rem}s`;
  }

  function relativeTime(ts) {
    if (!ts) return '';
    const diff = Date.now() - new Date(ts).getTime();
    const min = Math.round(diff / 60000);
    if (min < 1) return 'just now';
    if (min < 60) return `${min}m ago`;
    const hr = Math.round(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const days = Math.round(hr / 24);
    return `${days}d ago`;
  }

  onMount(async () => {
    try {
      runs = await listRuns();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<div class="page-header">
  <h1>Runs</h1>
  <div class="flex gap-sm">
    <select bind:value={filterStatus}>
      <option value="all">All</option>
      <option value="PASS">Pass</option>
      <option value="FAIL">Fail</option>
      <option value="RUNNING">Running</option>
    </select>
  </div>
</div>
<div class="page-content">
  {#if loading}
    <p class="text-muted">Loading runs...</p>
  {:else if error}
    <p class="status-fail">{error}</p>
  {:else if filtered.length === 0}
    <p class="text-muted">No runs found.</p>
  {:else}
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Workflow</th>
          <th>Status</th>
          <th>Duration</th>
          <th>Started</th>
        </tr>
      </thead>
      <tbody>
        {#each filtered as run}
          <tr class="clickable" on:click={() => push(`/run/${run.id}`)}>
            <td class="mono text-cyan">#{run.id}</td>
            <td class="mono">{run.workflow || run.name || ''}</td>
            <td><StatusBadge status={run.status} /></td>
            <td class="mono">{duration(run.started, run.finished)}</td>
            <td class="text-muted">{relativeTime(run.started)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .clickable { cursor: pointer; }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .text-cyan { color: var(--neon-cyan); }
</style>
```

- [ ] **Step 3: Rewrite RunView.svelte**

Replace the entire contents of `gui/src/routes/RunView.svelte` with:

```svelte
<script>
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { getRun, getKibanaRun } from '../lib/api.js';
  import { renderMarkdown } from '../lib/markdown.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import StatusBadge from '../lib/components/StatusBadge.svelte';

  let { params } = $props();
  let id = $derived(params?.id || '');

  let run = $state(null);
  let kibanaUrl = $state(null);
  let error = $state(null);
  let showTelemetry = $state(false);

  const breadcrumbs = $derived([
    { label: 'Runs', href: '#/runs' },
    { label: `#${id}${run?.workflow ? ' ' + run.workflow : ''}` },
  ]);

  function duration(started, finished) {
    if (!started) return '--';
    const start = new Date(started);
    const end = finished ? new Date(finished) : new Date();
    const sec = Math.round((end - start) / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  onMount(async () => {
    try {
      run = await getRun(id);
      try { kibanaUrl = (await getKibanaRun(id)).url; } catch (_) {}
    } catch (e) {
      error = e.message;
    }
  });
</script>

<div class="page-header">
  <Breadcrumb segments={breadcrumbs} onnavigate={(href) => push(href.replace('#', ''))} />
  <div class="flex items-center gap-sm">
    {#if run}
      <StatusBadge status={run.status} size="md" />
    {/if}
  </div>
</div>

<div class="page-content">
  {#if error}
    <p class="status-fail">{error}</p>
  {:else if !run}
    <p class="text-muted">Loading...</p>
  {:else}
    <div class="meta-row">
      <div class="meta-item">
        <span class="meta-label">Started</span>
        <span class="mono">{run.started ? new Date(run.started).toLocaleString() : '--'}</span>
      </div>
      <div class="meta-item">
        <span class="meta-label">Duration</span>
        <span class="mono">{duration(run.started, run.finished)}</span>
      </div>
      {#if run.model}
        <div class="meta-item">
          <span class="meta-label">Model</span>
          <span class="mono">{run.model}</span>
        </div>
      {/if}
      {#if run.tokens}
        <div class="meta-item">
          <span class="meta-label">Tokens</span>
          <span class="mono">{Number(run.tokens).toLocaleString()}</span>
        </div>
      {/if}
    </div>

    {#if run.steps?.length}
      <section class="section">
        <h3>Steps</h3>
        <table>
          <thead>
            <tr><th>#</th><th>Step</th><th>Model</th><th>Duration</th><th>Status</th></tr>
          </thead>
          <tbody>
            {#each run.steps as step, i}
              <tr>
                <td class="mono text-muted">{i + 1}</td>
                <td class="mono">{step.step_id || step.name || ''}</td>
                <td class="mono text-muted">{step.model || ''}</td>
                <td class="mono">{duration(step.started, step.finished)}</td>
                <td><StatusBadge status={step.status || 'pass'} /></td>
              </tr>
            {/each}
          </tbody>
        </table>
      </section>
    {/if}

    {#if run.output}
      <section class="section">
        <h3>Output</h3>
        <div class="output-content">{@html renderMarkdown(run.output)}</div>
      </section>
    {/if}

    {#if kibanaUrl}
      <section class="section">
        <button class="section-toggle" on:click={() => showTelemetry = !showTelemetry}>
          <h3>{@html icon(showTelemetry ? 'chevronDown' : 'chevronRight')} Telemetry</h3>
        </button>
        {#if showTelemetry}
          <iframe src={kibanaUrl} title="Kibana telemetry" class="kibana-frame"></iframe>
        {/if}
      </section>
    {/if}
  {/if}
</div>

<style>
  .meta-row {
    display: flex;
    gap: 32px;
    padding: 16px 0;
    border-bottom: 1px solid var(--border);
    margin-bottom: 24px;
    flex-wrap: wrap;
  }
  .meta-item { display: flex; flex-direction: column; gap: 4px; }
  .meta-label {
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }
  .mono { font-family: var(--font-mono); font-size: 12px; }
  .section { margin-bottom: 24px; }
  .section h3 { margin-bottom: 12px; display: flex; align-items: center; gap: 6px; }
  .section-toggle {
    background: none;
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    padding: 0;
  }
  .output-content { line-height: 1.6; }
  .output-content :global(pre) { margin: 12px 0; }
  .kibana-frame {
    width: 100%;
    height: 400px;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg-deep);
  }
  .text-muted { color: var(--text-muted); }
</style>
```

- [ ] **Step 4: Verify build**

Run: `cd gui && npm run build`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add gui/src/routes/RunList.svelte gui/src/routes/RunView.svelte gui/src/lib/markdown.js
git commit -m "feat(gui): redesign runs list and run detail with neon status badges"
```

---

## Task 8: Results Browser (Full File Browser)

**Files:**
- Create: `gui/src/lib/components/SplitPane.svelte`
- Create: `gui/src/lib/components/FileTree.svelte`
- Modify: `gui/src/routes/ResultsBrowser.svelte`
- Modify: `gui/src/lib/api.js`

This is the largest task — a real file browser with preview/edit modes.

- [ ] **Step 1: Create SplitPane component**

Create `gui/src/lib/components/SplitPane.svelte`:

```svelte
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
  <div class="split-left" style="width:{width}px">
    <slot name="left" />
  </div>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="split-handle" on:mousedown={onMouseDown}></div>
  <div class="split-right">
    <slot name="right" />
  </div>
</div>

<style>
  .split-pane {
    display: flex;
    flex: 1;
    overflow: hidden;
  }
  .split-pane.dragging { cursor: col-resize; user-select: none; }
  .split-left {
    flex-shrink: 0;
    overflow-y: auto;
    border-right: 1px solid var(--border);
  }
  .split-handle {
    width: 4px;
    cursor: col-resize;
    background: transparent;
    transition: background 0.15s;
    flex-shrink: 0;
  }
  .split-handle:hover, .dragging .split-handle {
    background: var(--neon-cyan);
    opacity: 0.3;
  }
  .split-right {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }
</style>
```

- [ ] **Step 2: Create FileTree component**

Create `gui/src/lib/components/FileTree.svelte`:

```svelte
<script>
  import { icon } from '../icons.js';

  let { entries = [], selectedPath = '', onselect, depth = 0 } = $props();

  let expanded = $state({});

  function toggle(name) {
    expanded[name] = !expanded[name];
    expanded = expanded;
  }

  function fileIcon(name) {
    if (name.endsWith('.md')) return 'text-cyan';
    if (name.endsWith('.diff') || name.endsWith('.patch')) return 'status-pass';
    if (name.endsWith('.json')) return 'text-amber';
    if (name.endsWith('.yaml') || name.endsWith('.yml')) return 'text-magenta';
    if (name.endsWith('.glitch')) return 'text-cyan';
    return 'text-muted';
  }
</script>

<ul class="tree" style="--depth:{depth}">
  {#each entries as entry}
    {#if entry.isDir}
      <li>
        <button class="tree-item dir" on:click={() => toggle(entry.name)}>
          <span class="tree-indent" style="width:{depth * 16}px"></span>
          <span class="tree-icon">{@html icon(expanded[entry.name] ? 'chevronDown' : 'chevronRight')}</span>
          <span class="tree-icon text-cyan">{@html icon(expanded[entry.name] ? 'folderOpen' : 'folder', 16)}</span>
          <span class="tree-name">{entry.name}</span>
        </button>
        {#if expanded[entry.name] && entry.children}
          <svelte:self entries={entry.children} {selectedPath} {onselect} depth={depth + 1} />
        {/if}
      </li>
    {:else}
      <li>
        <button
          class="tree-item file"
          class:selected={selectedPath === entry.path}
          on:click={() => onselect?.(entry.path)}
        >
          <span class="tree-indent" style="width:{depth * 16}px"></span>
          <span class="tree-icon {fileIcon(entry.name)}">{@html icon('file', 16)}</span>
          <span class="tree-name">{entry.name}</span>
        </button>
      </li>
    {/if}
  {/each}
</ul>

<style>
  .tree { list-style: none; padding: 0; margin: 0; }
  .tree-item {
    display: flex;
    align-items: center;
    gap: 4px;
    width: 100%;
    text-align: left;
    background: none;
    border: none;
    border-left: 3px solid transparent;
    color: var(--text-primary);
    font-size: 12px;
    font-family: var(--font-mono);
    padding: 4px 8px;
    cursor: pointer;
    white-space: nowrap;
  }
  .tree-item:hover { background: var(--bg-elevated); }
  .tree-item.selected {
    background: var(--bg-elevated);
    border-left-color: var(--neon-cyan);
  }
  .tree-indent { flex-shrink: 0; }
  .tree-icon { display: flex; align-items: center; flex-shrink: 0; }
  .tree-name { overflow: hidden; text-overflow: ellipsis; }
  .text-cyan { color: var(--neon-cyan); }
  .text-amber { color: var(--neon-amber); }
  .text-magenta { color: var(--neon-magenta); }
</style>
```

- [ ] **Step 3: Add saveResult and getResultText to api.js**

Add these two functions at the end of `gui/src/lib/api.js` (before the closing of the file):

```javascript
export async function saveResult(path, content) {
  const res = await fetch(`/api/results/${path}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getResultText(path) {
  const res = await fetch(`/api/results/${path}`);
  if (!res.ok) throw new Error(await res.text());
  return res.text();
}
```

- [ ] **Step 4: Rewrite ResultsBrowser.svelte**

Replace the entire contents of `gui/src/routes/ResultsBrowser.svelte` with:

```svelte
<script>
  import { onMount } from 'svelte';
  import { EditorView, basicSetup } from 'codemirror';
  import { EditorState } from '@codemirror/state';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { getResult, getResultText, saveResult } from '../lib/api.js';
  import { renderMarkdown } from '../lib/markdown.js';
  import { icon } from '../lib/icons.js';
  import Breadcrumb from '../lib/components/Breadcrumb.svelte';
  import SplitPane from '../lib/components/SplitPane.svelte';
  import FileTree from '../lib/components/FileTree.svelte';

  let tree = $state([]);
  let selectedPath = $state('');
  let fileContent = $state('');
  let mode = $state('preview'); // 'preview' | 'edit'
  let dirty = $state(false);
  let loading = $state(true);
  let error = $state(null);
  let editorEl = $state(null);
  let editorView = $state(null);

  const breadcrumbSegments = $derived(() => {
    const parts = selectedPath.split('/').filter(Boolean);
    return [
      { label: 'Results', href: '#/results' },
      ...parts.map((p, i) => ({
        label: p,
        href: i < parts.length - 1 ? undefined : undefined,
      })),
    ];
  });

  const isMarkdown = $derived(selectedPath.endsWith('.md'));
  const isDiff = $derived(selectedPath.endsWith('.diff') || selectedPath.endsWith('.patch'));
  const isJson = $derived(selectedPath.endsWith('.json'));

  async function loadTree(path) {
    try {
      const data = await getResult(path || '');
      if (Array.isArray(data)) {
        return buildTree(data, path || '');
      }
    } catch (e) {
      error = e.message;
    }
    return [];
  }

  function buildTree(entries, basePath) {
    return entries.map(entry => {
      const fullPath = basePath ? `${basePath}/${entry.name}` : entry.name;
      const node = { name: entry.name, path: fullPath, isDir: entry.is_dir };
      if (entry.is_dir) {
        node.children = null; // lazy load
        node.loadChildren = async () => {
          node.children = await loadTree(fullPath);
          tree = [...tree]; // trigger reactivity
        };
      }
      return node;
    });
  }

  async function selectFile(path) {
    if (editorView) { editorView.destroy(); editorView = null; }
    selectedPath = path;
    mode = 'preview';
    dirty = false;
    try {
      fileContent = await getResultText(path);
    } catch (e) {
      fileContent = `Error loading file: ${e.message}`;
    }
  }

  function switchToEdit() {
    mode = 'edit';
    // Wait for DOM update then init editor
    requestAnimationFrame(() => {
      if (!editorEl) return;
      if (editorView) editorView.destroy();
      editorView = new EditorView({
        state: EditorState.create({
          doc: fileContent,
          extensions: [
            basicSetup,
            oneDark,
            EditorView.updateListener.of(update => {
              if (update.docChanged) dirty = true;
            }),
          ],
        }),
        parent: editorEl,
      });
    });
  }

  async function handleSave() {
    if (!editorView) return;
    const content = editorView.state.doc.toString();
    try {
      await saveResult(selectedPath, content);
      fileContent = content;
      dirty = false;
    } catch (e) {
      error = e.message;
    }
  }

  // Lazy-load children on FileTree expand
  async function handleSelect(path) {
    // Check if it's a directory by finding it in tree
    const entry = findEntry(tree, path);
    if (entry?.isDir && entry.loadChildren && !entry.children) {
      await entry.loadChildren();
    } else if (entry && !entry.isDir) {
      await selectFile(path);
    }
  }

  function findEntry(entries, path) {
    for (const e of entries) {
      if (e.path === path) return e;
      if (e.children) {
        const found = findEntry(e.children, path);
        if (found) return found;
      }
    }
    return null;
  }

  onMount(async () => {
    tree = await loadTree('');
    loading = false;
  });
</script>

<div class="page-header">
  <Breadcrumb segments={breadcrumbSegments()} />
  {#if selectedPath}
    <div class="flex items-center gap-sm">
      <button
        class:primary={mode === 'preview'}
        on:click={() => { mode = 'preview'; if (editorView) { editorView.destroy(); editorView = null; } }}
      >
        {@html icon('eye', 14)} Preview
      </button>
      <button
        class:primary={mode === 'edit'}
        on:click={switchToEdit}
      >
        {@html icon('edit', 14)} Edit
      </button>
      {#if mode === 'edit' && dirty}
        <button class="primary" on:click={handleSave}>
          {@html icon('save', 14)} Save
        </button>
      {/if}
    </div>
  {/if}
</div>

{#if loading}
  <div class="page-content"><p class="text-muted">Loading...</p></div>
{:else if error}
  <div class="page-content"><p class="status-fail">{error}</p></div>
{:else}
  <SplitPane leftWidth={250}>
    <div slot="left" class="tree-pane">
      <FileTree entries={tree} {selectedPath} onselect={handleSelect} />
    </div>
    <div slot="right" class="preview-pane">
      {#if !selectedPath}
        <div class="empty-state">
          <span class="text-muted">{@html icon('folder', 48)}</span>
          <p class="text-muted">Select a file to preview</p>
        </div>
      {:else if mode === 'edit'}
        <div class="editor-wrap" bind:this={editorEl}></div>
      {:else if isMarkdown}
        <div class="rendered-content">{@html renderMarkdown(fileContent)}</div>
      {:else if isJson}
        <pre><code>{JSON.stringify(JSON.parse(fileContent), null, 2)}</code></pre>
      {:else if isDiff}
        <pre class="diff-content">{fileContent}</pre>
      {:else}
        <pre><code>{fileContent}</code></pre>
      {/if}
    </div>
  </SplitPane>
{/if}

<style>
  .tree-pane {
    padding: 8px 0;
    height: 100%;
  }
  .preview-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
  }
  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    opacity: 0.5;
  }
  .rendered-content {
    padding: 24px;
    line-height: 1.6;
  }
  .rendered-content :global(pre) { margin: 12px 0; }
  .editor-wrap {
    flex: 1;
  }
  .editor-wrap :global(.cm-editor) { height: 100%; }
  pre {
    padding: 24px;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
    font-family: var(--font-mono);
    font-size: 12px;
  }
  .diff-content :global(.diff-add) { color: var(--neon-green); }
  .diff-content :global(.diff-remove) { color: var(--neon-magenta); }
</style>
```

- [ ] **Step 5: Verify build**

Run: `cd gui && npm run build`
Expected: Build succeeds

- [ ] **Step 6: Commit**

```bash
git add gui/src/lib/components/SplitPane.svelte gui/src/lib/components/FileTree.svelte gui/src/routes/ResultsBrowser.svelte gui/src/lib/api.js
git commit -m "feat(gui): full file browser with preview/edit modes and resizable panes"
```

---

## Task 9: Action System — Go Backend

**Files:**
- Modify: `internal/pipeline/sexpr.go` — add `action` keyword parsing
- Modify: `internal/gui/api_workflows.go` — add `actions` to workflowEntry, add actions endpoint
- Modify: `internal/gui/server.go` — register new route

This task is entirely Go — independent of frontend tasks.

- [ ] **Step 1: Check the Workflow struct for the Action field**

Read `internal/pipeline/pipeline.go` (or wherever the Workflow struct is defined) and find the struct definition. You need to add an `Actions []string` field.

Look for the Workflow struct. It should have fields like Name, Description, Author, Version, Tags, Created. Add:

```go
Actions []string
```

- [ ] **Step 2: Add action keyword parsing in sexpr.go**

In `internal/pipeline/sexpr.go`, find the keyword switch block (around line 66-98). Add a new case before the `default`:

```go
case "action":
    w.Actions = append(w.Actions, resolveVal(val, defs))
```

- [ ] **Step 3: Add actions to workflow list API response**

In `internal/gui/api_workflows.go`, add `Actions` to the `workflowEntry` struct:

```go
type workflowEntry struct {
    Name        string   `json:"name"`
    File        string   `json:"file"`
    Description string   `json:"description,omitempty"`
    Tags        []string `json:"tags,omitempty"`
    Author      string   `json:"author,omitempty"`
    Version     string   `json:"version,omitempty"`
    Created     string   `json:"created,omitempty"`
    Actions     []string `json:"actions,omitempty"`
}
```

In `handleListWorkflows`, where the entry is populated, add:

```go
Actions: wf.Actions,
```

- [ ] **Step 4: Add actions endpoint**

In `internal/gui/api_workflows.go`, add a new handler:

```go
func (s *Server) handleGetWorkflowActions(w http.ResponseWriter, r *http.Request) {
    context := r.PathValue("context")
    prefix := context + ":"

    workflows, err := s.loadWorkflows()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var matches []workflowEntry
    for _, wf := range workflows {
        for _, action := range wf.Actions {
            if action == context || strings.HasPrefix(action, prefix) {
                matches = append(matches, wf)
                break
            }
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(matches)
}
```

You may need to extract the workflow loading logic from `handleListWorkflows` into a shared `loadWorkflows()` method if it doesn't already exist.

- [ ] **Step 5: Register the new route**

In `internal/gui/server.go`, in the route registration block, add:

```go
mux.HandleFunc("GET /api/workflows/actions/{context}", s.handleGetWorkflowActions)
```

Add this BEFORE the `GET /api/workflows/{name}` route to avoid path conflicts.

- [ ] **Step 6: Run Go tests**

Run: `cd ~/Projects/gl1tch && go test ./internal/gui/ -v`
Expected: All existing tests pass

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/pipeline.go internal/gui/api_workflows.go internal/gui/server.go
git commit -m "feat(gui): add (action ...) keyword and workflow actions API endpoint"
```

---

## Task 10: Action System — Frontend

**Files:**
- Create: `gui/src/lib/components/ActionBar.svelte`
- Modify: `gui/src/lib/api.js`
- Modify: `gui/src/routes/ResultsBrowser.svelte`

Depends on Task 8 (results browser) and Task 9 (Go backend).

- [ ] **Step 1: Add getWorkflowActions to api.js**

Add this function to `gui/src/lib/api.js`:

```javascript
export async function getWorkflowActions(context) {
  const res = await fetch(`/api/workflows/actions/${context}`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
```

- [ ] **Step 2: Create ActionBar component**

Create `gui/src/lib/components/ActionBar.svelte`:

```svelte
<script>
  import { onMount } from 'svelte';
  import { getWorkflowActions } from '../api.js';
  import { icon } from '../icons.js';

  let { context = '', resultPath = '', onrun } = $props();
  let actions = $state([]);

  onMount(async () => {
    try {
      actions = await getWorkflowActions(context) || [];
    } catch (_) {
      actions = [];
    }
  });
</script>

{#if actions.length > 0}
  <div class="action-bar">
    <span class="action-label text-muted">Actions:</span>
    {#each actions as wf}
      <button class="primary" on:click={() => onrun?.(wf)}>
        {@html icon('zap', 14)} {wf.name}
      </button>
    {/each}
  </div>
{/if}

<style>
  .action-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border);
    flex-wrap: wrap;
  }
  .action-label {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
</style>
```

- [ ] **Step 3: Wire ActionBar into ResultsBrowser**

In `gui/src/routes/ResultsBrowser.svelte`, add the ActionBar import at the top:

```javascript
import ActionBar from '../lib/components/ActionBar.svelte';
import RunDialog from './RunDialog.svelte';
```

Add state for the action workflow trigger:

```javascript
let actionWorkflow = $state(null);
```

Add the ActionBar just before the SplitPane in the template, and the RunDialog at the bottom:

After the page-header div but before the SplitPane:
```svelte
<ActionBar context="results" resultPath={selectedPath} onrun={(wf) => { actionWorkflow = wf; }} />
```

At the very bottom of the component, before the closing `</style>`:
```svelte
{#if actionWorkflow}
  <RunDialog
    name={actionWorkflow.name}
    params={actionWorkflow.params || []}
    onclose={() => { actionWorkflow = null; }}
  />
{/if}
```

- [ ] **Step 4: Verify build**

Run: `cd gui && npm run build`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add gui/src/lib/components/ActionBar.svelte gui/src/lib/api.js gui/src/routes/ResultsBrowser.svelte
git commit -m "feat(gui): dynamic workflow action bar in results browser"
```

---

## Task 11: Playwright Tests for New Features

**Files:**
- Modify: `gui/e2e/gui.spec.js`

- [ ] **Step 1: Add new test cases**

Append these test blocks to `gui/e2e/gui.spec.js`:

```javascript
test.describe('Sidebar navigation', () => {
  test('sidebar renders with nav items', async ({ page }) => {
    await page.goto('/');
    const sidebar = page.locator('aside.sidebar');
    await expect(sidebar).toBeVisible();
    await expect(page.locator('.nav-item')).toHaveCount(3);
  });

  test('sidebar expands on hover', async ({ page }) => {
    await page.goto('/');
    const sidebar = page.locator('aside.sidebar');
    await sidebar.hover();
    await expect(sidebar).toHaveClass(/expanded/);
  });

  test('active nav item is highlighted', async ({ page }) => {
    await page.goto('/');
    const workflowsNav = page.locator('.nav-item.active');
    await expect(workflowsNav).toBeVisible();
  });
});

test.describe('Workflow card grid', () => {
  test('renders workflow cards', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.card');
    const cards = page.locator('.card');
    await expect(cards.first()).toBeVisible();
  });

  test('search filters workflows', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.card');
    const countBefore = await page.locator('.card').count();
    await page.fill('input[placeholder="Search..."]', 'nonexistent-workflow-xyz');
    await expect(page.locator('.card')).toHaveCount(0);
  });

  test('clicking card navigates to editor', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.card');
    await page.locator('.card').first().click();
    await expect(page).toHaveURL(/\/#\/workflow\//);
  });
});

test.describe('Editor metadata panel', () => {
  test('shows metadata panel', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.card');
    await page.locator('.card').first().click();
    await expect(page.locator('.meta-panel')).toBeVisible();
  });

  test('metadata panel can be collapsed', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.card');
    await page.locator('.card').first().click();
    await page.locator('.meta-panel .close-btn').click();
    await expect(page.locator('.meta-panel')).not.toBeVisible();
  });
});

test.describe('Results file browser', () => {
  test('file tree renders directories', async ({ page }) => {
    await page.goto('/#/results');
    await page.waitForSelector('.tree-item');
    const items = page.locator('.tree-item');
    await expect(items.first()).toBeVisible();
  });

  test('preview/edit toggle exists when file selected', async ({ page }) => {
    await page.goto('/#/results');
    await page.waitForSelector('.tree-item.file');
    await page.locator('.tree-item.file').first().click();
    await expect(page.locator('button:has-text("Preview")')).toBeVisible();
    await expect(page.locator('button:has-text("Edit")')).toBeVisible();
  });
});
```

- [ ] **Step 2: Run Playwright tests**

Run: `cd ~/Projects/gl1tch-gui-e2e && npx playwright test`
Expected: All tests pass (existing 20 + new tests)

Note: Some existing tests may need selector updates since the nav structure changed from top nav to sidebar. If tests fail, update the selectors:
- Old: `.topnav` → New: `aside.sidebar`
- Old: `.topnav a` → New: `.nav-item`

- [ ] **Step 3: Fix any broken existing tests**

Update selectors in existing tests that reference `.topnav` or the old navigation structure.

- [ ] **Step 4: Commit**

```bash
git add gui/e2e/gui.spec.js
git commit -m "test(gui): add Playwright tests for sidebar, card grid, editor panel, file browser"
```

---

## Task Dependency Graph

```
Task 1 (CSS) ─────────────────────────────┐
Task 2 (Icons) ───────────────────────────┤
                                          ├─→ Task 5 (Workflow List)
Task 3 (Shared Components) ──────────────┤├─→ Task 6 (Editor)
Task 4 (Layout Shell) ───────────────────┤├─→ Task 7 (Runs)
                                          ├─→ Task 8 (Results Browser) ──→ Task 10 (Action Frontend)
                                          │                                        ↑
Task 9 (Action Go Backend) ──────────────────────────────────────────────────────────┘

Task 11 (Playwright) ──→ depends on ALL above
```

**Parallelizable groups:**
- Group A: Tasks 1, 2 (no deps, run first)
- Group B: Tasks 3, 4 (depend on A, run together)
- Group C: Tasks 5, 6, 7, 8, 9 (depend on B — tasks 5-8 are frontend, task 9 is Go-only, all parallel)
- Group D: Task 10 (depends on 8 and 9)
- Group E: Task 11 (depends on all)
