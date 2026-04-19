package pipeline

import (
	"bytes"
	"strings"
	"testing"
)

func TestREPL_BasicDefAndLookup(t *testing.T) {
	in := strings.NewReader("(def x \"42\")\nx\n")
	out := &bytes.Buffer{}
	r := NewREPL(NewEvaluator(), in, out)
	err := r.Run()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "42") {
		t.Fatalf("expected output to contain %q, got:\n%s", "42", out.String())
	}
}

func TestREPL_ParseError(t *testing.T) {
	in := strings.NewReader("(def x\n")
	out := &bytes.Buffer{}
	r := NewREPL(NewEvaluator(), in, out)
	_ = r.Run()
	if !strings.Contains(out.String(), "error:") {
		t.Fatalf("expected error output, got:\n%s", out.String())
	}
}

func TestREPL_MultilineInput(t *testing.T) {
	in := strings.NewReader("(str \"hello \"\n  \"world\")\n")
	out := &bytes.Buffer{}
	r := NewREPL(NewEvaluator(), in, out)
	_ = r.Run()
	if !strings.Contains(out.String(), "hello world") {
		t.Fatalf("expected %q, got:\n%s", "hello world", out.String())
	}
}

func TestREPL_Balanced(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{"empty", "", true},
		{"matched", "(def x 1)", true},
		{"unmatched open", "(def x", false},
		{"nested", "(if (eq x 1) (str \"yes\"))", true},
		{"string with parens", "(str \"(not a paren)\")", true},
		{"comment", "(def x 1) ; ignore this (", true},
		{"triple backtick", "(str ```hello (```)", true},
		{"unmatched in triple", "(str ```hello (", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := balanced(tt.src); got != tt.want {
				t.Errorf("balanced(%q) = %v, want %v", tt.src, got, tt.want)
			}
		})
	}
}

func TestREPL_PersistentEnv(t *testing.T) {
	in := strings.NewReader("(def a \"10\")\n(def b \"20\")\n(str a \" \" b)\n")
	out := &bytes.Buffer{}
	r := NewREPL(NewEvaluator(), in, out)
	_ = r.Run()
	if !strings.Contains(out.String(), "10 20") {
		t.Fatalf("expected persistent env to resolve a and b, got:\n%s", out.String())
	}
}
