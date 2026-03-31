## Why

Pipeline step outputs contain valuable research, analysis, and findings that disappear after a run — the brain store exists but agents write to it with an underspecified format that is hard to query or reuse. System prompts controlling brain writing, pipeline generation, prompt building, and clarification are hardcoded in Go constants, making them impossible for users to customize without recompiling.

## What Changes

- Replace `<brain_notes>` primary tag with `<brain>`, supporting structured attributes (`type`, `tags`, `title`) so stored notes are semantically rich and RAG-queryable; keep `<brain_notes>` as a backward-compatible alias
- Make brain-writing **block-level opt-in**: any step output containing a `<brain>` block is parsed and stored regardless of `write_brain` flag — the agent decides what's worth remembering, not just the pipeline author
- Update the brain write instruction (injected into prompts) to document the richer `<brain>` format with examples
- Externalize four hardcoded system prompt constants to `~/.config/orcai/prompts/` as Markdown files installed on first run, loaded at runtime with embedded fallback

## Capabilities

### New Capabilities

- `brain-block-semantics`: Rich `<brain type="..." tags="..." title="...">` block format, updated parser, block-level opt-in storage, updated write instruction
- `user-configurable-prompts`: Install-on-first-run system prompts at `~/.config/orcai/prompts/`, runtime loader with embedded fallback, covering brain-write, pipeline-generator, prompt-builder, and clarify instructions

### Modified Capabilities

- `pipeline-step-lifecycle`: Brain extraction now happens on every step (not just `write_brain` steps) when a `<brain>` block is present in output

## Impact

- `internal/pipeline/brain.go` — parser regex, write instruction constant, extraction call site
- `internal/pipeline/runner.go` — remove `write_brain` gate on brain extraction
- `internal/pipelineeditor/runner.go` — load `pipeline-generator.md` instead of const
- `internal/promptmgr/update.go` — load `prompt-builder.md` instead of const
- `internal/clarify/clarify.go` — load `clarify.md` instead of const
- New package `internal/systemprompts/` — loader with install-on-first-run and embedded fallback
- `brain_notes` table schema unchanged; `tags` column stores `type:` and `title:` metadata
