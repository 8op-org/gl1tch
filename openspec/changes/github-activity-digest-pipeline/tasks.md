## 1. gh Plugin Sidecar

- [x] 1.1 Create `orcai-plugins/plugins/gh/gh.yaml` sidecar descriptor wrapping `gh` CLI with `ORCAI_ARGS` passthrough
- [x] 1.2 Copy `gh.yaml` to `~/.config/orcai/wrappers/gh.yaml` to install the sidecar

## 2. Pipeline File

- [x] 2.1 Write `github-activity-digest.pipeline.yaml` with steps: fetch issues (both repos), fetch PRs (both repos), merge with jq, enrich via opencode/qwen2.5, summarise via claude haiku, validate with jq
- [x] 2.2 Copy pipeline to `~/.config/orcai/pipelines/github-activity-digest.pipeline.yaml`

## 3. Replace Demo Pipelines

- [x] 3.1 Delete all existing demo pipeline YAMLs from `~/.config/orcai/pipelines/` (httpbin, static snippet, dual-model-compare, etc.)
- [x] 3.2 Keep `opencode-code-review.pipeline.yaml` and `claude-review.pipeline.yaml` only if they are updated to reference real inputs (otherwise delete)

## 4. Live Test

- [ ] 4.1 Run `orcai pipeline run github-activity-digest` and verify all steps execute without error
- [ ] 4.2 Pipe final output through `jq '.issues.total_open, .prs.total_open, .action_items'` to confirm JSON is parseable
- [ ] 4.3 Fix any step failures found during live test (auth, model availability, JSON parsing)
