package pipeline

import (
	"testing"
)

func TestSexpr_MetadataExtraction(t *testing.T) {
	src := []byte(`
(workflow "test-wf"
  :description "a test workflow"
  :author "adam"
  :version "1.0"
  :tags ("foo" "bar")
  (step "s1"
    (run "echo hello")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "test-wf" {
		t.Fatalf("expected name %q, got %q", "test-wf", w.Name)
	}
	if w.Description != "a test workflow" {
		t.Fatalf("expected description %q, got %q", "a test workflow", w.Description)
	}
	if w.Author != "adam" {
		t.Fatalf("expected author %q, got %q", "adam", w.Author)
	}
	if w.Version != "1.0" {
		t.Fatalf("expected version %q, got %q", "1.0", w.Version)
	}
	if len(w.Tags) != 2 || w.Tags[0] != "foo" || w.Tags[1] != "bar" {
		t.Fatalf("expected tags [foo bar], got %v", w.Tags)
	}
	if w.Source == nil {
		t.Fatal("expected Source to be set")
	}
}

func TestSexpr_NoWorkflowForm(t *testing.T) {
	src := []byte(`(def x "hello")`)
	_, err := parseSexprWorkflow(src)
	if err == nil {
		t.Fatal("expected error for missing workflow form")
	}
}

func TestSexpr_NameOnly(t *testing.T) {
	src := []byte(`(workflow "minimal")`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "minimal" {
		t.Fatalf("expected name %q, got %q", "minimal", w.Name)
	}
}

func TestSexpr_DefResolution(t *testing.T) {
	src := []byte(`
(def my-desc "from def")
(workflow "test-def"
  :description my-desc
  (step "s1"
    (run "echo ok")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Description != "from def" {
		t.Fatalf("expected description %q, got %q", "from def", w.Description)
	}
}
