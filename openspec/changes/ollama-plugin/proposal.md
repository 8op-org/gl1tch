## Why

ORCAI currently relies on cloud-based AI providers, leaving users without a path for local, offline, or privacy-sensitive inference. An official Ollama plugin closes this gap by integrating with the Ollama daemon to run models like `llama3.2` and `qwen2.5` directly on the user's machine — no API keys, no network calls, no data leaving the host.

## What Changes

- New Go repository `adam-stokes/orcai-plugins` to house official third-party and local-provider plugins, starting with the Ollama plugin.
- New sidecar YAML descriptor (`orcai-ollama`) that declares the Ollama provider as a named plugin compatible with the existing CLI adapter discovery system.
- New `orcai-ollama` binary (CLI adapter) that speaks the orcai sidecar protocol, connects to the local Ollama HTTP daemon, and proxies pipeline step execution to the requested model.
- Sample pipeline YAML files that exercise the plugin against `llama3.2` and `qwen2.5`, runnable with `orcai run`.
- Plugin loading at startup: orcai picks up the sidecar automatically when `orcai-ollama` is on `$PATH` (no config changes required, per existing `plugin-binary-override` behaviour).

## Capabilities

### New Capabilities

- `ollama-provider`: Sidecar-based CLI adapter that connects to the Ollama daemon, accepts a `model` parameter, and streams completions back to orcai pipeline execution context.
- `ollama-sample-pipelines`: Ready-to-run `.pipeline.yaml` files demonstrating `llama3.2` and `qwen2.5` usage, including a simple prompt pipeline and a foreach pipeline.

### Modified Capabilities

_(none — all changes are additive)_

## Impact

- **New repository**: `adam-stokes/orcai-plugins` — hosts the Ollama plugin and future official plugins; versioned and released independently from the core `orcai` binary.
- **CLI adapter discovery**: The existing `LoadWrappersFromDir` / sidecar YAML mechanism is reused unchanged; no core modifications required.
- **Pipeline execution**: Pipelines using `provider: ollama` (or step type `ollama.<model>`) will invoke the adapter binary; the rest of the execution context is standard.
- **Dependencies**: The plugin binary depends on the Ollama HTTP API (`/api/generate`, `/api/chat`) — no Go SDK required, plain HTTP.
- **User prerequisite**: Ollama must be installed and running locally (`ollama serve`), with target models pulled (`ollama pull llama3.2`, `ollama pull qwen2.5`).
