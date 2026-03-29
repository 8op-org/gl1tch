## Context

The `orcai-ollama` plugin sends prompts to Ollama's `/api/generate` endpoint. The current `generateRequest` struct only passes `model`, `prompt`, and `stream` — no `options`. Ollama's API accepts an `options` object for runtime inference parameters including `num_ctx`, `temperature`, `top_p`, and others.

A separate concern is reusability: if a user sets `num_ctx 16384` for a session, the next pipeline run won't retain that setting. Ollama supports persistent named aliases via Modelfile (`FROM <base>\nPARAMETER <key> <value>`), letting users reference a pre-configured model by a short name.

The opencode plugin delegates entirely to the `opencode` binary via `opencode run --model <name>`. Since opencode passes the model name to Ollama, it automatically uses any named Ollama model — no binary changes are needed there.

## Goals / Non-Goals

**Goals:**
- Pass arbitrary Ollama `options` (e.g. `num_ctx`, `temperature`) at inference time via `orcai-ollama` flag
- Create named Ollama model aliases with baked-in parameters via `orcai-ollama --create-model`
- Update local wrapper configs to list high-context variants (`llama3.2-16k`, `qwen2.5-16k`) for both plugins
- Document the full workflow in plugin READMEs

**Non-Goals:**
- Managing Modelfile templates or multi-step parameter composition
- Supporting Ollama options in `orcai-opencode` binary (opencode handles this via model name)
- Any changes to the orcai core runner

## Decisions

### 1. `--option key=value` flag (repeatable) for orcai-ollama

**Decision**: Add a `-option` flag (Go `flag` package, repeatable via a custom `stringSlice` type) that populates an `options` map sent in the JSON request body.

**Alternatives considered**:
- Per-parameter flags (`--num-ctx 16384`): Tied to Ollama's specific parameters, breaks when new options are added. Rejected.
- `ORCAI_OLLAMA_OPTIONS` env var in JSON format: Hard to compose in pipeline `vars`. Rejected for the primary case.

**Rationale**: The `-option key=value` pattern is idiomatic (mirrors `docker run -e`), and Ollama's API accepts a generic map — so a single flag handles all current and future options without code changes.

### 2. Pipeline vars for options

**Decision**: Pipeline `vars` will support `option_<key>: <value>` pairs (e.g. `option_num_ctx: "16384"`) that the orcai runner passes as `--option num_ctx=16384`. This keeps the YAML readable and matches how model is already passed via vars.

**Alternatives considered**:
- Nested `options:` block in YAML: Requires orcai runner changes to pass structured data. Rejected.
- Flat `num_ctx: 16384` top-level vars: Too broad, would pollute the vars namespace. Rejected.

### 3. `--create-model <name>` flag for named Ollama aliases

**Decision**: Add `--create-model <name>` to `orcai-ollama` which: (1) reads `--model` as the base, (2) collects any `--option` flags as Modelfile `PARAMETER` lines, (3) POSTs to Ollama's `/api/create` endpoint (or shells out to `ollama create`). This creates a persistent model alias.

**Alternatives considered**:
- Shell script wrapper: Works but not composable in pipelines. Rejected.
- Only supporting `ollama create` via shell: Loses the ability to compose with existing flags in one invocation. Accepted as implementation detail — shell out to `ollama create -f -` via stdin pipe for simplicity.

### 4. Local config updates (immediate, outside proposal scope)

**Decision**: Update `~/.config/orcai/wrappers/ollama.yaml` and `opencode.yaml` directly as part of this change's implementation tasks. The custom model names (`llama3.2-16k`, `qwen2.5-16k`, `qwen3:8b-16k`) will be created in Ollama first via `--create-model`, then listed in wrappers.

## Risks / Trade-offs

- **Modelfile syntax drift** → Mitigation: Shell out to `ollama create` rather than calling the API directly; let Ollama validate the Modelfile format.
- **`num_ctx` silently ignored** if the model's architecture has a smaller max context → Mitigation: Document in README; Ollama logs a warning at pull time.
- **String-typed option values** require callers to know Ollama's expected types (int vs float) → Mitigation: The plugin passes values as raw JSON numbers when they parse as integers; strings otherwise.

## Migration Plan

1. Implement `--option` and `--create-model` in `orcai-ollama` binary.
2. Run `orcai-ollama --model llama3.2 --option num_ctx=16384 --create-model llama3.2-16k` to create the named alias.
3. Repeat for qwen2.5 and qwen3:8b (as `qwen2.5-16k` and `qwen3:8b-16k`).
4. Update `~/.config/orcai/wrappers/ollama.yaml` and `opencode.yaml` to list the new models.
5. Publish updated plugin binaries and YAML files.

Rollback: Delete the named Ollama models (`ollama rm llama3.2-16k` etc.) and revert wrapper YAML changes.

## Open Questions

- Should `--create-model` fail if the alias already exists, or overwrite? (Proposed default: overwrite, matching `ollama create` behavior.)
- Should the opencode plugin list also include `ollama/qwen3:8b-16k` or a separate `qwen3` base entry? (Proposed: include it as `ollama/qwen3:8b-16k` since that's the user's target model name.)
