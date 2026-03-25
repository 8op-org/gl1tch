## ADDED Requirements

### Requirement: Binary reads prompt from stdin
The `orcai-opencode` binary SHALL read its entire stdin as the prompt string before invoking `opencode run`.

#### Scenario: Stdin used as prompt
- **WHEN** the binary receives `"explain recursion"` on stdin
- **THEN** it passes `"explain recursion"` as the message argument to `opencode run`

#### Scenario: Empty stdin produces error
- **WHEN** stdin is closed immediately with no data
- **THEN** the binary exits non-zero and writes a descriptive error to stderr

### Requirement: Model resolved from --model flag or ORCAI_MODEL env var
The binary SHALL accept a `--model` flag. If not provided, it SHALL fall back to the `ORCAI_MODEL` environment variable. If neither is set, it SHALL exit non-zero with an error naming the missing configuration.

#### Scenario: --model flag used
- **WHEN** `--model ollama/llama3.2` is passed as a CLI flag
- **THEN** `opencode run --model ollama/llama3.2` is invoked

#### Scenario: ORCAI_MODEL env var used as fallback
- **WHEN** `--model` is not set and `ORCAI_MODEL=ollama/qwen3.5` is in the environment
- **THEN** `opencode run --model ollama/qwen3.5` is invoked

#### Scenario: Neither flag nor env var set
- **WHEN** neither `--model` flag nor `ORCAI_MODEL` is present
- **THEN** the binary exits non-zero with an error message referencing `--model` or `ORCAI_MODEL`

### Requirement: opencode run invoked in non-interactive mode
The binary SHALL always pass `--format default` (or equivalent flag) to `opencode run` to prevent TUI rendering and ensure stdout output is machine-readable.

#### Scenario: Non-interactive flag always present
- **WHEN** `orcai-opencode` invokes `opencode run`
- **THEN** the subprocess is called with a flag that prevents TUI rendering

### Requirement: opencode stdout written to binary stdout
The binary SHALL pipe `opencode run` stdout directly to its own stdout and exit with the same exit code as `opencode run`.

#### Scenario: Successful invocation exits zero
- **WHEN** `opencode run` completes successfully
- **THEN** `orcai-opencode` exits 0 and its stdout contains the opencode response

#### Scenario: opencode failure propagated
- **WHEN** `opencode run` exits non-zero
- **THEN** `orcai-opencode` exits non-zero and stderr contains the error output

### Requirement: Sidecar YAML declares the opencode provider
A sidecar file `opencode.yaml` SHALL be provided in `plugins/opencode/`. It SHALL declare `name: opencode`, a description, `command: orcai-opencode`, and document `vars.model` as the required model parameter.

#### Scenario: Sidecar loaded by orcai
- **WHEN** `opencode.yaml` is placed in `~/.config/orcai/wrappers/` and `orcai-opencode` is on PATH
- **THEN** `orcai` registers `opencode` as an available provider plugin

### Requirement: Plugin repository layout for opencode plugin
The `plugins/opencode/` directory SHALL contain `main.go`, `go.mod`, `Makefile`, and `opencode.yaml`. The Makefile SHALL provide `build`, `install`, and `test` targets. `make install` SHALL place `orcai-opencode` on PATH and `opencode.yaml` in `~/.config/orcai/wrappers/`.

#### Scenario: make install places binary and sidecar
- **WHEN** `make install` is run from `plugins/opencode/`
- **THEN** `orcai-opencode` is available via `which orcai-opencode` and `opencode.yaml` exists in `~/.config/orcai/wrappers/`
