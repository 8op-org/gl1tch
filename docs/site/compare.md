# Compare Runs

The `(compare ...)` form runs the same task through different branches and scores the results. A neutral local model judges each branch against your criteria.

This page builds on [Workflow Syntax](/docs/workflow-syntax). If you haven't read it yet, start there.

## Key concepts

- `(compare ...)` wraps `(branch ...)` blocks and a `(review ...)` block
- Each branch runs independently with its own model or strategy
- `(review :criteria (...))` scores branches on your terms
- `--variant` flag injects ad-hoc compare blocks without editing the workflow
- `--compare` discovers sibling variant workflows and cross-reviews them
- Results save to disk with `(save ...)` like any other step

## Tone cues

- Lead with the two real examples (compare-models.glitch, compare-branches.glitch)
- "your" framing throughout
- Explain scoring after showing the examples
- CLI flags section: --variant, --compare, --review-criteria
- End with saving/reading results and next steps
