// internal/pipeline/scope_test.go
package pipeline

import (
	"os"
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

func TestScopeEnvSet(t *testing.T) {
	os.Setenv("GL1TCH_TEST_VAR_X", "hello")
	defer os.Unsetenv("GL1TCH_TEST_VAR_X")
	s := NewScope()
	v, err := s.ResolvePath("env", []string{"GL1TCH_TEST_VAR_X"})
	if err != nil || v != "hello" {
		t.Errorf("got %q err=%v", v, err)
	}
}

func TestScopeEnvUnsetFailsLoud(t *testing.T) {
	os.Unsetenv("GL1TCH_DEFINITELY_NOT_SET_42")
	s := NewScope()
	_, err := s.ResolvePath("env", []string{"GL1TCH_DEFINITELY_NOT_SET_42"})
	if err == nil {
		t.Fatal("expected UndefinedRefError on unset env var")
	}
}

func TestScopeResourceHit(t *testing.T) {
	s := NewScope()
	s.SetResources(map[string]map[string]string{
		"my-repo": {"url": "https://github.com/x/y"},
	})
	v, err := s.ResolvePath("resource", []string{"my-repo", "url"})
	if err != nil || v != "https://github.com/x/y" {
		t.Errorf("got %q err=%v", v, err)
	}
}

func TestScopeResourceMissingFailsLoud(t *testing.T) {
	s := NewScope()
	_, err := s.ResolvePath("resource", []string{"nope", "url"})
	if err == nil {
		t.Fatal("expected UndefinedRefError on missing resource")
	}
}
