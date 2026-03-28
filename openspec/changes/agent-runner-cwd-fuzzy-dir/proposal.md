## Why

The agent runner overlay and pipeline launcher lack any concept of a working directory, forcing users to work exclusively from the directory orcai was launched in. The new session picker is also a separate, redundant UI that duplicates provider/model selection already present in the agent runner overlay — replacing it unifies the launch flow and reduces maintenance surface.

## What Changes

- Add a **WORKING DIRECTORY** field to the agent runner overlay, defaulting to the directory orcai was launched from, with fuzzy search across all directories under `~/`
- Replace the full-screen new session picker (`internal/picker/picker.go`) with the agent runner overlay — provider, model, prompt, and CWD are all configured there
- Add a reusable **directory picker overlay modal** (fuzzy search from `~/`) used by both the agent runner and pipeline launcher
- Expose directory selection in the pipeline run flow via the new overlay modal

## Capabilities

### New Capabilities

- `agent-runner-cwd`: CWD field in the agent runner overlay with fuzzy directory search, defaulting to orcai launch dir
- `dir-picker-overlay`: Reusable modal overlay for fuzzy directory selection from `~/`, shared by agent runner and pipeline launcher
- `new-session-agent-runner`: New session entry point uses the agent runner overlay instead of the legacy full-screen picker

### Modified Capabilities

- `pipeline-execution-context`: Pipeline launch flow gains directory selection via the new dir-picker-overlay modal before execution begins

## Impact

- `internal/switchboard/switchboard.go`: agent modal gains CWD field + fuzzy dir search trigger; new session keybinding launches agent modal instead of picker
- `internal/picker/picker.go` (and supporting files): legacy full-screen picker removed or reduced to worktree/git discovery utilities only
- New component: `internal/switchboard/dirpicker.go` (or similar) — reusable fuzzy dir picker overlay
- Pipeline run handler in switchboard: gains CWD selection step via the new overlay
