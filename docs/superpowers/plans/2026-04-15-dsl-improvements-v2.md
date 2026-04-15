# DSL Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 7 DSL improvements to unblock the support-triage workflow: optional `:query`, `(flatten)`, `pick` template function, `(llm :format "json")` post-processing, `(search :sort)`, `(search :ndjson)`, and `(index :upsert false)`.

**Architecture:** All changes live in the `internal/pipeline` package (types.go, sexpr.go, runner.go) plus a new `pick` function in the template funcMap. The esearch client needs a small addition for create-only indexing. Each improvement is independent — parser → struct → runner → test.

**Tech Stack:** Go, `encoding/json`, existing `internal/pipeline` and `internal/esearch` packages.

---

### Task 1: Make `(search)` `:query` optional — default to match_all

**Files:**
- Modify: `internal/pipeline/sexpr.go:893-894` (remove error check)
- Modify: `internal/pipeline/runner.go:1205-1212` (already handles empty Query — add match_all default)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_SearchQueryOptional(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (search :index "my-index" :size 5)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if s.Search.IndexName != "my-index" {
		t.Fatalf("expected index %q, got %q", "my-index", s.Search.IndexName)
	}
	if s.Search.Size != 5 {
		t.Fatalf("expected size 5, got %d", s.Search.Size)
	}
	if s.Search.Query != "" {
		t.Fatalf("expected empty query, got %q", s.Search.Query)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_SearchQueryOptional -v`
Expected: FAIL with "search missing :query"

- [ ] **Step 3: Remove the `:query` required check**

In `internal/pipeline/sexpr.go`, delete lines 893-895:

```go
// REMOVE these lines:
if sr.Query == "" {
	return nil, fmt.Errorf("line %d: search missing :query", n.Line)
}
```

The runner at line 1206 already checks `if step.Search.Query != ""` before adding query to the body, so an empty query naturally becomes match_all (Elasticsearch defaults).

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_SearchQueryOptional -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): make (search) :query optional — defaults to match_all"
```

---

### Task 2: Add `(flatten "step-id")` step form

**Files:**
- Modify: `internal/pipeline/types.go` (add `Flatten` field to Step)
- Modify: `internal/pipeline/sexpr.go` (add `flatten` to `convertStep` switch + new `convertFlatten`)
- Modify: `internal/pipeline/runner.go` (add flatten execution in `runSingleStep`)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_Flatten(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "fetch"
    (run "echo '[{\"a\":1},{\"b\":2}]'"))
  (step "flat"
    (flatten "fetch")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}
	s := w.Steps[1]
	if s.Flatten != "fetch" {
		t.Fatalf("expected flatten %q, got %q", "fetch", s.Flatten)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Flatten -v`
Expected: FAIL (Flatten field doesn't exist)

- [ ] **Step 3: Add `Flatten` field to Step struct**

In `internal/pipeline/types.go`, add after line 74 (the `Lines` field):

```go
Flatten string `yaml:"-"` // step ID whose JSON array output to flatten to NDJSON
```

- [ ] **Step 4: Add `flatten` to `convertStep` switch**

In `internal/pipeline/sexpr.go`, inside the `convertStep` function's switch block (around line 536), add a new case after the `"lines"` case:

```go
case "flatten":
	if len(child.Children) < 2 {
		return s, fmt.Errorf("line %d: (flatten) missing step ID", child.Line)
	}
	s.Flatten = resolveVal(child.Children[1], defs)
```

- [ ] **Step 5: Add flatten execution in runner**

In `internal/pipeline/runner.go`, add after the `step.Lines` block (after the closing `}` around line 1087) and before the `step.Merge` check:

```go
if step.Flatten != "" {
	from := stepsSnap[step.Flatten]
	var items []json.RawMessage
	if err := json.Unmarshal([]byte(from), &items); err != nil {
		return nil, fmt.Errorf("step %s: flatten: source is not a JSON array: %w", step.ID, err)
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = string(item)
	}
	ui.StepSDK(step.ID, "flatten")
	return &stepOutcome{output: strings.Join(lines, "\n")}, nil
}
```

- [ ] **Step 6: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Flatten -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add (flatten) step — JSON array to NDJSON"
```

---

### Task 3: Add `pick` template function

**Files:**
- Modify: `internal/pipeline/runner.go` (add `pick` to funcMap in `render`)
- Test: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write failing test**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRender_Pick(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"subject":"help me","from":"alice@example.com"}`,
		},
	}

	result, err := render(`{{.param.item | pick "subject"}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}
	if result != "help me" {
		t.Fatalf("expected %q, got %q", "help me", result)
	}
}

func TestRender_PickNested(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"email":{"subject":"nested"}}`,
		},
	}

	result, err := render(`{{.param.item | pick "email.subject"}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}
	if result != "nested" {
		t.Fatalf("expected %q, got %q", "nested", result)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRender_Pick" -v`
Expected: FAIL — function "pick" not defined

- [ ] **Step 3: Add `pick` to template funcMap**

In `internal/pipeline/runner.go`, inside the `render` function's `funcMap` (around line 651, after the `"hasSuffix"` entry), add:

```go
// pick extracts a field from a JSON string by key.
// Supports dot notation for nested access: pick "email.subject"
"pick": func(key, jsonStr string) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return ""
	}
	parts := strings.Split(key, ".")
	var cur any = obj
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[p]
	}
	switch v := cur.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
},
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRender_Pick" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat(dsl): add pick template function for JSON field extraction"
```

---

### Task 4: Add `(llm :format "json")` post-processing

**Files:**
- Modify: `internal/pipeline/runner.go` (add post-processing after LLM call)
- Test: `internal/pipeline/runner_test.go`

The tiered runner already validates structure via `CheckStructure`, but non-tiered LLM calls (direct ollama, lm-studio, agents) return raw output with no cleanup. When `:format "json"` is set, we need to strip `<think>` tags and markdown fences, then extract the first JSON object.

- [ ] **Step 1: Write test for the extraction function**

Add to `internal/pipeline/runner_test.go`:

```go
func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		err   bool
	}{
		{
			name:  "clean json",
			input: `{"category": "billing"}`,
			want:  `{"category": "billing"}`,
		},
		{
			name:  "with think tags",
			input: "<think>\nlet me analyze this\n</think>\n{\"category\": \"billing\"}",
			want:  `{"category": "billing"}`,
		},
		{
			name:  "with markdown fences",
			input: "```json\n{\"category\": \"billing\"}\n```",
			want:  `{"category": "billing"}`,
		},
		{
			name:  "with think and fences",
			input: "<think>\nhmm\n</think>\nHere is the result:\n```json\n{\"category\": \"billing\"}\n```\nDone.",
			want:  `{"category": "billing"}`,
		},
		{
			name:  "no json",
			input: "I cannot help with that",
			err:   true,
		},
		{
			name:  "json array",
			input: "[1,2,3]",
			want:  "[1,2,3]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJSON(tt.input)
			if tt.err {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestExtractJSON -v`
Expected: FAIL — undefined: extractJSON

- [ ] **Step 3: Implement `extractJSON` function**

Add to `internal/pipeline/runner.go` (near the other helper functions, before `render`):

```go
// extractJSON strips <think> tags and markdown fences, then extracts
// the first valid JSON object or array from LLM output.
func extractJSON(s string) (string, error) {
	// Strip <think>...</think> blocks (multiline)
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(s, "</think>")
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + s[end+len("</think>"):]
	}

	// Strip markdown fences
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			continue
		}
		cleaned = append(cleaned, line)
	}
	s = strings.TrimSpace(strings.Join(cleaned, "\n"))

	// Find first { or [ and extract valid JSON
	for i, ch := range s {
		if ch == '{' || ch == '[' {
			// Try progressively longer substrings from this point
			for j := len(s); j > i; j-- {
				candidate := s[i:j]
				var v any
				if json.Unmarshal([]byte(candidate), &v) == nil {
					return candidate, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no valid JSON found in LLM output")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestExtractJSON -v`
Expected: PASS

- [ ] **Step 5: Wire post-processing into LLM execution**

In `internal/pipeline/runner.go`, right after the `ui.StepLLMDone(...)` call (around line 1038) and before the `return &stepOutcome{...}` block, add:

```go
// Post-process structured output
if step.LLM.Format == "json" {
	extracted, err := extractJSON(out)
	if err != nil {
		return nil, fmt.Errorf("step %s: format json: %w", step.ID, err)
	}
	out = extracted
}
```

- [ ] **Step 6: Run all pipeline tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat(dsl): add (llm :format \"json\") post-processing — strips think tags and fences"
```

---

### Task 5: Add `(search :sort)` parameter

**Files:**
- Modify: `internal/pipeline/types.go` (add `Sort` to SearchStep)
- Modify: `internal/pipeline/sexpr.go` (parse `:sort` keyword in `convertSearch`)
- Modify: `internal/pipeline/runner.go` (add sort to query body)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_SearchSort(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (search :index "my-index" :size 10 :sort {"indexed_at" "desc"})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if s.Search.Sort == "" {
		t.Fatal("expected sort to be set")
	}
	// Sort should be valid JSON
	var v any
	if err := json.Unmarshal([]byte(s.Search.Sort), &v); err != nil {
		t.Fatalf("sort is not valid JSON: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_SearchSort -v`
Expected: FAIL — `s.Search.Sort` undefined

- [ ] **Step 3: Add `Sort` field to SearchStep**

In `internal/pipeline/types.go`, add to the `SearchStep` struct (after the `Fields` field, around line 144):

```go
Sort string // raw JSON sort clause
```

- [ ] **Step 4: Parse `:sort` keyword in `convertSearch`**

In `internal/pipeline/sexpr.go`, inside the `convertSearch` switch block (around line 880, after the `"es"` case), add:

```go
case "sort":
	b, err := nodeToJSON(val)
	if err != nil {
		return nil, fmt.Errorf("line %d: sort to JSON: %w", val.Line, err)
	}
	sr.Sort = string(b)
```

- [ ] **Step 5: Add sort to query body in runner**

In `internal/pipeline/runner.go`, inside the search execution block (after the `_source` fields check, around line 1216), add:

```go
if step.Search.Sort != "" {
	var sortClause any
	if err := json.Unmarshal([]byte(step.Search.Sort), &sortClause); err != nil {
		return nil, fmt.Errorf("step %s: sort parse: %w", step.ID, err)
	}
	queryBody["sort"] = []any{sortClause}
}
```

- [ ] **Step 6: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_SearchSort -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add (search :sort) parameter for ordered results"
```

---

### Task 6: Add `(search :ndjson)` flag

**Files:**
- Modify: `internal/pipeline/types.go` (add `NDJSON` to SearchStep)
- Modify: `internal/pipeline/sexpr.go` (parse `:ndjson` boolean flag in `convertSearch`)
- Modify: `internal/pipeline/runner.go` (conditionally output NDJSON)
- Test: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_SearchNDJSON(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (search :index "my-index" :ndjson)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if !s.Search.NDJSON {
		t.Fatal("expected ndjson to be true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_SearchNDJSON -v`
Expected: FAIL — `s.Search.NDJSON` undefined

- [ ] **Step 3: Add `NDJSON` field to SearchStep**

In `internal/pipeline/types.go`, add to the `SearchStep` struct:

```go
NDJSON bool // output as NDJSON instead of JSON array
```

- [ ] **Step 4: Parse `:ndjson` as a boolean flag in `convertSearch`**

`:ndjson` is a bare keyword with no value (like `:embed` in index). In `internal/pipeline/sexpr.go`, inside `convertSearch`, the keyword parsing loop needs to handle `:ndjson` as a flag that doesn't consume the next token.

Replace the keyword parsing section in `convertSearch` (lines 850-886). The full updated loop body:

```go
if child.IsAtom() && child.Atom.Type == sexpr.TokenKeyword {
	key := child.KeywordVal()
	switch key {
	case "ndjson":
		sr.NDJSON = true
		i++
		continue
	}
	i++
	if i >= len(children) {
		return nil, fmt.Errorf("line %d: keyword :%s missing value", child.Line, key)
	}
	val := children[i]
	switch key {
	case "index":
		sr.IndexName = resolveVal(val, defs)
	case "query":
		b, err := nodeToJSON(val)
		if err != nil {
			return nil, fmt.Errorf("line %d: query to JSON: %w", val.Line, err)
		}
		sr.Query = string(b)
	case "size":
		n, err := strconv.Atoi(resolveVal(val, defs))
		if err != nil {
			return nil, fmt.Errorf("line %d: :size must be an integer", val.Line)
		}
		sr.Size = n
	case "fields":
		if val.IsList() {
			for _, f := range val.Children {
				sr.Fields = append(sr.Fields, resolveVal(f, defs))
			}
		} else {
			sr.Fields = append(sr.Fields, resolveVal(val, defs))
		}
	case "es":
		sr.ESURL = resolveVal(val, defs)
	case "sort":
		b, err := nodeToJSON(val)
		if err != nil {
			return nil, fmt.Errorf("line %d: sort to JSON: %w", val.Line, err)
		}
		sr.Sort = string(b)
	default:
		return nil, fmt.Errorf("line %d: unknown search keyword :%s", child.Line, key)
	}
	i++
	continue
}
```

Note: This replaces the entire keyword block and also includes the `:sort` case from Task 5. If Task 5 was already applied, this is a clean replacement of the same block.

- [ ] **Step 5: Add NDJSON output mode in runner**

In `internal/pipeline/runner.go`, replace the search results marshaling block (around lines 1229-1237):

```go
sources := make([]json.RawMessage, 0)
for _, hit := range resp.Results {
	sources = append(sources, hit.Source)
}

if step.Search.NDJSON {
	lines := make([]string, len(sources))
	for i, s := range sources {
		lines[i] = string(s)
	}
	return &stepOutcome{output: strings.Join(lines, "\n")}, nil
}

out, err := json.Marshal(sources)
if err != nil {
	return nil, fmt.Errorf("step %s: marshal results: %w", step.ID, err)
}
return &stepOutcome{output: string(out)}, nil
```

- [ ] **Step 6: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_SearchNDJSON -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add (search :ndjson) flag for NDJSON output"
```

---

### Task 7: Add `(index :upsert false)` — skip-if-exists dedup

**Files:**
- Modify: `internal/pipeline/types.go` (add `Upsert` to IndexStep)
- Modify: `internal/pipeline/sexpr.go` (parse `:upsert` keyword in `convertIndex`)
- Modify: `internal/esearch/client.go` (add `IndexDocCreate` method using `op_type=create`)
- Modify: `internal/pipeline/runner.go` (use create-only when Upsert is false)
- Test: `internal/pipeline/sexpr_test.go`
- Test: `internal/esearch/client_test.go`

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_IndexUpsertFalse(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (index :index "my-index" :doc "{}" :id "doc1" :upsert false)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	s := w.Steps[0]
	if s.Index == nil {
		t.Fatal("expected index step")
	}
	if s.Index.Upsert {
		t.Fatal("expected upsert to be false")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_IndexUpsertFalse -v`
Expected: FAIL — `s.Index.Upsert` undefined

- [ ] **Step 3: Add `Upsert` field to IndexStep**

In `internal/pipeline/types.go`, add to `IndexStep`:

```go
Upsert *bool // nil = default (upsert), false = skip if exists (op_type=create)
```

Using `*bool` so we can distinguish "not set" (nil, default upsert behavior) from "explicitly false".

- [ ] **Step 4: Parse `:upsert` in `convertIndex`**

In `internal/pipeline/sexpr.go`, inside `convertIndex`'s default keyword switch (around line 940), add a new case:

```go
case "upsert":
	v := strings.ToLower(resolveVal(val, defs))
	b := v != "false" && v != "0" && v != "no"
	idx.Upsert = &b
```

- [ ] **Step 5: Add `IndexDocCreate` method to esearch client**

In `internal/esearch/client.go`, add after the `IndexDoc` method:

```go
// IndexDocCreate indexes a document only if it doesn't already exist (op_type=create).
// Returns (response, existed, error). If the doc already exists, existed=true and response is nil.
func (c *Client) IndexDocCreate(ctx context.Context, index, docID string, doc json.RawMessage) (*IndexDocResponse, bool, error) {
	path := "/" + index + "/_doc/" + docID + "?op_type=create"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, bytes.NewReader(doc))
	if err != nil {
		return nil, false, fmt.Errorf("index doc create: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("index doc create: %w", err)
	}
	defer resp.Body.Close()

	// 409 Conflict means doc already exists — not an error for dedup
	if resp.StatusCode == 409 {
		return nil, true, nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("index doc create: status %s — %s", resp.Status, truncate(string(body), 256))
	}

	var result IndexDocResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, fmt.Errorf("index doc create: decode response: %w", err)
	}
	return &result, false, nil
}
```

- [ ] **Step 6: Write esearch client test**

Add to `internal/esearch/client_test.go`:

```go
func TestIndexDocCreate_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "op_type=create") {
			t.Error("expected op_type=create in query")
		}
		w.WriteHeader(409)
		w.Write([]byte(`{"error":{"type":"version_conflict_engine_exception"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, existed, err := c.IndexDocCreate(context.Background(), "test-index", "doc1", []byte(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !existed {
		t.Fatal("expected existed=true for 409")
	}
	if resp != nil {
		t.Fatal("expected nil response for 409")
	}
}
```

- [ ] **Step 7: Use create-only in runner when Upsert is false**

In `internal/pipeline/runner.go`, replace the index execution block (around lines 1281-1287) with:

```go
esURL := resolveESURL(step.Index.ESURL, rctx)
es := esearch.NewClient(esURL)
ui.StepSDK(step.ID, "index")

if step.Index.Upsert != nil && !*step.Index.Upsert && idRendered != "" {
	resp, existed, err := es.IndexDocCreate(ctx, indexRendered, idRendered, docBytes)
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", step.ID, err)
	}
	if existed {
		return &stepOutcome{output: fmt.Sprintf(`{"result":"noop","_id":"%s"}`, idRendered)}, nil
	}
	out, _ := json.Marshal(resp)
	return &stepOutcome{output: string(out)}, nil
}

resp, err := es.IndexDoc(ctx, indexRendered, idRendered, docBytes)
if err != nil {
	return nil, fmt.Errorf("step %s: %w", step.ID, err)
}
out, _ := json.Marshal(resp)
return &stepOutcome{output: string(out)}, nil
```

- [ ] **Step 8: Run all tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ ./internal/esearch/ -v`
Expected: All PASS

- [ ] **Step 9: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/sexpr_test.go internal/esearch/client.go internal/esearch/client_test.go
git commit -m "feat(dsl): add (index :upsert false) for dedup-on-index"
```

---

### Task 8: Run full test suite and verify

- [ ] **Step 1: Run all pipeline and esearch tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ ./internal/esearch/ ./internal/provider/ -v -count=1`
Expected: All PASS

- [ ] **Step 2: Run go vet**

Run: `cd /Users/stokes/Projects/gl1tch && go vet ./internal/pipeline/ ./internal/esearch/`
Expected: No issues

- [ ] **Step 3: Verify build**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: Clean build
