## ADDED Requirements

### Requirement: Create named Ollama model alias with baked-in parameters
The `orcai-ollama` binary SHALL accept a `--create-model <name>` flag that creates a new Ollama model alias using a Modelfile derived from the `--model` base and any `--option` flags. The command SHALL print a confirmation message to stdout and exit without performing inference.

#### Scenario: Create a high-context alias from base model
- **WHEN** orcai-ollama is invoked with `--model llama3.2 --option num_ctx=16384 --create-model llama3.2-16k`
- **THEN** a Modelfile containing `FROM llama3.2` and `PARAMETER num_ctx 16384` SHALL be passed to `ollama create llama3.2-16k`
- **THEN** stdout SHALL print `Created model 'llama3.2-16k'`
- **THEN** the process SHALL exit with code 0

#### Scenario: Created model is usable by name in subsequent invocations
- **WHEN** an Ollama model alias `llama3.2-16k` has been created
- **THEN** `orcai-ollama --model llama3.2-16k` SHALL run inference using the alias (with num_ctx baked in)

#### Scenario: Existing alias is overwritten by default
- **WHEN** orcai-ollama is invoked with `--create-model <name>` for an already-existing model
- **THEN** `ollama create` SHALL overwrite the existing model definition without error

#### Scenario: --create-model without --model flag returns error
- **WHEN** orcai-ollama is invoked with `--create-model llama3.2-16k` but no `--model` flag and no `ORCAI_MODEL` env var
- **THEN** orcai-ollama SHALL exit with a non-zero code and print an error indicating that `--model` is required

### Requirement: New model aliases listed in wrapper configs
Both `ollama.yaml` and `opencode.yaml` wrapper configs SHALL list the following custom high-context model entries as defaults, making them available for pipeline step `vars.model` selection:
- `llama3.2-16k` (16384-token context, base: llama3.2)
- `qwen2.5-16k` (16384-token context, base: qwen2.5)
- `qwen3:8b-16k` (16384-token context, base: qwen3:8b)

#### Scenario: High-context models appear in ollama wrapper model list
- **WHEN** a user reads `~/.config/orcai/wrappers/ollama.yaml`
- **THEN** the `models` list SHALL include entries for `llama3.2-16k` and `qwen2.5-16k`

#### Scenario: High-context models appear in opencode wrapper model list
- **WHEN** a user reads `~/.config/orcai/wrappers/opencode.yaml`
- **THEN** the `models` list SHALL include entries for `ollama/llama3.2-16k`, `ollama/qwen2.5-16k`, and `ollama/qwen3:8b-16k`
