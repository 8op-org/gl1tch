---
title: "Prompts"
description: "Save the instructions you give your assistant most often — recall them instantly in any pipeline or live session."
order: 99
---

Some instructions you write once and never want to type again. Your code review persona. Your commit message style. Your preferred debugging approach. The prompts system is your personal library — save a prompt once, drop it into any pipeline with a single field, and update it in one place when you want to change it everywhere.


## Quick Start

Open the prompt manager:

```bash
glitch prompts
```

Press `n` to create a new prompt, fill in the title and body, press `ctrl+s` to save.

Use it in a pipeline:

```yaml
steps:
  - id: review
    executor: ollama
    model: qwen2.5-coder:latest
    prompt_id: "Code review persona"
    input: "{{steps.diff.output}}"
```

When the step runs, your saved prompt body prepends to the input automatically.


## The Prompt Manager

The manager has three panels. `tab` / `shift+tab` cycles between them.

```
┌─────────────────┬───────────────────────────────┐
│  Prompt list    │  Editor                        │
│                 │  Title ___________________     │
│  > Code review  │  Body                          │
│    Commit msg   │  ________________________      │
│    Debug helper │  Model: ollama/qwen2.5-coder   │
│                 │  CWD:   ~/Projects/myapp       │
│                 ├───────────────────────────────┤
│                 │  Runner output                 │
└─────────────────┴───────────────────────────────┘
```

### List panel (left)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate up and down |
| `n` | New prompt — opens editor with blank form |
| `e` / `enter` | Edit selected prompt |
| `d` | Delete selected prompt (asks to confirm) |
| `tab` | Move to editor panel |
| any other key | Types into the filter — fuzzy search by title or body |
| `q` / `esc` | Quit |

### Editor panel (top-right)

`tab` / `shift+tab` moves focus between Title → Body → Model → CWD.

| Key | Action |
|-----|--------|
| `ctrl+s` | Save prompt |
| `ctrl+r` | Run prompt against the selected model — output appears in runner panel |
| `esc` | Back to list |

The **Model** field selects which executor runs when you test. The **CWD** field scopes the run to a project directory — useful when your prompt references local files.

### Runner panel (bottom-right)

| Key | Action |
|-----|--------|
| `ctrl+r` | Run again from scratch |
| `r` | Open follow-up input — continue the conversation |
| `p` | Promote the response to the body editor for review and saving |
| `j` / `k` | Scroll output |
| `esc` | Back to editor |


## Using Prompts in Pipelines

Add `prompt_id` to any step that calls a model. The value is the title of your saved prompt (case-insensitive).

```yaml
steps:
  - id: review
    executor: ollama
    model: qwen2.5-coder:latest
    prompt_id: "Code review persona"
    input: |
      Review these changes:
      {{steps.diff.output}}
```

When this step runs, the executor receives:

```text
[your saved prompt body]

[your step input]
```

> **NOTE:** If `prompt_id` references a title that doesn't exist, the step fails with a clear error. Use `glitch prompts` to verify the title before running.


## Examples


### Code Review Pipeline

Save a prompt titled `"Code review persona"` with your preferred reviewer voice, then reuse it across multiple pipelines.

```yaml
name: weekly-review
version: "1"

steps:
  - id: diff
    executor: shell
    command: "git diff main --stat | head -40"

  - id: review
    executor: ollama
    model: qwen2.5-coder:latest
    needs: [diff]
    prompt_id: "Code review persona"
    input: |
      Review these recent changes:
      {{steps.diff.output}}

  - id: save
    executor: shell
    needs: [review]
    command: |
      echo "{{steps.review.output}}" > review-$(date +%Y%m%d).md
```


### Same Prompt Across Multiple Steps

Both steps share the same persona. Edit the prompt once to update both.

```yaml
name: layered-review
version: "1"

steps:
  - id: review-api
    executor: claude
    model: claude-sonnet-4-6
    prompt_id: "Code review persona"
    input: "Review the API layer for correctness"

  - id: review-db
    executor: claude
    model: claude-sonnet-4-6
    prompt_id: "Code review persona"
    input: "Review the database layer for correctness"
```


### Commit Message Generator

Save a prompt titled `"Commit message style"` that describes your project's conventions.

```yaml
name: commit-helper
version: "1"

steps:
  - id: diff
    executor: shell
    command: "git diff --cached"

  - id: message
    executor: ollama
    model: qwen2.5-coder:latest
    needs: [diff]
    prompt_id: "Commit message style"
    input: |
      Generate a commit message for this diff:
      {{steps.diff.output}}
```


## Reference

### Pipeline step fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt_id` | string | no | Title of the saved prompt to inject. Case-insensitive match. |

### Injection order

When a step has `prompt_id` set, the executor receives content in this order:

1. Your saved prompt body
2. Your step's `input` field
3. Brain context (if any brain notes exist for this run)


## See Also

- [Brain](/docs/pipelines/brain) — combine saved prompts with memory for smarter pipelines
- [Pipelines](/docs/pipelines/pipelines) — full pipeline step reference
- [Cron](/docs/pipelines/cron) — schedule pipelines that use your prompt library
