---
title: "Troubleshooting"
description: "Diagnose and fix common pipeline failures, template errors, and scheduling problems."
order: 7
---

Something's not working. Here's how to find out why.

## Check what would run first

Before running a pipeline for real, use `--dry-run` to see what gl1tch would execute without actually doing anything:

```bash
glitch pipeline run my-pipeline.pipeline.yaml --dry-run
glitch ask "summarize my PRs" --dry-run
```

Dry run shows you the resolved step order, template expressions, and which executor each step would call. Template errors surface here before they waste a real run.

## Read the step error

When a pipeline fails, the error message names the step and the executor. Look for the step `id` first — that tells you exactly where it broke.

```text
step [fetch-data] executor=gh: exit status 1
gh: unknown command "api repos/…"
```

Fix the command in that step, then re-run. You don't need to restart from the beginning — steps that already succeeded are cached in the run.

## Resume a failed run

Runs that fail mid-way can be resumed from the last successful step. Find the run ID in your inbox or from the signal board, then:

```bash
glitch pipeline resume --run-id <id>
```

Skips all completed steps and retries from the failure point.

## Template expression errors

Template errors look like `template: :1: unexpected "."` or an empty output where you expected data. Common causes:

| Symptom | Likely cause |
|---------|-------------|
| Step output is empty | Previous step produced no output |
| `unexpected "."` | Typo in expression — check the step ID |
| `undefined variable` | Step ID doesn't match what you typed |
| `index out of range` | Tried to access a key that doesn't exist |

Use `--dry-run` to see the resolved template before it runs. The dry-run output shows final expression values.

**Correct syntax:**

```yaml
prompt: "Summarize: {{steps.fetch.output}}"
input: "{{steps.extract.output}}"
vars:
  repo: "{{vars.repo}}"
```

Step IDs are case-sensitive. `{{steps.Fetch.output}}` won't match a step with `id: fetch`.

## Model not available

```text
no model available — is ollama running or a provider configured?
```

Check which model gl1tch would pick:

```bash
glitch model
glitch model --local   # only Ollama
```

If Ollama is not running, start it:

```bash
ollama serve
```

If you want to use a cloud provider, make sure it's authenticated and set it explicitly:

```bash
glitch ask "test" --provider claude
```

## Executor not found

```text
executor "ripgrep" not found
```

This means a pipeline step references a plugin executor that isn't installed. Install it:

```bash
glitch plugin install <owner/repo>
glitch plugin list   # verify it appears
```

## Cron jobs not firing

Check whether the scheduler is running:

```bash
glitch cron list
glitch cron logs
```

If there are no recent logs and jobs should have fired, the scheduler may not be running. Start it:

```bash
glitch cron start
```

If it's already running but jobs still don't fire, check the schedule expression:

```bash
# Test a cron expression manually
glitch cron run <name>   # force an immediate run
```

## Workflow stuck at a decision

Workflows that branch on conditions can get stuck if no condition matches and there's no `default` branch. Open the workflow file and add a fallback:

```yaml
branches:
  - condition: "{{steps.check.output | contains 'ok'}}"
    pipeline: success-path
  - condition: "default"
    pipeline: fallback-path
```

## Reset state

If something is deeply wrong and you want to start fresh without losing your pipelines and config:

```bash
# Back up everything first
glitch backup

# Restore only config, skip brain data
glitch restore ./backup.tar.gz --dry-run   # preview first
glitch restore ./backup.tar.gz
```

To wipe just the brain store and start with a clean context:

```bash
rm ~/.local/share/glitch/brain.db
```

This does not affect your pipelines, cron config, or prompts.

## See Also

- [CLI Reference](/docs/pipelines/cli-reference) — all flags including `--dry-run`
- [Pipelines](/docs/pipelines/pipelines) — pipeline YAML guide
- [Cron Scheduling](/docs/pipelines/cron) — scheduler setup and commands
- [Workflows](/docs/pipelines/workflows) — branching and resumption
