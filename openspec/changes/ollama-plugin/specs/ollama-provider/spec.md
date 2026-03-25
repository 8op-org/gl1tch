## ADDED Requirements

### Requirement: Binary reads prompt from stdin
The `orcai-ollama` binary SHALL read its entire stdin as the prompt string before making any HTTP request to the Ollama daemon.

#### Scenario: Stdin used as prompt
- **WHEN** the binary is invoked with `echo "tell me a joke" | orcai-ollama`
- **THEN** the string `"tell me a joke"` is sent as the `prompt` field in the Ollama API request

#### Scenario: Empty stdin produces error
- **WHEN** stdin is closed immediately with no data
- **THEN** the binary exits non-zero and writes a descriptive error to stderr

### Requirement: Model resolved from ORCAI_MODEL environment variable
The binary SHALL use the value of `ORCAI_MODEL` as the model name in the Ollama API request. If `ORCAI_MODEL` is unset or empty the binary SHALL exit non-zero with a message indicating the variable is required.

#### Scenario: ORCAI_MODEL set to valid model
- **WHEN** `ORCAI_MODEL=llama3.2` is in the environment
- **THEN** the API request body includes `"model": "llama3.2"`

#### Scenario: ORCAI_MODEL unset
- **WHEN** `ORCAI_MODEL` is not set
- **THEN** the binary exits with a non-zero code and writes "ORCAI_MODEL is required" (or similar) to stderr

### Requirement: Ollama base URL configurable via ORCAI_OLLAMA_URL
The binary SHALL use `ORCAI_OLLAMA_URL` as the base URL for all Ollama API calls. When the variable is unset or empty the binary SHALL default to `http://localhost:11434`.

#### Scenario: Default URL used when variable absent
- **WHEN** `ORCAI_OLLAMA_URL` is not set
- **THEN** HTTP requests target `http://localhost:11434/api/generate`

#### Scenario: Custom URL overrides default
- **WHEN** `ORCAI_OLLAMA_URL=http://gpu-box:11434` is in the environment
- **THEN** HTTP requests target `http://gpu-box:11434/api/generate`

### Requirement: Completion written to stdout
On a successful Ollama API response the binary SHALL write the completed text to stdout and exit zero.

#### Scenario: Successful completion exits zero
- **WHEN** the Ollama daemon returns a valid completion
- **THEN** the binary writes the response text to stdout and exits 0

#### Scenario: Non-streaming response collected in full
- **WHEN** `stream: false` is used in the request
- **THEN** the full response body is written as a single stdout write before exit

### Requirement: Ollama API errors surfaced as non-zero exit
If the Ollama HTTP endpoint returns a non-2xx status or a connection error occurs, the binary SHALL exit non-zero and write the error details to stderr.

#### Scenario: Connection refused
- **WHEN** no Ollama daemon is listening on the configured URL
- **THEN** the binary exits non-zero and writes a message containing "connection refused" (or equivalent) to stderr

#### Scenario: Non-2xx HTTP status
- **WHEN** the Ollama API returns HTTP 404 (e.g., model not found)
- **THEN** the binary exits non-zero and writes the status code and response body to stderr

### Requirement: Sidecar YAML declares the ollama provider
A sidecar file `ollama.yaml` SHALL be provided in the plugin repository. It SHALL declare `name: ollama`, a human-readable `description`, `command: orcai-ollama`, and document the required `ORCAI_MODEL` and optional `ORCAI_OLLAMA_URL` vars.

#### Scenario: Sidecar loaded by orcai
- **WHEN** `ollama.yaml` is placed in `~/.config/orcai/wrappers/` and `orcai-ollama` is on PATH
- **THEN** `orcai` registers `ollama` as an available provider plugin

#### Scenario: Sidecar includes vars documentation
- **WHEN** `ollama.yaml` is read
- **THEN** it contains comments or fields documenting `model` and `ollama_url` as accepted vars

### Requirement: Plugin repository layout
The `adam-stokes/orcai-plugins` repository SHALL contain a top-level `plugins/ollama/` directory with `main.go`, `go.mod`, `Makefile`, and `ollama.yaml`. The Makefile SHALL provide `build`, `install`, and `test` targets. `make install` SHALL copy the binary to a directory on `$PATH` and copy `ollama.yaml` to `~/.config/orcai/wrappers/`.

#### Scenario: make install places binary on PATH
- **WHEN** `make install` is run from `plugins/ollama/`
- **THEN** `orcai-ollama` is available via `which orcai-ollama`

#### Scenario: make install places sidecar
- **WHEN** `make install` is run from `plugins/ollama/`
- **THEN** `~/.config/orcai/wrappers/ollama.yaml` exists
