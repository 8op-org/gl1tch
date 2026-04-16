package pipeline

import "testing"

func TestParseInput_Basic(t *testing.T) {
	src := []byte(`(input :description "Free-form context" :example "fix latency spike")`)
	in, err := ParseInput(src)
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if in == nil {
		t.Fatal("ParseInput returned nil, want populated InputDef")
	}
	if in.Description != "Free-form context" {
		t.Errorf("Description = %q", in.Description)
	}
	if in.Example != "fix latency spike" {
		t.Errorf("Example = %q", in.Example)
	}
}

func TestParseInput_NoneReturnsNil(t *testing.T) {
	src := []byte(`(workflow "w" (step "s" (run "echo ok")))`)
	in, err := ParseInput(src)
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if in != nil {
		t.Errorf("expected nil InputDef when no (input ...), got %+v", in)
	}
}

func TestParseInput_RejectsMultiple(t *testing.T) {
	src := []byte(`(input :description "first") (input :description "second")`)
	_, err := ParseInput(src)
	if err == nil {
		t.Fatal("expected error for multiple (input ...) forms")
	}
}

func TestParseInput_RejectsPositionalName(t *testing.T) {
	src := []byte(`(input "foo" :description "bad")`)
	_, err := ParseInput(src)
	if err == nil {
		t.Fatal("expected error when (input ...) is given a name")
	}
}

func TestLoadBytes_WorkflowInputField(t *testing.T) {
	src := []byte(`
(input :description "Ctx" :example "e1")

(workflow "w" (step "s" (run "echo ~input")))
`)
	w, err := LoadBytes(src, "w.glitch")
	if err != nil {
		t.Fatalf("LoadBytes: %v", err)
	}
	if w.Input == nil {
		t.Fatal("Workflow.Input is nil")
	}
	if w.Input.Description != "Ctx" || w.Input.Example != "e1" {
		t.Errorf("Input = %+v", w.Input)
	}
}
