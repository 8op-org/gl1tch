## ADDED Requirements

### Requirement: Pass inference options to Ollama at call time
The `orcai-ollama` binary SHALL accept one or more `--option key=value` flags and forward them to the Ollama `/api/generate` endpoint as the `options` object. Integer-parseable values SHALL be sent as JSON numbers; all other values SHALL be sent as strings.

#### Scenario: num_ctx option is forwarded in generate request
- **WHEN** orcai-ollama is invoked with `--model llama3.2 --option num_ctx=16384`
- **THEN** the POST body to `/api/generate` SHALL include `"options": {"num_ctx": 16384}`

#### Scenario: Multiple options are all forwarded
- **WHEN** orcai-ollama is invoked with `--option num_ctx=16384 --option temperature=0.7`
- **THEN** the POST body SHALL include both `"num_ctx": 16384` and `"temperature": 0.7` under `options`

#### Scenario: No options flag leaves options absent from request
- **WHEN** orcai-ollama is invoked without any `--option` flags
- **THEN** the POST body SHALL NOT include an `options` field (preserving existing behavior)

### Requirement: Options passable via pipeline vars
Pipeline step `vars` SHALL support `option_<key>: <value>` entries that are translated by the orcai runner to `--option <key>=<value>` arguments.

#### Scenario: num_ctx set via pipeline vars
- **WHEN** a pipeline step sets `vars: { model: llama3.2, option_num_ctx: "16384" }`
- **THEN** orcai-ollama SHALL receive `--option num_ctx=16384` and use a 16384-token context window
