## 1. Repository Bootstrap

- [x] 1.1 Create `adam-stokes/orcai-plugins` GitHub repository with README and MIT license
- [x] 1.2 Create directory structure: `plugins/ollama/` with `go.mod` (module `github.com/adam-stokes/orcai-plugins/plugins/ollama`)
- [x] 1.3 Add top-level `README.md` explaining the plugin repo structure and how to add new plugins

## 2. Ollama Provider Binary

- [x] 2.1 Create `plugins/ollama/main.go` with CLI entrypoint: reads stdin, checks `ORCAI_MODEL`, resolves `ORCAI_OLLAMA_URL` (default `http://localhost:11434`)
- [x] 2.2 Implement `callOllama(baseURL, model, prompt string) (string, error)` using `net/http` POST to `/api/generate` with `stream: false`
- [x] 2.3 Handle non-2xx HTTP responses: exit non-zero, write status + body to stderr
- [x] 2.4 Handle connection errors (refused, timeout): exit non-zero with descriptive stderr message
- [x] 2.5 Write completion text to stdout and exit 0 on success
- [x] 2.6 Return error and exit non-zero when stdin is empty

## 3. Sidecar YAML

- [x] 3.1 Create `plugins/ollama/ollama.yaml` declaring `name: ollama`, `description`, `command: orcai-ollama`
- [x] 3.2 Add comment block in the YAML documenting `vars.model` (→ `ORCAI_MODEL`) and `vars.ollama_url` (→ `ORCAI_OLLAMA_URL`)

## 4. Build & Install Tooling

- [x] 4.1 Create `plugins/ollama/Makefile` with `build` target (`go build -o orcai-ollama .`)
- [x] 4.2 Add `install` target: copy `orcai-ollama` to `/usr/local/bin/` and copy `ollama.yaml` to `~/.config/orcai/wrappers/`
- [x] 4.3 Add `test` target (`go test ./...`)

## 5. Tests

- [x] 5.1 Write unit test: empty stdin returns error
- [x] 5.2 Write unit test: missing `ORCAI_MODEL` returns error
- [x] 5.3 Write unit test: `ORCAI_OLLAMA_URL` defaults to `http://localhost:11434`
- [x] 5.4 Write unit test: `callOllama` with mocked HTTP server returns parsed completion text
- [x] 5.5 Write unit test: `callOllama` with mocked 404 response returns non-nil error

## 6. Sample Pipelines

- [x] 6.1 Create `examples/llama3.2-prompt.pipeline.yaml` with a single `ollama` provider step, `model: llama3.2`, static prompt, and prerequisite comment block
- [x] 6.2 Create `examples/qwen2.5-prompt.pipeline.yaml` with a single `ollama` provider step, `model: qwen2.5`, static prompt, and prerequisite comment block
- [x] 6.3 Create `examples/ollama-foreach.pipeline.yaml` with a `foreach` step over a prompt list, `model` from pipeline vars (default `llama3.2`), and prerequisite comment block

## 7. End-to-End Validation

- [x] 7.1 Run `orcai run examples/llama3.2-prompt.pipeline.yaml` against local Ollama with `llama3.2` pulled — verify exit 0 and non-empty output
- [x] 7.2 Run `orcai run examples/qwen2.5-prompt.pipeline.yaml` against local Ollama with `qwen2.5` pulled — verify exit 0 and non-empty output
- [x] 7.3 Run `orcai run examples/ollama-foreach.pipeline.yaml` — verify all foreach items complete successfully
- [x] 7.4 Verify `orcai` lists `ollama` as a registered provider after sidecar installation
