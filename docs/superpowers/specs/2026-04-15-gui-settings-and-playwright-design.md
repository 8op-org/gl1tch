# GUI Settings Page & Playwright Tests

**Date:** 2026-04-15
**Status:** Draft

## Summary

Add a dedicated Settings page (`/#/settings`) to the gl1tch workflow GUI and expand Playwright E2E test coverage to include it. The settings page surfaces two categories of configuration: **Workflow Defaults** (model, provider, default parameter values) and **Workspace Config** (workspace name, Kibana URL, repositories).

All settings persist server-side via a new `PUT /api/workspace` endpoint that writes back to `workspace.glitch`.

## Architecture

### Frontend

**New files:**
- `gui/src/routes/Settings.svelte` — settings page with two sections
- `gui/e2e/settings.spec.js` — Playwright tests for settings

**Modified files:**
- `gui/src/App.svelte` — add `/settings` route
- `gui/src/lib/components/Sidebar.svelte` — wire settings footer as nav link to `/#/settings`, add active state
- `gui/src/lib/api.js` — add `getWorkspace()` and `updateWorkspace()` API functions
- `gui/src/routes/RunDialog.svelte` — pre-fill parameter inputs from workspace defaults

### Backend

**Modified files:**
- `internal/gui/api_workspace.go` — add `handlePutWorkspace` handler; extend `workspaceResponse` to include `elasticsearch` and `default_params`
- `internal/gui/server.go` — register `PUT /api/workspace` route
- `internal/workspace/workspace.go` — add `Serialize()` method to write workspace back to s-expression format; add `DefaultParams map[string]string` field

## Settings Page Design

### Section 1: Workflow Defaults

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Default model | Text input | `workspace.defaults.model` | e.g. `qwen2.5:7b` |
| Default provider | Dropdown | `workspace.defaults.provider` | Populated from provider registry via new `GET /api/providers` endpoint, or freeform text fallback |
| Default parameters | Key-value list | New `workspace.defaults.params` | Add/remove rows. Common params: `repo`, `results-dir`, `review-criteria`, `variant` |

Default parameters pre-fill the RunDialog inputs. User-entered values in RunDialog override defaults.

### Section 2: Workspace Config

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Workspace name | Text input | `workspace.name` | Display name for this workspace |
| Kibana URL | Text input | `workspace.defaults.elasticsearch` | Base URL for telemetry links. Currently hardcoded to `localhost:5601` in `api_kibana.go` |
| Repositories | Editable list | `workspace.repos` | Add/remove repo paths |

### Save behavior

- Single "Save" button at the bottom of the page
- `PUT /api/workspace` sends the full workspace object
- Server serializes back to `workspace.glitch` s-expression format
- Success toast or inline status indicator (matches existing `saveStatus` pattern from Editor)
- No auto-save — explicit save only

### Kibana URL wiring

When saved, the Kibana URL replaces the hardcoded `defaultKibanaURL` constant in `api_kibana.go`. The server reads the workspace's `defaults.elasticsearch` field at request time instead of using the const.

## Workspace S-expression Serialization

The workspace parser currently only reads. A new `Serialize(*Workspace) []byte` function writes back to the canonical format:

```
(workspace "my-workspace"
  :description "workspace description"
  :owner "adam"
  (repos
    "elastic/kibana"
    "elastic/observability-robots")
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :elasticsearch "http://localhost:9200"
    (params
      :repo "elastic/kibana"
      :results-dir "results/kibana")))
```

The `(params ...)` form is new — keyword pairs inside `(defaults ...)`. The parser needs to handle this new child form.

## API Changes

### `GET /api/workspace` (updated response)

```json
{
  "name": "stokagent",
  "description": "...",
  "owner": "adam",
  "repos": ["elastic/kibana", "..."],
  "defaults": {
    "model": "qwen2.5:7b",
    "provider": "ollama",
    "elasticsearch": "http://localhost:9200",
    "params": {
      "repo": "elastic/kibana",
      "results-dir": "results/kibana"
    }
  }
}
```

### `PUT /api/workspace` (new)

Accepts the same JSON shape as the GET response. Serializes to s-expression and writes `workspace.glitch`.

### `GET /api/providers` (new)

Returns list of configured provider names from the provider registry. Used to populate the provider dropdown. If the provider registry directory doesn't exist or is empty, returns an empty array and the frontend renders a freeform text input instead.

```json
["ollama", "openai", "anthropic"]
```

## RunDialog Default Parameter Pre-fill

When RunDialog opens:
1. Fetch workspace defaults (or use cached value from a prior fetch)
2. For each workflow parameter, check if a default exists in `workspace.defaults.params`
3. Pre-fill the input with the default value
4. User can override by typing in the field
5. Existing `autoParams` (from results browser context) take priority over workspace defaults

Priority order: `autoParams` > `workspace defaults` > empty

## Playwright Tests

### New file: `gui/e2e/settings.spec.js`

```
Settings page
  - navigates to settings via sidebar
  - sidebar settings link shows active state on /settings
  - page loads without JS errors
  - shows Workflow Defaults section
  - shows Workspace Config section
  - displays current workspace name
  - displays current default model
  - displays current default provider
  - displays Kibana URL field
  - displays repositories list
  - save button is disabled when no changes made
  - editing a field enables save button
  - saving workspace config persists and reloads correctly
  - adding a default parameter shows key-value row
  - removing a default parameter removes the row
  - invalid Kibana URL shows validation feedback
  - breadcrumbs show Settings label
  - no JS errors during interaction
```

### Updates to existing tests

- `gui.spec.js` — update sidebar nav count from 3 to 4 (Settings is now a nav item)
- Add test: RunDialog pre-fills defaults from workspace config

## Out of Scope

- Theme/font/editor display preferences
- Per-workflow setting overrides (handled in workflow metadata)
- Authentication or user management
- Import/export of settings
- Migration or backwards-compat for old workspace files (pre-1.0 policy: wipe and restart)
