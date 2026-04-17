package indexer

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func goExtractor() *LanguageExtractor {
	return &LanguageExtractor{
		Language:   "go",
		Grammar:    golang.GetLanguage(),
		Extensions: []string{".go"},
		SymbolQueries: []SymbolQuery{
			{
				Query: `(function_declaration name: (identifier) @name) @decl`,
				Kind:  KindFunction,
			},
			{
				Query: `(method_declaration name: (field_identifier) @name) @decl`,
				Kind:  KindMethod,
			},
		},
		ImportQuery: `(import_spec path: (_) @path)`,
		CallQuery:   `(call_expression function: (identifier) @callee)`,
	}
}

func TestExtractorParseGo(t *testing.T) {
	src := []byte(`package main

func Hello() string {
	return "hello"
}

func Add(a, b int) int {
	return a + b
}
`)
	ext := goExtractor()
	syms, err := ext.Extract(src, "main.go", "myrepo", "abc123")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(syms))
	}

	// Verify first function
	hello := syms[0]
	if hello.Name != "Hello" {
		t.Errorf("expected name Hello, got %s", hello.Name)
	}
	if hello.Kind != KindFunction {
		t.Errorf("expected kind %s, got %s", KindFunction, hello.Kind)
	}
	if hello.StartLine != 3 {
		t.Errorf("expected start line 3, got %d", hello.StartLine)
	}
	if hello.Language != "go" {
		t.Errorf("expected language go, got %s", hello.Language)
	}
	if hello.File != "main.go" {
		t.Errorf("expected file main.go, got %s", hello.File)
	}
	if hello.Repo != "myrepo" {
		t.Errorf("expected repo myrepo, got %s", hello.Repo)
	}
	if hello.FileHash != "abc123" {
		t.Errorf("expected file hash abc123, got %s", hello.FileHash)
	}
	if hello.Signature == "" {
		t.Error("expected non-empty signature")
	}

	// Verify second function
	add := syms[1]
	if add.Name != "Add" {
		t.Errorf("expected name Add, got %s", add.Name)
	}
	if add.Kind != KindFunction {
		t.Errorf("expected kind %s, got %s", KindFunction, add.Kind)
	}
	if add.StartLine != 7 {
		t.Errorf("expected start line 7, got %d", add.StartLine)
	}

	// Verify IDs are deterministic and distinct
	if hello.ID == "" || add.ID == "" {
		t.Error("expected non-empty IDs")
	}
	if hello.ID == add.ID {
		t.Error("expected distinct IDs")
	}
}

func TestExtractorEmptySource(t *testing.T) {
	ext := goExtractor()
	syms, err := ext.Extract([]byte{}, "empty.go", "myrepo", "deadbeef")
	if err != nil {
		t.Fatalf("Extract on empty source: %v", err)
	}
	if len(syms) != 0 {
		t.Fatalf("expected 0 symbols from empty source, got %d", len(syms))
	}
}

func TestExtractImports(t *testing.T) {
	src := []byte(`package main

import (
	"fmt"
	"os"
)
`)
	ext := goExtractor()
	imports, err := ext.ExtractImports(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractImports: %v", err)
	}
	if len(imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(imports))
	}
	if imports[0].Path != "fmt" {
		t.Errorf("expected path fmt, got %s", imports[0].Path)
	}
	if imports[1].Path != "os" {
		t.Errorf("expected path os, got %s", imports[1].Path)
	}
}

func TestExtractCalls(t *testing.T) {
	src := []byte(`package main

func main() {
	fmt.Println("hello")
	doStuff()
}
`)
	ext := goExtractor()
	calls, err := ext.ExtractCalls(src, "main.go")
	if err != nil {
		t.Fatalf("ExtractCalls: %v", err)
	}
	// Only bare identifier calls are matched (not selector calls like fmt.Println)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].CalleeName != "doStuff" {
		t.Errorf("expected callee doStuff, got %s", calls[0].CalleeName)
	}
	if calls[0].Line < 1 {
		t.Errorf("expected positive line number, got %d", calls[0].Line)
	}
}

func TestExtractNoImportQuery(t *testing.T) {
	ext := &LanguageExtractor{
		Language:   "go",
		Grammar:    golang.GetLanguage(),
		Extensions: []string{".go"},
	}
	imports, err := ext.ExtractImports([]byte(`package main`), "main.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(imports) != 0 {
		t.Fatalf("expected 0 imports when no query set, got %d", len(imports))
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`"fmt"`, "fmt"},
		{`'fmt'`, "fmt"},
		{"`fmt`", "fmt"},
		{"fmt", "fmt"},
		{`""`, ""},
	}
	for _, tt := range tests {
		got := trimQuotes(tt.in)
		if got != tt.want {
			t.Errorf("trimQuotes(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// Ensure the types compile with the expected sitter package.
var _ *sitter.Language = golang.GetLanguage()
