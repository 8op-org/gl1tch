# glitch-github Plugin Design

Personal GitHub activity plugin for gl1tch. Replaces inline GraphQL + jq
plumbing in stokagent workflows with clean subcommands that output JSON.

## Problem

Five workflows in stokagent (`dashboard-activity`, `activity-report`,
`review-dashboard`, `morning-briefing`, `task-list`) copy-paste the same
GraphQL queries, jq transforms, repo filter lists, and date math. The
`morning-briefing.yaml` duplicates the entire `review-dashboard.yaml` fetch
layer verbatim. Adding a repo or changing the username means editing every
file.

## Solution

A Go binary plugin (`glitch-github`) that encapsulates GitHub data fetching
behind subcommands. Workflows call it like any shell command. The plugin
handles GraphQL, JSON shaping, date math, and repo filtering internally.

## Naming

- **Repo:** `gl1tch-github` (per gl1tch repo naming convention)
- **Binary:** `glitch-github` (per gl1tch binary naming convention)
- **CLI usage:** `glitch github <subcommand>` (auto-discovered by gl1tch plugin system)

## Subcommands

All subcommands output JSON to stdout. Exit 0 on success, exit 1 on error
with a message to stderr.

### `glitch-github prs`

Fetch pull requests.

```
glitch-github prs --authored --since yesterday    # PRs you authored
glitch-github prs --reviewing --since yesterday   # PRs requesting your review
```

Flags:
- `--authored` — PRs authored by you (created or merged in range, deduplicated by PR number)
- `--reviewing` — open PRs requesting your review
- `--since` — time range (default: `yesterday`)

Output shape (authored):
```json
[
  {
    "number": 123,
    "title": "Fix flaky test",
    "url": "https://github.com/elastic/ensemble/pull/123",
    "repo": "elastic/ensemble",
    "state": "MERGED",
    "additions": 10,
    "deletions": 5,
    "created_at": "2026-04-09T14:00:00Z",
    "merged_at": "2026-04-09T16:00:00Z"
  }
]
```

Output shape (reviewing):
```json
[
  {
    "number": 456,
    "title": "Add retry logic",
    "url": "https://github.com/elastic/oblt-cli/pull/456",
    "repo": "elastic/oblt-cli",
    "author": "coworker",
    "additions": 30,
    "deletions": 12,
    "created_at": "2026-04-08T10:00:00Z"
  }
]
```

### `glitch-github reviews`

Fetch reviews you gave on other people's PRs.

```
glitch-github reviews --since yesterday
```

Output shape:
```json
[
  {
    "number": 789,
    "title": "Refactor pipeline",
    "url": "https://github.com/elastic/observability-robots/pull/789",
    "repo": "elastic/observability-robots",
    "author": "teammate",
    "review_state": "APPROVED"
  }
]
```

`review_state` is the latest review you submitted in the time range:
`APPROVED`, `CHANGES_REQUESTED`, or `COMMENTED`.

Self-authored PRs are excluded (you don't review your own PRs).

### `glitch-github issues`

Fetch issues.

```
glitch-github issues --assigned                  # open issues assigned to you
glitch-github issues --closed --since yesterday  # issues you closed
```

Flags:
- `--assigned` — open issues assigned to you
- `--closed` — issues closed in range
- `--since` — time range (default: `yesterday`, only used with `--closed`)

Output shape (assigned):
```json
[
  {
    "number": 100,
    "title": "Investigate memory leak",
    "url": "https://github.com/elastic/ensemble/issues/100",
    "repo": "elastic/ensemble",
    "created_at": "2026-04-01T09:00:00Z",
    "labels": ["bug", "P1"]
  }
]
```

Output shape (closed):
```json
[
  {
    "number": 101,
    "title": "Update CI config",
    "url": "https://github.com/elastic/oblt-cli/issues/101",
    "repo": "elastic/oblt-cli"
  }
]
```

### `glitch-github mentions`

Fetch issues and PRs where you were mentioned.

```
glitch-github mentions --since 7d
```

Output shape:
```json
[
  {
    "number": 200,
    "title": "Deploy schedule change",
    "url": "https://github.com/elastic/observability-test-environments/issues/200",
    "repo": "elastic/observability-test-environments",
    "type": "issue",
    "updated_at": "2026-04-09T11:00:00Z"
  }
]
```

### `glitch-github activity`

Combined dashboard feed. Calls the authored-prs, reviews, and closed-issues
queries internally and merges into one response.

```
glitch-github activity --since yesterday
```

Output shape:
```json
{
  "authored": [...],
  "reviews": [...],
  "closed": [...]
}
```

This matches the shape that `dashboard-activity.yaml` currently produces,
so `build.sh` needs no changes beyond swapping the workflow content.

## `--since` Flag

Accepts:
- `yesterday` — previous calendar day (in US/Eastern)
- `week` — Monday of current week through today
- `Nd` — last N days (e.g. `7d`, `30d`)
- `YYYY-MM-DD` — specific start date through today

All date math uses US/Eastern timezone. The end date is always today.

## Baked-in Defaults

These are Go constants, not config:

```go
const (
    username = "adam-stokes"
    timezone = "US/Eastern"
)

var repos = []string{
    "elastic/observability-test-environments",
    "elastic/observability-robots",
    "elastic/oblt-cli",
    "elastic/ensemble",
}
```

To change repos or username, edit the source and cut a new release.

## GitHub Access

Shells out to `gh api graphql` for all queries. This means:
- Auth handled by `gh auth` (no tokens in plugin)
- GraphQL queries are Go string constants with `fmt.Sprintf` for dates
- JSON response parsing and reshaping done in Go (`encoding/json`)
- No dependency on `jq`

## Error Handling

- `gh` not on PATH: exit 1, print `"glitch-github: gh CLI not found — install from https://cli.github.com"`
- `gh` not authenticated: exit 1, forward `gh` error message
- GitHub rate limit: exit 1, forward `gh` error message
- Empty results: output valid JSON with empty arrays (never null, never exit 1)
- Invalid `--since` value: exit 1, print usage hint

## Project Structure

```
gl1tch-github/
├── main.go              # cobra root + subcommands wiring
├── github.go            # GraphQL queries, JSON shaping, repo filtering
├── dateparse.go         # --since flag parsing
├── go.mod
├── .goreleaser.yaml
└── .github/workflows/release.yml
```

## Workflow Impact

### dashboard-activity.yaml (74 lines → 4)

Before:
```yaml
steps:
  - id: fetch
    run: |
      YESTERDAY=$(TZ='US/Eastern' date -v-1d +%Y-%m-%d)
      # ... 70 lines of GraphQL + jq ...
```

After:
```yaml
steps:
  - id: fetch
    run: glitch-github activity --since yesterday
```

### review-dashboard.yaml (66 lines → 28)

Before: 3 fetch steps, each with inline GraphQL + jq (~20 lines each)

After:
```yaml
steps:
  - id: fetch-reviews
    run: glitch-github prs --reviewing
  - id: fetch-issues
    run: glitch-github issues --assigned
  - id: fetch-mentions
    run: glitch-github mentions --since 7d
  - id: format
    llm:
      prompt: ...
```

### morning-briefing.yaml (148 lines → ~50)

Stops duplicating review-dashboard fetch steps. Same one-liner calls.

### activity-report.yaml (114 lines → ~50)

Fetch steps become one-liners. Date-range step stays (daily vs weekly
routing is workflow logic). LLM format step stays.

## Release

Standard gl1tch plugin release pipeline:
- GoReleaser builds darwin-arm64 and darwin-amd64
- GitHub Actions tags trigger release
- Homebrew tap formula auto-updated via GoReleaser
- Install: `brew install 8op-org/tap/glitch-github`
