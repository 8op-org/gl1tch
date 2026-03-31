## Why

Codex is a first-class AI coding agent alongside Claude, Copilot, and Gemini, but ORCAI has no official plugin for it — users cannot use Codex through the agent runner. Additionally, all provider plugins live in a separate repo with manual install steps, creating friction; and the agent runner grid has no defined ordering, so priority providers like Claude and Copilot appear in arbitrary positions.

## What Changes

- **New `orcai-codex` plugin** in `orcai-plugins/plugins/codex/`: Go binary wrapper + `codex.yaml` sidecar + `Makefile` following the existing claude/opencode/copilot pattern
- **Bundled release archives**: goreleaser builds all core plugin binaries (orcai-claude, orcai-copilot, orcai-codex, orcai-gemini, orcai-opencode, orcai-ollama) alongside the `orcai` binary and includes them in every release tarball
- **Install script** updated to extract and place plugin binaries into `~/.local/bin` and sidecar YAMLs into `~/.config/orcai/wrappers/`
- **Agent runner priority ordering**: picker enforces a canonical priority list — Claude → Copilot → Codex → Gemini → OpenCode → Ollama → Shell → others

## Capabilities

### New Capabilities

- `codex-plugin`: Sidecar plugin for OpenAI Codex CLI — wraps `codex` binary, reads stdin prompt, streams output, supports `--model` flag
- `core-plugin-bundling`: Mechanism to build and bundle core provider plugin binaries in goreleaser release archives, with install script support
- `agent-runner-ordering`: Canonical priority ordering for the agent runner provider grid (most popular first)

### Modified Capabilities

- `cli-adapter-sidecar`: Install script and release archive now include sidecar YAML files alongside binaries

## Impact

- `orcai-plugins/plugins/codex/` — new directory with `main.go`, `codex.yaml`, `Makefile`
- `/Users/stokes/Projects/orcai/.goreleaser.yml` — add plugin build targets and extra_files
- `/Users/stokes/Projects/orcai/internal/picker/picker.go` — priority ordering in `buildProviders()`
- Release archives grow by ~6 plugin binaries per platform (small Go binaries, ~2–5 MB each)
- No breaking changes to existing plugin API or sidecar schema
