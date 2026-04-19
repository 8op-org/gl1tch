package indexer

import (
	"testing"
)

func TestRegistryHasGo(t *testing.T) {
	ext := ExtractorForLanguage("go")
	if ext == nil {
		t.Fatal("expected non-nil extractor for go")
	}
	if ext.Language != "go" {
		t.Errorf("expected language go, got %s", ext.Language)
	}
}

func TestRegistryHasPython(t *testing.T) {
	ext := ExtractorForLanguage("python")
	if ext == nil {
		t.Fatal("expected non-nil extractor for python")
	}
	if ext.Language != "python" {
		t.Errorf("expected language python, got %s", ext.Language)
	}
}

func TestRegistryHasTypeScript(t *testing.T) {
	ext := ExtractorForLanguage("typescript")
	if ext == nil {
		t.Fatal("expected non-nil extractor for typescript")
	}
	if ext.Language != "typescript" {
		t.Errorf("expected language typescript, got %s", ext.Language)
	}
}

func TestRegistryHasJavaScript(t *testing.T) {
	ext := ExtractorForLanguage("javascript")
	if ext == nil {
		t.Fatal("expected non-nil extractor for javascript")
	}
	if ext.Language != "javascript" {
		t.Errorf("expected language javascript, got %s", ext.Language)
	}
}

func TestRegistryReturnsNilForUnknown(t *testing.T) {
	ext := ExtractorForLanguage("brainfuck")
	if ext != nil {
		t.Fatal("expected nil for unknown language")
	}
}

func TestGoExtractorFullParse(t *testing.T) {
	ext := ExtractorForLanguage("go")
	if ext == nil {
		t.Fatal("expected non-nil extractor for go")
	}

	src := []byte(`package main

type Server struct {
	Port int
}

type Handler interface {
	Handle()
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	return nil
}

const Version = "1.0"

var Debug = false
`)

	syms, err := ext.Extract(src, "main.go", "myrepo", "abc123")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	// Expect: 1 type (Server), 1 interface (Handler), 1 function (NewServer),
	// 1 method (Start), 1 const (Version), 1 var (Debug) = 6
	kindCounts := map[string]int{}
	for _, s := range syms {
		kindCounts[s.Kind]++
	}

	want := map[string]int{
		KindType:      1,
		KindInterface: 1,
		KindFunction:  1,
		KindMethod:    1,
		KindConst:     1,
		KindVar:       1,
	}

	for kind, wantN := range want {
		if gotN := kindCounts[kind]; gotN != wantN {
			t.Errorf("kind %s: got %d, want %d", kind, gotN, wantN)
		}
	}

	if len(syms) != 6 {
		t.Errorf("expected 6 symbols total, got %d", len(syms))
		for _, s := range syms {
			t.Logf("  %s %s (line %d)", s.Kind, s.Name, s.StartLine)
		}
	}
}

func TestGoExtractorImports(t *testing.T) {
	ext := ExtractorForLanguage("go")
	if ext == nil {
		t.Fatal("expected non-nil extractor for go")
	}

	src := []byte(`package main

import (
	"fmt"
	"net/http"
	"os"
)
`)

	imports, err := ext.ExtractImports(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractImports: %v", err)
	}
	if len(imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(imports))
	}

	paths := map[string]bool{}
	for _, imp := range imports {
		paths[imp.Path] = true
	}
	for _, want := range []string{"fmt", "net/http", "os"} {
		if !paths[want] {
			t.Errorf("missing import %s", want)
		}
	}
}

func TestGoExtractorCalls(t *testing.T) {
	ext := ExtractorForLanguage("go")
	if ext == nil {
		t.Fatal("expected non-nil extractor for go")
	}

	src := []byte(`package main

func main() {
	fmt.Println("hello")
	doStuff()
	result := compute(1, 2)
	_ = result
}
`)

	calls, err := ext.ExtractCalls(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}

	// Should capture: Println (selector), doStuff, compute = 3 calls
	// (both bare identifiers and selector fields)
	callees := map[string]bool{}
	for _, c := range calls {
		callees[c.CalleeName] = true
	}

	if !callees["doStuff"] {
		t.Error("missing call to doStuff")
	}
	if !callees["compute"] {
		t.Error("missing call to compute")
	}
	if !callees["Println"] {
		t.Error("missing call to Println (selector)")
	}
}

func TestPythonExtractorFullParse(t *testing.T) {
	ext := ExtractorForLanguage("python")
	if ext == nil {
		t.Fatal("expected non-nil extractor for python")
	}

	src := []byte(`class MyClass:
    def method(self):
        pass

def standalone():
    pass

def helper(x, y):
    return x + y
`)

	syms, err := ext.Extract(src, "main.py", "myrepo", "abc123")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	kindCounts := map[string]int{}
	for _, s := range syms {
		kindCounts[s.Kind]++
	}

	// 1 class, 3 functions (method, standalone, helper)
	// Note: tree-sitter python treats methods as function_definition too,
	// so we capture them as functions unless we specifically filter nested ones.
	if kindCounts[KindClass] != 1 {
		t.Errorf("expected 1 class, got %d", kindCounts[KindClass])
	}

	// Total functions captured (both top-level and methods)
	totalFuncs := kindCounts[KindFunction] + kindCounts[KindMethod]
	if totalFuncs < 2 {
		t.Errorf("expected at least 2 functions/methods, got %d", totalFuncs)
	}
}

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

// TestAllLanguageQueriesCompile verifies that every registered language's
// queries can be compiled against its grammar without error.
func TestAllLanguageQueriesCompile(t *testing.T) {
	languages := []string{
		"go", "python", "javascript", "typescript", "tsx", "jsx",
		"ruby", "rust", "java", "c", "cpp", "csharp", "php",
		"scala", "swift", "kotlin",
	}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			ext := ExtractorForLanguage(lang)
			if ext == nil {
				t.Fatalf("no extractor for %s", lang)
			}

			// Try extracting from minimal valid source to compile all queries.
			// Use an empty-ish source; the important thing is queries compile.
			src := []byte("// empty")
			_, err := ext.Extract(src, "test.txt", "repo", "hash")
			if err != nil {
				t.Fatalf("symbol query compile error for %s: %v", lang, err)
			}

			if ext.ImportQuery != "" {
				_, err := ext.ExtractImports(src, "test.txt")
				if err != nil {
					t.Fatalf("import query compile error for %s: %v", lang, err)
				}
			}

			if ext.CallQuery != "" {
				_, err := ext.ExtractCalls(src, "test.txt")
				if err != nil {
					t.Fatalf("call query compile error for %s: %v", lang, err)
				}
			}

			if ext.RefQuery != "" {
				_, err := ext.ExtractRefs(src, "test.txt")
				if err != nil {
					t.Fatalf("ref query compile error for %s: %v", lang, err)
				}
			}
		})
	}
}
