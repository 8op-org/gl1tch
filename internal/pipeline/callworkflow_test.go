package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCallWorkflowChildRun(t *testing.T) {
	dir := t.TempDir()
	parent := `(workflow "parent" (step "out" (call-workflow "child" :input "hi")))`
	child := `(workflow "child" (step "echo" (run "echo got:{{.input}}")))`
	_ = os.WriteFile(filepath.Join(dir, "parent.glitch"), []byte(parent), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "child.glitch"), []byte(child), 0o644)

	w, err := ParseSexprWorkflowFromFile(filepath.Join(dir, "parent.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := Run(w, "hi", "", map[string]string{}, nil, RunOpts{WorkflowsDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Steps["out"] != "got:hi" {
		t.Fatalf("unexpected output: %q", res.Steps["out"])
	}
}

func TestCallWorkflowCycleRejected(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.glitch"),
		[]byte(`(workflow "a" (step "x" (call-workflow "b")))`), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.glitch"),
		[]byte(`(workflow "b" (step "x" (call-workflow "a")))`), 0o644)

	w, err := ParseSexprWorkflowFromFile(filepath.Join(dir, "a.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Run(w, "", "", nil, nil, RunOpts{WorkflowsDir: dir})
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}
