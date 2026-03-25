## ADDED Requirements

### Requirement: Test harness entrypoint runs all tests and reports results
A script `tests/run_all.sh` SHALL execute every `tests/*.sh` test script in sequence, print PASS / SKIP / FAIL for each, and exit non-zero if any test FAILs.

#### Scenario: All tests pass
- **WHEN** `tests/run_all.sh` is executed and every test script exits 0
- **THEN** the runner prints `All N tests passed` and exits 0

#### Scenario: One test fails
- **WHEN** one test script exits non-zero (not 77)
- **THEN** the runner prints `FAIL: <name>` and exits 1

#### Scenario: One test is skipped
- **WHEN** a test script exits 77 (skip code)
- **THEN** the runner prints `SKIP: <name>` and continues without failing

### Requirement: Each test script checks prerequisites before running
Every test script SHALL check for required tools via `command -v` at the top and exit 77 (skip) if any prerequisite is missing.

#### Scenario: Missing tool causes skip
- **WHEN** a test requires `jq` and `jq` is not on PATH
- **THEN** the test exits 77 and the harness reports SKIP

#### Scenario: All prerequisites present
- **WHEN** all required tools are available
- **THEN** the test proceeds to execution

### Requirement: test-ollama-simple verifies the Ollama plugin end-to-end
A test `tests/test-ollama-simple.sh` SHALL run `orcai pipeline run examples/llama3.2-prompt.pipeline.yaml`, capture output, assert exit code is 0, and assert output is non-empty.

#### Scenario: Pipeline exits 0 and produces output
- **WHEN** `tests/test-ollama-simple.sh` is executed with Ollama running and `llama3.2` pulled
- **THEN** it exits 0 (PASS)

#### Scenario: Ollama not running
- **WHEN** the Ollama daemon is not listening
- **THEN** `orcai pipeline run` exits non-zero and the test exits 1 (FAIL)

### Requirement: test-jq-transform verifies JSON extraction via the jq plugin
A test `tests/test-jq-transform.sh` SHALL run `orcai pipeline run examples/jq-transform.pipeline.yaml`, assert exit code is 0, and use `jq -e` to validate the pipeline output is valid JSON with an expected field.

#### Scenario: jq pipeline produces expected JSON field
- **WHEN** `tests/test-jq-transform.sh` is executed with `jq` and `orcai-jq` installed
- **THEN** the pipeline output contains the expected extracted field and the test exits 0

### Requirement: test-opencode-local verifies the opencode plugin end-to-end
A test `tests/test-opencode-local.sh` SHALL run `orcai pipeline run examples/opencode-local.pipeline.yaml`, assert exit code is 0, and assert output is non-empty.

#### Scenario: opencode pipeline exits 0 and produces output
- **WHEN** `tests/test-opencode-local.sh` is executed with opencode installed and Ollama running with llama3.2
- **THEN** it exits 0 (PASS)

### Requirement: test-builtin-assert verifies pipeline assertion steps
A test `tests/test-builtin-assert.sh` SHALL run a pipeline that uses `builtin.assert` to validate intermediate step output, exercising the assert step with both passing and failing conditions.

#### Scenario: Assert pipeline passes all checks
- **WHEN** `tests/test-builtin-assert.sh` is executed
- **THEN** it exits 0 (PASS)

### Requirement: Root Makefile test-e2e target runs the harness
The root `Makefile` in `orcai-plugins` SHALL add a `test-e2e` target that runs `tests/run_all.sh`. It SHALL print a prerequisites reminder before running.

#### Scenario: make test-e2e invokes the harness
- **WHEN** `make test-e2e` is run from the repo root
- **THEN** `tests/run_all.sh` is executed

### Requirement: jq-transform example pipeline fetches JSON and extracts a field
A pipeline `examples/jq-transform.pipeline.yaml` SHALL use `builtin.http_get` to fetch a public JSON API endpoint, pass the response body through an `orcai-jq` step with a filter expression, then use `builtin.assert` to verify the result is non-empty.

#### Scenario: Pipeline produces a non-empty extracted value
- **WHEN** `orcai pipeline run examples/jq-transform.pipeline.yaml` is executed with internet access and `orcai-jq` installed
- **THEN** the pipeline exits 0 and outputs the extracted JSON value

### Requirement: opencode-local example pipeline runs an agentic local-model step
A pipeline `examples/opencode-local.pipeline.yaml` SHALL define a single `opencode` provider step with `model: ollama/llama3.2` and a static prompt, demonstrating agentic local inference.

#### Scenario: Pipeline runs opencode against local ollama model
- **WHEN** `orcai pipeline run examples/opencode-local.pipeline.yaml` is executed with opencode installed and Ollama running
- **THEN** the pipeline exits 0 and outputs a non-empty completion
