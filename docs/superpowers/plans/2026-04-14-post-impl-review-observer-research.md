# Post-Impl Review + Observer Context + Research Fallback

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close three gaps — (1) add post-implementation code review to the issue-to-pr pipeline, (2) give the observer repo/index awareness so answers are grounded, (3) wire `glitch ask` fallback to the research tool loop when no workflow matches.

**Architecture:** All three changes follow the same pattern: extend existing systems (workflows, observer, ask command) rather than introducing new subsystems. Post-impl review is a new workflow file. Observer gets repo context via a new `--repo` flag and index metadata query. Research fallback is a 15-line addition to `cmd/ask.go` that calls the existing `ToolLoop`.

**Tech Stack:** Go, Elasticsearch, Ollama (qwen2.5:7b), gh CLI, existing pipeline runner

---

## File Map

| Task | File | Action | Responsibility |
|------|------|--------|----------------|
| 1 | `~/.config/glitch/workflows/post-impl-review.glitch` | Create | Workflow that reviews a PR diff against its issue after implementation |
| 1 | `internal/pipeline/runner_test.go` | Modify | Test parseWorkflowName handles new workflow name |
| 2 | `internal/observer/query.go` | Modify | Accept repo context, add IndexStats method, enrich synthesize prompt |
| 2 | `internal/observer/query_test.go` | Modify | Tests for IndexStats parsing and repo-scoped query building |
| 2 | `internal/esearch/client.go` | Modify | Add IndexStats method (GET /_cat/indices) |
| 2 | `internal/esearch/client_test.go` | Modify | Test IndexStats response parsing |
| 2 | `cmd/observe.go` | Modify | Add --repo flag, pass to QueryEngine |
| 3 | `cmd/ask.go` | Modify | Wire research loop fallback when no workflow matches |
| 3 | `cmd/ask.go` | Modify | Test: verify fallback is reachable (unit test parseability) |

---

### Task 1: Post-Implementation Review Workflow

The existing `pr-review.glitch` reviews PRs standalone. This new workflow is designed to be run *after* the issue-to-pr workflow — it takes an issue number and PR number and reviews the code against the issue requirements, checking for the specific bugs we've seen: API compat, convention drift, test quality.

**Files:**
- Create: `~/.config/glitch/workflows/post-impl-review.glitch`

- [ ] **Step 1: Write the post-impl-review workflow**

```lisp
(def provider "claude")
(def model "sonnet")

(workflow "post-impl-review"
  :description "Review a PR's code against its issue — catch API compat, convention, and test quality bugs"

  (step "fetch-issue"
    (run "gh issue view {{.param.issue}} --repo {{.param.repo}} --json number,title,body,labels,comments"))

  (step "fetch-pr"
    (run "gh pr view {{.param.pr}} --repo {{.param.repo}} --json number,title,body,files,additions,deletions,state"))

  (step "fetch-diff"
    (run "gh pr diff {{.param.pr}} --repo {{.param.repo}} | head -3000"))

  (step "repo-conventions"
    (run "REPO_NAME=$(echo '{{.param.repo}}' | cut -d/ -f2); REPO_PATH=\"$HOME/Projects/$REPO_NAME\"; if [ ! -d \"$REPO_PATH\" ]; then echo 'repo not cloned locally'; exit 0; fi; echo '=== LANGUAGE & DEPS ==='; for f in go.mod pyproject.toml package.json Gemfile; do if [ -f \"$REPO_PATH/$f\" ]; then echo \"--- $f ---\"; head -30 \"$REPO_PATH/$f\"; echo; fi; done; echo '=== TEST PATTERNS ==='; find \"$REPO_PATH\" -name '*_test.go' -o -name 'test_*.py' -o -name '*.test.ts' 2>/dev/null | head -20; echo; echo '=== CI CONFIG ==='; for f in .github/workflows/*.yml Makefile Jenkinsfile; do if [ -f \"$REPO_PATH/$f\" ]; then echo \"--- $f ---\"; head -20 \"$REPO_PATH/$f\"; echo; fi; done"))

  (step "review"
    (llm :provider provider :model model
      :prompt ```
You are a post-implementation code reviewer. Your job is to catch bugs that self-review misses.

CONTEXT: An AI agent implemented this PR from the issue below. Plan review passed, but the CODE may have bugs. You are the safety net.

ISSUE (what was requested):
{{step "fetch-issue"}}

PR METADATA:
{{step "fetch-pr"}}

REPO CONVENTIONS:
{{step "repo-conventions"}}

CODE DIFF:
{{step "fetch-diff"}}

Review with HIGH RIGOR against these criteria:

1. API COMPATIBILITY
   - Are all API calls using endpoints/methods that exist in the target version?
   - Any deprecated APIs? Any removed auth methods (e.g. http_auth in ES 9.x)?
   - Are SDK versions compatible with the code?

2. DATA HANDLING
   - Bulk operations used where appropriate (not single-doc indexing in loops)?
   - Document IDs unique (no collision risk)?
   - Error handling on API responses (not just connection errors)?

3. REPO CONVENTIONS
   - Does the code match existing patterns in the repo?
   - Python version, dependency versions, project structure consistent?
   - Naming conventions followed?

4. TEST QUALITY
   - Do tests verify real behavior or just mock everything?
   - Are edge cases covered (empty input, API errors, malformed data)?
   - Would these tests catch a regression?

5. MISSING REQUIREMENTS
   - Does the PR address ALL acceptance criteria from the issue?
   - Any requirements mentioned in issue comments that were missed?

OUTPUT FORMAT:

For each criterion: criterion name — PASS or FAIL — one-line reason
If FAIL, include: file path, what's wrong, what the fix should be

OVERALL: PASS or FAIL

If FAIL, list the specific changes needed to fix it, in priority order.
```))

  (step "save-review"
    (run "RESULTS_DIR=\"results/{{.param.issue}}\"; mkdir -p \"$RESULTS_DIR\"; cp '{{stepfile \"review\"}}' \"$RESULTS_DIR/post-impl-review.md\"; echo 'Post-impl review saved: '\"$RESULTS_DIR/post-impl-review.md\"")))
```

Save this file to `~/.config/glitch/workflows/post-impl-review.glitch`.

- [ ] **Step 2: Verify the workflow parses correctly**

Run: `go run . workflow list 2>&1 | grep post-impl`
Expected: `post-impl-review` appears in the workflow list

- [ ] **Step 3: Test the workflow against a real PR**

Pick a recent PR from a known repo and run:
```bash
go run . ask "review elastic/observability-robots#<pr-number>" --variant ""
```
Or directly:
```bash
go run . workflow run post-impl-review --param repo=elastic/observability-robots --param issue=<issue> --param pr=<pr>
```
Expected: Review output with PASS/FAIL per criterion.

- [ ] **Step 4: Wire issue-to-pr handoff**

The issue-to-pr workflow's final step already prints "To create the PR: claude ...". Update the handoff message in `cmd/ask.go` to also suggest running the post-impl review after the PR is created.

In `cmd/ask.go:runSingleIssue()`, after the existing handoff message (around line 202):

```go
fmt.Printf("\nAfter creating the PR, run post-impl review:\n")
fmt.Printf("  glitch workflow run post-impl-review --param repo=%s --param issue=%s --param pr=<PR_NUMBER>\n", repo, issue)
```

- [ ] **Step 5: Commit**

```bash
git add cmd/ask.go
git commit -m "feat: add post-impl review workflow and handoff hint"
```

Note: The workflow file lives in ~/.config/glitch/workflows/ (user-global), not in the repo. Copy to the repo's examples/ if desired.

---

### Task 2: Observer Repo Context

The observer's `synthesize` prompt only gets raw ES hits — no awareness of which repos exist or what's been indexed. We add: (a) an ES index stats query so the observer knows what data exists, (b) a `--repo` flag so the user can scope queries, (c) enriched synthesize prompt.

**Files:**
- Modify: `internal/esearch/client.go`
- Modify: `internal/esearch/client_test.go` (create if needed)
- Modify: `internal/observer/query.go`
- Modify: `internal/observer/query_test.go`
- Modify: `cmd/observe.go`

- [ ] **Step 1: Write the failing test for IndexStats**

In `internal/esearch/client_test.go`:

```go
package esearch

import (
	"testing"
)

func TestParseIndexStats(t *testing.T) {
	// Simulates /_cat/indices?format=json response
	raw := `[
		{"index":"glitch-events","docs.count":"150","store.size":"1mb"},
		{"index":"glitch-llm-calls","docs.count":"42","store.size":"512kb"},
		{"index":".kibana_1","docs.count":"10","store.size":"100kb"}
	]`

	stats := parseIndexStats([]byte(raw))

	if len(stats) != 2 {
		t.Fatalf("expected 2 glitch indices, got %d", len(stats))
	}

	found := false
	for _, s := range stats {
		if s.Index == "glitch-events" && s.DocCount == "150" {
			found = true
		}
	}
	if !found {
		t.Error("expected glitch-events with 150 docs")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/esearch/ -run TestParseIndexStats -v`
Expected: FAIL — `parseIndexStats` not defined

- [ ] **Step 3: Implement IndexStats in esearch client**

In `internal/esearch/client.go`, add:

```go
// IndexStat holds basic metadata about an ES index.
type IndexStat struct {
	Index    string `json:"index"`
	DocCount string `json:"docs.count"`
	StoreSize string `json:"store.size"`
}

// IndexStats returns doc counts for all glitch-* indices.
func (c *Client) IndexStats(ctx context.Context) ([]IndexStat, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/_cat/indices/glitch-*?format=json&h=index,docs.count,store.size", nil)
	if err != nil {
		return nil, fmt.Errorf("index stats: build request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("index stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("index stats: status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("index stats: read body: %w", err)
	}

	return parseIndexStats(body), nil
}

// parseIndexStats parses /_cat/indices JSON, filtering to glitch-* indices.
func parseIndexStats(data []byte) []IndexStat {
	var raw []IndexStat
	json.Unmarshal(data, &raw)

	var filtered []IndexStat
	for _, s := range raw {
		if strings.HasPrefix(s.Index, "glitch-") {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/esearch/ -run TestParseIndexStats -v`
Expected: PASS

- [ ] **Step 5: Write failing test for repo-scoped query building**

In `internal/observer/query_test.go`, add:

```go
func TestDefaultQueryWithRepo(t *testing.T) {
	q := defaultQueryWithRepo("what broke today", "elastic/ensemble")
	raw, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Should contain a repo filter
	s := string(raw)
	if !strings.Contains(s, "elastic/ensemble") {
		t.Error("expected repo filter in query")
	}
	if !strings.Contains(s, "filter") {
		t.Error("expected bool filter clause")
	}
}

func TestDefaultQueryWithRepoEmpty(t *testing.T) {
	// Empty repo should produce the same query as defaultQuery
	q := defaultQueryWithRepo("what broke today", "")
	raw, _ := json.Marshal(q)
	qOrig := defaultQuery("what broke today")
	rawOrig, _ := json.Marshal(qOrig)

	if string(raw) != string(rawOrig) {
		t.Error("empty repo should produce same query as defaultQuery")
	}
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `go test ./internal/observer/ -run TestDefaultQueryWithRepo -v`
Expected: FAIL — `defaultQueryWithRepo` not defined

- [ ] **Step 7: Implement repo-scoped query and enriched synthesize**

In `internal/observer/query.go`:

1. Add `repo` field to `QueryEngine`:

```go
type QueryEngine struct {
	es    *esearch.Client
	model string
	repo  string
}

func NewQueryEngine(es *esearch.Client, model string) *QueryEngine {
	if model == "" {
		model = defaultModel
	}
	return &QueryEngine{es: es, model: model}
}

// WithRepo returns a copy of the engine scoped to a repo.
func (q *QueryEngine) WithRepo(repo string) *QueryEngine {
	return &QueryEngine{es: q.es, model: q.model, repo: repo}
}
```

2. Add `defaultQueryWithRepo`:

```go
func defaultQueryWithRepo(question, repo string) map[string]any {
	q := defaultQuery(question)
	if repo == "" {
		return q
	}
	// Add repo filter to the bool query
	boolQ := q["query"].(map[string]any)["bool"].(map[string]any)
	boolQ["filter"] = []any{
		map[string]any{
			"bool": map[string]any{
				"should": []any{
					map[string]any{"term": map[string]any{"repo": repo}},
					map[string]any{"term": map[string]any{"source": repo}},
				},
				"minimum_should_match": 1,
			},
		},
	}
	return q
}
```

3. Update `searchWithFallback` to use repo-scoped fallback:

```go
func (q *QueryEngine) searchWithFallback(ctx context.Context, question string) (*esearch.SearchResponse, error) {
	esQuery, err := q.generateQuery(ctx, question)
	if err != nil {
		esQuery = defaultQueryWithRepo(question, q.repo)
	}
	// ... rest unchanged
}
```

4. Update `synthesize` to include index context:

```go
func (q *QueryEngine) synthesize(ctx context.Context, question string, results *esearch.SearchResponse) (string, error) {
	formatted := formatResults(results)

	// Build index context if ES is available
	indexContext := ""
	if stats, err := q.es.IndexStats(ctx); err == nil && len(stats) > 0 {
		var sb strings.Builder
		sb.WriteString("Indexed data available:\n")
		for _, s := range stats {
			sb.WriteString(fmt.Sprintf("- %s: %s documents (%s)\n", s.Index, s.DocCount, s.StoreSize))
		}
		indexContext = sb.String()
	}

	repoContext := ""
	if q.repo != "" {
		repoContext = fmt.Sprintf("The user is asking about repo: %s\n", q.repo)
	}

	prompt := fmt.Sprintf(`You are an observability assistant. Answer the question based ONLY on the data below.

Rules:
- Be direct and concise.
- Cite specific repos, timestamps, or pipeline names from the data.
- Never fabricate information not present in the data.
- Only reference data shown below.
- If the data doesn't contain what was asked about, say so clearly and suggest what to index.

%s%sObserved data:
%s

Question: %s

Answer:`, repoContext, indexContext, formatted, question)

	answer, err := ollamaGenerate(ctx, q.model, prompt)
	if err != nil {
		return "", fmt.Errorf("synthesize: %w", err)
	}
	return strings.TrimSpace(answer), nil
}
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `go test ./internal/observer/ -v`
Expected: All pass including new tests

- [ ] **Step 9: Add --repo flag to cmd/observe.go**

```go
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/observer"
)

var observeRepo string

func init() {
	observeCmd.Flags().StringVar(&observeRepo, "repo", "", "scope query to a specific repo (e.g. elastic/ensemble)")
	rootCmd.AddCommand(observeCmd)
}

var observeCmd = &cobra.Command{
	Use:   "observe [question]",
	Short: "query indexed activity via elasticsearch",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := strings.Join(args, " ")

		es := esearch.NewClient("http://localhost:9200")
		if err := es.Ping(context.Background()); err != nil {
			return fmt.Errorf("elasticsearch is not running — start with: glitch up")
		}

		engine := observer.NewQueryEngine(es, "")
		if observeRepo != "" {
			engine = engine.WithRepo(observeRepo)
		}
		answer, err := engine.Answer(cmd.Context(), question)
		if err != nil {
			return err
		}
		fmt.Println(answer)
		return nil
	},
}
```

- [ ] **Step 10: Run all tests**

Run: `go test ./...`
Expected: All pass

- [ ] **Step 11: Commit**

```bash
git add internal/esearch/client.go internal/esearch/client_test.go internal/observer/query.go internal/observer/query_test.go cmd/observe.go
git commit -m "feat: add repo context and index stats to observer queries"
```

---

### Task 3: Research Loop Fallback in `glitch ask`

When `glitch ask "some question"` doesn't match any workflow, it currently prints "No matching workflow found" and exits. Instead, fall through to the existing `ToolLoop` in `internal/research/`. Per user direction: this is about making the existing research loop reachable, not building a new system. Future work will create workflows for specific research patterns.

**Files:**
- Modify: `cmd/ask.go`
- Modify: `cmd/research_helpers.go`

- [ ] **Step 1: Write the failing test**

In `cmd/ask_test.go` (create if needed), or verify manually. The key behavior: when router.Match returns nil, `ask` should attempt the research loop instead of printing "No matching workflow found."

This is a wiring change, so we verify via integration:

Run: `go build -o /tmp/glitch-test . && /tmp/glitch-test ask "what is the project structure of this repo" 2>&1 | head -5`
Expected currently: "No matching workflow found."

- [ ] **Step 2: Wire the research fallback**

In `cmd/ask.go`, replace the "No match" block (around line 147-161) with:

```go
		// No match — fall through to research loop
		fmt.Println(">> research (no workflow matched)")

		// Detect repo from question or current directory
		org, repoName := research.ParseRepoFromQuestion(input)
		repoPath := ""
		if repoName != "" {
			repoPath, _ = research.EnsureRepo(org, repoName, "")
		}
		if repoPath == "" {
			// Use cwd as repo
			cwd, _ := os.Getwd()
			repoPath = cwd
		}

		tl, err := buildToolLoop(repoPath)
		if err != nil {
			return fmt.Errorf("research loop: %w", err)
		}

		doc := research.ResearchDocument{
			Source: "question",
			Title:  input,
			Body:   input,
			Repo:   org + "/" + repoName,
			RepoPath: repoPath,
		}

		result, err := tl.Run(context.Background(), doc, research.GoalSummarize)
		if err != nil {
			return fmt.Errorf("research loop: %w", err)
		}

		fmt.Println(result.Output)

		// Save if substantive
		if research.IsSubstantive(result.Output) {
			savedPath, _ := research.SaveLoopResult(result)
			if savedPath != "" {
				fmt.Fprintf(os.Stderr, "\nSaved: %s\n", savedPath)
			}
		}

		return nil
```

Add the import for `"github.com/8op-org/gl1tch/internal/research"` to the imports block.

- [ ] **Step 3: Verify the fallback works**

Run: `go run . ask "what is the directory structure of this project"`
Expected: Tool calls appear on stderr (e.g. `> list_files`, `> grep_code`), followed by a research summary.

- [ ] **Step 4: Run all tests**

Run: `go test ./...`
Expected: All pass

- [ ] **Step 5: Commit**

```bash
git add cmd/ask.go
git commit -m "feat: wire research loop as fallback when no workflow matches in glitch ask"
```

---

### Task 4: Verify End-to-End

- [ ] **Step 1: Run the full test suite**

```bash
go test ./...
```
Expected: All packages pass.

- [ ] **Step 2: Smoke test each feature**

1. Post-impl review: `go run . workflow run post-impl-review --param repo=elastic/observability-robots --param issue=3916 --param pr=3930`
2. Observer with repo: `go run . observe --repo elastic/ensemble "what's been indexed recently"`
3. Research fallback: `go run . ask "explain the pipeline runner in this codebase"`

- [ ] **Step 3: Final commit if any fixups needed**

```bash
git add -A
git commit -m "fix: address smoke test findings"
```
