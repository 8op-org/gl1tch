---
title: "CLI Reference"
description: "Every glitch command — ask, pipeline, workflow, cron, config, plugin, backup, and more."
order: 6
---

Every subcommand gl1tch exposes is listed here. The binary is `glitch`; all commands follow the pattern `glitch <command> [subcommand] [flags] [args]`.


## glitch ask

Send a prompt from the terminal. Routes to a matching pipeline automatically, or generates one on the fly when nothing matches.

```bash
glitch ask "sync my docs with the latest code changes"
glitch ask "what PRs need my review"
glitch ask --pipeline code-review "focus on the auth package"
```

Defaults to the first available local provider (ollama). No remote API calls unless you ask for them explicitly.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-p`, `--provider` | *(local)* | Provider ID: `ollama`, `claude`, etc. |
| `-m`, `--model` | *(provider default)* | Model name, e.g. `llama3.2`, `mistral`. |
| `--pipeline` | *(none)* | Run a named pipeline or file path instead of routing. |
| `--input key=value` | *(none)* | Pass input vars into the pipeline. Repeatable. |
| `--brain` | `true` | Inject brain context from the store to ground the response. |
| `--write-brain` | `false` | Write the response back to the brain store. |
| `--synthesize` | `false` | Run the response through claude to clean it up without adding new information. |
| `--synthesize-model` | *(claude default)* | Model used for the synthesis pass. |
| `--json` | `false` | Output the response as a JSON envelope. |
| `--route` | `true` | Route the prompt through the intent classifier to find a matching pipeline. |
| `--auto`, `-y` | `false` | Skip confirmation when a pipeline is generated on the fly. |
| `--dry-run` | `false` | Show what would run without executing. |

> [!TIP]
> Use `--dry-run` to preview intent routing decisions before committing to a run.


## glitch pipeline

Run and manage AI pipelines.


### glitch pipeline run

Run a saved pipeline by name or file path.

```bash
glitch pipeline run hello
glitch pipeline run ./my-pipeline.pipeline.yaml
glitch pipeline run code-review --input "focus=auth"
```

Named pipelines are looked up in `~/.config/glitch/pipelines/`. File paths (containing `/` or ending in `.yaml`) are used as-is.

| Flag | Default | Description |
|------|---------|-------------|
| `--input` | *(none)* | User input passed to the pipeline as `{{param.input}}`. |


### glitch pipeline resume

Resume a pipeline run that was interrupted or paused.

```bash
glitch pipeline resume --run-id 42
```

| Flag | Required | Description |
|------|----------|-------------|
| `--run-id` | yes | Store run ID to resume. |


## glitch workflow

Run and manage multi-step workflows. Workflows sequence multiple pipelines and can make branching decisions at runtime.


### glitch workflow run

Run a workflow by name. Workflow files live in `~/.config/glitch/workflows/`.

```bash
glitch workflow run triage-and-fix
glitch workflow run morning-prep --input "date=Monday"
```

| Flag | Default | Description |
|------|---------|-------------|
| `--input` | *(none)* | Input string passed to the workflow as `temp.input`. |


### glitch workflow resume

Resume a workflow run from its last checkpoint.

```bash
glitch workflow resume --run-id 7
```

| Flag | Required | Description |
|------|----------|-------------|
| `--run-id` | yes | Workflow run ID to resume. |


## glitch cron

Manage recurring pipeline and agent schedules. The cron daemon runs in a detached tmux session named `glitch-cron`.


### glitch cron start

Start the cron daemon.

```bash
glitch cron start
glitch cron start --force   # kill an existing session first
```

| Flag | Default | Description |
|------|---------|-------------|
| `--force` | `false` | Kill an existing `glitch-cron` session before starting. |


### glitch cron stop

Stop the running cron daemon.

```bash
glitch cron stop
```


### glitch cron list

List all configured cron schedules.

```bash
glitch cron list
```


### glitch cron logs

Show logs from the cron daemon.

```bash
glitch cron logs
```


### glitch cron run

Run a scheduled cron job immediately.

```bash
glitch cron run <job-name>
```


## glitch config

Manage glitch configuration.


### glitch config init

Write default `layout.yaml` and `keybindings.yaml` to `~/.config/glitch/`. Safe to run on an existing install — it will prompt before overwriting.

```bash
glitch config init
```

The default layout creates a `welcome` pane and a `sysop` pane. The default keybindings bind `M-n`, `M-t`, `M-w`, and vim-style pane resizing.


## glitch plugin

Install, remove, and list gl1tch plugins. Plugins are installed from GitHub repositories and their sidecar definitions are written to `~/.config/glitch/wrappers/`.


### glitch plugin install

Install a plugin from a GitHub repository.

```bash
glitch plugin install owner/repo
glitch plugin install owner/repo@v1.2.0
```

After install, restart glitch or run `glitch ask` to activate the new executor.


### glitch plugin remove

Remove an installed plugin.

```bash
glitch plugin remove my-plugin
```

Aliases: `rm`, `uninstall`.


### glitch plugin list

List all installed plugins.

```bash
glitch plugin list
```

Alias: `ls`.


## glitch backup

Backup config files, pipelines, prompts, and brain data to a compressed archive.

```bash
glitch backup
glitch backup --output ~/backups/glitch-$(date +%Y%m%d).tar.gz
glitch backup --no-agents
```

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `glitch-backup-<date>.tar.gz` | Output path for the backup archive. |
| `--no-agents` | `false` | Exclude auto-generated agent pipelines from `pipelines/.agents/`. |

Output on success:

```
Backup created: glitch-backup-2026-04-02.tar.gz
  Config files:  12
  Brain notes:   47
  Saved prompts: 8
```


## glitch restore

Restore config and brain data from a backup archive.

```bash
glitch restore glitch-backup-2026-04-02.tar.gz
glitch restore glitch-backup-2026-04-02.tar.gz --overwrite
glitch restore glitch-backup-2026-04-02.tar.gz --dry-run
```

| Flag | Default | Description |
|------|---------|-------------|
| `--overwrite` | `false` | Overwrite existing config files. By default, existing files are skipped. |
| `--dry-run` | `false` | Preview what would be restored without writing anything. |

> [!WARNING]
> Without `--overwrite`, existing files are left in place. Run `--dry-run` first to see what would change.


## glitch widget

Reusable TUI widget subcommands.


### glitch widget jump-window

Open the jump window TUI in standalone mode. This is the same overlay launched by `M-n` inside the main glitch session — useful for scripting or launching from outside a running session.

```bash
glitch widget jump-window
```


## See Also

- [Your First Pipeline](/docs/pipelines/quickstart) — write and run your first pipeline in five minutes
- [Pipeline YAML Reference](/docs/pipelines/yaml-reference) — every pipeline field documented
- [Executors and Plugins](/docs/pipelines/executors) — native executors, sidecar wrappers, and APM plugins
- [Brain](/docs/pipelines/brain) — how context is stored and injected automatically
