## Context

ORCAI pipelines are YAML-defined multi-step workflows where each step is an executor (builtin, plugin, or model). The existing `~/.config/orcai/pipelines/` contains demo pipelines that serve no practical purpose. The `orcai-plugins` repo holds sidecar YAML descriptors that map a named plugin to a CLI binary — `jq` and `opencode` sidecars already exist. The `gh` CLI is installed and authenticated but has no sidecar, blocking it from use in pipelines.

The target pipeline must:
1. Call `gh` to pull live GitHub data (issues + PRs)
2. Normalise it with `jq`
3. Run a local LLM enrichment pass via `opencode` (qwen2.5 via Ollama)
4. Call Claude Haiku for structured JSON output
5. Validate the JSON with `jq`

## Goals / Non-Goals

**Goals:**
- `gh` sidecar plugin at `orcai-plugins/plugins/gh/gh.yaml` — wraps `gh` CLI, passes args via `ORCAI_ARGS` env var
- `github-activity-digest.pipeline.yaml` replacing the existing demo pipelines
- Strict JSON output schema consumable via `jq` downstream
- Work with both `elastic/ensemble` and `elastic/observability-robots`

**Non-Goals:**
- GitHub Actions / CI integration
- Webhook-triggered pipelines
- Pagination beyond default `gh` limits (first 30 items per query is sufficient for a daily digest)
- Authentication management (user must have `gh auth status` passing)

## Decisions

### Decision: `gh` sidecar uses `ORCAI_ARGS` env var

The `jq` sidecar uses `ORCAI_FILTER` for its single parameter. `gh` takes variable subcommands and flags, so we need a general-purpose args passthrough. Passing `ORCAI_ARGS` and expanding it in a `sh -c` invocation (`gh $ORCAI_ARGS`) is the simplest approach consistent with the existing sidecar pattern.

**Alternative considered:** Named vars per subcommand (e.g., `ORCAI_SUBCOMMAND`, `ORCAI_REPO`). Rejected — too rigid; `gh` has dozens of subcommands.

### Decision: Two-stage LLM — opencode (qwen2.5) then Claude Haiku

- **Stage 1 (opencode/qwen2.5):** Local model that runs offline, produces a free-text enrichment of the raw data — identifies patterns, flags stale items, adds context. This runs fast and costs nothing.
- **Stage 2 (Claude Haiku):** Strict JSON schema enforcement. Haiku is cheap, fast, and follows structured output instructions reliably. Takes Stage 1 narrative + raw counts and emits the final JSON digest.

**Alternative considered:** Single Claude call only. Rejected — loses the local enrichment step the user specifically requested.

### Decision: Merge both repos' data before LLM steps

Rather than running parallel LLM calls per repo, we merge issue and PR lists with `jq` into a single JSON array (tagged by `repo` field) before the LLM steps. This gives both models a holistic view across repos and halves LLM round-trips.

### Decision: Replace all existing demo pipelines

The existing pipelines (httpbin, static code snippets) have no production value. Replacing them wholesale rather than keeping them reduces confusion. The new set is: `github-activity-digest`, plus a lean `claude-code-review` and `opencode-code-review` kept only if they reference real inputs.

## Risks / Trade-offs

- **gh rate limits** → Mitigation: default `gh issue list` fetches ≤30 items; well within REST API limits for authenticated calls
- **Ollama not running** → Mitigation: opencode step will fail with clear error; add `builtin.assert` after to surface the message
- **Claude API key not set** → Mitigation: Claude executor will fail; documented in pipeline comment header
- **qwen2.5 hallucinating JSON** → Mitigation: qwen2.5 only produces free-text narrative; Haiku produces the JSON, with a downstream `jq` assert that validates it parses

## Migration Plan

1. Install `gh` sidecar: copy `gh.yaml` to `~/.config/orcai/wrappers/gh.yaml`
2. Replace pipelines: copy new YAMLs to `~/.config/orcai/pipelines/`, delete old demo files
3. Verify: run `orcai pipeline run github-activity-digest`
4. Rollback: old pipeline files are tracked in git; restore with `git checkout`

## Open Questions

- Should the digest pipeline run on a schedule (cron) or on-demand only? → Out of scope for this change; user can add a cron trigger later.
