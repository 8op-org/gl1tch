# WebSearch Form Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `(websearch ...)` workflow form backed by SearXNG, with the endpoint configured as a workspace default (`:websearch`).

**Architecture:** Follows the same pattern as Elasticsearch: workspace config holds the URL, `runCtx` threads it through, and a new step type executes the search via SearXNG's JSON API. The parser, types, runner, and docs are each touched once.

**Tech Stack:** Go, SearXNG JSON API, net/http, encoding/json

---

### Task 1: Add `WebSearch` field to workspace Defaults

**Files:**
- Modify: `internal/workspace/workspace.go:20-25` (Defaults struct)
- Modify: `internal/workspace/workspace.go:123-177` (convertDefaults)
- Modify: `internal/workspace/serialize.go:29-56` (Serialize defaults block)
- Test: `internal/workspace/workspace_test.go`

- [ ] **Step 1: Write the failing test**

In `internal/workspace/workspace_test.go`, add:

```go
func TestParseWorkspace_WebSearch(t *testing.T) {
	src := []byte(`
(workspace "test"
  (defaults
    :model "qwen2.5:7b"
    :provider "ollama"
    :websearch "http://localhost:8080"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.WebSearch != "http://localhost:8080" {
		t.Fatalf("websearch = %q, want http://localhost:8080", w.Defaults.WebSearch)
	}
}

func TestParseWorkspace_WebSearchDefault(t *testing.T) {
	src := []byte(`
(workspace "test"
  (defaults :model "qwen2.5:7b"))
`)
	w, err := ParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.WebSearch != "" {
		t.Fatalf("websearch should be empty by default, got %q", w.Defaults.WebSearch)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -run TestParseWorkspace_WebSearch -v`
Expected: FAIL — `w.Defaults.WebSearch` not a field

- [ ] **Step 3: Add WebSearch field to Defaults struct**

In `internal/workspace/workspace.go`, add the field to the `Defaults` struct:

```go
type Defaults struct {
	Model         string
	Provider      string
	Elasticsearch string
	WebSearch     string
	Params        map[string]string
}
```

- [ ] **Step 4: Parse `:websearch` keyword in convertDefaults**

In `internal/workspace/workspace.go` inside `convertDefaults`, add a case to the keyword switch:

```go
case "websearch":
	d.WebSearch = val.StringVal()
```

This goes right after the `"elasticsearch"` case.

- [ ] **Step 5: Serialize `:websearch` in Serialize function**

In `internal/workspace/serialize.go`, add after the Elasticsearch serialization block:

```go
if w.Defaults.WebSearch != "" {
	b.WriteString(fmt.Sprintf("\n    :websearch %q", w.Defaults.WebSearch))
}
```

Also update the `hasDefaults` check to include `w.Defaults.WebSearch != ""`.

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/workspace/ -v`
Expected: PASS — all tests including the new WebSearch tests

- [ ] **Step 7: Commit**

```bash
git add internal/workspace/workspace.go internal/workspace/serialize.go internal/workspace/workspace_test.go
git commit -m "feat: add :websearch to workspace defaults"
```

---

### Task 2: Thread WebSearch URL through RunOpts and runCtx

**Files:**
- Modify: `internal/pipeline/runner.go:107-137` (RunOpts struct)
- Modify: `internal/pipeline/runner.go:42-69` (runCtx struct)
- Modify: `internal/pipeline/runner.go:250-310` (Run function — extract and thread the URL)
- Modify: `internal/pipeline/runner.go:1320-1340` (call-workflow child runCtx creation)
- Modify: `internal/pipeline/runner.go:1650-1660` (call-workflow RunOpts forwarding)
- Modify: `cmd/run.go:150-159` (extract WebSearch URL from workspace)
- Modify: `cmd/run.go:260-272` (pass WebSearchURL in RunOpts)

- [ ] **Step 1: Add WebSearchURL to RunOpts**

In `internal/pipeline/runner.go`, add to the `RunOpts` struct after `ESURL`:

```go
WebSearchURL string // default SearXNG URL from workspace config
```

- [ ] **Step 2: Add webSearchURL to runCtx**

In `internal/pipeline/runner.go`, add to the `runCtx` struct after `esURL`:

```go
webSearchURL string
```

- [ ] **Step 3: Thread WebSearchURL in the Run function**

In the `Run` function, after the `esURL` extraction block (around line 267-269), add:

```go
var webSearchURL string
if len(opts) > 0 && opts[0].WebSearchURL != "" {
	webSearchURL = opts[0].WebSearchURL
}
```

And add `webSearchURL: webSearchURL,` to the `rctx` initialization struct.

- [ ] **Step 4: Forward webSearchURL in call-workflow child runCtx**

In the call-workflow child creation (around line 1327-1333), add `webSearchURL: rctx.webSearchURL,` to the child `runCtx` struct.

- [ ] **Step 5: Forward WebSearchURL in call-workflow RunOpts**

In the call-workflow `RunOpts` forwarding (around line 1654-1660), add `WebSearchURL: rctx.webSearchURL,`.

- [ ] **Step 6: Extract WebSearch URL in cmd/run.go**

In `cmd/run.go`, after the `wsESURL` extraction (line 155-157), add:

```go
var wsWebSearchURL string
```

Then inside the workspace parse block, add:

```go
if ws.Defaults.WebSearch != "" {
	wsWebSearchURL = ws.Defaults.WebSearch
}
```

And pass it in the RunOpts at line 265:

```go
WebSearchURL: wsWebSearchURL,
```

- [ ] **Step 7: Verify compilation**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: clean build

- [ ] **Step 8: Commit**

```bash
git add internal/pipeline/runner.go cmd/run.go
git commit -m "feat: thread websearch URL through RunOpts and runCtx"
```

---

### Task 3: Add WebSearchStep type and parser

**Files:**
- Modify: `internal/pipeline/types.go:100-126` (Step struct — add WebSearch field)
- Modify: `internal/pipeline/sexpr.go` (add `case "websearch"` and `convertWebSearch` function)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write the failing test**

In `internal/pipeline/sexpr_test.go`, add:

```go
func TestSexprWorkflow_WebSearch(t *testing.T) {
	src := []byte(`
(workflow "web-research"
  (step "find"
    (websearch "kubernetes pod eviction"
      :engines ("google" "stackoverflow")
      :results 3
      :lang "en")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	s := w.Steps[0]
	if s.WebSearch == nil {
		t.Fatal("expected WebSearch step")
	}
	if s.WebSearch.Query != "kubernetes pod eviction" {
		t.Fatalf("query = %q, want %q", s.WebSearch.Query, "kubernetes pod eviction")
	}
	if len(s.WebSearch.Engines) != 2 {
		t.Fatalf("engines len = %d, want 2", len(s.WebSearch.Engines))
	}
	if s.WebSearch.Engines[0] != "google" || s.WebSearch.Engines[1] != "stackoverflow" {
		t.Fatalf("engines = %v, want [google stackoverflow]", s.WebSearch.Engines)
	}
	if s.WebSearch.Results != 3 {
		t.Fatalf("results = %d, want 3", s.WebSearch.Results)
	}
	if s.WebSearch.Lang != "en" {
		t.Fatalf("lang = %q, want %q", s.WebSearch.Lang, "en")
	}
}

func TestSexprWorkflow_WebSearchDefaults(t *testing.T) {
	src := []byte(`
(workflow "web-minimal"
  (step "find"
    (websearch "test query")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.WebSearch == nil {
		t.Fatal("expected WebSearch step")
	}
	if s.WebSearch.Query != "test query" {
		t.Fatalf("query = %q, want %q", s.WebSearch.Query, "test query")
	}
	if s.WebSearch.Results != 5 {
		t.Fatalf("results = %d, want default 5", s.WebSearch.Results)
	}
	if s.WebSearch.Lang != "en" {
		t.Fatalf("lang = %q, want default %q", s.WebSearch.Lang, "en")
	}
	if len(s.WebSearch.Engines) != 0 {
		t.Fatalf("engines should be empty by default, got %v", s.WebSearch.Engines)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_WebSearch -v`
Expected: FAIL — `WebSearch` not a field on Step

- [ ] **Step 3: Add WebSearchStep type to types.go**

In `internal/pipeline/types.go`, add the struct after `EmbedStep`:

```go
// WebSearchStep queries a SearXNG instance and returns results as JSON.
type WebSearchStep struct {
	Query   string   // search query (template-rendered)
	Engines []string // SearXNG engine names (empty = SearXNG defaults)
	Results int      // max results (default 5)
	Lang    string   // language filter (default "en")
}
```

And add the field to the `Step` struct, after the `Embed` field:

```go
// Web search
WebSearch *WebSearchStep `yaml:"-"`
```

- [ ] **Step 4: Add `convertWebSearch` function to sexpr.go**

In `internal/pipeline/sexpr.go`, add the converter function:

```go
func convertWebSearch(n *sexpr.Node, defs map[string]string) (*WebSearchStep, error) {
	children := n.Children[1:] // skip "websearch"
	if len(children) < 1 {
		return nil, fmt.Errorf("line %d: (websearch) missing query", n.Line)
	}
	ws := &WebSearchStep{
		Query:   resolveVal(children[0], defs),
		Results: 5,
		Lang:    "en",
	}
	for i := 1; i < len(children); i++ {
		child := children[i]
		if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
			key := child.KeywordVal()
			i++
			if i >= len(children) {
				return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
			}
			val := children[i]
			switch key {
			case "engines":
				if val.IsList() {
					for _, e := range val.Children {
						ws.Engines = append(ws.Engines, resolveVal(e, defs))
					}
				} else {
					ws.Engines = append(ws.Engines, resolveVal(val, defs))
				}
			case "results":
				n, err := strconv.Atoi(resolveVal(val, defs))
				if err != nil {
					return nil, fmt.Errorf("line %d: :results must be an integer", val.Line)
				}
				ws.Results = n
			case "lang":
				ws.Lang = resolveVal(val, defs)
			default:
				return nil, fmt.Errorf("line %d: unknown websearch keyword :%s", child.Line, key)
			}
		}
	}
	return ws, nil
}
```

- [ ] **Step 5: Wire `"websearch"` case in the step converter switch**

In `internal/pipeline/sexpr.go`, in the form dispatch switch (around line 1373, after the `"search"` case), add:

```go
case "websearch":
	ws, err := convertWebSearch(child, defs)
	if err != nil {
		return s, err
	}
	s.WebSearch = ws
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_WebSearch -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat: add websearch step type and sexpr parser"
```

---

### Task 4: Execute WebSearchStep in the runner

**Files:**
- Modify: `internal/pipeline/runner.go` (add execution block after the `step.Embed` block)
- Test: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write the failing test**

In `internal/pipeline/runner_test.go`, add a test that uses a local HTTP test server to simulate SearXNG:

```go
func TestWebSearchStep(t *testing.T) {
	// Simulate SearXNG JSON API
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			http.Error(w, "not found", 404)
			return
		}
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "missing q", 400)
			return
		}
		format := r.URL.Query().Get("format")
		if format != "json" {
			http.Error(w, "format must be json", 400)
			return
		}
		resp := map[string]any{
			"results": []map[string]any{
				{
					"title":   "Test Result",
					"url":     "https://example.com",
					"content": "A test result snippet",
					"engine":  "google",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	src := fmt.Sprintf(`
(workflow "websearch-test"
  (step "find"
    (websearch "test query" :results 1)))
`, )
	w, err := LoadBytes([]byte(src), "test.glitch")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Run(w, "", "", nil, provider.NewRegistry(), RunOpts{
		WebSearchURL: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}

	var results []map[string]any
	if err := json.Unmarshal([]byte(result.Output), &results); err != nil {
		t.Fatalf("output is not valid JSON array: %v\noutput: %s", err, result.Output)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["title"] != "Test Result" {
		t.Fatalf("title = %v, want %q", results[0]["title"], "Test Result")
	}
}

func TestWebSearchStep_NoURL(t *testing.T) {
	src := []byte(`
(workflow "websearch-nourl"
  (step "find"
    (websearch "test query")))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}

	_, err = Run(w, "", "", nil, provider.NewRegistry(), RunOpts{})
	if err == nil {
		t.Fatal("expected error for missing websearch URL")
	}
	if !strings.Contains(err.Error(), "no websearch endpoint configured") {
		t.Fatalf("error = %q, want mention of 'no websearch endpoint configured'", err.Error())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestWebSearchStep -v`
Expected: FAIL — no execution path for WebSearch steps

- [ ] **Step 3: Implement WebSearch execution in runner.go**

In `internal/pipeline/runner.go`, add the execution block after the `step.Embed` block (around line 2200):

```go
if step.WebSearch != nil {
	if rctx.webSearchURL == "" {
		return nil, fmt.Errorf("step %s: no websearch endpoint configured — add :websearch to workspace defaults", step.ID)
	}

	queryRendered, err := renderInStep(step.WebSearch.Query, scopeFromData(data), stepsSnap, rctx.workflowObj, &step)
	if err != nil {
		return nil, fmt.Errorf("step %s: websearch query render: %w", step.ID, err)
	}

	searchURL := fmt.Sprintf("%s/search", strings.TrimRight(rctx.webSearchURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("step %s: websearch request: %w", step.ID, err)
	}

	q := req.URL.Query()
	q.Set("q", queryRendered)
	q.Set("format", "json")
	if step.WebSearch.Lang != "" {
		q.Set("language", step.WebSearch.Lang)
	}
	if len(step.WebSearch.Engines) > 0 {
		q.Set("engines", strings.Join(step.WebSearch.Engines, ","))
	}
	req.URL.RawQuery = q.Encode()

	ui.StepSDK(step.ID, "websearch")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("step %s: websearch: %w", step.ID, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("step %s: websearch read body: %w", step.ID, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("step %s: websearch %d: %s", step.ID, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Parse SearXNG response and extract results array
	var searxResp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
			Engine  string `json:"engine"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &searxResp); err != nil {
		return nil, fmt.Errorf("step %s: websearch parse response: %w", step.ID, err)
	}

	// Limit to requested number of results
	results := searxResp.Results
	if len(results) > step.WebSearch.Results {
		results = results[:step.WebSearch.Results]
	}

	// Build clean output array
	type searchResult struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
		Engine  string `json:"engine"`
	}
	out := make([]searchResult, len(results))
	for i, r := range results {
		out[i] = searchResult{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
			Engine:  r.Engine,
		}
	}
	outJSON, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("step %s: websearch marshal: %w", step.ID, err)
	}
	return &stepOutcome{output: string(outJSON)}, nil
}
```

- [ ] **Step 4: Add required imports if missing**

Ensure `"net/http/httptest"` is in the test file imports, and that `"fmt"` is available in the test.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestWebSearchStep -v`
Expected: PASS

- [ ] **Step 6: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v -count=1`
Expected: PASS — no regressions

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat: execute websearch steps via SearXNG JSON API"
```

---

### Task 5: Add `websearch` to gate-hallucinations and docs

**Files:**
- Modify: `scripts/gate-hallucinations.py:59-86` (VALID_FORMS set)
- Modify: `docs/site/workflow-syntax.md` (add websearch section + form reference)
- Modify: `docs/site/workspaces.md` (document `:websearch` default)

- [ ] **Step 1: Add `websearch` to VALID_FORMS**

In `scripts/gate-hallucinations.py`, add `"websearch"` to the `VALID_FORMS` set. Place it in the SDK forms group:

```python
"http-get", "http-post", "read-file", "write-file", "glob", "websearch",
```

- [ ] **Step 2: Add websearch section to workflow-syntax.md**

In `docs/site/workflow-syntax.md`, add a new section after `## Elasticsearch forms` → `### embed` (after line ~590), before `## ES connection`:

```markdown
## Web search

### websearch

Query a SearXNG instance and return results as JSON:

\`\`\`glitch
(step "find-sources"
  (websearch "kubernetes pod eviction causes"
    :engines ("google" "stackoverflow")
    :results 5
    :lang "en"))
\`\`\`

| keyword | required | default | description |
|---------|----------|---------|-------------|
| query (1st arg) | yes | — | Search query. Supports template refs. |
| `:engines` | no | SearXNG defaults | List of engine names to target. |
| `:results` | no | 5 | Max number of results. |
| `:lang` | no | `"en"` | Language filter. |

Output is a JSON array:

\`\`\`json
[{"title": "...", "url": "...", "content": "...", "engine": "..."}]
\`\`\`

Requires `:websearch` in workspace defaults. See [workspaces](workspaces.md).
```

- [ ] **Step 3: Add `websearch` to the form reference table**

In `docs/site/workflow-syntax.md` under `### Step-level forms (inside a step)`, add `websearch` to the list.

- [ ] **Step 4: Document `:websearch` in workspaces.md**

In `docs/site/workspaces.md`, add `:websearch` to the defaults documentation, following the pattern of the `:elasticsearch` entry.

- [ ] **Step 5: Run gate-hallucinations to verify**

Run: `cd /Users/stokes/Projects/gl1tch && python3 scripts/gate-hallucinations.py docs/site/workflow-syntax.md`
Expected: PASS (or at least no failure on "websearch")

- [ ] **Step 6: Commit**

```bash
git add scripts/gate-hallucinations.py docs/site/workflow-syntax.md docs/site/workspaces.md
git commit -m "docs: add websearch form to workflow syntax and workspaces"
```

---

### Task 6: Add example workflow

**Files:**
- Create: `examples/websearch.glitch`

- [ ] **Step 1: Create the example workflow**

Create `examples/websearch.glitch`:

```glitch
(workflow "websearch"
  :description "Search the web and summarize results"

  (step "search"
    (websearch "~input"
      :results 3))

  (step "summarize"
    (llm :prompt "Summarize these search results for the user:\n\n~(step search)")))
```

- [ ] **Step 2: Verify it parses**

Run: `cd /Users/stokes/Projects/gl1tch && go run . workflow show examples/websearch.glitch`
Expected: workflow loads without parse errors

- [ ] **Step 3: Commit**

```bash
git add examples/websearch.glitch
git commit -m "feat: add websearch example workflow"
```
