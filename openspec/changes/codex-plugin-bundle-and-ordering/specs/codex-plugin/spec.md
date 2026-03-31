## ADDED Requirements

### Requirement: Codex plugin binary wraps the codex CLI
The `orcai-codex` binary SHALL read the user prompt from stdin, read the model from the `ORCAI_MODEL` environment variable, and execute `codex --prompt <prompt> [--model <model>]`, streaming stdout and stderr to its own stdout/stderr. If the `codex` binary is not found on PATH, it SHALL exit with a non-zero code and a clear error message.

#### Scenario: Prompt forwarded to codex CLI
- **WHEN** `orcai-codex` receives `"explain this code"` on stdin and `ORCAI_MODEL` is unset
- **THEN** it executes `codex --prompt "explain this code"` and streams the output

#### Scenario: Model flag injected when ORCAI_MODEL is set
- **WHEN** `ORCAI_MODEL=o4-mini` is in the environment
- **THEN** it executes `codex --prompt <prompt> --model o4-mini`

#### Scenario: codex binary not found returns error
- **WHEN** `codex` is not on PATH
- **THEN** `orcai-codex` exits with code 1 and prints an installation hint to stderr

#### Scenario: Empty stdin returns error
- **WHEN** stdin is empty or whitespace-only
- **THEN** `orcai-codex` exits with code 1 and prints "prompt is required" to stderr

### Requirement: Codex plugin supports --list-models
The `orcai-codex` binary SHALL respond to the `--list-models` flag by printing a JSON array of `{"id": "...", "label": "..."}` objects to stdout and exiting 0, without invoking the `codex` binary. The static model list SHALL include at minimum: `codex-mini-latest`, `o4-mini`, `o3`.

#### Scenario: --list-models outputs JSON array
- **WHEN** `orcai-codex --list-models` is invoked
- **THEN** stdout contains a valid JSON array with at least one model object

#### Scenario: --list-models does not require codex on PATH
- **WHEN** `codex` is absent from PATH and `--list-models` is passed
- **THEN** `orcai-codex` still exits 0 with the model list

### Requirement: Codex sidecar YAML declares the plugin
A `codex.yaml` sidecar file SHALL exist in `orcai-plugins/plugins/codex/` declaring `name: codex`, `kind: agent`, `command: orcai-codex`, and the same model list as the binary's static list.

#### Scenario: Sidecar loaded by orcai wrappers discovery
- **WHEN** `codex.yaml` is present in `~/.config/orcai/wrappers/`
- **THEN** `loadSidecarMeta` returns a `sidecarMeta` with `Kind: "agent"` and the declared models

### Requirement: Codex plugin Makefile follows the standard plugin pattern
The `orcai-plugins/plugins/codex/Makefile` SHALL provide `build`, `install`, `test`, and `clean` targets matching the convention in `orcai-github-copilot/Makefile`. `install` SHALL copy the binary to `~/.local/bin/orcai-codex` and the sidecar to `~/.config/orcai/wrappers/codex.yaml`.

#### Scenario: make install places binary and sidecar
- **WHEN** `make install` is run in `plugins/codex/`
- **THEN** `orcai-codex` is present at `~/.local/bin/orcai-codex` and `codex.yaml` at `~/.config/orcai/wrappers/codex.yaml`
