## 1. Codex Plugin (orcai-plugins)

- [x] 1.1 Create `orcai-plugins/plugins/codex/` directory with `go.mod` (module `github.com/adam-stokes/orcai-plugins/codex`)
- [x] 1.2 Implement `main.go` — `run()` function reads stdin as prompt, reads `ORCAI_MODEL`, handles `--list-models` flag, execs `codex --prompt <prompt> [--model <model>]`
- [x] 1.3 Write `codex.yaml` sidecar declaring `name: codex`, `kind: agent`, `command: orcai-codex`, and static model list
- [x] 1.4 Write `Makefile` with `build`, `install`, `test`, `clean` targets matching the copilot plugin convention
- [x] 1.5 Add unit tests in `main_test.go` covering: prompt forwarding, model injection, `--list-models`, missing binary error, empty stdin error

## 2. Agent Runner Priority Ordering (orcai)

- [x] 2.1 Add `providerPriority` slice to `internal/picker/picker.go`: `["claude", "copilot", "codex", "gemini", "opencode", "ollama", "shell"]`
- [x] 2.2 Sort the `extras` slice before the append loop using `providerPriority` — providers in the list sort by rank; unknown providers preserve their relative discovery order after all ranked entries
- [x] 2.3 Update or add tests in `picker_test.go` asserting Claude → Copilot → Codex → Gemini order and that unknown providers appear last

## 3. Release Bundling (.goreleaser.yml)

- [x] 3.1 Add a `builds` entry for each core plugin (`orcai-claude`, `orcai-github-copilot`, `orcai-codex`, `orcai-gemini`, `orcai-opencode`, `orcai-ollama`) using `dir: ../orcai-plugins/plugins/<name>` and `main: .`, with the same `goos`/`goarch` matrix and `windows/arm64` ignore
- [x] 3.2 Add `extra_files` to the `archives` section globbing all core sidecar YAMLs from `../orcai-plugins/plugins/*/` into the archive
- [x] 3.3 Write `install.sh` at the repo root — detects write access to `/usr/local/bin` vs `~/.local/bin`, copies plugin binaries and sidecar YAMLs, prints each destination, idempotent
- [x] 3.4 Add `install.sh` as an `extra_files` entry in goreleaser archives

## 4. Verification

- [x] 4.1 Run `goreleaser build --snapshot --clean` and confirm all 6 plugin binaries appear in `dist/` for each platform
- [x] 4.2 Extract a snapshot archive and run `./install.sh` — verify binaries land in `~/.local/bin` and sidecars in `~/.config/orcai/wrappers/`
- [x] 4.3 Start orcai with all core plugins installed and confirm agent runner grid order: Claude → Copilot → Codex → Gemini → OpenCode → Ollama → Shell
- [x] 4.4 Run `orcai-codex --list-models` and confirm JSON output; run with a prompt and confirm codex CLI is invoked
