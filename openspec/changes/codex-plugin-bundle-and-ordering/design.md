## Context

ORCAI's plugin ecosystem lives in a sibling repo (`orcai-plugins`). Each provider plugin is a small Go binary plus a sidecar YAML. Currently users must clone `orcai-plugins` and run `make install` per plugin. There is no official Codex plugin, and the agent runner grid has no enforced provider ordering. GoReleaser today only builds the `orcai` binary.

The `buildProviders()` function in `internal/picker/picker.go` appends discovered sidecar plugins in filesystem iteration order — non-deterministic across platforms.

## Goals / Non-Goals

**Goals:**
- Official `orcai-codex` plugin in `orcai-plugins/plugins/codex/` following the copilot pattern
- All core plugins (claude, copilot, codex, gemini, opencode, ollama) bundled in every goreleaser release archive alongside `orcai`
- Deterministic agent runner ordering: Claude → Copilot → Codex → Gemini → OpenCode → Ollama → Shell → others
- Install script that places plugin binaries and sidecar YAMLs in the right locations

**Non-Goals:**
- Moving `orcai-plugins` into a monorepo with `orcai`
- Auto-detection of installed vs. uninstalled plugins for hiding cards
- Dynamic priority configuration by the user

## Decisions

### 1. Codex plugin modeled on `orcai-github-copilot`

Codex CLI (`codex`) accepts a prompt via `--prompt` flag and `--model`. The plugin binary reads stdin as the prompt and `ORCAI_MODEL` env var, then execs `codex --prompt <prompt> [--model <model>]`. This mirrors exactly how `orcai-github-copilot` wraps the `copilot` CLI.

**Alternatives considered**: reading prompt from a temp file (rejected — stdin is the established convention); using `codex run` subcommand (rejected — `codex --prompt` is the direct interface per Codex CLI docs).

Static model list matches OpenAI's current Codex/o-series offerings: `codex-mini-latest`, `o4-mini`, `o3`.

### 2. Bundle plugins via goreleaser `builds` + `extra_files`

GoReleaser supports a `dir` field on each `builds` entry that changes the working directory for `go build`. Each plugin is a self-contained Go module, so adding a build entry with `dir: ../orcai-plugins/plugins/<name>` and `main: .` natively cross-compiles the plugin for every target platform.

Sidecar YAML files are not platform-specific. They are included via goreleaser `extra_files` using a glob across all plugin directories. Each archive then contains: `orcai`, all plugin binaries, all sidecar YAMLs, and `install.sh`.

**Alternatives considered**:
- `before.hooks` to run `make build-all`: rejected — goreleaser's before hooks run once without per-platform env injection, making cross-compilation awkward
- Embedding plugin binaries with `go:embed`: rejected — platform-specific binaries cannot be embedded in a cross-compiled binary cleanly; adds significant binary size complexity
- Separate goreleaser config in `orcai-plugins`: rejected — users would need to download two release artifacts; bundling keeps the UX simple

### 3. Priority ordering via a static ranked list in picker

A `providerPriority` slice in `picker.go` defines canonical rank: `["claude", "copilot", "codex", "gemini", "opencode", "ollama", "shell"]`. After `extras` are collected, they are sorted using this priority before appending to `out`. Providers not in the priority list sort after all known providers in their original discovery order.

**Alternatives considered**:
- Sorting by label alphabetically: rejected — "Claude" should lead regardless of alphabet
- User-configurable ordering via a config file: rejected — premature; add when there is demand

### 4. `install.sh` script in every archive

A shell script at the root of each archive detects the OS, copies plugin binaries to `~/.local/bin` (or `/usr/local/bin` if writable), and copies sidecar YAMLs to `~/.config/orcai/wrappers/`. This is generated once and included as a goreleaser `extra_files` entry.

## Risks / Trade-offs

- **`dir` path is relative to goreleaser project root**: `../orcai-plugins` must exist as a sibling directory at release time. CI pipelines need to clone both repos. → Mitigation: document in release runbook; add a CI check that the sibling repo is present.
- **Plugin binary name collisions**: goreleaser names binaries by the `binary` field, not directory. Each plugin must declare a unique `binary` name (e.g. `orcai-codex`). → Mitigation: enforce naming convention in plugin `Makefile` and goreleaser config.
- **Extra archive size**: 6 plugins × ~5 MB × supported platforms adds ~30 MB per platform archive. → Acceptable trade-off for zero-friction install.
- **Static Codex model list goes stale**: OpenAI adds models; the static list in `orcai-codex/main.go` must be updated manually. → Mitigation: implement `--list-models` via `codex models list` if/when that CLI subcommand becomes available.

## Migration Plan

1. Users who already have plugins installed via `make install` are unaffected — binaries and sidecars remain at the same paths.
2. New installs: download release archive, run `./install.sh`.
3. No changes to sidecar schema or plugin API — fully backward-compatible.

## Open Questions

- Does `codex` CLI support `--prompt` as a flag, or only stdin? Verify against current Codex CLI docs before implementing `execCodex()`.
- Should `install.sh` be a Makefile target instead, for users who prefer `make install` after extracting the archive?
