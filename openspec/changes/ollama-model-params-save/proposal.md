## Why

Local LLM inference via Ollama defaults to a small context window (2048 tokens), which is too limiting for agentic and coding workflows. There is currently no way to pass model parameters (like `num_ctx`) through the orcai-ollama or orcai-opencode plugins, nor to create and reference named model variants with custom settings baked in.

## What Changes

- **orcai-ollama plugin**: Add support for Ollama `options` (e.g. `num_ctx`, `temperature`) passed at inference time via a `--option key=value` flag, usable from pipeline `vars`.
- **orcai-ollama plugin**: Add `--create-model <name>` flag to create a named Ollama model via Modelfile (`FROM <base> PARAMETER <key> <value>`) so the alias is reusable by name from any plugin.
- **orcai-opencode plugin**: No binary changes needed — once a named model is created in Ollama, opencode can reference it as `ollama/<name>` directly.
- **Local wrapper configs**: Update `~/.config/orcai/wrappers/ollama.yaml` and `opencode.yaml` model lists to include custom-named high-context variants for llama3.2 and qwen2.5 (e.g. `llama3.2-16k`, `qwen2.5-16k`).
- **Plugin repo documentation**: Update READMEs for both plugins documenting the new flags and custom model workflow.

## Capabilities

### New Capabilities

- `ollama-model-options`: Pass Ollama inference options (e.g. `num_ctx`) at call time via plugin flags and pipeline vars.
- `ollama-model-create`: Create named Ollama model aliases with baked-in parameters via Modelfile, accessible by both ollama and opencode plugins.

### Modified Capabilities

<!-- No existing spec-level behavior changes. -->

## Impact

- `orcai-plugins/plugins/ollama/main.go` — add `--option` and `--create-model` flags
- `orcai-plugins/plugins/ollama/ollama.yaml` — add custom high-context model entries
- `orcai-plugins/plugins/opencode/opencode.yaml` — add `ollama/<name>-16k` model entries
- `~/.config/orcai/wrappers/ollama.yaml` — update local install with new models
- `~/.config/orcai/wrappers/opencode.yaml` — update local install with new models
- `orcai-plugins/plugins/ollama/README.md` — new docs
- `orcai-plugins/plugins/opencode/README.md` — updated docs for custom model usage
