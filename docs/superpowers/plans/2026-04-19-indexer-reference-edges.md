# Indexer Reference Edges Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `references` edges to the glitch indexer so `glitch index` can detect dead code (symbols with no inbound call or reference edges).

**Architecture:** Add a `RefQuery` field to `LanguageExtractor`, an `ExtractRefs` method mirroring `ExtractCalls`, and `ResolveRefs` on `Resolver`. Wire into the `IndexRepoGraph` pipeline. Reference edges capture non-call symbol usage: selector/member access (method values, field reads) and type identifiers (struct literals, params, returns, type assertions).

**Tech Stack:** Go, tree-sitter (smacker/go-tree-sitter), Elasticsearch

**Languages:** Go, Python, JavaScript/JSX, TypeScript/TSX, C, Rust, Java

---

### Task 1: Add EdgeReferences constant and RefQuery field

**Files:**
- Modify: `internal/indexer/symbol.go:22-30`
- Modify: `internal/indexer/extractor.go:34-44`

- [ ] **Step 1: Write the failing test**

Add to `internal/indexer/symbol_test.go`:

```go
func TestEdgeReferencesConstant(t *testing.T) {
	if EdgeReferences != "references" {
		t.Errorf("expected EdgeReferences = %q, got %q", "references", EdgeReferences)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/indexer/ -run TestEdgeReferencesConstant -v`
Expected: FAIL — `EdgeReferences` undefined

- [ ] **Step 3: Add EdgeReferences constant**

In `internal/indexer/symbol.go`, add to the edge kind const block:

```go
const (
	EdgeContains   = "contains"
	EdgeImports    = "imports"
	EdgeExports    = "exports"
	EdgeExtends    = "extends"
	EdgeImplements = "implements"
	EdgeCalls      = "calls"
	EdgeReferences = "references"
)
```

- [ ] **Step 4: Add RefQuery field to LanguageExtractor**

In `internal/indexer/extractor.go`, add `RefQuery` to the struct:

```go
type LanguageExtractor struct {
	Language      string
	Grammar       *sitter.Language
	Extensions    []string
	SymbolQueries []SymbolQuery
	ImportQuery   string
	ExportQuery   string
	CallQuery     string
	RefQuery      string
	PathResolver  func(importPath, fromFile, repoRoot string) string
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/indexer/ -run TestEdgeReferencesConstant -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/indexer/symbol.go internal/indexer/symbol_test.go internal/indexer/extractor.go
git commit -m "feat(indexer): add EdgeReferences constant and RefQuery field"
```

---

### Task 2: Add ExtractRefs method

**Files:**
- Modify: `internal/indexer/extractor.go`
- Modify: `internal/indexer/extractor_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/indexer/extractor_test.go`:

```go
func TestExtractRefs(t *testing.T) {
	src := []byte(`package main

type Server struct{}

func (s *Server) Start() {}

func main() {
	handler := s.Start
	var srv Server
}
`)
	ext := &LanguageExtractor{
		Language:   "go",
		Grammar:    golang.GetLanguage(),
		Extensions: []string{".go"},
		RefQuery: strings.Join([]string{
			`(selector_expression field: (field_identifier) @ref)`,
			`(type_identifier) @ref`,
		}, "\n"),
	}
	refs, err := ext.ExtractRefs(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractRefs: %v", err)
	}
	// Should capture: Start (selector), Server (type_identifier x2 — struct decl + var decl)
	// The exact count depends on tree-sitter; just verify we got some refs.
	if len(refs) == 0 {
		t.Fatal("expected at least 1 ref, got 0")
	}

	names := make(map[string]bool)
	for _, r := range refs {
		names[r.CalleeName] = true
	}
	if !names["Start"] {
		t.Error("expected ref to Start (selector expression)")
	}
	if !names["Server"] {
		t.Error("expected ref to Server (type identifier)")
	}
}

func TestExtractRefsEmptyQuery(t *testing.T) {
	ext := &LanguageExtractor{
		Language:   "go",
		Grammar:    golang.GetLanguage(),
		Extensions: []string{".go"},
	}
	refs, err := ext.ExtractRefs([]byte(`package main`), "main.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("expected 0 refs when no query set, got %d", len(refs))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/indexer/ -run TestExtractRefs -v`
Expected: FAIL — `ExtractRefs` undefined

- [ ] **Step 3: Implement ExtractRefs**

Add to `internal/indexer/extractor.go`, after `ExtractCalls`:

```go
// ExtractRefs runs the RefQuery against source and returns reference sites
// (non-call symbol references like selector access and type identifiers).
// If RefQuery is empty, returns nil.
func (le *LanguageExtractor) ExtractRefs(source []byte, file string) ([]CallSite, error) {
	if le.RefQuery == "" {
		return nil, nil
	}
	if len(source) == 0 {
		return nil, nil
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(le.Grammar)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, err
	}
	defer tree.Close()
	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(le.RefQuery), le.Grammar)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(q, root)

	var refs []CallSite

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, cap := range match.Captures {
			capName := q.CaptureNameForId(cap.Index)
			if capName == "ref" {
				refs = append(refs, CallSite{
					CalleeName: cap.Node.Content(source),
					File:       file,
					Line:       int(cap.Node.StartPoint().Row) + 1,
				})
			}
		}
	}

	return refs, nil
}
```

Note: we reuse the `CallSite` struct — it has the same shape (name + file + line). No need for a new type.

- [ ] **Step 4: Add import for strings to test file**

The test file needs `"strings"` imported. Add it to the import block in `extractor_test.go`.

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/indexer/ -run TestExtractRefs -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/indexer/extractor.go internal/indexer/extractor_test.go
git commit -m "feat(indexer): add ExtractRefs method for non-call symbol references"
```

---

### Task 3: Add ResolveRefs to Resolver

**Files:**
- Modify: `internal/indexer/resolve.go`
- Modify: `internal/indexer/resolve_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/indexer/resolve_test.go`:

```go
func TestResolveRefsLocalScope(t *testing.T) {
	repo := "example/repo"
	repoRoot := "/src/repo"

	callerID := SymbolID("pkg/foo.go", KindFunction, "main", 1)
	targetID := SymbolID("pkg/foo.go", KindMethod, "Start", 10)

	symbols := []SymbolDoc{
		{
			ID:        callerID,
			File:      "pkg/foo.go",
			Kind:      KindFunction,
			Name:      "main",
			StartLine: 1,
			EndLine:   8,
			Repo:      repo,
		},
		{
			ID:        targetID,
			File:      "pkg/foo.go",
			Kind:      KindMethod,
			Name:      "Start",
			StartLine: 10,
			EndLine:   15,
			Repo:      repo,
		},
	}

	r := NewResolver(symbols, repo, repoRoot, nil)

	refs := []CallSite{
		{
			CalleeName: "Start",
			File:       "pkg/foo.go",
			Line:       5,
		},
	}

	edges := r.ResolveRefs(refs)

	if len(edges) != 1 {
		t.Fatalf("expected 1 references edge, got %d", len(edges))
	}

	e := edges[0]
	if e.SourceID != callerID {
		t.Errorf("expected source %q, got %q", callerID, e.SourceID)
	}
	if e.TargetID != targetID {
		t.Errorf("expected target %q, got %q", targetID, e.TargetID)
	}
	if e.Kind != EdgeReferences {
		t.Errorf("expected kind %q, got %q", EdgeReferences, e.Kind)
	}
}

func TestResolveRefsMatchesTypes(t *testing.T) {
	repo := "example/repo"
	repoRoot := "/src/repo"

	callerID := SymbolID("pkg/foo.go", KindFunction, "main", 1)
	typeID := SymbolID("pkg/foo.go", KindType, "Server", 10)

	symbols := []SymbolDoc{
		{
			ID:        callerID,
			File:      "pkg/foo.go",
			Kind:      KindFunction,
			Name:      "main",
			StartLine: 1,
			EndLine:   8,
			Repo:      repo,
		},
		{
			ID:        typeID,
			File:      "pkg/foo.go",
			Kind:      KindType,
			Name:      "Server",
			StartLine: 10,
			EndLine:   20,
			Repo:      repo,
		},
	}

	r := NewResolver(symbols, repo, repoRoot, nil)

	refs := []CallSite{
		{
			CalleeName: "Server",
			File:       "pkg/foo.go",
			Line:       5,
		},
	}

	edges := r.ResolveRefs(refs)

	if len(edges) != 1 {
		t.Fatalf("expected 1 references edge, got %d", len(edges))
	}
	if edges[0].TargetID != typeID {
		t.Errorf("expected target %q, got %q", typeID, edges[0].TargetID)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/indexer/ -run TestResolveRefs -v`
Expected: FAIL — `ResolveRefs` undefined

- [ ] **Step 3: Implement ResolveRefs**

Add to `internal/indexer/resolve.go`, after `ResolveCalls`:

```go
// ResolveRefs resolves reference sites into "references" edges. Unlike
// ResolveCalls which only matches functions/methods, this also matches types,
// interfaces, and classes — capturing non-call symbol usage like method values,
// type identifiers in struct fields, params, returns, and composite literals.
func (r *Resolver) ResolveRefs(refs []CallSite) []EdgeDoc {
	var edges []EdgeDoc
	for _, ref := range refs {
		targetID := r.findRefTarget(ref)
		if targetID == "" {
			continue
		}
		callerID := r.findEnclosing(ref.File, ref.Line)
		if callerID == "" {
			continue
		}
		edges = append(edges, EdgeDoc{
			SourceID: callerID,
			TargetID: targetID,
			Kind:     EdgeReferences,
			File:     ref.File,
			Repo:     r.repo,
		})
	}
	return edges
}

// findRefTarget finds the target symbol for a reference. Unlike findCallTarget,
// this matches any symbol kind (types, interfaces, classes, functions, methods).
func (r *Resolver) findRefTarget(ref CallSite) string {
	// Local scope: same file, matching name.
	if syms, ok := r.byFile[ref.File]; ok {
		for _, s := range syms {
			if s.Name == ref.CalleeName {
				return s.ID
			}
		}
	}
	// Global: any symbol by name.
	if syms, ok := r.byName[ref.CalleeName]; ok {
		if len(syms) > 0 {
			return syms[0].ID
		}
	}
	return ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/indexer/ -run TestResolveRefs -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/indexer/resolve.go internal/indexer/resolve_test.go
git commit -m "feat(indexer): add ResolveRefs for non-call reference edges"
```

---

### Task 4: Add RefQuery to language extractors

**Files:**
- Modify: `internal/indexer/languages.go`
- Modify: `internal/indexer/languages_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/indexer/languages_test.go`:

```go
func TestRefQueryRegistered(t *testing.T) {
	langs := []string{"go", "python", "javascript", "jsx", "typescript", "tsx", "c", "rust", "java"}
	for _, lang := range langs {
		ext := ExtractorForLanguage(lang)
		if ext == nil {
			t.Errorf("no extractor for %s", lang)
			continue
		}
		if ext.RefQuery == "" {
			t.Errorf("RefQuery not set for %s", lang)
		}
	}
}

func TestGoRefQueryExtractsMethodValue(t *testing.T) {
	src := []byte(`package main

type Handler struct{}

func (h *Handler) Serve() {}

func register() {
	h := &Handler{}
	callback := h.Serve
	_ = callback
}
`)
	ext := ExtractorForLanguage("go")
	refs, err := ext.ExtractRefs(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractRefs: %v", err)
	}

	names := make(map[string]bool)
	for _, r := range refs {
		names[r.CalleeName] = true
	}
	if !names["Serve"] {
		t.Error("expected ref to Serve (method value)")
	}
	if !names["Handler"] {
		t.Error("expected ref to Handler (type identifier)")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/indexer/ -run "TestRefQuery|TestGoRefQuery" -v`
Expected: FAIL — RefQuery is empty for all languages

- [ ] **Step 3: Add RefQuery to all target languages**

In `internal/indexer/languages.go`, add `RefQuery` to each extractor definition:

**Go** (add after `CallQuery` in goExt):
```go
RefQuery: strings.Join([]string{
	`(selector_expression field: (field_identifier) @ref)`,
	`(type_identifier) @ref`,
}, "\n"),
```

**Python** (add after `CallQuery` in pyExt):
```go
RefQuery: `(attribute attribute: (identifier) @ref)`,
```

**JavaScript** (add after `CallQuery` in jsExt):
```go
RefQuery: `(member_expression property: (property_identifier) @ref)`,
```

Note: JSX inherits from jsExt via `copyExtractor`, so it gets RefQuery automatically.

**TypeScript** (add after `CallQuery` in tsExt):
```go
RefQuery: strings.Join([]string{
	`(member_expression property: (property_identifier) @ref)`,
	`(type_identifier) @ref`,
}, "\n"),
```

**TSX** (add after `CallQuery` in tsxExt):
```go
RefQuery: strings.Join([]string{
	`(member_expression property: (property_identifier) @ref)`,
	`(type_identifier) @ref`,
}, "\n"),
```

**C** (add after `CallQuery` in c extractor):
```go
RefQuery: strings.Join([]string{
	`(field_expression field: (field_identifier) @ref)`,
	`(type_identifier) @ref`,
}, "\n"),
```

**Rust** (add after `CallQuery` in rust extractor):
```go
RefQuery: strings.Join([]string{
	`(field_expression field: (field_identifier) @ref)`,
	`(type_identifier) @ref`,
}, "\n"),
```

**Java** (add after `CallQuery` in java extractor):
```go
RefQuery: strings.Join([]string{
	`(field_access field: (identifier) @ref)`,
	`(type_identifier) @ref`,
}, "\n"),
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/indexer/ -run "TestRefQuery|TestGoRefQuery" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/indexer/languages.go internal/indexer/languages_test.go
git commit -m "feat(indexer): add RefQuery for Go, Python, JS/TS, C, Rust, Java"
```

---

### Task 5: Wire ExtractRefs + ResolveRefs into IndexRepoGraph pipeline

**Files:**
- Modify: `internal/indexer/indexer.go:377-449`

- [ ] **Step 1: Write the failing test**

This is an integration test. Add to `internal/indexer/indexer_test.go`:

```go
func TestExtractRefsPhase(t *testing.T) {
	// Verify that the extraction phase collects refs alongside calls.
	// We test this at the extractor level since IndexRepoGraph needs ES.
	src := []byte(`package main

type Config struct{}

func NewConfig() *Config {
	return &Config{}
}

func setup() {
	c := NewConfig()
	_ = c
}
`)
	ext := ExtractorForLanguage("go")

	calls, err := ext.ExtractCalls(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}

	refs, err := ext.ExtractRefs(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractRefs: %v", err)
	}

	// Calls should include NewConfig
	callNames := make(map[string]bool)
	for _, c := range calls {
		callNames[c.CalleeName] = true
	}
	if !callNames["NewConfig"] {
		t.Error("expected call to NewConfig")
	}

	// Refs should include Config (type identifier)
	refNames := make(map[string]bool)
	for _, r := range refs {
		refNames[r.CalleeName] = true
	}
	if !refNames["Config"] {
		t.Error("expected ref to Config (type identifier)")
	}
}
```

- [ ] **Step 2: Run test to verify it passes** (this validates the plumbing is correct at extractor level)

Run: `go test ./internal/indexer/ -run TestExtractRefsPhase -v`
Expected: PASS

- [ ] **Step 3: Wire into IndexRepoGraph**

In `internal/indexer/indexer.go`, modify the Phase 1 extraction loop (around line 377-414). Add `allRefs` collection alongside `allCalls`:

After the `var allCalls []CallSite` declaration (line 381), add:
```go
var allRefs []CallSite
```

Inside the file loop, after the calls extraction block (after line 413), add:
```go
		refs, err := ext.ExtractRefs(fe.content, fe.relPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warn: extract refs %s: %v\n", fe.relPath, err)
		} else {
			allRefs = append(allRefs, refs...)
		}
```

In Phase 2+3 edge resolution (around line 444-449), add after the `ResolveCalls` line:
```go
	allEdges = append(allEdges, resolver.ResolveRefs(allRefs)...)
```

- [ ] **Step 4: Run full test suite**

Run: `go test ./internal/indexer/ -v`
Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/indexer/indexer.go internal/indexer/indexer_test.go
git commit -m "feat(indexer): wire ExtractRefs + ResolveRefs into IndexRepoGraph pipeline"
```

---

### Task 6: Validate with glitch index

- [ ] **Step 1: Build and run**

```bash
go build -o /tmp/glitch-ref-test . && /tmp/glitch-ref-test index . --symbols-only --languages go --full --stats
```

Expected: higher edge count than before (was 11076 edges).

- [ ] **Step 2: Query for unreferenced pipeline symbols**

```bash
# Get all symbol IDs in internal/pipeline (non-test)
curl -s 'http://localhost:9200/glitch-symbols-lisp-eval-cleanup/_search' -H 'Content-Type: application/json' -d '{
  "size": 500,
  "query": {
    "bool": {
      "must": [
        {"prefix": {"file": "internal/pipeline/"}},
        {"terms": {"kind": ["function", "method", "type", "interface"]}}
      ],
      "must_not": [{"wildcard": {"file": "*_test.go"}}]
    }
  },
  "_source": ["id", "name", "kind", "file", "start_line"]
}' | jq -r '.hits.hits[]._source | "\(.id)\t\(.name)\t\(.kind)\t\(.file):\(.start_line)"' | sort -t$'\t' -k2 > /tmp/pipeline_symbols.tsv

# Get all target_ids from call + reference edges
curl -s 'http://localhost:9200/glitch-edges-lisp-eval-cleanup/_search' -H 'Content-Type: application/json' -d '{
  "size": 10000,
  "query": {"terms": {"kind": ["calls", "references"]}},
  "_source": ["target_id"]
}' | jq -r '.hits.hits[]._source.target_id' | sort -u > /tmp/all_targets.txt

# Find unreferenced symbols
while IFS=$'\t' read -r id name kind location; do
  if ! grep -qF "$id" /tmp/all_targets.txt; then
    echo "$kind  $name  $location"
  fi
done < /tmp/pipeline_symbols.tsv | sort
```

Expected: dramatically fewer false positives — builtins and types should no longer show as dead code. Remaining unreferenced symbols are actual dead code candidates.

- [ ] **Step 3: Commit any fixes if queries need adjustment**

Only if tree-sitter queries produce unexpected results during validation.
