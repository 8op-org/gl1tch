## ADDED Requirements

### Requirement: Simple prompt pipeline for llama3.2
A pipeline file `examples/llama3.2-prompt.pipeline.yaml` SHALL be provided in the plugin repository. It SHALL define a single step that invokes the `ollama` provider with `model: llama3.2` and a static prompt demonstrating basic text generation. The file SHALL include comments explaining each field.

#### Scenario: Pipeline runs against llama3.2
- **WHEN** `orcai run examples/llama3.2-prompt.pipeline.yaml` is executed with Ollama running and `llama3.2` pulled
- **THEN** the pipeline exits 0 and the step output contains a non-empty text completion

#### Scenario: Pipeline YAML is valid
- **WHEN** the YAML file is parsed by orcai
- **THEN** no schema validation errors are reported

### Requirement: Simple prompt pipeline for qwen2.5
A pipeline file `examples/qwen2.5-prompt.pipeline.yaml` SHALL be provided. It SHALL define a single step invoking the `ollama` provider with `model: qwen2.5` and a static prompt. The file SHALL include comments.

#### Scenario: Pipeline runs against qwen2.5
- **WHEN** `orcai run examples/qwen2.5-prompt.pipeline.yaml` is executed with Ollama running and `qwen2.5` pulled
- **THEN** the pipeline exits 0 and the step output contains a non-empty text completion

#### Scenario: Pipeline YAML is valid
- **WHEN** the YAML file is parsed by orcai
- **THEN** no schema validation errors are reported

### Requirement: foreach pipeline demonstrates multi-step chaining
A pipeline file `examples/ollama-foreach.pipeline.yaml` SHALL be provided. It SHALL define a `foreach` step that iterates over a list of prompts and invokes the `ollama` provider (model configurable via pipeline variable) for each. This exercises the `pipeline-for-each` execution path with the plugin.

#### Scenario: foreach pipeline iterates all items
- **WHEN** `orcai run examples/ollama-foreach.pipeline.yaml` is executed
- **THEN** a completion is produced for each item in the foreach list and all steps exit 0

#### Scenario: Pipeline uses pipeline variable for model
- **WHEN** the pipeline YAML declares a `vars` block with `model` defaulting to `llama3.2`
- **THEN** each step passes `ORCAI_MODEL=llama3.2` to the adapter binary

### Requirement: Sample pipelines document prerequisites in comments
Each sample pipeline YAML file SHALL include a comment block at the top listing prerequisites: Ollama installed and running, required models pulled, and the `orcai-ollama` binary installed.

#### Scenario: Prerequisites comment present
- **WHEN** a sample pipeline file is opened
- **THEN** the first lines contain a comment block listing `ollama serve`, the required `ollama pull` command, and `make install` from the plugin repo
