package pipeline

import (
	"os"
	"testing"

	"github.com/8op-org/gl1tch/internal/sexpr"
)

func checkHelper(t *testing.T, src string) []Diagnostic {
	t.Helper()
	nodes, err := sexpr.Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return Check(nodes)
}

func TestCheck_UnknownForm(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" (frobnicate "x"))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagUnknownSymbol {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagUnknownSymbol for 'frobnicate'")
	}
}

func TestCheck_StepBadArity(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" (step))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagBadArity {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagBadArity for step with no args")
	}
}

func TestCheck_UndefinedRef(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" (step "a" (str "~undefined")))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagUndefinedRef {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagUndefinedRef for ~undefined")
	}
}

func TestCheck_CleanWorkflow(t *testing.T) {
	diags := checkHelper(t, `(workflow "test" :description "ok" (step "a" (run "echo hi")))`)
	for _, d := range diags {
		if d.Severity == SeverityError {
			t.Fatalf("unexpected error diagnostic: %s", d.Message)
		}
	}
}

func TestCheck_DefShadowsBuiltin(t *testing.T) {
	diags := checkHelper(t, `(def str "oops") (workflow "test" (step "a" (str "hi")))`)
	found := false
	for _, d := range diags {
		if d.Kind == DiagShadowsBuiltin {
			found = true
		}
	}
	if !found {
		t.Fatal("expected DiagShadowsBuiltin for def of 'str'")
	}
}

func TestCheck_ParDemo(t *testing.T) {
	src, err := os.ReadFile("testdata/par-demo.glitch")
	if err != nil {
		t.Skip("testdata not found")
	}
	nodes, err := sexpr.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	diags := Check(nodes)
	for _, d := range diags {
		if d.Severity == SeverityError {
			t.Errorf("unexpected error: %s", d)
		}
	}
}

func TestCheck_PhaseGate(t *testing.T) {
	src, err := os.ReadFile("testdata/phase-gate.glitch")
	if err != nil {
		t.Skip("testdata not found")
	}
	nodes, err := sexpr.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	diags := Check(nodes)
	for _, d := range diags {
		if d.Severity == SeverityError {
			t.Errorf("unexpected error: %s", d)
		}
	}
}
