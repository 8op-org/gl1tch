## ADDED Requirements

### Requirement: DecisionNode calls Ollama with format:json

`DecisionNode.Evaluate(ctx, wctx WorkflowContext)` SHALL POST to Ollama's `/api/generate` endpoint with `"format": "json"` and `"stream": false`. The request body MUST include the model name from the step definition and the prompt after `{{ctx.*}}` template expansion. The call MUST time out after `timeout` seconds (default 30).

#### Scenario: Successful Ollama call returns branch
- **WHEN** Ollama responds with `{"branch": "feature"}` and `on.feature` is defined
- **THEN** `Evaluate` returns `"feature"` with no error

#### Scenario: Request times out
- **WHEN** Ollama does not respond within the configured timeout
- **THEN** `Evaluate` returns an error: "decision node timeout after <n>s"

#### Scenario: Ollama returns non-200 status
- **WHEN** Ollama returns HTTP 503
- **THEN** `Evaluate` returns an error containing the status code

### Requirement: Structured JSON response validated before branch lookup

The JSON response from Ollama MUST contain a `branch` field of type string. Any response missing this field or with a non-string value SHALL be treated as an error regardless of other fields present.

#### Scenario: Missing branch field is an error
- **WHEN** Ollama returns `{"confidence": 0.9}` with no `branch` field
- **THEN** `Evaluate` returns an error: "decision node: response missing 'branch' field"

#### Scenario: Non-string branch field is an error
- **WHEN** Ollama returns `{"branch": 42}`
- **THEN** `Evaluate` returns an error: "decision node: 'branch' must be a string"

### Requirement: default_branch used on Ollama error

If a `decision` step declares `default_branch` and `Evaluate` returns any error, `ConductorRunner` SHALL use `default_branch` as the resolved branch and log a warning. If no `default_branch` is declared, the workflow fails.

#### Scenario: default_branch used on timeout
- **WHEN** Ollama times out and `default_branch: "fallback"` is declared
- **THEN** the workflow continues via the `"fallback"` branch and a warning is logged to stderr

#### Scenario: No default_branch propagates error
- **WHEN** Ollama returns an error and no `default_branch` is declared
- **THEN** the workflow fails with the Ollama error

### Requirement: Decision node prompt supports WorkflowContext template references

The `prompt` field in a `decision` step MUST be expanded via `WorkflowContext` template substitution before the Ollama request is made. The expanded prompt is what is sent; the raw template is never sent to Ollama.

#### Scenario: Prompt expanded before send
- **WHEN** a decision step declares `prompt: "Classify this: {{ctx.fetch.output}}"` and context has `fetch.output="log data"`
- **THEN** Ollama receives the prompt `"Classify this: log data"`
