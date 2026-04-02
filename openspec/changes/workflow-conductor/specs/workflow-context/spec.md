## ADDED Requirements

### Requirement: WorkflowContext is a shared key-value store

`WorkflowContext` SHALL be a `map[string]string` threaded through all steps in a workflow run. It MUST be safe for concurrent reads and writes (protected by a `sync.RWMutex`). Keys are namespaced as `<step_id>.output` for step outputs and `<step_id>.<field>` for structured decision outputs.

#### Scenario: Step output stored by ID
- **WHEN** a `pipeline-ref` step with `id: "fetch"` completes with output `"result text"`
- **THEN** `WorkflowContext["fetch.output"]` equals `"result text"`

#### Scenario: Decision field stored
- **WHEN** a `decision` step with `id: "triage"` returns `{"branch": "bugfix", "reason": "..."}`
- **THEN** `WorkflowContext["triage.branch"]` equals `"bugfix"` and `WorkflowContext["triage.reason"]` equals `"..."`

### Requirement: Template expansion in step inputs

Step `input` and `prompt` fields SHALL support `{{ctx.<key>}}` template references. The runner MUST expand these against `WorkflowContext` before passing the value to `pipeline.Run()` or `DecisionNode.Evaluate()`. Unknown keys expand to an empty string.

#### Scenario: Known key expanded
- **WHEN** a step declares `input: "Summarize: {{ctx.fetch.output}}"` and `WorkflowContext["fetch.output"]` is `"hello"`
- **THEN** the effective input passed to the executor is `"Summarize: hello"`

#### Scenario: Unknown key expands to empty string
- **WHEN** a step references `{{ctx.missing.output}}` and the key is not in `WorkflowContext`
- **THEN** the reference is replaced with `""` and no error is returned

### Requirement: Context values truncated at 16 KB per key

When a step output exceeds 16,384 bytes, `WorkflowContext` SHALL store only the first 16,384 bytes of the value. The full output SHALL remain accessible in the pipeline run store. A warning SHALL be logged to stderr.

#### Scenario: Large output truncated in context
- **WHEN** a step produces 32,000 bytes of output
- **THEN** `WorkflowContext["<step_id>.output"]` contains exactly 16,384 bytes and a warning is written to stderr

#### Scenario: Small output stored intact
- **WHEN** a step produces 500 bytes of output
- **THEN** `WorkflowContext["<step_id>.output"]` contains all 500 bytes without truncation

### Requirement: Context serialized to JSON for checkpointing

`WorkflowContext.Marshal()` SHALL return a JSON byte slice of the full map. `WorkflowContext.Unmarshal([]byte)` SHALL restore the map from JSON. Both operations MUST be lossless for values within the 16 KB truncation limit.

#### Scenario: Round-trip marshal/unmarshal
- **WHEN** a context with three keys is marshalled then unmarshalled
- **THEN** the resulting context has identical keys and values
