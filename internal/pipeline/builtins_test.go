// internal/pipeline/builtins_test.go
package pipeline

import "testing"

func TestBuiltinStep(t *testing.T) {
	s := NewScope()
	s.SetSteps(map[string]string{"fetch": "hello"})
	v, err := callBuiltin("step", []string{"fetch"}, s)
	if err != nil || v != "hello" {
		t.Errorf("got %q err=%v", v, err)
	}
}

func TestBuiltinStepUnknown(t *testing.T) {
	s := NewScope()
	s.SetSteps(map[string]string{})
	_, err := callBuiltin("step", []string{"missing"}, s)
	if err == nil {
		t.Fatal("expected error on unknown step")
	}
}

func TestBuiltinOr(t *testing.T) {
	s := NewScope()
	v, _ := callBuiltin("or", []string{"", "fallback"}, s)
	if v != "fallback" {
		t.Errorf("got %q", v)
	}
	v2, _ := callBuiltin("or", []string{"first", "second"}, s)
	if v2 != "first" {
		t.Errorf("got %q", v2)
	}
}

func TestBuiltinUnknown(t *testing.T) {
	_, err := callBuiltin("doesnotexist", nil, NewScope())
	if err == nil {
		t.Fatal("expected error")
	}
}
