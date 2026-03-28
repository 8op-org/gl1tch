## ADDED Requirements

### Requirement: Fetch recent GitHub issues from both repos
The pipeline SHALL fetch issues updated within the last 3 days from `elastic/ensemble` and `elastic/observability-robots` using `gh issue list` with JSON output fields: `number`, `title`, `state`, `updatedAt`, `labels`, `assignees`, `url`.

#### Scenario: Issues fetched successfully
- **WHEN** the pipeline runs with valid `gh` authentication
- **THEN** steps `fetch-issues-ensemble` and `fetch-issues-robots` each return a JSON array of issues

#### Scenario: No recent issues
- **WHEN** no issues were updated in the last 3 days
- **THEN** the step returns an empty JSON array `[]` and the pipeline continues without error

### Requirement: Fetch open PRs from both repos
The pipeline SHALL fetch open pull requests from both repos using `gh pr list` with JSON output fields: `number`, `title`, `state`, `updatedAt`, `labels`, `reviewDecision`, `isDraft`, `url`.

#### Scenario: PRs fetched successfully
- **WHEN** the pipeline runs with valid `gh` authentication
- **THEN** steps `fetch-prs-ensemble` and `fetch-prs-robots` each return a JSON array of open PRs

### Requirement: Merge and tag data by repo
A `jq` step SHALL merge all issue and PR arrays into two unified objects — one for issues, one for PRs — tagging each item with its source `repo` field.

#### Scenario: Data merged with repo tags
- **WHEN** both fetch steps succeed
- **THEN** `merge-issues` produces `[{"repo":"elastic/ensemble",...}, {"repo":"elastic/observability-robots",...}]`

### Requirement: opencode LLM enrichment pass
A pipeline step SHALL pass the merged issues and PRs to `opencode` with the `ollama/qwen2.5:latest` model, prompting it to identify stale items (no activity >2 days), items needing review, and any cross-repo patterns.

#### Scenario: Enrichment returns narrative
- **WHEN** Ollama is running with qwen2.5 loaded
- **THEN** step `enrich` returns a free-text narrative summary (non-empty string)

#### Scenario: Ollama unavailable
- **WHEN** Ollama is not running
- **THEN** a `builtin.assert` step fails with message "opencode enrichment returned empty — is Ollama running?"

### Requirement: Claude Haiku produces structured JSON digest
A pipeline step SHALL pass the enrichment narrative and raw counts to `claude-haiku-4-5-20251001`, prompting it to return strict JSON matching this schema:

```json
{
  "generated_at": "<ISO8601>",
  "repos": ["elastic/ensemble", "elastic/observability-robots"],
  "issues": {
    "total_open": <int>,
    "updated_last_3d": <int>,
    "needs_attention": [{"repo": str, "number": int, "title": str, "reason": str, "url": str}]
  },
  "prs": {
    "total_open": <int>,
    "draft": <int>,
    "awaiting_review": <int>,
    "needs_attention": [{"repo": str, "number": int, "title": str, "reason": str, "url": str}]
  },
  "summary": "<one paragraph narrative>",
  "action_items": ["<imperative string>"]
}
```

#### Scenario: Valid JSON digest produced
- **WHEN** Claude Haiku returns a response
- **THEN** a downstream `jq` step successfully parses `.issues.total_open` and `.prs.total_open` without error

#### Scenario: Claude API key missing
- **WHEN** `ANTHROPIC_API_KEY` is not set
- **THEN** the Claude step fails and the pipeline surfaces an error message

### Requirement: Final jq validation and pretty-print
The last pipeline step SHALL pipe the Haiku output through `jq '.'` to validate it is parseable JSON and pretty-print it to stdout.

#### Scenario: Valid JSON passes through
- **WHEN** Haiku returns valid JSON
- **THEN** `jq '.'` outputs indented JSON and exits 0

#### Scenario: Invalid JSON caught
- **WHEN** Haiku returns malformed JSON (e.g., trailing narrative text)
- **THEN** `jq '.'` exits non-zero, triggering a `builtin.assert` failure with the raw output visible
