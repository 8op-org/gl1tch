// internal/pipeline/scope_test.go
package pipeline

import (
	"strings"
	"testing"
)

func TestScopeResolvesBareSymbol(t *testing.T) {
	s := NewScope()
	s.SetLet("model", "qwen2.5:7b")
	v, err := s.Resolve("model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "qwen2.5:7b" {
		t.Errorf("got %q, want %q", v, "qwen2.5:7b")
	}
}

func TestScopeLetShadowsDef(t *testing.T) {
	s := NewScope()
	s.SetDef("model", "llama3")
	s.SetLet("model", "qwen2.5:7b")
	v, _ := s.Resolve("model")
	if v != "qwen2.5:7b" {
		t.Errorf("let should shadow def; got %q", v)
	}
}

func TestScopeUndefinedSymbolReturnsError(t *testing.T) {
	s := NewScope()
	s.SetDef("model", "llama3")
	_, err := s.Resolve("modle")
	if err == nil {
		t.Fatal("expected UndefinedRefError, got nil")
	}
	var ure *UndefinedRefError
	if !errorsAs(err, &ure) {
		t.Fatalf("want UndefinedRefError, got %T", err)
	}
	if !strings.Contains(err.Error(), "modle") {
		t.Errorf("error should mention symbol; got: %v", err)
	}
	if !strings.Contains(err.Error(), "model") {
		t.Errorf("error should include suggestion; got: %v", err)
	}
}

// errorsAs is a tiny wrapper so the test file doesn't need to import "errors".
func errorsAs(err error, target any) bool {
	return errorsAsStd(err, target)
}
