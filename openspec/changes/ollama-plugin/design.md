## Context

ORCAI's plugin system uses CLI adapter sidecars: a YAML descriptor in `~/.config/orcai/wrappers/` declares a named plugin that wraps a local binary. The binary receives `input` on stdin and `vars` as `ORCAI_*` environment variables, and writes its result to stdout. The existing `plugin-binary-override` and `cli-adapter-discovery` specs mean no core changes are required — the Ollama plugin fits cleanly into the established extension model.

Ollama exposes a local HTTP API (default `http://localhost:11434`). The plugin binary communicates with this daemon directly via `POST /api/generate` or `POST /api/chat`, streams the response, and writes the completed output to stdout.

This change lives in a new repository (`adam-stokes/orcai-plugins`) so the plugin ecosystem can be versioned, released, and installed independently of the core binary.

## Goals / Non-Goals

**Goals:**
- Provide a working `orcai-ollama` binary (Go, single static binary) that adapts the Ollama HTTP API to the orcai CLI adapter calling convention.
- Provide a sidecar YAML (`ollama.yaml`) installable into `~/.config/orcai/wrappers/`.
- Provide sample pipeline YAML files runnable with `orcai run` against `llama3.2` and `qwen2.5`.
- Establish `adam-stokes/orcai-plugins` as the canonical home for official orcai plugins.
- Support configuring the Ollama base URL and model name via `ORCAI_` env vars.

**Non-Goals:**
- Streaming output to the TUI in real time (the adapter waits for the full completion before writing stdout — streaming is a future enhancement).
- Support for multimodal / vision models in this iteration.
- GUI for model selection — model is specified in the pipeline YAML or via `vars`.
- Bundling Ollama itself — the user is responsible for installation.

## Decisions

### Decision: Separate repository (`adam-stokes/orcai-plugins`)

**Chosen**: New `adam-stokes/orcai-plugins` repo with one directory per plugin (`plugins/ollama/`).

**Rationale**: The plugin list will grow. Keeping plugins out of core avoids bloating the main repo, allows independent release cadences, and lets community contributors focus on a smaller codebase. Each plugin directory has its own `go.mod`, `Makefile`, and sidecar YAML.

**Alternative considered**: Monorepo subdirectory in `orcai`. Rejected — complicates versioning and couples plugin releases to core releases.

---

### Decision: Plain HTTP client, no Ollama Go SDK

**Chosen**: Direct `net/http` calls to the Ollama REST API (`/api/generate`).

**Rationale**: Keeps the plugin binary dependency-light and avoids SDK version drift. The `/api/generate` endpoint is stable, well-documented, and sufficient for text completion tasks.

**Alternative considered**: `github.com/ollama/ollama/api` client package. Rejected — adds a large transitive dependency for a small surface area.

---

### Decision: Model and URL configurable via `ORCAI_` env vars

**Chosen**: `ORCAI_MODEL` sets the model name; `ORCAI_OLLAMA_URL` overrides the base URL (default `http://localhost:11434`).

**Rationale**: Consistent with the existing calling convention (vars → `ORCAI_*` env vars). Pipeline authors set `vars.model` and optionally `vars.ollama_url` in their YAML.

---

### Decision: Sidecar installed manually, not auto-discovered from the plugin repo

**Chosen**: Users install the binary and sidecar YAML themselves (`make install` in the plugin repo copies both to the right locations).

**Rationale**: Zero changes to orcai core. Matches the existing sidecar pattern. A future plugin manager can automate this.

---

### Decision: Sample pipelines use `provider: ollama` step type

**Chosen**: Sample pipelines declare steps with `provider: ollama` and pass `model` in `vars`.

**Rationale**: Exercises the full dispatch path through the sidecar and validates the plugin end-to-end.

## Risks / Trade-offs

- **Ollama not running** → The adapter binary exits non-zero with a clear error message. The pipeline fails with the adapter's stderr surfaced (per `plugin-binary-override` spec). Mitigation: document prerequisite in README; add a health-check step option in a future version.
- **Model not pulled** → Ollama returns a 404-like error. The adapter propagates this as a non-zero exit. Mitigation: document required `ollama pull` commands in sample pipeline comments.
- **No streaming to TUI** → Long completions block without visual feedback. Mitigation: accepted for v1; streaming is a tracked follow-up.
- **New repo bootstrap overhead** → CI, releases, and `go.mod` setup are one-time costs. Mitigation: minimal Makefile with `build`, `install`, and `test` targets from day one.
