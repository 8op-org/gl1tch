## Why

The existing pipelines in `~/.config/orcai` are toy demos (httpbin fetches, static code snippets) with no real-world data. We need production-grade pipelines that pull live GitHub activity from repos the team actually cares about — `elastic/ensemble` and `elastic/observability-robots` — and distill it into an actionable digest using a local LLM pass followed by a Claude Haiku JSON summary.

## What Changes

- **Replace** all existing pipelines in `~/.config/orcai/pipelines/` with real-world pipelines
- **Add** a `gh` CLI sidecar plugin at `orcai-plugins/plugins/gh/` (YAML descriptor + install docs) so pipelines can call `gh` as a first-class plugin
- **Add** `github-activity-digest.pipeline.yaml` — the primary pipeline:
  1. Fetches issues (updated in last 3 days) from both repos via `gh`
  2. Fetches open PRs from both repos via `gh`
  3. Extracts/normalises data with `jq`
  4. Passes structured data to `opencode` (qwen2.5) for an LLM-enriched narrative summary
  5. Passes that summary + raw counts to `claude haiku` which outputs strict JSON with digest totals
  6. Final `jq` step validates/pretty-prints the JSON output

## Capabilities

### New Capabilities

- `gh-plugin-sidecar`: ORCAI plugin sidecar wrapping the `gh` CLI so pipelines can call `gh issue list`, `gh pr list`, etc. as a pipeline step
- `github-activity-digest`: End-to-end pipeline that fetches GitHub issues + PRs, enriches with a local LLM, and produces a structured JSON digest via Claude Haiku

### Modified Capabilities

_(none — existing demo pipelines are replaced wholesale, no spec-level behavior changes)_

## Impact

- `~/.config/orcai/pipelines/` — existing YAML files replaced; old demo pipelines removed
- `orcai-plugins/plugins/gh/` — new sidecar plugin (YAML only, no binary required)
- Requires: `gh` CLI authenticated (`gh auth status`), `jq` on PATH, `orcai-opencode` installed, Claude API key set
