package pipeline

import (
	"strings"
	"testing"

	"github.com/8op-org/gl1tch/internal/provider"
)

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

func TestRunWithParams(t *testing.T) {
	w := &Workflow{
		Name: "test-params",
		Steps: []Step{
			{
				ID:  "echo-param",
				Run: `echo "issue={{.param.issue}} repo={{.param.repo}}"`,
			},
		},
	}
	params := map[string]string{
		"issue": "3642",
		"repo":  "elastic/observability-robots",
	}
	result, err := Run(w, "", "", params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "issue=3642 repo=elastic/observability-robots"
	if strings.TrimSpace(result.Output) != expected {
		t.Fatalf("expected %q, got %q", expected, strings.TrimSpace(result.Output))
	}
}

func TestRun_OpenAICompatibleProvider(t *testing.T) {
	called := false
	resolver := func(name string) (provider.ProviderFunc, bool) {
		if name == "openrouter" {
			return func(model, prompt string) (provider.LLMResult, error) {
				called = true
				return provider.LLMResult{
					Provider: "openrouter",
					Model:    model,
					Response: "llm-output",
					TokensIn: 10, TokensOut: 5,
				}, nil
			}, true
		}
		return nil, false
	}

	w := &Workflow{
		Name: "test-openai",
		Steps: []Step{
			{
				ID: "ask",
				LLM: &LLMStep{
					Provider: "openrouter",
					Model:    "meta-llama/llama-4-scout:free",
					Prompt:   "say hello",
				},
			},
		},
	}

	reg, _ := provider.LoadProviders(t.TempDir())
	result, err := Run(w, "", "qwen3:8b", nil, reg, RunOpts{
		ProviderResolver: resolver,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Fatal("resolver was not called for openrouter provider")
	}
	if result.Output != "llm-output" {
		t.Errorf("output = %q, want llm-output", result.Output)
	}
}
