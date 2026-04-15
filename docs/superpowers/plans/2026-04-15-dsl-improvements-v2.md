# DSL Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 12 DSL improvements: 7 from the original spec (optional `:query`, `(flatten)`, `pick`, `(llm :format "json")`, `(search :sort)`, `(search :ndjson)`, `(index :upsert false)`) plus 5 Clojure/Babashka-inspired forms (`->` threading, `(filter)`, `(reduce)`, `(when)`, `assoc` template function).

**Architecture:** All changes live in the `internal/pipeline` package (types.go, sexpr.go, runner.go) plus new template functions (`pick`, `assoc`) in the funcMap. The esearch client needs a small addition for create-only indexing. Each improvement is independent — parser → struct → runner → test. The 5 new control-flow forms (`->`, `filter`, `reduce`, `when`) follow the same pattern as existing `(map)` and `(cond)`: parsed in `convertForm`, executed as dedicated `execute*` functions in the runner.

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

### Task 8: Add `assoc` template function

**Files:**
- Modify: `internal/pipeline/runner.go` (add `assoc` to funcMap in `render`)
- Test: `internal/pipeline/runner_test.go`

`assoc` is the complement to `pick` — it sets a field on a JSON string and returns the modified JSON. Inspired by Clojure's `assoc`.

- [ ] **Step 1: Write failing test**

Add to `internal/pipeline/runner_test.go`:

```go
func TestRender_Assoc(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"subject":"help me","from":"alice@example.com"}`,
		},
	}

	result, err := render(`{{.param.item | assoc "status" "triaged"}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if obj["status"] != "triaged" {
		t.Fatalf("expected status %q, got %v", "triaged", obj["status"])
	}
	if obj["subject"] != "help me" {
		t.Fatalf("expected subject preserved, got %v", obj["subject"])
	}
}

func TestRender_AssocOverwrite(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"status":"new"}`,
		},
	}

	result, err := render(`{{.param.item | assoc "status" "closed"}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if obj["status"] != "closed" {
		t.Fatalf("expected status %q, got %v", "closed", obj["status"])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRender_Assoc" -v`
Expected: FAIL — function "assoc" not defined

- [ ] **Step 3: Add `assoc` to template funcMap**

In `internal/pipeline/runner.go`, inside the `render` function's `funcMap` (right after the `pick` entry), add:

```go
// assoc sets a key on a JSON object string and returns the updated JSON.
// Usage: {{.param.item | assoc "status" "triaged"}}
"assoc": func(key, val, jsonStr string) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return jsonStr
	}
	obj[key] = val
	b, err := json.Marshal(obj)
	if err != nil {
		return jsonStr
	}
	return string(b)
},
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRender_Assoc" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat(dsl): add assoc template function — set fields on JSON strings"
```

---

### Task 9: Add `(when)` conditional form

**Files:**
- Modify: `internal/pipeline/types.go` (add `WhenPred` and `WhenNot` fields to Step)
- Modify: `internal/pipeline/sexpr.go` (add `when`/`when-not` to `convertForm`, add `convertWhen`)
- Modify: `internal/pipeline/runner.go` (add `executeWhen` handler in `executeStep`)
- Test: `internal/pipeline/sexpr_test.go`

`(when pred (step ...))` is a simpler alternative to `(cond)` for single-branch conditionals. The predicate is a shell command (exit 0 = true) or a step reference (non-empty output = true). `(when-not)` inverts the predicate.

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_When(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "check" (run "echo found"))
  (when "check"
    (step "notify" (run "echo notifying"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}
	s := w.Steps[1]
	if s.Form != "when" {
		t.Fatalf("expected form %q, got %q", "when", s.Form)
	}
	if s.WhenPred != "check" {
		t.Fatalf("expected when pred %q, got %q", "check", s.WhenPred)
	}
}

func TestSexprWorkflow_WhenNot(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "check" (run "echo found"))
  (when-not "check"
    (step "fallback" (run "echo fallback"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	s := w.Steps[1]
	if s.Form != "when" {
		t.Fatalf("expected form %q, got %q", "when", s.Form)
	}
	if !s.WhenNot {
		t.Fatal("expected when-not to be true")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestSexprWorkflow_When" -v`
Expected: FAIL — `WhenPred` undefined

- [ ] **Step 3: Add `WhenPred` and `WhenNot` fields to Step**

In `internal/pipeline/types.go`, add after the `MapBody` field (around line 58):

```go
WhenPred string `yaml:"-"` // when: step ID or shell command predicate
WhenBody *Step  `yaml:"-"` // when: step to execute if predicate is true
WhenNot  bool   `yaml:"-"` // when-not: invert the predicate
```

- [ ] **Step 4: Add `convertWhen` and register in `convertForm`**

In `internal/pipeline/sexpr.go`, add the converter function (near `convertCond`):

```go
// convertWhen: (when "pred" (step ...)) or (when-not "pred" (step ...))
func convertWhen(n *sexpr.Node, defs map[string]string, negate bool) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (when) needs predicate and body step", n.Line)
	}
	pred := resolveVal(children[0], defs)
	body, err := convertStep(children[1], defs)
	if err != nil {
		// Body might be a compound form (map, par, etc.), not just a step
		head := children[1].Children[0].SymbolVal()
		if head == "" {
			head = children[1].Children[0].StringVal()
		}
		steps, formErr := convertForm(children[1], head, defs)
		if formErr != nil {
			return Step{}, fmt.Errorf("line %d: when body: %w", n.Line, err)
		}
		if len(steps) != 1 {
			return Step{}, fmt.Errorf("line %d: when body must be a single form", n.Line)
		}
		body = steps[0]
	}
	return Step{
		ID:       fmt.Sprintf("when-%d", n.Line),
		Form:     "when",
		WhenPred: pred,
		WhenBody: &body,
		WhenNot:  negate,
	}, nil
}
```

In `convertForm`, add cases for `when` and `when-not`:

```go
case "when":
	s, err := convertWhen(n, defs, false)
	if err != nil {
		return nil, err
	}
	return []Step{s}, nil
case "when-not":
	s, err := convertWhen(n, defs, true)
	if err != nil {
		return nil, err
	}
	return []Step{s}, nil
```

- [ ] **Step 5: Add `executeWhen` in runner**

In `internal/pipeline/runner.go`, add the executor function (near `executeCond`):

```go
// executeWhen checks a predicate and runs the body step if it passes.
// Predicate is a step ID (non-empty output = true) or a shell command (exit 0 = true).
func executeWhen(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	snap := rctx.stepsSnapshot()
	matched := false

	// Check if predicate is a step reference
	if output, ok := snap[step.WhenPred]; ok {
		matched = strings.TrimSpace(output) != ""
	} else {
		// Treat as shell command
		data := map[string]any{"input": rctx.input, "param": rctx.params, "workspace": rctx.workspace}
		rendered, err := render(step.WhenPred, data, snap)
		if err != nil {
			return nil, fmt.Errorf("when %s: template: %w", step.ID, err)
		}
		cmd := exec.CommandContext(ctx, "sh", "-c", rendered)
		matched = cmd.Run() == nil
	}

	if step.WhenNot {
		matched = !matched
	}

	if matched {
		return executeStep(ctx, rctx, *step.WhenBody)
	}

	// Not matched — empty outcome
	rctx.mu.Lock()
	rctx.steps[step.ID] = ""
	rctx.mu.Unlock()
	return &stepOutcome{}, nil
}
```

Wire it into `executeStep`'s switch block (around line 678):

```go
case "when":
	return executeWhen(ctx, rctx, step)
```

- [ ] **Step 6: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestSexprWorkflow_When" -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add (when) and (when-not) conditional forms"
```

---

### Task 10: Add `(filter)` collection form

**Files:**
- Modify: `internal/pipeline/types.go` (add `FilterOver` and `FilterBody` fields to Step)
- Modify: `internal/pipeline/sexpr.go` (add `filter` to `convertForm`, add `convertFilter`)
- Modify: `internal/pipeline/runner.go` (add `executeFilter` handler)
- Test: `internal/pipeline/sexpr_test.go`

`(filter "step-id" (step "pred" ...))` iterates over NDJSON lines from a source step, runs the body step per item, and keeps only items where the body output is truthy (non-empty, not "false", not "0"). Like Clojure's `filter`, but each predicate is a full step (can be LLM, shell, etc.).

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_Filter(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "data" (run "echo 'a\nb\nc'"))
  (filter "data"
    (step "keep" (run "test '{{.param.item}}' = 'b' && echo true"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}
	s := w.Steps[1]
	if s.Form != "filter" {
		t.Fatalf("expected form %q, got %q", "filter", s.Form)
	}
	if s.FilterOver != "data" {
		t.Fatalf("expected filter over %q, got %q", "data", s.FilterOver)
	}
	if s.FilterBody == nil {
		t.Fatal("expected filter body")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Filter -v`
Expected: FAIL — `FilterOver` undefined

- [ ] **Step 3: Add filter fields to Step struct**

In `internal/pipeline/types.go`, add after the `WhenNot` field:

```go
FilterOver string `yaml:"-"` // filter: step ID whose output to iterate
FilterBody *Step  `yaml:"-"` // filter: predicate step (truthy output = keep)
```

- [ ] **Step 4: Add `convertFilter` and register in `convertForm`**

In `internal/pipeline/sexpr.go`:

```go
// convertFilter: (filter "step-id" (step "pred" ...))
func convertFilter(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (filter) needs source step ID and predicate step", n.Line)
	}
	source := resolveVal(children[0], defs)
	body, err := convertStep(children[1], defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: filter body: %w", n.Line, err)
	}
	return Step{
		ID:         fmt.Sprintf("filter-%d", n.Line),
		Form:       "filter",
		FilterOver: source,
		FilterBody: &body,
	}, nil
}
```

In `convertForm`:

```go
case "filter":
	s, err := convertFilter(n, defs)
	if err != nil {
		return nil, err
	}
	return []Step{s}, nil
```

- [ ] **Step 5: Add `executeFilter` in runner**

In `internal/pipeline/runner.go`:

```go
// executeFilter iterates over NDJSON lines from a source step, runs the predicate
// step per item, and keeps items where the output is truthy (non-empty, not "false", not "0").
func executeFilter(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	snap := rctx.stepsSnapshot()
	source, ok := snap[step.FilterOver]
	if !ok {
		return nil, fmt.Errorf("filter %s: source step %q has no output", step.ID, step.FilterOver)
	}

	lines := strings.Split(strings.TrimSpace(source), "\n")
	var kept []string

	for idx, item := range lines {
		if item == "" {
			continue
		}
		body := *step.FilterBody
		body.ID = fmt.Sprintf("%s-%d", step.FilterBody.ID, idx)

		origParams := rctx.params
		filterParams := make(map[string]string, len(origParams)+2)
		for k, v := range origParams {
			filterParams[k] = v
		}
		filterParams["item"] = item
		filterParams["item_index"] = fmt.Sprintf("%d", idx)
		rctx.params = filterParams

		outcome, err := executeStep(ctx, rctx, body)
		rctx.params = origParams
		if err != nil {
			return nil, fmt.Errorf("filter %s item %d: %w", step.ID, idx, err)
		}

		result := strings.TrimSpace(outcome.output)
		if result != "" && result != "false" && result != "0" {
			kept = append(kept, item)
		}
	}

	combined := strings.Join(kept, "\n")
	rctx.mu.Lock()
	rctx.steps[step.ID] = combined
	rctx.mu.Unlock()
	return &stepOutcome{output: combined}, nil
}
```

Wire into `executeStep`:

```go
case "filter":
	return executeFilter(ctx, rctx, step)
```

- [ ] **Step 6: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Filter -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add (filter) collection form — keep items matching predicate"
```

---

### Task 11: Add `(reduce)` collection form

**Files:**
- Modify: `internal/pipeline/types.go` (add `ReduceOver` and `ReduceBody` fields to Step)
- Modify: `internal/pipeline/sexpr.go` (add `reduce` to `convertForm`, add `convertReduce`)
- Modify: `internal/pipeline/runner.go` (add `executeReduce` handler)
- Test: `internal/pipeline/sexpr_test.go`

`(reduce "step-id" (step "fold" ...))` iterates over NDJSON lines, running the body step per item with `{{.param.item}}` (current item) and `{{.param.accumulator}}` (output of previous iteration, empty string initially). Final output is the last accumulator value. Like Clojure's `reduce`.

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_Reduce(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "data" (run "echo 'a\nb\nc'"))
  (reduce "data"
    (step "fold" (run "echo '{{.param.accumulator}},{{.param.item}}'"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}
	s := w.Steps[1]
	if s.Form != "reduce" {
		t.Fatalf("expected form %q, got %q", "reduce", s.Form)
	}
	if s.ReduceOver != "data" {
		t.Fatalf("expected reduce over %q, got %q", "data", s.ReduceOver)
	}
	if s.ReduceBody == nil {
		t.Fatal("expected reduce body")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Reduce -v`
Expected: FAIL — `ReduceOver` undefined

- [ ] **Step 3: Add reduce fields to Step struct**

In `internal/pipeline/types.go`, add after the `FilterBody` field:

```go
ReduceOver string `yaml:"-"` // reduce: step ID whose output to iterate
ReduceBody *Step  `yaml:"-"` // reduce: body step (receives item + accumulator)
```

- [ ] **Step 4: Add `convertReduce` and register in `convertForm`**

In `internal/pipeline/sexpr.go`:

```go
// convertReduce: (reduce "step-id" (step "fold" ...))
func convertReduce(n *sexpr.Node, defs map[string]string) (Step, error) {
	children := n.Children[1:]
	if len(children) < 2 {
		return Step{}, fmt.Errorf("line %d: (reduce) needs source step ID and body step", n.Line)
	}
	source := resolveVal(children[0], defs)
	body, err := convertStep(children[1], defs)
	if err != nil {
		return Step{}, fmt.Errorf("line %d: reduce body: %w", n.Line, err)
	}
	return Step{
		ID:         fmt.Sprintf("reduce-%d", n.Line),
		Form:       "reduce",
		ReduceOver: source,
		ReduceBody: &body,
	}, nil
}
```

In `convertForm`:

```go
case "reduce":
	s, err := convertReduce(n, defs)
	if err != nil {
		return nil, err
	}
	return []Step{s}, nil
```

- [ ] **Step 5: Add `executeReduce` in runner**

In `internal/pipeline/runner.go`:

```go
// executeReduce folds over NDJSON lines from a source step.
// Each iteration receives {{.param.item}} and {{.param.accumulator}}.
// The accumulator starts as "" and is updated with each step's output.
func executeReduce(ctx context.Context, rctx *runCtx, step Step) (*stepOutcome, error) {
	snap := rctx.stepsSnapshot()
	source, ok := snap[step.ReduceOver]
	if !ok {
		return nil, fmt.Errorf("reduce %s: source step %q has no output", step.ID, step.ReduceOver)
	}

	lines := strings.Split(strings.TrimSpace(source), "\n")
	accumulator := ""

	for idx, item := range lines {
		if item == "" {
			continue
		}
		body := *step.ReduceBody
		body.ID = fmt.Sprintf("%s-%d", step.ReduceBody.ID, idx)

		origParams := rctx.params
		reduceParams := make(map[string]string, len(origParams)+3)
		for k, v := range origParams {
			reduceParams[k] = v
		}
		reduceParams["item"] = item
		reduceParams["item_index"] = fmt.Sprintf("%d", idx)
		reduceParams["accumulator"] = accumulator
		rctx.params = reduceParams

		outcome, err := executeStep(ctx, rctx, body)
		rctx.params = origParams
		if err != nil {
			return nil, fmt.Errorf("reduce %s item %d: %w", step.ID, idx, err)
		}
		accumulator = outcome.output
	}

	rctx.mu.Lock()
	rctx.steps[step.ID] = accumulator
	rctx.mu.Unlock()
	return &stepOutcome{output: accumulator}, nil
}
```

Wire into `executeStep`:

```go
case "reduce":
	return executeReduce(ctx, rctx, step)
```

- [ ] **Step 6: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run TestSexprWorkflow_Reduce -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/types.go internal/pipeline/sexpr.go internal/pipeline/runner.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add (reduce) collection form — fold items with accumulator"
```

---

### Task 12: Add `(->)` threading macro

**Files:**
- Modify: `internal/pipeline/sexpr.go` (add `->` to `convertForm`, add `convertThread`)
- Test: `internal/pipeline/sexpr_test.go`

`(-> (search ...) (flatten) (each (step ...)))` threads the output of each form into the next. This is syntactic sugar — the converter expands it into named steps with auto-generated IDs. Each form in the thread implicitly references the previous form's output.

The threading macro works at parse time, not runtime. It desugars into regular steps:
- `(-> (search ...) (flatten) (each ...))` becomes:
  - `(step "thread-N-0" (search ...))`
  - `(step "thread-N-1" (flatten "thread-N-0"))`
  - `(each "thread-N-1" ...)`

- [ ] **Step 1: Write failing parser test**

Add to `internal/pipeline/sexpr_test.go`:

```go
func TestSexprWorkflow_Thread(t *testing.T) {
	src := []byte(`
(workflow "test"
  (-> (search :index "my-index" :ndjson)
      (each
        (step "classify" (run "echo classified")))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// -> with 2 forms: search becomes a step, each references it
	if len(w.Steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(w.Steps))
	}
	// First step should be the search
	if w.Steps[0].Search == nil {
		t.Fatal("expected first step to be search")
	}
	// Second step should be map/each referencing first
	s1 := w.Steps[1]
	if s1.Form != "map" {
		t.Fatalf("expected form %q, got %q", "map", s1.Form)
	}
	if s1.MapOver != w.Steps[0].ID {
		t.Fatalf("expected map over %q, got %q", w.Steps[0].ID, s1.MapOver)
	}
}

func TestSexprWorkflow_ThreadWithFlatten(t *testing.T) {
	src := []byte(`
(workflow "test"
  (-> (search :index "my-index")
      (flatten)
      (each
        (step "process" (run "echo done")))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(w.Steps))
	}
	// Step 0: search
	if w.Steps[0].Search == nil {
		t.Fatal("expected step 0 to be search")
	}
	// Step 1: flatten referencing step 0
	if w.Steps[1].Flatten != w.Steps[0].ID {
		t.Fatalf("expected flatten over %q, got %q", w.Steps[0].ID, w.Steps[1].Flatten)
	}
	// Step 2: each referencing step 1
	if w.Steps[2].MapOver != w.Steps[1].ID {
		t.Fatalf("expected each over %q, got %q", w.Steps[1].ID, w.Steps[2].MapOver)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestSexprWorkflow_Thread" -v`
Expected: FAIL — unknown form "->"

- [ ] **Step 3: Add `convertThread` function**

In `internal/pipeline/sexpr.go`:

```go
// convertThread: (-> form1 form2 form3 ...)
// Desugars into a sequence of steps where each form implicitly references the previous.
// SDK forms (search, index, etc.) are wrapped in auto-named steps.
// Collection forms (each/map, filter, reduce) have their source set to the previous step.
// (flatten) with no args becomes (flatten "prev-step-id").
func convertThread(n *sexpr.Node, defs map[string]string) ([]Step, error) {
	children := n.Children[1:] // skip "->"
	if len(children) < 2 {
		return nil, fmt.Errorf("line %d: (->) needs at least 2 forms", n.Line)
	}

	var steps []Step
	prevID := ""

	for i, child := range children {
		if !child.IsList() || len(child.Children) == 0 {
			return nil, fmt.Errorf("line %d: (->) children must be forms", child.Line)
		}
		head := child.Children[0].SymbolVal()
		if head == "" {
			head = child.Children[0].StringVal()
		}
		threadID := fmt.Sprintf("thread-%d-%d", n.Line, i)

		switch head {
		case "search", "index", "delete", "embed", "run", "llm",
			"json-pick", "pick", "lines", "merge", "http-get", "fetch",
			"http-post", "send", "read-file", "read", "write-file", "write", "glob", "plugin":
			// SDK/primitive form — wrap in a step
			wrappedNode := &sexpr.Node{
				Children: append([]*sexpr.Node{
					{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: "step"}},
					{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: threadID}},
				}, child),
				Line: child.Line,
			}
			s, err := convertStep(wrappedNode, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = threadID

		case "flatten":
			// (flatten) with no args or (flatten "explicit") — auto-fill source if missing
			s := Step{ID: threadID, Flatten: prevID}
			if len(child.Children) >= 2 {
				s.Flatten = resolveVal(child.Children[1], defs)
			}
			steps = append(steps, s)
			prevID = threadID

		case "each", "map":
			// Rewrite source to prevID
			if prevID == "" {
				return nil, fmt.Errorf("line %d: (->) each/map has no preceding step", child.Line)
			}
			// Build a new node with the source injected
			newChildren := make([]*sexpr.Node, 0, len(child.Children)+1)
			newChildren = append(newChildren, child.Children[0]) // "each"/"map"
			newChildren = append(newChildren, &sexpr.Node{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: prevID}})
			newChildren = append(newChildren, child.Children[1:]...) // body
			rewritten := &sexpr.Node{Children: newChildren, Line: child.Line}
			s, err := convertMap(rewritten, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		case "filter":
			if prevID == "" {
				return nil, fmt.Errorf("line %d: (->) filter has no preceding step", child.Line)
			}
			newChildren := make([]*sexpr.Node, 0, len(child.Children)+1)
			newChildren = append(newChildren, child.Children[0])
			newChildren = append(newChildren, &sexpr.Node{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: prevID}})
			newChildren = append(newChildren, child.Children[1:]...)
			rewritten := &sexpr.Node{Children: newChildren, Line: child.Line}
			s, err := convertFilter(rewritten, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		case "reduce":
			if prevID == "" {
				return nil, fmt.Errorf("line %d: (->) reduce has no preceding step", child.Line)
			}
			newChildren := make([]*sexpr.Node, 0, len(child.Children)+1)
			newChildren = append(newChildren, child.Children[0])
			newChildren = append(newChildren, &sexpr.Node{Atom: &sexpr.Token{Type: sexpr.TokenString, Val: prevID}})
			newChildren = append(newChildren, child.Children[1:]...)
			rewritten := &sexpr.Node{Children: newChildren, Line: child.Line}
			s, err := convertReduce(rewritten, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		case "step":
			// Named step — just convert normally
			s, err := convertStep(child, defs)
			if err != nil {
				return nil, fmt.Errorf("line %d: thread form %d: %w", child.Line, i, err)
			}
			steps = append(steps, s)
			prevID = s.ID

		default:
			return nil, fmt.Errorf("line %d: unknown form %q in (->)", child.Line, head)
		}
	}
	return steps, nil
}
```

- [ ] **Step 4: Register `->` in `convertForm`**

In `internal/pipeline/sexpr.go`, add to the `convertForm` switch:

```go
case "->":
	return convertThread(n, defs)
```

- [ ] **Step 5: Run tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestSexprWorkflow_Thread" -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat(dsl): add (->) threading macro — pipe data between forms"
```

---

### Task 13: Run full test suite and verify

- [ ] **Step 1: Run all pipeline and esearch tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ ./internal/esearch/ ./internal/provider/ -v -count=1`
Expected: All PASS

- [ ] **Step 2: Run go vet**

Run: `cd /Users/stokes/Projects/gl1tch && go vet ./internal/pipeline/ ./internal/esearch/`
Expected: No issues

- [ ] **Step 3: Verify build**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: Clean build

---

## Form Summary

After all tasks, the DSL supports these forms (~22 total):

| Form | Type | Description |
|------|------|-------------|
| `(step)` | core | Named unit of work |
| `(run)` | primitive | Shell command |
| `(llm)` | primitive | LLM call (now with `:format "json"`) |
| `(save)` | primitive | Write to file |
| `(search)` | ES | Query ES (now with optional `:query`, `:sort`, `:ndjson`) |
| `(index)` | ES | Index doc (now with `:upsert false`) |
| `(delete)` | ES | Delete by query |
| `(embed)` | ES | Generate embedding |
| `(flatten)` | transform | **NEW** JSON array → NDJSON |
| `(json-pick)` | transform | jq expression |
| `(lines)` | transform | Split by newlines |
| `(merge)` | transform | Merge step outputs |
| `(map)`/`(each)` | collection | Iterate items |
| `(filter)` | collection | **NEW** Keep items matching predicate |
| `(reduce)` | collection | **NEW** Fold items with accumulator |
| `(par)` | control | Concurrent execution |
| `(cond)` | control | Multi-branch conditional |
| `(when)`/`(when-not)` | control | **NEW** Single-branch conditional |
| `(catch)` | control | Error handling |
| `(retry)` | control | Retry on failure |
| `(timeout)` | control | Deadline |
| `(->)` | macro | **NEW** Threading — pipe data between forms |
| `(let)` | binding | Local definitions |
| `(phase)` | structure | Retriable unit with gates |

Template functions: `step`, `stepfile`, `split`, `join`, `last`, `first`, `upper`, `lower`, `trim`, `trimPrefix`, `trimSuffix`, `replace`, `truncate`, `contains`, `hasPrefix`, `hasSuffix`, **`pick`** (NEW), **`assoc`** (NEW)
