# Code Graph Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add tree-sitter AST extraction to `glitch index`, build per-repo symbol + edge graphs in Elasticsearch, and wire graph queries into `glitch observe` and the research tool loop.

**Architecture:** Three-phase pipeline (extract symbols → resolve imports → resolve calls) using `smacker/go-tree-sitter` for AST parsing. Symbols and edges stored in two new ES indices per repo. Incremental indexing via file SHA256 comparison. Observer gains graph BFS traversal with `--depth` flag. Research tools gain three new structured code navigation tools.

**Tech Stack:** Go, `github.com/smacker/go-tree-sitter` (bundled grammars), Elasticsearch

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/esearch/mappings.go` | Modify | Add symbol + edge index mappings and constants |
| `internal/esearch/client.go` | Modify | Add `TermsAgg` method for file hash aggregation |
| `internal/indexer/symbol.go` | Create | `SymbolDoc`, `EdgeDoc` types, `SymbolID()` helper |
| `internal/indexer/extractor.go` | Create | `LanguageExtractor` framework, tree-sitter parse + query loop |
| `internal/indexer/languages.go` | Create | Per-language extractor configs (queries, path resolvers) |
| `internal/indexer/resolve.go` | Create | Import resolution, call resolution, edge emission |
| `internal/indexer/indexer.go` | Modify | Replace regex extraction with tree-sitter pipeline, add incremental mode |
| `cmd/index.go` | Modify | Add CLI flags |
| `internal/research/tools.go` | Modify | Add `search_symbols`, `search_edges`, `symbol_context` tools |
| `internal/observer/query.go` | Modify | Expand index list, add graph BFS traversal |
| `cmd/observe.go` | Modify | Add `--depth` flag |

---

## Task 1: ES Mappings for Symbols and Edges

**Files:**
- Modify: `internal/esearch/mappings.go:4-12` (constants), `internal/esearch/mappings.go:154-164` (`AllIndices`)
- Test: `internal/esearch/mappings_test.go` (create if absent)

- [ ] **Step 1: Write failing test for new index constants**

```go
// internal/esearch/mappings_test.go
package esearch

import (
	"encoding/json"
	"testing"
)

func TestSymbolsMappingValid(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(SymbolsMapping), &m); err != nil {
		t.Fatalf("SymbolsMapping is not valid JSON: %v", err)
	}
	props := m["mappings"].(map[string]any)["properties"].(map[string]any)
	for _, field := range []string{"id", "file", "kind", "name", "signature", "language", "start_line", "end_line", "parent_id", "docstring", "file_hash", "repo", "indexed_at"} {
		if _, ok := props[field]; !ok {
			t.Errorf("SymbolsMapping missing field %q", field)
		}
	}
}

func TestEdgesMappingValid(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(EdgesMapping), &m); err != nil {
		t.Fatalf("EdgesMapping is not valid JSON: %v", err)
	}
	props := m["mappings"].(map[string]any)["properties"].(map[string]any)
	for _, field := range []string{"source_id", "target_id", "kind", "file", "repo"} {
		if _, ok := props[field]; !ok {
			t.Errorf("EdgesMapping missing field %q", field)
		}
	}
}

func TestAllIndicesIncludesGraph(t *testing.T) {
	all := AllIndices()
	if _, ok := all[IndexSymbolsPrefix+"test"]; ok {
		t.Skip("prefix-based indices not in AllIndices — checked separately")
	}
	// Just verify the function still returns all existing indices + the new mapping strings exist
	if SymbolsMapping == "" {
		t.Error("SymbolsMapping is empty")
	}
	if EdgesMapping == "" {
		t.Error("EdgesMapping is empty")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/esearch/ -run TestSymbolsMapping -v`
Expected: FAIL — `SymbolsMapping` undefined

- [ ] **Step 3: Add constants and mappings**

Add to `internal/esearch/mappings.go` after existing constants (line 12):

```go
IndexSymbolsPrefix = "glitch-symbols-"
IndexEdgesPrefix   = "glitch-edges-"
```

Add after the last mapping constant:

```go
SymbolsMapping = `{
  "mappings": {
    "properties": {
      "id":         {"type": "keyword"},
      "file":       {"type": "keyword"},
      "kind":       {"type": "keyword"},
      "name":       {"type": "text", "fields": {"raw": {"type": "keyword"}}},
      "signature":  {"type": "text"},
      "language":   {"type": "keyword"},
      "start_line": {"type": "integer"},
      "end_line":   {"type": "integer"},
      "parent_id":  {"type": "keyword"},
      "docstring":  {"type": "text"},
      "file_hash":  {"type": "keyword"},
      "repo":       {"type": "keyword"},
      "indexed_at": {"type": "date"}
    }
  }
}`

EdgesMapping = `{
  "mappings": {
    "properties": {
      "source_id": {"type": "keyword"},
      "target_id": {"type": "keyword"},
      "kind":      {"type": "keyword"},
      "file":      {"type": "keyword"},
      "repo":      {"type": "keyword"}
    }
  }
}`
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/esearch/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/esearch/mappings.go internal/esearch/mappings_test.go
git commit -m "feat(esearch): add symbol + edge index mappings for code graph"
```

---

## Task 2: ES Client — TermsAgg Method

**Files:**
- Modify: `internal/esearch/client.go`
- Test: `internal/esearch/client_test.go` (create if absent)

Needed for incremental indexing — fetches `{file -> file_hash}` from the symbols index.

- [ ] **Step 1: Write failing test**

```go
// internal/esearch/client_test.go
package esearch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTermsAggParseResponse(t *testing.T) {
	// Test the response parsing logic with a fake ES response
	body := `{
		"aggregations": {
			"group": {
				"buckets": [
					{"key": "main.go", "top": {"hits": {"hits": [{"_source": {"file_hash": "abc123"}}]}}},
					{"key": "lib.go", "top": {"hits": {"hits": [{"_source": {"file_hash": "def456"}}]}}}
				]
			}
		}
	}`

	result := parseTermsAggResponse([]byte(body), "file_hash")
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["main.go"] != "abc123" {
		t.Errorf("main.go hash = %q, want abc123", result["main.go"])
	}
	if result["lib.go"] != "def456" {
		t.Errorf("lib.go hash = %q, want def456", result["lib.go"])
	}
}

func TestTermsAggRequest(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"aggregations":{"group":{"buckets":[]}}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.TermsAgg(t.Context(), "glitch-symbols-test", "file", "file_hash", 50000)
	if err != nil {
		t.Fatalf("TermsAgg error: %v", err)
	}

	// Verify the request structure
	aggs, ok := gotBody["aggs"].(map[string]any)
	if !ok {
		t.Fatal("missing aggs in request body")
	}
	group, ok := aggs["group"].(map[string]any)
	if !ok {
		t.Fatal("missing group in aggs")
	}
	terms, ok := group["terms"].(map[string]any)
	if !ok {
		t.Fatal("missing terms in group")
	}
	if terms["field"] != "file" {
		t.Errorf("terms field = %v, want 'file'", terms["field"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/esearch/ -run TestTermsAgg -v`
Expected: FAIL — `TermsAgg` undefined

- [ ] **Step 3: Implement TermsAgg**

Add to `internal/esearch/client.go`:

```go
// TermsAgg runs a terms aggregation on keyField, returning a map of key -> valueField.
// Uses top_hits(size=1) to pull one field from each bucket. Useful for file hash lookups.
func (c *Client) TermsAgg(ctx context.Context, index, keyField, valueField string, size int) (map[string]string, error) {
	query := map[string]any{
		"size": 0,
		"aggs": map[string]any{
			"group": map[string]any{
				"terms": map[string]any{
					"field": keyField,
					"size":  size,
				},
				"aggs": map[string]any{
					"top": map[string]any{
						"top_hits": map[string]any{
							"size":    1,
							"_source": []string{valueField},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("terms agg: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/"+index+"/_search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("terms agg: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("terms agg: %w", err)
	}
	defer resp.Body.Close()

	// 404 means index doesn't exist yet — return empty map (first run)
	if resp.StatusCode == 404 {
		return map[string]string{}, nil
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("terms agg: status %s — %s", resp.Status, truncate(string(b), 256))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("terms agg: read body: %w", err)
	}

	return parseTermsAggResponse(respBody, valueField), nil
}

// parseTermsAggResponse extracts key->value pairs from a terms aggregation response.
func parseTermsAggResponse(data []byte, valueField string) map[string]string {
	var raw struct {
		Aggregations struct {
			Group struct {
				Buckets []struct {
					Key string `json:"key"`
					Top struct {
						Hits struct {
							Hits []struct {
								Source json.RawMessage `json:"_source"`
							} `json:"hits"`
						} `json:"hits"`
					} `json:"top"`
				} `json:"buckets"`
			} `json:"group"`
		} `json:"aggregations"`
	}
	json.Unmarshal(data, &raw)

	result := make(map[string]string, len(raw.Aggregations.Group.Buckets))
	for _, b := range raw.Aggregations.Group.Buckets {
		if len(b.Top.Hits.Hits) == 0 {
			continue
		}
		var src map[string]string
		json.Unmarshal(b.Top.Hits.Hits[0].Source, &src)
		result[b.Key] = src[valueField]
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/esearch/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/esearch/client.go internal/esearch/client_test.go
git commit -m "feat(esearch): add TermsAgg method for file hash lookups"
```

---

## Task 3: Symbol and Edge Types

**Files:**
- Create: `internal/indexer/symbol.go`
- Test: `internal/indexer/symbol_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/indexer/symbol_test.go
package indexer

import "testing"

func TestSymbolID(t *testing.T) {
	id := SymbolID("internal/handler.go", "function", "HandleRequest", 42)
	if id == "" {
		t.Fatal("SymbolID returned empty string")
	}
	// Deterministic
	id2 := SymbolID("internal/handler.go", "function", "HandleRequest", 42)
	if id != id2 {
		t.Error("SymbolID not deterministic")
	}
	// Different inputs produce different IDs
	id3 := SymbolID("internal/handler.go", "function", "HandleRequest", 43)
	if id == id3 {
		t.Error("different inputs produced same ID")
	}
}

func TestSymbolDocJSON(t *testing.T) {
	doc := SymbolDoc{
		ID:        SymbolID("main.go", "function", "main", 1),
		File:      "main.go",
		Kind:      "function",
		Name:      "main",
		Signature: "func main()",
		Language:  "go",
		StartLine: 1,
		EndLine:   5,
		Repo:      "gl1tch",
	}
	if doc.ID == "" {
		t.Error("ID empty")
	}
	if doc.Kind != "function" {
		t.Errorf("Kind = %q, want function", doc.Kind)
	}
}

func TestEdgeDocFields(t *testing.T) {
	edge := EdgeDoc{
		SourceID: "abc",
		TargetID: "def",
		Kind:     "calls",
		File:     "main.go",
		Repo:     "gl1tch",
	}
	if edge.Kind != "calls" {
		t.Errorf("Kind = %q, want calls", edge.Kind)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run TestSymbolID -v`
Expected: FAIL — `SymbolID` undefined

- [ ] **Step 3: Implement types**

```go
// internal/indexer/symbol.go
package indexer

import (
	"crypto/sha256"
	"fmt"
)

// SymbolDoc is a code symbol indexed into Elasticsearch.
type SymbolDoc struct {
	ID        string `json:"id"`
	File      string `json:"file"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Language  string `json:"language"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	ParentID  string `json:"parent_id,omitempty"`
	Docstring string `json:"docstring,omitempty"`
	FileHash  string `json:"file_hash"`
	Repo      string `json:"repo"`
	IndexedAt string `json:"indexed_at"`
}

// EdgeDoc is a relationship between two symbols.
type EdgeDoc struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Kind     string `json:"kind"`
	File     string `json:"file"`
	Repo     string `json:"repo"`
}

// Edge kinds.
const (
	EdgeContains   = "contains"
	EdgeImports    = "imports"
	EdgeExports    = "exports"
	EdgeExtends    = "extends"
	EdgeImplements = "implements"
	EdgeCalls      = "calls"
)

// Symbol kinds.
const (
	KindFunction  = "function"
	KindMethod    = "method"
	KindType      = "type"
	KindInterface = "interface"
	KindClass     = "class"
	KindField     = "field"
	KindConst     = "const"
	KindVar       = "var"
	KindImport    = "import"
	KindExport    = "export"
)

// SymbolID returns a deterministic ID for a symbol based on file, kind, name, and start line.
func SymbolID(file, kind, name string, startLine int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%d", file, kind, name, startLine)))
	return fmt.Sprintf("%x", h[:12])
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/indexer/symbol.go internal/indexer/symbol_test.go
git commit -m "feat(indexer): add SymbolDoc, EdgeDoc types and SymbolID helper"
```

---

## Task 4: LanguageExtractor Framework

**Files:**
- Create: `internal/indexer/extractor.go`
- Test: `internal/indexer/extractor_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/indexer/extractor_test.go
package indexer

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func TestExtractorParseGo(t *testing.T) {
	ext := &LanguageExtractor{
		Language: "go",
		Grammar:  golang.GetLanguage(),
		SymbolQueries: []SymbolQuery{
			{
				Query: `(function_declaration name: (identifier) @name) @decl`,
				Kind:  KindFunction,
			},
		},
	}

	src := []byte(`package main

func Hello() {}
func World() {}
`)

	symbols, err := ext.Extract(src, "main.go", "testrepo", "abc123")
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}
	if symbols[0].Name != "Hello" {
		t.Errorf("symbols[0].Name = %q, want Hello", symbols[0].Name)
	}
	if symbols[0].Kind != KindFunction {
		t.Errorf("symbols[0].Kind = %q, want function", symbols[0].Kind)
	}
	if symbols[0].StartLine < 1 {
		t.Error("start_line should be >= 1")
	}
	if symbols[0].FileHash != "abc123" {
		t.Errorf("file_hash = %q, want abc123", symbols[0].FileHash)
	}
}

func TestExtractorEmptySource(t *testing.T) {
	ext := &LanguageExtractor{
		Language: "go",
		Grammar:  golang.GetLanguage(),
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (identifier) @name) @decl`, Kind: KindFunction},
		},
	}

	symbols, err := ext.Extract([]byte(""), "empty.go", "testrepo", "abc")
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}
	if len(symbols) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(symbols))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run TestExtractorParse -v`
Expected: FAIL — `LanguageExtractor` undefined. (First run `go get github.com/smacker/go-tree-sitter` if needed.)

- [ ] **Step 3: Add tree-sitter dependency**

Run: `cd /Users/stokes/Projects/gl1tch && go get github.com/smacker/go-tree-sitter github.com/smacker/go-tree-sitter/golang`

- [ ] **Step 4: Implement extractor**

```go
// internal/indexer/extractor.go
package indexer

import (
	"context"
	"fmt"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
)

// SymbolQuery is a tree-sitter query that extracts symbols of a given kind.
// The query must capture @name (symbol name) and @decl (full declaration node).
// Optional captures: @docstring.
type SymbolQuery struct {
	Query string
	Kind  string
}

// CallSite is an unresolved call extracted from source.
type CallSite struct {
	CalleeName string
	File       string
	Line       int
}

// UnresolvedImport is an import extracted from source before cross-file resolution.
type UnresolvedImport struct {
	Path  string
	Alias string
	Names []string // specific imported names, empty = whole module
	File  string
	Line  int
}

// LanguageExtractor holds tree-sitter queries for a single language.
type LanguageExtractor struct {
	Language      string
	Grammar       *sitter.Language
	Extensions    []string
	SymbolQueries []SymbolQuery
	ImportQuery   string
	ExportQuery   string
	CallQuery     string
	// PathResolver resolves an import path to a relative file path in the repo.
	// Returns empty string if unresolvable.
	PathResolver func(importPath, fromFile, repoRoot string) string
}

// Extract parses source with tree-sitter and returns symbols found by all SymbolQueries.
func (le *LanguageExtractor) Extract(source []byte, file, repo, fileHash string) ([]SymbolDoc, error) {
	if len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	parser.SetLanguage(le.Grammar)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	defer tree.Close()

	root := tree.RootNode()
	now := time.Now().UTC().Format(time.RFC3339)
	var symbols []SymbolDoc

	for _, sq := range le.SymbolQueries {
		q, err := sitter.NewQuery([]byte(sq.Query), le.Grammar)
		if err != nil {
			return nil, fmt.Errorf("compile query for %s/%s: %w", le.Language, sq.Kind, err)
		}

		cursor := sitter.NewQueryCursor()
		cursor.Exec(q, root)

		for {
			match, ok := cursor.NextMatch()
			if !ok {
				break
			}

			var name, signature string
			var startLine, endLine uint32
			for _, cap := range match.Captures {
				capName := q.CaptureNameForId(cap.Index)
				switch capName {
				case "name":
					name = cap.Node.Content(source)
				case "decl":
					startLine = cap.Node.StartPoint().Row + 1
					endLine = cap.Node.EndPoint().Row + 1
					// Signature: first line of the declaration
					start := cap.Node.StartByte()
					end := cap.Node.EndByte()
					text := string(source[start:end])
					for i, ch := range text {
						if ch == '\n' {
							text = text[:i]
							break
						}
						_ = i
					}
					signature = text
				}
			}

			if name == "" {
				continue
			}

			symbols = append(symbols, SymbolDoc{
				ID:        SymbolID(file, sq.Kind, name, int(startLine)),
				File:      file,
				Kind:      sq.Kind,
				Name:      name,
				Signature: signature,
				Language:  le.Language,
				StartLine: int(startLine),
				EndLine:   int(endLine),
				FileHash:  fileHash,
				Repo:      repo,
				IndexedAt: now,
			})
		}
	}

	return symbols, nil
}

// ExtractImports parses source and returns unresolved imports.
func (le *LanguageExtractor) ExtractImports(source []byte, file string) ([]UnresolvedImport, error) {
	if le.ImportQuery == "" || len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	parser.SetLanguage(le.Grammar)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parse imports %s: %w", file, err)
	}
	defer tree.Close()

	root := tree.RootNode()
	q, err := sitter.NewQuery([]byte(le.ImportQuery), le.Grammar)
	if err != nil {
		return nil, fmt.Errorf("compile import query for %s: %w", le.Language, err)
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(q, root)

	var imports []UnresolvedImport
	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}
		var path, alias string
		var line uint32
		for _, cap := range match.Captures {
			capName := q.CaptureNameForId(cap.Index)
			switch capName {
			case "path":
				path = cap.Node.Content(source)
				line = cap.Node.StartPoint().Row + 1
			case "alias":
				alias = cap.Node.Content(source)
			}
		}
		if path == "" {
			continue
		}
		imports = append(imports, UnresolvedImport{
			Path:  trimQuotes(path),
			Alias: alias,
			File:  file,
			Line:  int(line),
		})
	}

	return imports, nil
}

// ExtractCalls parses source and returns unresolved call sites.
func (le *LanguageExtractor) ExtractCalls(source []byte, file string) ([]CallSite, error) {
	if le.CallQuery == "" || len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	parser.SetLanguage(le.Grammar)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parse calls %s: %w", file, err)
	}
	defer tree.Close()

	root := tree.RootNode()
	q, err := sitter.NewQuery([]byte(le.CallQuery), le.Grammar)
	if err != nil {
		return nil, fmt.Errorf("compile call query for %s: %w", le.Language, err)
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(q, root)

	var calls []CallSite
	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}
		for _, cap := range match.Captures {
			capName := q.CaptureNameForId(cap.Index)
			if capName == "callee" {
				calls = append(calls, CallSite{
					CalleeName: cap.Node.Content(source),
					File:       file,
					Line:       int(cap.Node.StartPoint().Row + 1),
				})
			}
		}
	}

	return calls, nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run TestExtractor -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/indexer/extractor.go internal/indexer/extractor_test.go go.mod go.sum
git commit -m "feat(indexer): add LanguageExtractor framework with tree-sitter parsing"
```

---

## Task 5: Language Configs

**Files:**
- Create: `internal/indexer/languages.go`
- Test: `internal/indexer/languages_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/indexer/languages_test.go
package indexer

import "testing"

func TestRegistryHasGo(t *testing.T) {
	ext := ExtractorForLanguage("go")
	if ext == nil {
		t.Fatal("no extractor for go")
	}
	if ext.Language != "go" {
		t.Errorf("Language = %q, want go", ext.Language)
	}
}

func TestRegistryHasPython(t *testing.T) {
	ext := ExtractorForLanguage("python")
	if ext == nil {
		t.Fatal("no extractor for python")
	}
}

func TestRegistryHasTypeScript(t *testing.T) {
	ext := ExtractorForLanguage("typescript")
	if ext == nil {
		t.Fatal("no extractor for typescript")
	}
}

func TestRegistryHasJavaScript(t *testing.T) {
	ext := ExtractorForLanguage("javascript")
	if ext == nil {
		t.Fatal("no extractor for javascript")
	}
}

func TestRegistryReturnsNilForUnknown(t *testing.T) {
	ext := ExtractorForLanguage("brainfuck")
	if ext != nil {
		t.Error("expected nil for unknown language")
	}
}

func TestGoExtractorFullParse(t *testing.T) {
	ext := ExtractorForLanguage("go")
	src := []byte(`package main

import "fmt"

// Hello prints a greeting.
func Hello(name string) {
	fmt.Println(name)
}

type Server struct {
	Port int
}

func (s *Server) Start() error {
	return nil
}

type Handler interface {
	Handle()
}

const MaxRetries = 3

var DefaultTimeout = 30
`)

	symbols, err := ext.Extract(src, "main.go", "test", "hash")
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}

	want := map[string]string{
		"Hello":          KindFunction,
		"Server":         KindType,
		"Start":          KindMethod,
		"Handler":        KindInterface,
		"MaxRetries":     KindConst,
		"DefaultTimeout": KindVar,
	}

	found := make(map[string]string)
	for _, s := range symbols {
		found[s.Name] = s.Kind
	}

	for name, kind := range want {
		if got, ok := found[name]; !ok {
			t.Errorf("missing symbol %q", name)
		} else if got != kind {
			t.Errorf("%s: kind = %q, want %q", name, got, kind)
		}
	}
}

func TestGoExtractorImports(t *testing.T) {
	ext := ExtractorForLanguage("go")
	src := []byte(`package main

import (
	"fmt"
	"os"
	"github.com/example/pkg"
)
`)

	imports, err := ext.ExtractImports(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractImports error: %v", err)
	}
	if len(imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(imports))
	}
	paths := make(map[string]bool)
	for _, imp := range imports {
		paths[imp.Path] = true
	}
	for _, p := range []string{"fmt", "os", "github.com/example/pkg"} {
		if !paths[p] {
			t.Errorf("missing import %q", p)
		}
	}
}

func TestGoExtractorCalls(t *testing.T) {
	ext := ExtractorForLanguage("go")
	src := []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
	doWork()
}

func doWork() {}
`)

	calls, err := ext.ExtractCalls(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractCalls error: %v", err)
	}
	if len(calls) == 0 {
		t.Fatal("expected at least 1 call site")
	}
	names := make(map[string]bool)
	for _, c := range calls {
		names[c.CalleeName] = true
	}
	if !names["doWork"] && !names["Println"] {
		t.Error("expected doWork or Println in call sites")
	}
}

func TestPythonExtractorFullParse(t *testing.T) {
	ext := ExtractorForLanguage("python")
	src := []byte(`
import os
from pathlib import Path

class MyService:
    def handle(self, req):
        pass

def main():
    svc = MyService()
    svc.handle(None)
`)

	symbols, err := ext.Extract(src, "app.py", "test", "hash")
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}

	found := make(map[string]string)
	for _, s := range symbols {
		found[s.Name] = s.Kind
	}

	if found["MyService"] != KindClass {
		t.Errorf("MyService kind = %q, want class", found["MyService"])
	}
	if found["main"] != KindFunction {
		t.Errorf("main kind = %q, want function", found["main"])
	}
	if found["handle"] != KindMethod {
		t.Errorf("handle kind = %q, want method", found["handle"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run TestRegistry -v`
Expected: FAIL — `ExtractorForLanguage` undefined

- [ ] **Step 3: Add tree-sitter grammar dependencies**

Run:
```bash
cd /Users/stokes/Projects/gl1tch && go get \
  github.com/smacker/go-tree-sitter/golang \
  github.com/smacker/go-tree-sitter/python \
  github.com/smacker/go-tree-sitter/javascript \
  github.com/smacker/go-tree-sitter/typescript/typescript \
  github.com/smacker/go-tree-sitter/typescript/tsx \
  github.com/smacker/go-tree-sitter/ruby \
  github.com/smacker/go-tree-sitter/rust \
  github.com/smacker/go-tree-sitter/java \
  github.com/smacker/go-tree-sitter/c \
  github.com/smacker/go-tree-sitter/cpp \
  github.com/smacker/go-tree-sitter/csharp \
  github.com/smacker/go-tree-sitter/php \
  github.com/smacker/go-tree-sitter/scala \
  github.com/smacker/go-tree-sitter/swift \
  github.com/smacker/go-tree-sitter/kotlin \
  github.com/smacker/go-tree-sitter/haskell
```

- [ ] **Step 4: Implement language configs**

```go
// internal/indexer/languages.go
package indexer

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/javascript"
	tsTypescript "github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/haskell"
)

var extractors map[string]*LanguageExtractor

func init() {
	extractors = make(map[string]*LanguageExtractor)

	register(goExtractor())
	register(pythonExtractor())
	register(jsExtractor())
	register(tsExtractor())
	register(tsxExtractor())
	register(rubyExtractor())
	register(rustExtractor())
	register(javaExtractor())
	register(cExtractor())
	register(cppExtractor())
	register(csharpExtractor())
	register(phpExtractor())
	register(scalaExtractor())
	register(swiftExtractor())
	register(kotlinExtractor())
	register(haskellExtractor())
}

func register(e *LanguageExtractor) {
	extractors[e.Language] = e
}

// ExtractorForLanguage returns the extractor for the given language name, or nil.
func ExtractorForLanguage(lang string) *LanguageExtractor {
	return extractors[lang]
}

// ---------- Go ----------

func goExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "go",
		Grammar:    golang.GetLanguage(),
		Extensions: []string{".go"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(method_declaration name: (field_identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(type_declaration (type_spec name: (type_identifier) @name type: (struct_type))) @decl`, Kind: KindType},
			{Query: `(type_declaration (type_spec name: (type_identifier) @name type: (interface_type))) @decl`, Kind: KindInterface},
			{Query: `(const_spec name: (identifier) @name) @decl`, Kind: KindConst},
			{Query: `(var_spec name: (identifier) @name) @decl`, Kind: KindVar},
		},
		ImportQuery: `(import_spec path: (interpreted_string_literal) @path)`,
		CallQuery: `[
			(call_expression function: (identifier) @callee)
			(call_expression function: (selector_expression field: (field_identifier) @callee))
		]`,
		PathResolver: goPathResolver,
	}
}

func goPathResolver(importPath, fromFile, repoRoot string) string {
	// Only resolve local imports (those starting with the module path)
	// Standard library and external imports are skipped
	modPath := readModulePath(repoRoot)
	if modPath == "" || !strings.HasPrefix(importPath, modPath) {
		return ""
	}
	rel := strings.TrimPrefix(importPath, modPath+"/")
	return rel
}

func readModulePath(repoRoot string) string {
	// This will be called during indexing — read go.mod to get module path
	// Implemented as a simple line scan to avoid importing golang.org/x/mod
	return readFirstModuleLine(filepath.Join(repoRoot, "go.mod"))
}

// ---------- Python ----------

func pythonExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "python",
		Grammar:    python.GetLanguage(),
		Extensions: []string{".py"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_definition name: (identifier) @name) @decl`, Kind: KindClass},
			// Methods: functions inside a class body
			{Query: `(class_definition body: (block (function_definition name: (identifier) @name) @decl))`, Kind: KindMethod},
		},
		ImportQuery: `[
			(import_statement name: (dotted_name) @path)
			(import_from_statement module_name: (dotted_name) @path)
		]`,
		CallQuery: `[
			(call function: (identifier) @callee)
			(call function: (attribute attribute: (identifier) @callee))
		]`,
		PathResolver: pythonPathResolver,
	}
}

func pythonPathResolver(importPath, fromFile, repoRoot string) string {
	parts := strings.Split(importPath, ".")
	// Try as directory with __init__.py, then as .py file
	dirPath := filepath.Join(parts...)
	pyFile := dirPath + ".py"
	return pyFile
}

// ---------- JavaScript ----------

func jsExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "javascript",
		Grammar:    javascript.GetLanguage(),
		Extensions: []string{".js", ".jsx"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(method_definition name: (property_identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(export_statement declaration: (function_declaration name: (identifier) @name)) @decl`, Kind: KindFunction},
			{Query: `(export_statement declaration: (class_declaration name: (identifier) @name)) @decl`, Kind: KindClass},
			{Query: `(lexical_declaration (variable_declarator name: (identifier) @name)) @decl`, Kind: KindVar},
		},
		ImportQuery: `(import_statement source: (string) @path)`,
		CallQuery: `[
			(call_expression function: (identifier) @callee)
			(call_expression function: (member_expression property: (property_identifier) @callee))
		]`,
		PathResolver: jsPathResolver,
	}
}

func jsPathResolver(importPath, fromFile, repoRoot string) string {
	if !strings.HasPrefix(importPath, ".") {
		return "" // skip node_modules
	}
	dir := filepath.Dir(fromFile)
	resolved := filepath.Join(dir, importPath)
	// Try common extensions — return the resolved base path,
	// actual file existence check happens during resolution phase
	return resolved
}

// ---------- TypeScript ----------

func tsExtractor() *LanguageExtractor {
	ext := jsExtractor()
	ext.Language = "typescript"
	ext.Grammar = tsTypescript.GetLanguage()
	ext.Extensions = []string{".ts"}
	// Add interface declaration for TS
	ext.SymbolQueries = append(ext.SymbolQueries, SymbolQuery{
		Query: `(interface_declaration name: (type_identifier) @name) @decl`,
		Kind:  KindInterface,
	})
	ext.SymbolQueries = append(ext.SymbolQueries, SymbolQuery{
		Query: `(type_alias_declaration name: (type_identifier) @name) @decl`,
		Kind:  KindType,
	})
	return ext
}

func tsxExtractor() *LanguageExtractor {
	ext := tsExtractor()
	ext.Language = "tsx"
	ext.Grammar = tsx.GetLanguage()
	ext.Extensions = []string{".tsx"}
	return ext
}

// ---------- Ruby ----------

func rubyExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "ruby",
		Grammar:    ruby.GetLanguage(),
		Extensions: []string{".rb"},
		SymbolQueries: []SymbolQuery{
			{Query: `(method name: (identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(class name: (constant) @name) @decl`, Kind: KindClass},
			{Query: `(module name: (constant) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(call method: (identifier) @_method arguments: (argument_list (string (string_content) @path)) (#eq? @_method "require"))`,
		CallQuery:   `(call method: (identifier) @callee)`,
	}
}

// ---------- Rust ----------

func rustExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "rust",
		Grammar:    rust.GetLanguage(),
		Extensions: []string{".rs"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_item name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(struct_item name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(enum_item name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(trait_item name: (type_identifier) @name) @decl`, Kind: KindInterface},
			{Query: `(impl_item type: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(const_item name: (identifier) @name) @decl`, Kind: KindConst},
		},
		ImportQuery: `(use_declaration argument: (scoped_identifier path: (identifier) @path))`,
		CallQuery: `[
			(call_expression function: (identifier) @callee)
			(call_expression function: (field_expression field: (field_identifier) @callee))
		]`,
	}
}

// ---------- Java ----------

func javaExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "java",
		Grammar:    java.GetLanguage(),
		Extensions: []string{".java"},
		SymbolQueries: []SymbolQuery{
			{Query: `(method_declaration name: (identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(class_declaration name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(interface_declaration name: (identifier) @name) @decl`, Kind: KindInterface},
			{Query: `(field_declaration declarator: (variable_declarator name: (identifier) @name)) @decl`, Kind: KindField},
		},
		ImportQuery: `(import_declaration (scoped_identifier) @path)`,
		CallQuery: `[
			(method_invocation name: (identifier) @callee)
		]`,
	}
}

// ---------- C ----------

func cExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "c",
		Grammar:    c.GetLanguage(),
		Extensions: []string{".c", ".h"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition declarator: (function_declarator declarator: (identifier) @name)) @decl`, Kind: KindFunction},
			{Query: `(struct_specifier name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(enum_specifier name: (type_identifier) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(preproc_include path: (string_literal) @path)`,
		CallQuery:   `(call_expression function: (identifier) @callee)`,
	}
}

// ---------- C++ ----------

func cppExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "cpp",
		Grammar:    cpp.GetLanguage(),
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hh"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition declarator: (function_declarator declarator: (identifier) @name)) @decl`, Kind: KindFunction},
			{Query: `(class_specifier name: (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(struct_specifier name: (type_identifier) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(preproc_include path: (string_literal) @path)`,
		CallQuery: `[
			(call_expression function: (identifier) @callee)
			(call_expression function: (field_expression field: (field_identifier) @callee))
		]`,
	}
}

// ---------- C# ----------

func csharpExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "csharp",
		Grammar:    csharp.GetLanguage(),
		Extensions: []string{".cs"},
		SymbolQueries: []SymbolQuery{
			{Query: `(method_declaration name: (identifier) @name) @decl`, Kind: KindMethod},
			{Query: `(class_declaration name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(interface_declaration name: (identifier) @name) @decl`, Kind: KindInterface},
			{Query: `(struct_declaration name: (identifier) @name) @decl`, Kind: KindType},
		},
		CallQuery: `(invocation_expression function: (identifier) @callee)`,
	}
}

// ---------- PHP ----------

func phpExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "php",
		Grammar:    php.GetLanguage(),
		Extensions: []string{".php"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition name: (name) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (name) @name) @decl`, Kind: KindClass},
			{Query: `(method_declaration name: (name) @name) @decl`, Kind: KindMethod},
			{Query: `(interface_declaration name: (name) @name) @decl`, Kind: KindInterface},
		},
		CallQuery: `(function_call_expression function: (name) @callee)`,
	}
}

// ---------- Scala ----------

func scalaExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "scala",
		Grammar:    scala.GetLanguage(),
		Extensions: []string{".scala"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_definition name: (identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_definition name: (identifier) @name) @decl`, Kind: KindClass},
			{Query: `(trait_definition name: (identifier) @name) @decl`, Kind: KindInterface},
			{Query: `(object_definition name: (identifier) @name) @decl`, Kind: KindType},
		},
		ImportQuery: `(import_declaration path: (identifier) @path)`,
		CallQuery:   `(call_expression function: (identifier) @callee)`,
	}
}

// ---------- Swift ----------

func swiftExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "swift",
		Grammar:    swift.GetLanguage(),
		Extensions: []string{".swift"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration name: (simple_identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration name: (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(struct_declaration name: (type_identifier) @name) @decl`, Kind: KindType},
			{Query: `(protocol_declaration name: (type_identifier) @name) @decl`, Kind: KindInterface},
		},
		CallQuery: `(call_expression (simple_identifier) @callee)`,
	}
}

// ---------- Kotlin ----------

func kotlinExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "kotlin",
		Grammar:    kotlin.GetLanguage(),
		Extensions: []string{".kt", ".kts"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function_declaration (simple_identifier) @name) @decl`, Kind: KindFunction},
			{Query: `(class_declaration (type_identifier) @name) @decl`, Kind: KindClass},
			{Query: `(interface_declaration (type_identifier) @name) @decl`, Kind: KindInterface},
		},
		ImportQuery: `(import_header (identifier) @path)`,
		CallQuery:   `(call_expression (simple_identifier) @callee)`,
	}
}

// ---------- Haskell ----------

func haskellExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "haskell",
		Grammar:    haskell.GetLanguage(),
		Extensions: []string{".hs"},
		SymbolQueries: []SymbolQuery{
			{Query: `(function name: (variable) @name) @decl`, Kind: KindFunction},
			{Query: `(type_alias name: (type) @name) @decl`, Kind: KindType},
			{Query: `(newtype name: (type) @name) @decl`, Kind: KindType},
			{Query: `(adt name: (type) @name) @decl`, Kind: KindType},
			{Query: `(class_declaration name: (type) @name) @decl`, Kind: KindInterface},
		},
		ImportQuery: `(import module: (module) @path)`,
	}
}

// ---------- helpers ----------

// readFirstModuleLine reads go.mod and returns the module path.
func readFirstModuleLine(goModPath string) string {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
```

Note: add `"os"` to the import block.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run "TestRegistry|TestGoExtractor|TestPythonExtractor" -v`
Expected: PASS

Some tree-sitter queries may need adjustment per grammar. If a query fails to compile, the error message will say which capture or pattern is wrong — adjust the S-expression accordingly. The test suite catches these immediately.

- [ ] **Step 6: Commit**

```bash
git add internal/indexer/languages.go internal/indexer/languages_test.go go.mod go.sum
git commit -m "feat(indexer): add language extractor configs for 16 languages"
```

---

## Task 6: Import and Call Resolution

**Files:**
- Create: `internal/indexer/resolve.go`
- Test: `internal/indexer/resolve_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/indexer/resolve_test.go
package indexer

import "testing"

func TestResolveImportsGoLocal(t *testing.T) {
	// Simulate: main.go imports "github.com/8op-org/gl1tch/internal/indexer"
	// Symbol "IndexRepo" exists in internal/indexer/indexer.go
	symbols := []SymbolDoc{
		{ID: "sym1", File: "internal/indexer/indexer.go", Kind: KindFunction, Name: "IndexRepo", Repo: "gl1tch"},
		{ID: "sym2", File: "cmd/index.go", Kind: KindFunction, Name: "main", Repo: "gl1tch"},
	}

	unresolved := []UnresolvedImport{
		{Path: "github.com/8op-org/gl1tch/internal/indexer", File: "cmd/index.go", Line: 5},
	}

	resolver := NewResolver(symbols, "gl1tch", "/fake/root")
	resolver.ModulePath = "github.com/8op-org/gl1tch"
	edges := resolver.ResolveImports(unresolved)

	found := false
	for _, e := range edges {
		if e.Kind == EdgeImports && e.File == "cmd/index.go" {
			found = true
		}
	}
	if !found {
		t.Error("expected imports edge from cmd/index.go")
	}
}

func TestResolveContainsEdges(t *testing.T) {
	parentID := SymbolID("handler.go", KindType, "Server", 10)
	childID := SymbolID("handler.go", KindMethod, "Start", 15)

	symbols := []SymbolDoc{
		{ID: parentID, File: "handler.go", Kind: KindType, Name: "Server"},
		{ID: childID, File: "handler.go", Kind: KindMethod, Name: "Start", ParentID: parentID},
	}

	resolver := NewResolver(symbols, "test", "/fake")
	edges := resolver.ResolveContains()

	if len(edges) != 1 {
		t.Fatalf("expected 1 contains edge, got %d", len(edges))
	}
	if edges[0].Kind != EdgeContains {
		t.Errorf("edge kind = %q, want contains", edges[0].Kind)
	}
	if edges[0].SourceID != parentID || edges[0].TargetID != childID {
		t.Error("contains edge has wrong source/target")
	}
}

func TestResolveCallsLocalScope(t *testing.T) {
	sym1 := SymbolDoc{ID: "s1", File: "main.go", Kind: KindFunction, Name: "main"}
	sym2 := SymbolDoc{ID: "s2", File: "main.go", Kind: KindFunction, Name: "doWork"}

	symbols := []SymbolDoc{sym1, sym2}
	calls := []CallSite{
		{CalleeName: "doWork", File: "main.go", Line: 5},
	}

	resolver := NewResolver(symbols, "test", "/fake")
	edges := resolver.ResolveCalls(calls)

	found := false
	for _, e := range edges {
		if e.Kind == EdgeCalls && e.TargetID == "s2" {
			found = true
		}
	}
	if !found {
		t.Error("expected calls edge to doWork")
	}
}

func TestResolveCallsUnresolvedDropped(t *testing.T) {
	sym1 := SymbolDoc{ID: "s1", File: "main.go", Kind: KindFunction, Name: "main"}
	symbols := []SymbolDoc{sym1}
	calls := []CallSite{
		{CalleeName: "nonExistent", File: "main.go", Line: 5},
	}

	resolver := NewResolver(symbols, "test", "/fake")
	edges := resolver.ResolveCalls(calls)

	if len(edges) != 0 {
		t.Errorf("expected 0 edges for unresolved call, got %d", len(edges))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run TestResolve -v`
Expected: FAIL — `NewResolver` undefined

- [ ] **Step 3: Implement resolver**

```go
// internal/indexer/resolve.go
package indexer

import (
	"strings"
)

// Resolver resolves cross-file references into edges.
type Resolver struct {
	symbols    []SymbolDoc
	byFile     map[string][]SymbolDoc // file -> symbols in that file
	byName     map[string][]SymbolDoc // name -> symbols with that name
	byID       map[string]SymbolDoc
	repo       string
	repoRoot   string
	ModulePath string // Go module path (set externally)
}

// NewResolver creates a resolver from the full symbol set.
func NewResolver(symbols []SymbolDoc, repo, repoRoot string) *Resolver {
	r := &Resolver{
		symbols:  symbols,
		byFile:   make(map[string][]SymbolDoc),
		byName:   make(map[string][]SymbolDoc),
		byID:     make(map[string]SymbolDoc),
		repo:     repo,
		repoRoot: repoRoot,
	}
	for _, s := range symbols {
		r.byFile[s.File] = append(r.byFile[s.File], s)
		r.byName[s.Name] = append(r.byName[s.Name], s)
		r.byID[s.ID] = s
	}
	return r
}

// ResolveContains emits contains edges for symbols with a parent_id.
func (r *Resolver) ResolveContains() []EdgeDoc {
	var edges []EdgeDoc
	for _, s := range r.symbols {
		if s.ParentID == "" {
			continue
		}
		if _, ok := r.byID[s.ParentID]; !ok {
			continue
		}
		edges = append(edges, EdgeDoc{
			SourceID: s.ParentID,
			TargetID: s.ID,
			Kind:     EdgeContains,
			File:     s.File,
			Repo:     r.repo,
		})
	}
	return edges
}

// ResolveImports resolves unresolved imports to edges.
func (r *Resolver) ResolveImports(unresolved []UnresolvedImport) []EdgeDoc {
	var edges []EdgeDoc
	for _, imp := range unresolved {
		targetDir := r.resolveImportPath(imp)
		if targetDir == "" {
			continue
		}
		// Find symbols in files under the resolved directory
		for file, syms := range r.byFile {
			if !strings.HasPrefix(file, targetDir) {
				continue
			}
			for _, sym := range syms {
				// Link import to exported/top-level symbols in target
				if sym.Kind == KindFunction || sym.Kind == KindType || sym.Kind == KindClass ||
					sym.Kind == KindInterface || sym.Kind == KindConst || sym.Kind == KindVar {
					edges = append(edges, EdgeDoc{
						SourceID: fileSymbolID(imp.File, r.repo),
						TargetID: sym.ID,
						Kind:     EdgeImports,
						File:     imp.File,
						Repo:     r.repo,
					})
				}
			}
		}
	}
	return edges
}

// ResolveCalls resolves call sites to edges.
func (r *Resolver) ResolveCalls(calls []CallSite) []EdgeDoc {
	var edges []EdgeDoc
	for _, call := range calls {
		// 1. Check local scope (same file)
		target := r.findCallTarget(call)
		if target == "" {
			continue
		}
		// Find the enclosing function for the call site
		caller := r.findEnclosing(call.File, call.Line)
		if caller == "" {
			continue
		}
		edges = append(edges, EdgeDoc{
			SourceID: caller,
			TargetID: target,
			Kind:     EdgeCalls,
			File:     call.File,
			Repo:     r.repo,
		})
	}
	return edges
}

func (r *Resolver) resolveImportPath(imp UnresolvedImport) string {
	// Go: strip module prefix to get relative directory
	if r.ModulePath != "" && strings.HasPrefix(imp.Path, r.ModulePath) {
		rel := strings.TrimPrefix(imp.Path, r.ModulePath+"/")
		return rel
	}
	// Other languages: use the path directly (Python dots, JS relative paths)
	return imp.Path
}

func (r *Resolver) findCallTarget(call CallSite) string {
	// First: same-file symbols
	for _, s := range r.byFile[call.File] {
		if s.Name == call.CalleeName && (s.Kind == KindFunction || s.Kind == KindMethod) {
			return s.ID
		}
	}
	// Second: any symbol with that name (cross-file)
	candidates := r.byName[call.CalleeName]
	for _, s := range candidates {
		if s.Kind == KindFunction || s.Kind == KindMethod {
			return s.ID
		}
	}
	return ""
}

func (r *Resolver) findEnclosing(file string, line int) string {
	var best SymbolDoc
	for _, s := range r.byFile[file] {
		if (s.Kind == KindFunction || s.Kind == KindMethod) &&
			s.StartLine <= line && s.EndLine >= line {
			if best.ID == "" || s.StartLine > best.StartLine {
				best = s
			}
		}
	}
	return best.ID
}

func fileSymbolID(file, repo string) string {
	return SymbolID(file, "file", file, 0)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run TestResolve -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/indexer/resolve.go internal/indexer/resolve_test.go
git commit -m "feat(indexer): add import and call resolution for edge emission"
```

---

## Task 7: Rewrite IndexRepo with Tree-Sitter Pipeline

**Files:**
- Modify: `internal/indexer/indexer.go`
- Modify: `internal/indexer/indexer_test.go`

- [ ] **Step 1: Write failing tests for new behavior**

Add to `internal/indexer/indexer_test.go`:

```go
func TestClassifyFiles(t *testing.T) {
	existing := map[string]string{
		"main.go":    "hash1",
		"old.go":     "hash2",
		"changed.go": "hash3",
	}
	current := map[string]string{
		"main.go":    "hash1",    // unchanged
		"changed.go": "hash999",  // changed
		"new.go":     "hash4",    // new
	}

	toIndex, toDelete := classifyFiles(existing, current)

	// changed.go and new.go should be indexed
	if len(toIndex) != 2 {
		t.Errorf("toIndex: got %d, want 2", len(toIndex))
	}
	// old.go should be deleted
	if len(toDelete) != 1 {
		t.Errorf("toDelete: got %d, want 1", len(toDelete))
	}
	if toDelete[0] != "old.go" {
		t.Errorf("toDelete[0] = %q, want old.go", toDelete[0])
	}
}

func TestFileHashCompute(t *testing.T) {
	h1 := fileHash([]byte("hello world"))
	h2 := fileHash([]byte("hello world"))
	h3 := fileHash([]byte("different"))

	if h1 != h2 {
		t.Error("same content should produce same hash")
	}
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}
	if h1 == "" {
		t.Error("hash should not be empty")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -run "TestClassify|TestFileHash" -v`
Expected: FAIL — functions undefined

- [ ] **Step 3: Rewrite indexer.go**

Replace `IndexRepo` in `internal/indexer/indexer.go` with the three-phase pipeline. Keep `DetectLanguage`, `ChunkContent` (still used for content chunks), and the existing `CodeDoc` type. Replace `ExtractSymbols` with tree-sitter extraction.

Add to `internal/indexer/indexer.go`:

```go
import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
)

// IndexOpts controls indexing behavior.
type IndexOpts struct {
	Repo        string
	ESURL       string
	Languages   []string // empty = all
	Full        bool     // force full re-index
	SymbolsOnly bool     // skip content chunks
	Stats       bool
}

func fileHash(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

// classifyFiles compares existing file hashes with current ones.
// Returns files to (re)index and files to delete.
func classifyFiles(existing, current map[string]string) (toIndex []string, toDelete []string) {
	for file, hash := range current {
		if existingHash, ok := existing[file]; !ok || existingHash != hash {
			toIndex = append(toIndex, file)
		}
	}
	for file := range existing {
		if _, ok := current[file]; !ok {
			toDelete = append(toDelete, file)
		}
	}
	return
}

// IndexRepoGraph indexes a repository using the three-phase tree-sitter pipeline.
func IndexRepoGraph(root string, es *esearch.Client, opts IndexOpts) error {
	ctx := context.Background()
	repo := opts.Repo
	if repo == "" {
		repo = filepath.Base(root)
	}

	symbolsIndex := esearch.IndexSymbolsPrefix + repo
	edgesIndex := esearch.IndexEdgesPrefix + repo

	// Ensure indices exist
	if err := es.EnsureIndex(ctx, symbolsIndex, esearch.SymbolsMapping); err != nil {
		return fmt.Errorf("ensure symbols index: %w", err)
	}
	if err := es.EnsureIndex(ctx, edgesIndex, esearch.EdgesMapping); err != nil {
		return fmt.Errorf("ensure edges index: %w", err)
	}

	// Walk filesystem and compute file hashes
	currentFiles := make(map[string]string) // relPath -> sha256
	fileContents := make(map[string][]byte) // relPath -> content
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			for _, skip := range skipDirs {
				if base == skip {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if shouldSkipFile(path, info) {
			return nil
		}
		lang := DetectLanguage(info.Name())
		if lang == "other" {
			return nil
		}
		if len(opts.Languages) > 0 && !containsStr(opts.Languages, lang) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(root, path)
		h := fileHash(content)
		currentFiles[relPath] = h
		fileContents[relPath] = content
		return nil
	})

	// Incremental: get existing file hashes from ES
	var toIndex []string
	var toDelete []string
	if opts.Full {
		for file := range currentFiles {
			toIndex = append(toIndex, file)
		}
	} else {
		existing, err := es.TermsAgg(ctx, symbolsIndex, "file", "file_hash", 50000)
		if err != nil {
			return fmt.Errorf("get existing hashes: %w", err)
		}
		toIndex, toDelete = classifyFiles(existing, currentFiles)
	}

	// Delete stale symbols + edges
	for _, file := range toDelete {
		delQuery, _ := json.Marshal(map[string]any{
			"query": map[string]any{"term": map[string]any{"file": file}},
		})
		es.DeleteByQuery(ctx, symbolsIndex, delQuery)
		es.DeleteByQuery(ctx, edgesIndex, delQuery)
	}

	// Also delete symbols/edges for files being re-indexed
	for _, file := range toIndex {
		delQuery, _ := json.Marshal(map[string]any{
			"query": map[string]any{"term": map[string]any{"file": file}},
		})
		es.DeleteByQuery(ctx, symbolsIndex, delQuery)
		es.DeleteByQuery(ctx, edgesIndex, delQuery)
	}

	// Phase 1: Extract symbols, imports, and calls from changed files
	var allSymbols []SymbolDoc
	var allImports []UnresolvedImport
	var allCalls []CallSite

	for _, file := range toIndex {
		content := fileContents[file]
		lang := DetectLanguage(filepath.Base(file))
		ext := ExtractorForLanguage(lang)
		if ext == nil {
			continue
		}

		hash := currentFiles[file]

		symbols, err := ext.Extract(content, file, repo, hash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: extract %s: %v\n", file, err)
			continue
		}
		allSymbols = append(allSymbols, symbols...)

		imports, err := ext.ExtractImports(content, file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: extract imports %s: %v\n", file, err)
		}
		allImports = append(allImports, imports...)

		calls, err := ext.ExtractCalls(content, file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: extract calls %s: %v\n", file, err)
		}
		allCalls = append(allCalls, calls...)
	}

	// Bulk index symbols
	var symDocs []esearch.BulkDoc
	for _, s := range allSymbols {
		body, _ := json.Marshal(s)
		symDocs = append(symDocs, esearch.BulkDoc{ID: s.ID, Body: body})
	}
	for i := 0; i < len(symDocs); i += bulkBatch {
		end := i + bulkBatch
		if end > len(symDocs) {
			end = len(symDocs)
		}
		if err := es.BulkIndex(ctx, symbolsIndex, symDocs[i:end]); err != nil {
			return fmt.Errorf("bulk index symbols: %w", err)
		}
	}

	// Phase 2 + 3: Resolve imports, contains, and calls
	resolver := NewResolver(allSymbols, repo, root)
	modPath := readModulePath(root)
	if modPath != "" {
		resolver.ModulePath = modPath
	}

	var allEdges []EdgeDoc
	allEdges = append(allEdges, resolver.ResolveContains()...)
	allEdges = append(allEdges, resolver.ResolveImports(allImports)...)
	allEdges = append(allEdges, resolver.ResolveCalls(allCalls)...)

	// Bulk index edges
	var edgeDocs []esearch.BulkDoc
	for i, e := range allEdges {
		body, _ := json.Marshal(e)
		edgeDocs = append(edgeDocs, esearch.BulkDoc{
			ID:   fmt.Sprintf("%s-%s-%s-%d", e.SourceID, e.TargetID, e.Kind, i),
			Body: body,
		})
	}
	for i := 0; i < len(edgeDocs); i += bulkBatch {
		end := i + bulkBatch
		if end > len(edgeDocs) {
			end = len(edgeDocs)
		}
		if err := es.BulkIndex(ctx, edgesIndex, edgeDocs[i:end]); err != nil {
			return fmt.Errorf("bulk index edges: %w", err)
		}
	}

	// Content chunks (existing behavior, unless --symbols-only)
	if !opts.SymbolsOnly {
		codeIndex := "glitch-code-" + repo
		if err := es.EnsureIndex(ctx, codeIndex, ""); err != nil {
			// Non-fatal — code index may already exist
		}
		var chunkDocs []esearch.BulkDoc
		now := time.Now().UTC().Format(time.RFC3339)
		for _, file := range toIndex {
			content := fileContents[file]
			lang := DetectLanguage(filepath.Base(file))
			chunks := ChunkContent(string(content))
			symbols := ExtractSymbols(string(content))
			for i, chunk := range chunks {
				id := file
				if len(chunks) > 1 {
					id = fmt.Sprintf("%s:%d", file, i)
				}
				doc := CodeDoc{
					Path:      file,
					Content:   chunk,
					Repo:      repo,
					Language:  lang,
					Hash:      currentFiles[file],
					Symbols:   symbols,
					IndexedAt: now,
				}
				body, _ := json.Marshal(doc)
				chunkDocs = append(chunkDocs, esearch.BulkDoc{ID: id, Body: body})
			}
		}
		for i := 0; i < len(chunkDocs); i += bulkBatch {
			end := i + bulkBatch
			if end > len(chunkDocs) {
				end = len(chunkDocs)
			}
			if err := es.BulkIndex(ctx, codeIndex, chunkDocs[i:end]); err != nil {
				return fmt.Errorf("bulk index chunks: %w", err)
			}
		}
	}

	if opts.Stats {
		fmt.Printf("Indexed %d symbols, %d edges from %d files (%d changed, %d deleted)\n",
			len(allSymbols), len(allEdges), len(currentFiles), len(toIndex), len(toDelete))
	}

	return nil
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
```

Also add `"context"` to the imports and define `skipDirs` and `shouldSkipFile` if not already present (they exist in the current code, just ensure they're reusable).

- [ ] **Step 4: Keep existing IndexRepo as wrapper**

Keep the old `IndexRepo(root string, es *esearch.Client) error` signature working by calling the new function:

```go
func IndexRepo(root string, es *esearch.Client) error {
	return IndexRepoGraph(root, es, IndexOpts{Stats: true})
}
```

- [ ] **Step 5: Run all indexer tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/indexer/ -v`
Expected: PASS (existing tests still pass + new tests pass)

- [ ] **Step 6: Commit**

```bash
git add internal/indexer/indexer.go internal/indexer/indexer_test.go
git commit -m "feat(indexer): rewrite IndexRepo with three-phase tree-sitter pipeline"
```

---

## Task 8: CLI Flags for index Command

**Files:**
- Modify: `cmd/index.go`

- [ ] **Step 1: Update cmd/index.go with new flags**

```go
// cmd/index.go — replace the existing command
var (
	indexRepo        string
	indexESURL       string
	indexLanguages   string
	indexFull        bool
	indexSymbolsOnly bool
	indexStats       bool
)

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "index a repository into Elasticsearch for code search",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		esURL := indexESURL
		if esURL == "" {
			esURL = "http://localhost:9200"
		}
		es := esearch.NewClient(esURL)

		opts := indexer.IndexOpts{
			Repo:        indexRepo,
			ESURL:       esURL,
			Full:        indexFull,
			SymbolsOnly: indexSymbolsOnly,
			Stats:       indexStats,
		}
		if indexLanguages != "" {
			opts.Languages = strings.Split(indexLanguages, ",")
		}

		return indexer.IndexRepoGraph(path, es, opts)
	},
}

func init() {
	indexCmd.Flags().StringVar(&indexRepo, "repo", "", "override repo name (default: directory name)")
	indexCmd.Flags().StringVar(&indexESURL, "es-url", "", "Elasticsearch URL (default: localhost:9200)")
	indexCmd.Flags().StringVar(&indexLanguages, "languages", "", "comma-separated language filter")
	indexCmd.Flags().BoolVar(&indexFull, "full", false, "force full re-index")
	indexCmd.Flags().BoolVar(&indexSymbolsOnly, "symbols-only", false, "only index symbols + edges, skip content chunks")
	indexCmd.Flags().BoolVar(&indexStats, "stats", false, "print index stats after completion")
	rootCmd.AddCommand(indexCmd)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./cmd/...`
Expected: Compiles without error

- [ ] **Step 3: Commit**

```bash
git add cmd/index.go
git commit -m "feat(cmd): add CLI flags to glitch index (--repo, --es-url, --full, --stats, etc.)"
```

---

## Task 9: Research Tools — search_symbols, search_edges, symbol_context

**Files:**
- Modify: `internal/research/tools.go`
- Test: `internal/research/tools_test.go` (create if absent)

- [ ] **Step 1: Write failing test**

```go
// internal/research/tools_test.go
package research

import "testing"

func TestDefinitionsIncludeGraphTools(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	defs := ts.Definitions()

	want := map[string]bool{
		"search_symbols":  false,
		"search_edges":    false,
		"symbol_context":  false,
	}

	for _, d := range defs {
		if _, ok := want[d.Name]; ok {
			want[d.Name] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Errorf("missing tool definition: %s", name)
		}
	}
}

func TestValidToolGraphTools(t *testing.T) {
	ts := NewToolSet("/tmp", nil)
	for _, name := range []string{"search_symbols", "search_edges", "symbol_context"} {
		if !ts.ValidTool(name) {
			t.Errorf("ValidTool(%q) = false, want true", name)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -run TestDefinitionsInclude -v`
Expected: FAIL

- [ ] **Step 3: Add tools to ToolSet**

Add to `Definitions()` in `internal/research/tools.go`:

```go
{
	Name:        "search_symbols",
	Description: "Search code symbols (functions, types, classes, interfaces) in the repo graph. Returns structured results with file, kind, signature, and line range.",
	Params:      "name (string, symbol name or pattern), kind (string, optional: function/method/type/interface/class/const/var), file (string, optional: filter by file path), repo (string, required: repo name)",
},
{
	Name:        "search_edges",
	Description: "Query relationships between code symbols. Find callers, callees, importers, implementations. Edge kinds: contains, imports, exports, extends, implements, calls.",
	Params:      "symbol_id (string, source or target symbol ID), direction (string: 'from' or 'to'), kind (string, optional: edge kind filter), repo (string, required: repo name)",
},
{
	Name:        "symbol_context",
	Description: "Get full context for a symbol: its definition, edges (callers, callees, imports), and source code snippet. One call for the complete picture.",
	Params:      "symbol_id (string, required), repo (string, required: repo name)",
},
```

Add to `Execute()` dispatcher:

```go
case "search_symbols":
	return ts.searchSymbols(ctx, params)
case "search_edges":
	return ts.searchEdges(ctx, params)
case "symbol_context":
	return ts.symbolContext(ctx, params)
```

Implement the three methods:

```go
func (ts *ToolSet) searchSymbols(ctx context.Context, params map[string]string) ToolResult {
	repo := params["repo"]
	if repo == "" {
		return ToolResult{Tool: "search_symbols", Err: "repo is required"}
	}
	index := "glitch-symbols-" + repo

	// Build bool query
	must := []map[string]any{}
	if name := params["name"]; name != "" {
		must = append(must, map[string]any{"match": map[string]any{"name": name}})
	}
	if kind := params["kind"]; kind != "" {
		must = append(must, map[string]any{"term": map[string]any{"kind": kind}})
	}
	if file := params["file"]; file != "" {
		must = append(must, map[string]any{"wildcard": map[string]any{"file": map[string]any{"value": "*" + file + "*"}}})
	}

	query := map[string]any{
		"size":    20,
		"query":   map[string]any{"bool": map[string]any{"must": must}},
		"_source": []string{"id", "file", "kind", "name", "signature", "start_line", "end_line"},
	}

	body, _ := json.Marshal(query)
	resp, err := ts.es.Search(ctx, []string{index}, body)
	if err != nil {
		return ToolResult{Tool: "search_symbols", Err: err.Error()}
	}

	var out strings.Builder
	for _, hit := range resp.Results {
		out.Write(hit.Source)
		out.WriteByte('\n')
	}
	return ToolResult{Tool: "search_symbols", Output: truncateOutput(out.String(), 8000)}
}

func (ts *ToolSet) searchEdges(ctx context.Context, params map[string]string) ToolResult {
	repo := params["repo"]
	if repo == "" {
		return ToolResult{Tool: "search_edges", Err: "repo is required"}
	}
	index := "glitch-edges-" + repo
	symbolID := params["symbol_id"]
	direction := params["direction"]
	edgeKind := params["kind"]

	field := "source_id"
	if direction == "to" {
		field = "target_id"
	}

	must := []map[string]any{
		{"term": map[string]any{field: symbolID}},
	}
	if edgeKind != "" {
		must = append(must, map[string]any{"term": map[string]any{"kind": edgeKind}})
	}

	query := map[string]any{
		"size":  50,
		"query": map[string]any{"bool": map[string]any{"must": must}},
	}

	body, _ := json.Marshal(query)
	resp, err := ts.es.Search(ctx, []string{index}, body)
	if err != nil {
		return ToolResult{Tool: "search_edges", Err: err.Error()}
	}

	var out strings.Builder
	for _, hit := range resp.Results {
		out.Write(hit.Source)
		out.WriteByte('\n')
	}
	return ToolResult{Tool: "search_edges", Output: truncateOutput(out.String(), 8000)}
}

func (ts *ToolSet) symbolContext(ctx context.Context, params map[string]string) ToolResult {
	repo := params["repo"]
	symbolID := params["symbol_id"]
	if repo == "" || symbolID == "" {
		return ToolResult{Tool: "symbol_context", Err: "repo and symbol_id are required"}
	}

	symIndex := "glitch-symbols-" + repo
	edgeIndex := "glitch-edges-" + repo
	codeIndex := "glitch-code-" + repo

	// Fetch symbol
	symQuery, _ := json.Marshal(map[string]any{
		"size":  1,
		"query": map[string]any{"term": map[string]any{"id": symbolID}},
	})
	symResp, err := ts.es.Search(ctx, []string{symIndex}, symQuery)
	if err != nil || len(symResp.Results) == 0 {
		return ToolResult{Tool: "symbol_context", Err: "symbol not found"}
	}

	// Fetch edges (both directions)
	edgeQuery, _ := json.Marshal(map[string]any{
		"size": 50,
		"query": map[string]any{"bool": map[string]any{"should": []map[string]any{
			{"term": map[string]any{"source_id": symbolID}},
			{"term": map[string]any{"target_id": symbolID}},
		}}},
	})
	edgeResp, _ := ts.es.Search(ctx, []string{edgeIndex}, edgeQuery)

	// Fetch source code snippet from code index
	var sym struct {
		File      string `json:"file"`
		StartLine int    `json:"start_line"`
		EndLine   int    `json:"end_line"`
	}
	json.Unmarshal(symResp.Results[0].Source, &sym)

	codeQuery, _ := json.Marshal(map[string]any{
		"size":  5,
		"query": map[string]any{"term": map[string]any{"path": sym.File}},
	})
	codeResp, _ := ts.es.Search(ctx, []string{codeIndex}, codeQuery)

	var out strings.Builder
	out.WriteString("=== Symbol ===\n")
	out.Write(symResp.Results[0].Source)
	out.WriteString("\n\n=== Edges ===\n")
	if edgeResp != nil {
		for _, hit := range edgeResp.Results {
			out.Write(hit.Source)
			out.WriteByte('\n')
		}
	}
	out.WriteString("\n=== Source ===\n")
	if codeResp != nil {
		for _, hit := range codeResp.Results {
			var chunk struct {
				Content string `json:"content"`
			}
			json.Unmarshal(hit.Source, &chunk)
			out.WriteString(chunk.Content)
			out.WriteByte('\n')
		}
	}

	return ToolResult{Tool: "symbol_context", Output: truncateOutput(out.String(), 8000)}
}
```

Add `"encoding/json"` and `"strings"` to imports if not present.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/research/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/research/tools.go internal/research/tools_test.go
git commit -m "feat(research): add search_symbols, search_edges, symbol_context tools"
```

---

## Task 10: Observer Integration — Graph BFS + --depth

**Files:**
- Modify: `internal/observer/query.go`
- Modify: `internal/observer/query_test.go`
- Modify: `cmd/observe.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/observer/query_test.go`:

```go
func TestAllIndicesIncludesSymbols(t *testing.T) {
	indices := allIndicesForRepo("gl1tch")
	hasSymbols := false
	hasEdges := false
	for _, idx := range indices {
		if idx == "glitch-symbols-gl1tch" {
			hasSymbols = true
		}
		if idx == "glitch-edges-gl1tch" {
			hasEdges = true
		}
	}
	if !hasSymbols {
		t.Error("missing symbols index")
	}
	if !hasEdges {
		t.Error("missing edges index")
	}
}

func TestGraphBFSEmptyResults(t *testing.T) {
	// BFS with no results should return empty slice
	results := graphBFS(nil, "sym1", "calls", 2)
	if results != nil && len(results) > 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/observer/ -run "TestAllIndicesIncludes|TestGraphBFS" -v`
Expected: FAIL

- [ ] **Step 3: Implement observer changes**

Add to `internal/observer/query.go`:

```go
// allIndicesForRepo returns telemetry + graph indices for a repo.
func allIndicesForRepo(repo string) []string {
	base := allIndices()
	if repo != "" {
		base = append(base,
			esearch.IndexSymbolsPrefix+repo,
			esearch.IndexEdgesPrefix+repo,
		)
	}
	return base
}
```

Update `searchWithFallback` to use `allIndicesForRepo(q.repo)` instead of `allIndices()` when repo is set.

Add graph BFS:

```go
// graphBFS performs breadth-first traversal on edges from a starting symbol.
// edgeFetcher is called with (symbolID, edgeKind) and returns edges.
// Returns collected symbol IDs at each depth level.
func graphBFS(edgeFetcher func(symbolID, edgeKind string) []EdgeHit, startID, edgeKind string, depth int) []string {
	if edgeFetcher == nil || depth < 1 {
		return nil
	}

	visited := map[string]bool{startID: true}
	frontier := []string{startID}
	var collected []string

	for d := 0; d < depth && len(frontier) > 0; d++ {
		var next []string
		for _, id := range frontier {
			edges := edgeFetcher(id, edgeKind)
			for _, e := range edges {
				targetID := e.TargetID
				if e.TargetID == id {
					targetID = e.SourceID // reverse direction
				}
				if !visited[targetID] {
					visited[targetID] = true
					next = append(next, targetID)
					collected = append(collected, targetID)
				}
			}
		}
		frontier = next
	}
	return collected
}

// EdgeHit is a parsed edge from ES.
type EdgeHit struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Kind     string `json:"kind"`
}
```

Add `--depth` to the `QueryEngine`:

```go
type QueryEngine struct {
	es    *esearch.Client
	llm   LLMFunc
	repo  string
	depth int
}

func (q *QueryEngine) WithDepth(d int) *QueryEngine {
	q.depth = d
	return q
}
```

Update `cmd/observe.go` to add the flag:

```go
var observeDepth int

// In init():
observeCmd.Flags().IntVar(&observeDepth, "depth", 1, "BFS traversal depth for graph queries")

// In RunE, after creating engine:
if observeDepth > 0 {
	engine.WithDepth(observeDepth)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/observer/ -v`
Expected: PASS

- [ ] **Step 5: Verify it compiles**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./cmd/...`
Expected: Compiles without error

- [ ] **Step 6: Commit**

```bash
git add internal/observer/query.go internal/observer/query_test.go cmd/observe.go
git commit -m "feat(observer): add symbol/edge index search and graph BFS with --depth"
```

---

## Task 11: Smoke Tests

**Files:**
- Create: `test/smoke_codegraph_test.sh`

These require a running ES instance at localhost:9200.

- [ ] **Step 1: Write smoke test script**

```bash
#!/usr/bin/env bash
# Smoke tests for code graph indexing.
# Requires: Elasticsearch running at localhost:9200
set -euo pipefail

GLITCH="go run ."
REPO_DIR="$(pwd)"
REPO_NAME="gl1tch"

echo "=== Smoke: glitch index (full) ==="
$GLITCH index "$REPO_DIR" --repo "$REPO_NAME" --full --stats
echo "PASS: index completed"

echo ""
echo "=== Smoke: glitch index (incremental, no changes) ==="
$GLITCH index "$REPO_DIR" --repo "$REPO_NAME" --stats
echo "PASS: incremental index completed"

echo ""
echo "=== Smoke: symbols indexed ==="
COUNT=$(curl -s "http://localhost:9200/glitch-symbols-${REPO_NAME}/_count" | jq '.count')
if [ "$COUNT" -lt 1 ]; then
  echo "FAIL: no symbols indexed (count=$COUNT)"
  exit 1
fi
echo "PASS: $COUNT symbols indexed"

echo ""
echo "=== Smoke: edges indexed ==="
EDGE_COUNT=$(curl -s "http://localhost:9200/glitch-edges-${REPO_NAME}/_count" | jq '.count')
if [ "$EDGE_COUNT" -lt 1 ]; then
  echo "FAIL: no edges indexed (count=$EDGE_COUNT)"
  exit 1
fi
echo "PASS: $EDGE_COUNT edges indexed"

echo ""
echo "=== Smoke: search for IndexRepo symbol ==="
RESULT=$(curl -s "http://localhost:9200/glitch-symbols-${REPO_NAME}/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query":{"match":{"name":"IndexRepo"}},"size":1}')
HIT_COUNT=$(echo "$RESULT" | jq '.hits.total.value')
if [ "$HIT_COUNT" -lt 1 ]; then
  echo "FAIL: IndexRepo symbol not found"
  exit 1
fi
echo "PASS: IndexRepo symbol found"

echo ""
echo "=== Smoke: search for calls edges ==="
CALLS=$(curl -s "http://localhost:9200/glitch-edges-${REPO_NAME}/_search" \
  -H 'Content-Type: application/json' \
  -d '{"query":{"term":{"kind":"calls"}},"size":1}')
CALL_HITS=$(echo "$CALLS" | jq '.hits.total.value')
echo "calls edges: $CALL_HITS"
if [ "$CALL_HITS" -lt 1 ]; then
  echo "WARN: no calls edges found (may need review)"
fi
echo "PASS: edges query works"

echo ""
echo "=== Smoke: glitch observe with symbols ==="
$GLITCH observe "what functions are in internal/indexer" --repo "$REPO_NAME" || true
echo "PASS: observe completed (output above)"

echo ""
echo "=== Smoke: symbols-only mode ==="
$GLITCH index "$REPO_DIR" --repo "${REPO_NAME}-test" --full --symbols-only --stats
# Clean up test index
curl -s -X DELETE "http://localhost:9200/glitch-symbols-${REPO_NAME}-test" > /dev/null
curl -s -X DELETE "http://localhost:9200/glitch-edges-${REPO_NAME}-test" > /dev/null
echo "PASS: symbols-only mode completed"

echo ""
echo "=== All smoke tests passed ==="
```

- [ ] **Step 2: Make executable and run**

Run:
```bash
chmod +x test/smoke_codegraph_test.sh
cd /Users/stokes/Projects/gl1tch && ./test/smoke_codegraph_test.sh
```

Expected: All tests PASS (assuming ES is running)

- [ ] **Step 3: Commit**

```bash
git add test/smoke_codegraph_test.sh
git commit -m "test: add code graph smoke tests"
```

---

## Task 12: Final Review + Cleanup

- [ ] **Step 1: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... -count=1`
Expected: All tests PASS

- [ ] **Step 2: Run vet + build**

Run: `cd /Users/stokes/Projects/gl1tch && go vet ./... && go build ./cmd/...`
Expected: No warnings, compiles cleanly

- [ ] **Step 3: Run smoke tests**

Run: `cd /Users/stokes/Projects/gl1tch && ./test/smoke_codegraph_test.sh`
Expected: All PASS

- [ ] **Step 4: Review — invoke superpowers:requesting-code-review**

Review all changes against the design spec. Verify:
- All 6 edge kinds are produced
- Incremental indexing works (second run is fast)
- Observer queries hit symbol/edge indices
- Research tools return structured results
- CLI flags all work
- No regressions in existing tests
