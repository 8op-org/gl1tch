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

func TestCallWorkflowCallsChildRunCreator(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "parent.glitch"),
		[]byte(`(workflow "parent" (step "o" (call-workflow "child")))`), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "child.glitch"),
		[]byte(`(workflow "child" (step "e" (run "echo ok")))`), 0o644)
	w, err := ParseSexprWorkflowFromFile(filepath.Join(dir, "parent.glitch"))
	if err != nil {
		t.Fatal(err)
	}
	var calls []struct {
		Parent   int64
		Workflow string
	}
	creator := func(parent int64, name string) (int64, error) {
		calls = append(calls, struct {
			Parent   int64
			Workflow string
		}{parent, name})
		return int64(len(calls) + 1000), nil // fake child row id
	}
	_, err = Run(w, "", "", nil, nil, RunOpts{
		WorkflowsDir:    dir,
		ParentRunID:     42,
		ChildRunCreator: creator,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("expected one creator call, got %d", len(calls))
	}
	if calls[0].Parent != 42 {
		t.Errorf("expected child to be linked to parent 42, got %d", calls[0].Parent)
	}
	if calls[0].Workflow != "child" {
		t.Errorf("expected workflow name 'child', got %q", calls[0].Workflow)
	}
}
