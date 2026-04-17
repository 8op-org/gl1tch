<script>
  import { getRun } from '../api.js';
  import ELK from 'elkjs/lib/elk.bundled.js';
  import GraphNode from './GraphNode.svelte';
  import NodePanel from './NodePanel.svelte';

  let { runId } = $props();
  let run = $state(null);
  let steps = $state([]);
  let layoutNodes = $state([]);
  let layoutEdges = $state([]);
  let selectedStep = $state(null);
  let loading = $state(true);
  let error = $state(null);
  let svgWidth = $state(800);
  let svgHeight = $state(400);

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

    const nodeWidth = 200;
    const nodeHeight = 72;

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
        // (we'll skip to j in the loop — but since edges for par→next are added,
        // we just need to not add duplicate sequential edges)
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
        return {
          path: edgePath(adjusted),
          color: statusColor(st),
          running: st === 'running',
        };
      });

      // Set SVG dimensions with padding — add extra space for node widths at edges
      svgWidth = (laid.width || 800) + padding * 2 + 40;
      svgHeight = (laid.height || 400) + padding * 2 + 40;
    } catch (e) {
      console.error('ELK layout failed:', e);
    }
  }

  // Fetch data and layout
  async function loadGraph() {
    loading = true;
    error = null;
    selectedStep = null;
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

  // Reload when runId changes
  $effect(() => {
    if (runId) {
      loadGraph();
    }
  });
</script>

<div class="graph-container">
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
    <div class="graph-scroll">
      <div class="graph-viewport" style="width: {svgWidth}px; height: {svgHeight}px; position: relative;">
        <svg width={svgWidth} height={svgHeight} style="position: absolute; top: 0; left: 0;">
          {#each layoutEdges as edge}
            <path d={edge.path} fill="none" stroke={edge.color} stroke-width="2" class:running={edge.running} />
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
    </div>
  {/if}

  {#if selectedStep}
    <NodePanel step={selectedStep} {runId} onclose={() => selectedStep = null} />
  {/if}
</div>

<style>
  .graph-container {
    display: flex;
    flex-direction: row;
    min-height: 300px;
    max-height: 500px;
    overflow: hidden;
  }

  .graph-scroll {
    flex: 1;
    overflow: auto;
    min-width: 0;
  }

  .graph-viewport {
    min-width: 100%;
    min-height: 100%;
    padding: 20px;
    box-sizing: content-box;
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
