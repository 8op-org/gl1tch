package pipeline

import "testing"

func TestRender_WithParams(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"input": "work on issue 3442",
		"param": map[string]string{
			"repo":  "elastic/observability-robots",
			"issue": "3442",
		},
	}
	result, err := render(`gh issue view {{.param.issue}} --repo {{.param.repo}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}
	expected := "gh issue view 3442 --repo elastic/observability-robots"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestLoadBytes_Sexpr(t *testing.T) {
	src := []byte(`
(workflow "test-sexpr"
  :description "loaded from sexpr"
  (step "s1"
    (run "echo hello")))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "test-sexpr" {
		t.Fatalf("expected name %q, got %q", "test-sexpr", w.Name)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	if w.Steps[0].Run != "echo hello" {
		t.Fatalf("expected run %q, got %q", "echo hello", w.Steps[0].Run)
	}
}

func TestRender_WithStepRefs(t *testing.T) {
	steps := map[string]string{
		"fetch": `{"title": "fix bug"}`,
	}
	data := map[string]any{
		"input": "test",
	}
	result, err := render(`Issue: {{step "fetch"}}`, data, steps)
	if err != nil {
		t.Fatal(err)
	}
	expected := `Issue: {"title": "fix bug"}`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}
