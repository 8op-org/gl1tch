<script>
  import { getRun } from '../api.js';
  import { icon } from '../icons.js';
  import ELK from 'elkjs/lib/elk.bundled.js';
  import GraphNode from './GraphNode.svelte';
  import NodePanel from './NodePanel.svelte';

  let { runId, externalSteps = undefined } = $props();
  let run = $state(null);
  let steps = $state([]);
  let layoutNodes = $state([]);
  let layoutEdges = $state([]);
  let selectedStep = $state(null);
  let loading = $state(true);
  let error = $state(null);
  let svgWidth = $state(800);
  let svgHeight = $state(400);

  // Zoom/pan state
  let scale = $state(1);
  let panX = $state(0);
  let panY = $state(0);
  let naturalWidth = $state(800);
  let naturalHeight = $state(400);
  let containerEl = $state(null);
  let dragging = $state(false);
  let dragStartX = $state(0);
  let dragStartY = $state(0);
  let panStartX = $state(0);
  let panStartY = $state(0);

  const elk = new ELK();

  // Status helpers
  function stepStatus(step) {
    if (step.exit_status === undefined || step.exit_status === null) {
      if (step.gate_passed) return 'pass';
      return 'running';
    }
    return step.exit_status === 0 ? 'pass' : 'fail';
  }

  function statusColor(status) {
    switch (status) {
      case 'pass': return '#00e5ff';
      case 'fail': return '#ff2d6f';
      case 'running': return '#ffb800';
      default: return '#1e2a3a';
    }
  }

  // Map edge color to SVG filter
  function glowFilter(color) {
    switch (color) {
      case '#00e5ff': return 'url(#glow-cyan)';
      case '#ff2d6f': return 'url(#glow-magenta)';
      case '#ffb800': return 'url(#glow-amber)';
      default: return 'none';
    }
  }

  // Edge path generation from ELK layout
  function edgePath(edge) {
    const sections = edge.sections || [];
    if (sections.length === 0) return '';
    const s = sections[0];
    const start = s.startPoint;
    const end = s.endPoint;
    const bends = s.bendPoints || [];

    if (bends.length === 0) {
      const mx = (start.x + end.x) / 2;
      return `M ${start.x} ${start.y} C ${mx} ${start.y}, ${mx} ${end.y}, ${end.x} ${end.y}`;
    }

    let path = `M ${start.x} ${start.y}`;
    let prev = start;
    for (const bp of bends) {
      const mx = (prev.x + bp.x) / 2;
      path += ` C ${mx} ${prev.y}, ${mx} ${bp.y}, ${bp.x} ${bp.y}`;
      prev = bp;
    }
    const mx = (prev.x + end.x) / 2;
    path += ` C ${mx} ${prev.y}, ${mx} ${end.y}, ${end.x} ${end.y}`;
    return path;
  }

  // Build ELK graph from steps
  function buildElkGraph(steps) {
    if (!steps || steps.length === 0) return null;

    const nodeWidth = 180;
    const nodeHeight = 68;

    // Identify par groups: consecutive steps with kind === 'par'
    // Fan out from previous non-par step, converge to next non-par step
    const nodes = steps.map((s) => ({
      id: s.step_id,
      width: nodeWidth,
      height: nodeHeight,
    }));

    const edges = [];
    for (let i = 0; i < steps.length - 1; i++) {
      const current = steps[i];
      const next = steps[i + 1];

      // Check for par fan-out: if next step is par, and current is not par,
      // we connect current to all consecutive par steps
      if (next.kind === 'par' && current.kind !== 'par') {
        // Find all consecutive par steps starting from i+1
        let j = i + 1;
        while (j < steps.length && steps[j].kind === 'par') j++;
        // Connect current to each par step
        for (let k = i + 1; k < j; k++) {
          edges.push({
            id: `e-${current.step_id}-${steps[k].step_id}`,
            sources: [current.step_id],
            targets: [steps[k].step_id],
          });
        }
        // Connect each par step to the step after the par group (if exists)
        if (j < steps.length) {
          for (let k = i + 1; k < j; k++) {
            edges.push({
              id: `e-${steps[k].step_id}-${steps[j].step_id}`,
              sources: [steps[k].step_id],
              targets: [steps[j].step_id],
            });
          }
        }
        // Skip the par group in normal sequential iteration
        continue;
      }

      // Skip edges from within a par group (already handled above)
      if (current.kind === 'par') continue;

      // Normal sequential edge
      edges.push({
        id: `e-${current.step_id}-${next.step_id}`,
        sources: [current.step_id],
        targets: [next.step_id],
      });
    }

    return {
      id: 'root',
      layoutOptions: {
        'elk.algorithm': 'layered',
        'elk.direction': 'RIGHT',
        'elk.spacing.nodeNode': '40',
        'elk.layered.spacing.nodeNodeBetweenLayers': '60',
        'elk.edgeRouting': 'SPLINES',
      },
      children: nodes,
      edges: edges,
    };
  }

  // Auto-fit: compute scale and pan to center graph in container
  function autoFit() {
    if (!containerEl || naturalWidth <= 0 || naturalHeight <= 0) return;
    const rect = containerEl.getBoundingClientRect();
    const cw = rect.width;
    const ch = rect.height;
    if (cw <= 0 || ch <= 0) return;

    const s = Math.min(cw / naturalWidth, ch / naturalHeight, 1.5);
    scale = s;
    panX = (cw - naturalWidth * s) / 2;
    panY = (ch - naturalHeight * s) / 2;
  }

  // Perform layout and extract positions
  async function doLayout(steps) {
    const graph = buildElkGraph(steps);
    if (!graph) {
      layoutNodes = [];
      layoutEdges = [];
      return;
    }

    try {
      const laid = await elk.layout(graph);
      const padding = 20;

      // Map step data by id for lookup
      const stepMap = {};
      for (const s of steps) stepMap[s.step_id] = s;

      layoutNodes = (laid.children || []).map((n) => ({
        x: n.x + padding,
        y: n.y + padding,
        width: n.width,
        height: n.height,
        step: stepMap[n.id],
      }));

      layoutEdges = (laid.edges || []).map((e) => {
        // Offset edge points by padding
        const adjusted = {
          ...e,
          sections: (e.sections || []).map((s) => ({
            ...s,
            startPoint: { x: s.startPoint.x + padding, y: s.startPoint.y + padding },
            endPoint: { x: s.endPoint.x + padding, y: s.endPoint.y + padding },
            bendPoints: (s.bendPoints || []).map((bp) => ({
              x: bp.x + padding,
              y: bp.y + padding,
            })),
          })),
        };
        const sourceStep = stepMap[e.sources?.[0]];
        const st = sourceStep ? stepStatus(sourceStep) : 'default';
        const lastSection = adjusted.sections?.[adjusted.sections.length - 1];
        return {
          path: edgePath(adjusted),
          color: statusColor(st),
          running: st === 'running',
          endX: lastSection?.endPoint?.x || 0,
          endY: lastSection?.endPoint?.y || 0,
        };
      });

      // Set SVG/natural dimensions — ELK width/height already includes all nodes
      svgWidth = (laid.width || 800) + padding * 4;
      svgHeight = (laid.height || 400) + padding * 4;
      naturalWidth = svgWidth;
      naturalHeight = svgHeight;

      // Auto-fit after layout
      requestAnimationFrame(() => autoFit());
    } catch (e) {
      console.error('ELK layout failed:', e);
    }
  }

  // Fetch data and layout
  async function loadGraph() {
    loading = true;
    error = null;
    selectedStep = null;

    // If externalSteps provided, skip fetch
    if (externalSteps) {
      steps = externalSteps;
      await doLayout(steps);
      loading = false;
      return;
    }

    try {
      const data = await getRun(runId);
      run = data.run || data;
      steps = data.steps || [];
      await doLayout(steps);
    } catch (e) {
      error = e.message;
      console.error('Failed to load run:', e);
    } finally {
      loading = false;
    }
  }

  // Reload when runId or externalSteps changes
  $effect(() => {
    if (externalSteps) {
      loadGraph();
    } else if (runId) {
      loadGraph();
    }
  });

  // ResizeObserver for auto-fit on container resize
  $effect(() => {
    if (!containerEl) return;
    const ro = new ResizeObserver(() => {
      if (layoutNodes.length > 0) {
        autoFit();
      }
    });
    ro.observe(containerEl);
    return () => ro.disconnect();
  });

  // Wheel zoom handler
  function handleWheel(e) {
    e.preventDefault();
    const rect = containerEl.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    const zoomFactor = e.deltaY < 0 ? 1.1 : 0.9;
    const newScale = Math.max(0.1, Math.min(5, scale * zoomFactor));

    // Zoom toward cursor: adjust pan so the point under cursor stays fixed
    panX = mouseX - (mouseX - panX) * (newScale / scale);
    panY = mouseY - (mouseY - panY) * (newScale / scale);
    scale = newScale;
  }

  // Mouse drag for panning
  function handleMouseDown(e) {
    // Don't pan when clicking a graph node
    if (e.target.closest('.graph-node')) return;
    if (e.button !== 0) return;
    dragging = true;
    dragStartX = e.clientX;
    dragStartY = e.clientY;
    panStartX = panX;
    panStartY = panY;
  }

  function handleMouseMove(e) {
    if (!dragging) return;
    panX = panStartX + (e.clientX - dragStartX);
    panY = panStartY + (e.clientY - dragStartY);
  }

  function handleMouseUp() {
    dragging = false;
  }

  function resetZoom() {
    autoFit();
  }
</script>

<div class="graph-container" class:panel-open={selectedStep}>
  {#if loading}
    <div class="graph-loading">
      <p class="text-muted">Loading graph...</p>
    </div>
  {:else if error}
    <div class="graph-loading">
      <p class="text-muted">Failed to load run: {error}</p>
    </div>
  {:else if steps.length === 0}
    <div class="graph-loading">
      <p class="text-muted">No steps in this run</p>
    </div>
  {:else}
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="graph-canvas"
      bind:this={containerEl}
      onwheel={handleWheel}
      onmousedown={handleMouseDown}
      onmousemove={handleMouseMove}
      onmouseup={handleMouseUp}
      onmouseleave={handleMouseUp}
    >
      <div
        class="graph-transform"
        style="transform: translate({panX}px, {panY}px) scale({scale}); transform-origin: 0 0;"
      >
        <svg width={svgWidth} height={svgHeight} style="position: absolute; top: 0; left: 0;">
          <defs>
            <filter id="glow-cyan" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur" />
              <feFlood flood-color="#00e5ff" flood-opacity="0.6" result="color" />
              <feComposite in="color" in2="blur" operator="in" result="glow" />
              <feMerge>
                <feMergeNode in="glow" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
            <filter id="glow-magenta" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur" />
              <feFlood flood-color="#ff2d6f" flood-opacity="0.6" result="color" />
              <feComposite in="color" in2="blur" operator="in" result="glow" />
              <feMerge>
                <feMergeNode in="glow" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
            <filter id="glow-amber" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur" />
              <feFlood flood-color="#ffb800" flood-opacity="0.6" result="color" />
              <feComposite in="color" in2="blur" operator="in" result="glow" />
              <feMerge>
                <feMergeNode in="glow" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
          </defs>
          {#each layoutEdges as edge}
            <path
              d={edge.path}
              fill="none"
              stroke={edge.color}
              stroke-width="2"
              stroke-opacity="0.8"
              filter={glowFilter(edge.color)}
              class:running={edge.running}
            />
            <!-- Arrow marker at end -->
            <circle cx={edge.endX} cy={edge.endY} r="4" fill={edge.color} opacity="0.8" />
          {/each}
        </svg>
        {#each layoutNodes as node}
          <div style="position: absolute; left: {node.x}px; top: {node.y}px; width: {node.width}px; height: {node.height}px;">
            <GraphNode
              step={node.step}
              selected={selectedStep?.step_id === node.step.step_id}
              onclick={() => selectedStep = selectedStep?.step_id === node.step.step_id ? null : node.step}
            />
          </div>
        {/each}
      </div>

      <button class="zoom-reset" onclick={resetZoom} type="button">
        {@html icon('search', 14)}
        Reset
      </button>
    </div>
  {/if}

  {#if selectedStep}
    <div class="slide-over">
      <NodePanel step={selectedStep} {runId} onclose={() => selectedStep = null} />
    </div>
  {/if}
</div>

<style>
  .graph-container {
    display: flex;
    flex-direction: row;
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  .graph-canvas {
    flex: 1;
    overflow: hidden;
    position: relative;
    cursor: grab;
    min-height: 200px;
  }

  .graph-canvas:active {
    cursor: grabbing;
  }

  .graph-transform {
    position: absolute;
    top: 0;
    left: 0;
    will-change: transform;
  }

  .zoom-reset {
    position: absolute;
    bottom: 12px;
    right: 12px;
    padding: 6px 10px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 6px;
    color: var(--text-muted);
    cursor: pointer;
    font-size: 12px;
    display: flex;
    align-items: center;
    gap: 4px;
    z-index: 10;
  }

  .zoom-reset:hover {
    color: var(--text-primary);
    border-color: var(--neon-cyan);
  }

  .slide-over {
    width: 420px;
    flex-shrink: 0;
    border-left: 1px solid var(--border);
    background: var(--bg-surface);
    overflow-y: auto;
    height: 100%;
  }

  .graph-loading {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
  }

  .text-muted {
    color: var(--text-muted);
    font-size: 13px;
  }

  /* Running edge animation */
  @keyframes dash {
    to { stroke-dashoffset: -24; }
  }
  path.running {
    stroke-dasharray: 8 4;
    animation: dash 1s linear infinite;
  }
</style>
