# Research System v2 — Tool-Use Loop with Tiered Escalation

**Date:** 2026-04-13
**Status:** Draft
**Author:** adam-stokes + claude

## Problem

The current research system picks researchers from a menu via a single LLM call, gathers evidence in parallel, and drafts a response. This produces shallow, generic results because:

- The local LLM planner picks one researcher and stops
- YAML researchers don't receive repo context
- Research events go to a local JSONL file, not Elasticsearch
- Pipeline runs and LLM calls aren't tracked (no cost/token visibility)
- The system only handles GitHub issues — not PRs, Google Docs, or other sources

## Goals

1. Point glitch at any input source (issue, PR, Google Doc) and get a useful research summary with links to fuller context
2. Ask glitch to produce an implementation plan + patch in `results/` that the user can manually hand to claude/copilot for testing
3. Track every LLM call, tool call, and research run in Elasticsearch so Kibana dashboards show cost trends, escalation rates, and tool effectiveness
4. Use the cheapest LLM tier that works, escalating from local → free-tier paid → paid only when needed

## Non-Goals

- Autonomous PR creation (future work)
- Replacing direct workflows (morning-briefing, dashboards, etc.) — those stay as pipeline YAML
- Real-time streaming of research progress to a UI

---

## Architecture Overview

```
User Input (URL or question)
    │
    ▼
Input Adapter (detect source, fetch content, normalize)
    │
    ▼
ResearchDocument (title, body, repo, links, metadata)
    │
    ▼
Tool-Use Loop (LLM selects tools iteratively)
    │   ├── grep_code, read_file, git_log, git_diff
    │   ├── search_es, list_files
    │   ├── fetch_issue, fetch_pr
    │   └── (every call logged to ES)
    │
    ▼
LLM Escalation (tier 0 → 1 → 2 if confidence low)
    │
    ▼
Output (summary.md, evidence/, run.json, implementation/)
    │
    ▼
Elasticsearch (research-runs, tool-calls, llm-calls indices)
    │
    ▼
Kibana Dashboards (cost, escalation, tool effectiveness)
```

---

## Section 1: Input Adapters

Every input source gets normalized into a ResearchDocument before the research loop touches it.

### ResearchDocument

```go
type ResearchDocument struct {
    Source    string            // "github-issue", "github-pr", "google-doc"
    SourceURL string           // original URL
    Title     string           // extracted title
    Body      string           // full text content
    Repo      string           // "elastic/ensemble" (if applicable)
    RepoPath  string           // local clone path (resolved by EnsureRepo)
    Metadata  map[string]string // source-specific extras (labels, author, state)
    Links     []Link           // related URLs from the source
}

type Link struct {
    URL   string
    Label string
}
```

### Adapters

Adapters are functions: `func(url string) (ResearchDocument, error)`. Three to start:

- **GitHubIssueAdapter** — `gh issue view <url> --json number,title,body,comments,labels,assignees` → parse into ResearchDocument, extract linked PRs/issues into Links
- **GitHubPRAdapter** — `gh pr view <url> --json number,title,body,comments,files,additions,deletions,reviews` → parse, include diff stats and changed file list in Metadata
- **GoogleDocAdapter** — `gws docs get <id>` → parse document content into Body

### Detection

URL pattern matching, not LLM routing:
- `github.com/*/issues/*` → GitHubIssueAdapter
- `github.com/*/pull/*` → GitHubPRAdapter
- `docs.google.com/*` → GoogleDocAdapter
- No match → treat raw text as Body with no adapter

---

## Section 2: Tool-Use Research Loop

The current plan→gather→draft cycle is replaced with an iterative tool-use loop. The LLM calls tools one at a time until it has enough evidence to produce a final output.

### Tools

| Tool | Input | Output | Cost |
|------|-------|--------|------|
| `grep_code` | pattern, path (optional), file_glob (optional) | matching lines with file:line context | free |
| `read_file` | path, start_line (optional), end_line (optional) | file content | free |
| `git_log` | query (optional), path (optional), limit | commit list | free |
| `git_diff` | ref1, ref2 (optional), path (optional) | diff output | free |
| `search_es` | query, index (optional) | ES search results | free |
| `list_files` | path, depth | directory tree | free |
| `fetch_issue` | repo, number | issue JSON | free |
| `fetch_pr` | repo, number | PR JSON with diff stats | free |

All tools execute locally (shell commands, ES queries). No LLM inside any tool. Tools always run against `ResearchDocument.RepoPath` as the working directory.

### Loop Contract

```
1. LLM receives: ResearchDocument + tool descriptions + goal ("summarize" or "implement")
2. LLM emits: tool call (e.g., grep_code {pattern: "error", path: "cmd/"})
3. Glitch executes tool, returns result to LLM
4. LLM decides: call another tool, or produce final output
5. Repeat until: LLM says done, or budget exhausted (max 15 tool calls)
```

### Budget

- Max 15 tool executions per run (tool execution is cheap — local commands). LLM calls to select tools do not count toward this budget.
- LLM calls are the expensive part and tracked separately in ES
- If tool budget exhausted, LLM must produce output with whatever evidence it has

### What This Replaces

- The `Researcher` interface (GitResearcher, FSResearcher, ESResearcher, YAMLResearcher) — removed
- The plan→gather→draft→score loop — removed
- YAML researchers — removed (YAML workflows stay for direct pipelines only)
- PlanPrompt, DraftPrompt, CritiquePrompt — replaced by tool-use system prompt

---

## Section 3: LLM Escalation Tiers

Every LLM call goes through a tiered provider chain. Cheapest first, escalate on failure.

### Tier Definitions

| Tier | Providers | Cost | When |
|------|-----------|------|------|
| 0 | Local Ollama (qwen3:8b) | Free | First attempt at everything |
| 1 | Codex, Gemini (Vertex free tier) | Free | Tier 0 returned low-confidence or malformed output |
| 2 | Copilot, Claude | Paid | Tier 1 failed or task requires synthesis/implementation |

### Confidence Detection

Structural checks, not LLM self-assessment:

- Did the LLM return a valid tool call or valid structured output? (parseable JSON)
- Did it hallucinate a tool name that doesn't exist?
- Did it produce an empty or trivially short response for a synthesis task?
- For the final draft: does the output reference actual evidence from tool calls, or is it generic?

If any check fails → escalate one tier. If tier 2 also fails → return best attempt with a warning.

### Configuration

`~/.config/glitch/config.yaml`:

```yaml
tiers:
  - providers: [ollama]
    model: qwen3:8b
  - providers: [codex, gemini]
  - providers: [copilot, claude]
```

User can reorder, remove tiers, or pin a provider. Each tier tries its providers in order until one succeeds.

### Tracking

Every LLM call logs to ES:
- Tier used, provider, model
- Tokens in/out (parsed from provider response where available)
- Latency
- Whether it escalated and why
- Cost estimate (token counts x known pricing)

---

## Section 4: ES Telemetry & Kibana

Every meaningful action gets indexed into Elasticsearch.

### New Indices

**`glitch-research-runs`** — one doc per research session

| Field | Type | Description |
|-------|------|-------------|
| run_id | keyword | unique run identifier |
| input_source | keyword | "github-issue", "github-pr", "google-doc" |
| source_url | keyword | original URL |
| goal | keyword | "summarize" or "implement" |
| total_tool_calls | integer | number of tool invocations |
| total_llm_calls | integer | number of LLM invocations |
| total_tokens_in | long | sum of input tokens across all LLM calls |
| total_tokens_out | long | sum of output tokens across all LLM calls |
| total_cost_usd | float | estimated total cost |
| duration_ms | long | wall clock time |
| final_tier_used | integer | highest tier reached (0, 1, 2) |
| escalation_count | integer | number of tier escalations |
| confidence_pass | boolean | did the final output pass confidence checks |
| timestamp | date | when the run started |

**`glitch-tool-calls`** — one doc per tool call

| Field | Type | Description |
|-------|------|-------------|
| run_id | keyword | parent run |
| tool_name | keyword | grep_code, read_file, etc. |
| input_summary | text | truncated input params |
| output_size_bytes | integer | size of tool output |
| latency_ms | long | execution time |
| success | boolean | did the tool return results |
| timestamp | date | when called |

**`glitch-llm-calls`** — one doc per LLM invocation

| Field | Type | Description |
|-------|------|-------------|
| run_id | keyword | parent run |
| step | keyword | tool_select, draft, synthesize |
| tier | integer | 0, 1, or 2 |
| provider | keyword | ollama, codex, claude, etc. |
| model | keyword | specific model used |
| tokens_in | long | input tokens |
| tokens_out | long | output tokens |
| cost_usd | float | estimated cost |
| latency_ms | long | response time |
| escalated | boolean | did this call trigger escalation |
| escalation_reason | keyword | malformed_output, empty_response, hallucinated_tool, etc. |
| timestamp | date | when called |

### Index Lifecycle

The existing `glitch-events`, `glitch-code-*` indices remain unchanged. The unused `glitch-pipelines`, `glitch-summaries`, `glitch-insights` mappings are replaced by the three new indices above.

### Kibana Dashboards

Auto-provisioned NDJSON files in `deploy/kibana/`, imported via Kibana saved objects API on `glitch up`.

| Dashboard | Panels |
|-----------|--------|
| Research Overview | Runs over time, success rate, avg tool calls per run, runs by input source |
| Cost & Tokens | Total spend over time, tokens by provider, cost per run trending down |
| Escalation Funnel | % of calls at each tier, escalation reasons breakdown, tier distribution over time |
| Tool Effectiveness | Tool call frequency, which tools produce evidence cited in drafts, avg output size |
| Provider Comparison | Latency by provider, token efficiency, success rate, cost per successful call |

---

## Section 5: Results Output & User Flow

### Directory Structure

```
results/<org>/<repo>/<issue-number>/
├── summary.md          # human-readable summary with links
├── evidence/           # raw tool call outputs
│   ├── 001-grep_code.txt
│   ├── 002-read_file.txt
│   └── ...
├── run.json            # full run metadata (run_id, tiers, cost)
└── implementation/     # only when goal=implement
    ├── plan.md         # what files to change and why
    └── patch.diff      # proposed changes
```

### summary.md Format

```markdown
# <Issue Title> (#<number>)

<2-3 paragraph summary of what the issue asks for and what glitch found>

## Key Findings
- <specific finding with file:line references>
- <related PRs or prior work>

## Links
- [Issue](<github url>)
- [<file>:<lines>](<github permalink>)
- [Research run in Kibana](<kibana discover link with run_id>)
- [Full evidence](./evidence/)

## Cost
Tier 0 (ollama): N calls, Nk tokens, $0.00
Tier 2 (claude): N calls, Nk tokens, ~$X.XX
```

### User Flow

1. **`glitch ask "<issue-url>"`** — adapter fetches + normalizes, tool-use loop researches, summary printed to stdout and saved to `results/`
2. **User reads summary**, follows links for deeper context (GitHub, Kibana, local evidence files)
3. **`glitch ask "implement a fix for <issue-ref>"`** — glitch loads prior research from `results/`, runs tool-use loop with goal=implement, produces `implementation/plan.md` and `patch.diff`
4. **User takes output to claude/copilot** manually, tests, iterates
5. **(Future)** `glitch ask "submit fix for <issue-ref>"` — creates branch, commits, opens PR

Step 3 picks up prior research — it does not re-fetch the issue or re-grep the codebase. The existing `results/` directory serves as context for the implementation run.

---

## Migration Notes

### What Gets Removed
- `internal/research/researcher.go` (Researcher interface)
- `internal/research/git_researcher.go`, `fs_researcher.go`, `es_researcher.go`
- `internal/research/yaml_researcher.go`
- `internal/research/prompts.go` (PlanPrompt, DraftPrompt, CritiquePrompt, JudgePrompt)
- `internal/research/score.go` (ComputeScore, critique system)
- `internal/research/registry.go` (Registry)
- `researchers/` directory (YAML researchers)
- `~/.config/glitch/researchers/` (user YAML researchers)

### What Stays
- `internal/pipeline/` — unchanged, still powers direct workflows
- `~/.config/glitch/workflows/` — unchanged
- `internal/router/` — updated to detect input sources and route to research vs workflows
- `internal/esearch/` — extended with new index mappings
- `internal/provider/` — extended with tier support and token tracking
- `internal/research/loop.go` — rewritten as tool-use loop
- `internal/research/types.go` — updated with ResearchDocument, new Result format
- `internal/research/events.go` — replaced with ES-native telemetry
- `internal/research/repo.go` — unchanged (EnsureRepo, ParseRepoFromQuestion)
- `internal/research/results.go` — updated for new directory structure
- `cmd/ask.go` — updated to use adapters + new loop
- `cmd/research_helpers.go` — rewritten to build tool-use loop

### What Gets Added
- `internal/research/adapter.go` — input adapters
- `internal/research/tools.go` — tool definitions and execution
- `internal/research/toolloop.go` — tool-use loop engine
- `internal/provider/tiers.go` — tiered provider escalation
- `internal/provider/tokens.go` — token counting and cost estimation
- `internal/esearch/telemetry.go` — research run/tool/llm call indexing
- `deploy/kibana/` — dashboard NDJSON files
- Updated `docker-compose.yml` or `glitch up` to import dashboards
