## Why

The Ollama plugin proves the sidecar model works, but real workflows need more capable local-AI adapters. OpenCode (`opencode run --model ollama/<model>`) provides agentic, tool-using inference on top of local models — something raw Ollama can't offer. Additionally, `orcai-plugins` has no automated testing beyond unit tests for the Ollama binary; end-to-end validation currently requires manual steps and only exercises the Ollama plugin. Both gaps need to close before the plugin repo is usable as a reference ecosystem.

## What Changes

- New `plugins/opencode/` directory in `adam-stokes/orcai-plugins` with a `orcai-opencode` binary, sidecar YAML, and Makefile — wrapping `opencode run` for non-interactive pipeline use.
- New `tests/` directory containing a shell-based e2e test harness that runs `orcai pipeline run` against real pipelines and validates outputs using `jq`, `grep`, and `orcai`'s own `builtin.*` steps.
- New example pipelines demonstrating real CLI tool chaining: `builtin.http_get` → `jq` (JSON extraction), `orcai-opencode` (agentic local-model step), and multi-step assertion pipelines.
- New `examples/jq-transform.pipeline.yaml` — fetches JSON from a public API, pipes it through a `jq` sidecar step to extract a field, then asserts the result is non-empty.
- New `examples/opencode-local.pipeline.yaml` — single agentic step using `opencode run --model ollama/llama3.2`.
- Updated top-level `Makefile` and `README` in `orcai-plugins` with `test-e2e` target and prerequisites.

## Capabilities

### New Capabilities

- `opencode-provider`: Sidecar-based CLI adapter wrapping `opencode run`; reads prompt from stdin, passes it as a positional arg, routes model via `--model` flag (resolved from `ORCAI_MODEL` env var or `--model` flag); exits non-zero on failure.
- `jq-sidecar`: Sidecar descriptor wrapping the system `jq` binary; accepts a jq filter expression via `ORCAI_FILTER` env var (set from `vars.filter`), applies it to stdin JSON, writes result to stdout.
- `pipeline-e2e-tests`: Shell-based e2e test harness in `tests/` that runs sample pipelines against real installed tools (`orcai`, `ollama`, `opencode`, `jq`, `curl`) and asserts expected output properties.

### Modified Capabilities

_(none — all changes are additive)_

## Impact

- **`adam-stokes/orcai-plugins`**: two new plugin directories, one new test directory, new example pipelines, updated root `Makefile`.
- **`orcai` core**: no changes required.
- **Prerequisites for e2e tests**: `orcai` built and on `$PATH`, `ollama` running with `llama3.2` pulled, `opencode` installed (`opencode run` available), `jq` installed.
- **CI potential**: the `test-e2e` target can be run in any environment with the prerequisites; designed to be runnable locally and in a future GitHub Actions workflow.
