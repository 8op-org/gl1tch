package pipeline

import (
	"testing"
)

func TestSexprWorkflow_Basic(t *testing.T) {
	src := []byte(`
(workflow "my-pipeline"
  :description "a test pipeline"
  (step "fetch"
    (run "echo hello"))
  (step "analyze"
    (llm :prompt "summarize: {{step \"fetch\"}}")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "my-pipeline" {
		t.Fatalf("expected name %q, got %q", "my-pipeline", w.Name)
	}
	if w.Description != "a test pipeline" {
		t.Fatalf("expected description %q, got %q", "a test pipeline", w.Description)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}

	s0 := w.Steps[0]
	if s0.ID != "fetch" {
		t.Fatalf("step 0: expected id %q, got %q", "fetch", s0.ID)
	}
	if s0.Run != "echo hello" {
		t.Fatalf("step 0: expected run %q, got %q", "echo hello", s0.Run)
	}

	s1 := w.Steps[1]
	if s1.ID != "analyze" {
		t.Fatalf("step 1: expected id %q, got %q", "analyze", s1.ID)
	}
	if s1.LLM == nil {
		t.Fatal("step 1: expected LLM step")
	}
	if s1.LLM.Prompt != `summarize: {{step "fetch"}}` {
		t.Fatalf("step 1: expected prompt with template, got %q", s1.LLM.Prompt)
	}
}

func TestSexprWorkflow_LLMWithProviderAndModel(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (llm
      :provider "claude"
      :model "opus"
      :prompt "hello")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.LLM.Provider != "claude" {
		t.Fatalf("expected provider %q, got %q", "claude", s.LLM.Provider)
	}
	if s.LLM.Model != "opus" {
		t.Fatalf("expected model %q, got %q", "opus", s.LLM.Model)
	}
}

func TestSexprWorkflow_Save(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "gen"
    (llm :prompt "write something"))
  (step "write"
    (save "output/{{.param.repo}}/result.md" :from "gen")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.Save != "output/{{.param.repo}}/result.md" {
		t.Fatalf("expected save path, got %q", s.Save)
	}
	if s.SaveStep != "gen" {
		t.Fatalf("expected save_step %q, got %q", "gen", s.SaveStep)
	}
}

func TestSexprWorkflow_MultilinePrompt(t *testing.T) {
	src := "(workflow \"test\"\n  (step \"s1\"\n    (llm :prompt ```\n      hello\n      world\n      ```)))"
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].LLM.Prompt != "hello\nworld" {
		t.Fatalf("expected dedented multiline, got %q", w.Steps[0].LLM.Prompt)
	}
}

func TestSexprWorkflow_DiscardedStep(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "keep"
    (run "echo yes"))
  #_(step "skip"
    (run "echo no"))
  (step "also-keep"
    (run "echo yes2")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps (one discarded), got %d", len(w.Steps))
	}
	if w.Steps[0].ID != "keep" {
		t.Fatalf("expected first step 'keep', got %q", w.Steps[0].ID)
	}
	if w.Steps[1].ID != "also-keep" {
		t.Fatalf("expected second step 'also-keep', got %q", w.Steps[1].ID)
	}
}

func TestSexprWorkflow_NotAWorkflow(t *testing.T) {
	_, err := parseSexprWorkflow([]byte(`(notworkflow "test")`))
	if err == nil {
		t.Fatal("expected error for non-workflow form")
	}
}

func TestSexprWorkflow_MissingName(t *testing.T) {
	_, err := parseSexprWorkflow([]byte(`(workflow)`))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestSexprWorkflow_DefBindings(t *testing.T) {
	src := []byte(`
(def model "qwen2.5:7b")
(def provider "ollama")

(workflow "test"
  (step "s1"
    (llm
      :provider provider
      :model model
      :prompt "hello")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.LLM.Provider != "ollama" {
		t.Fatalf("expected provider %q, got %q", "ollama", s.LLM.Provider)
	}
	if s.LLM.Model != "qwen2.5:7b" {
		t.Fatalf("expected model %q, got %q", "qwen2.5:7b", s.LLM.Model)
	}
}

func TestSexprWorkflow_DefChaining(t *testing.T) {
	src := []byte(`
(def base-model "qwen2.5")
(def model "qwen2.5:7b")

(workflow "test"
  (step "s1"
    (llm :model model :prompt "hello")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].LLM.Model != "qwen2.5:7b" {
		t.Fatalf("expected %q, got %q", "qwen2.5:7b", w.Steps[0].LLM.Model)
	}
}

func TestSexprWorkflow_DefInRun(t *testing.T) {
	src := []byte(`
(def cmd "echo hello")

(workflow "test"
  (step "s1"
    (run cmd)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].Run != "echo hello" {
		t.Fatalf("expected %q, got %q", "echo hello", w.Steps[0].Run)
	}
}

func TestSexprWorkflow_UnresolvedSymbolPassesThrough(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (llm :model unknown-thing :prompt "hello")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].LLM.Model != "unknown-thing" {
		t.Fatalf("expected %q, got %q", "unknown-thing", w.Steps[0].LLM.Model)
	}
}

func TestLoadBytes_Sexpr_LLMTierAndFormat(t *testing.T) {
	src := []byte(`
(workflow "test-tier"
  :description "test tier and format"
  (step "classify"
    (llm
      :tier 1
      :format "json"
      :prompt "classify this")))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	llm := w.Steps[0].LLM
	if llm == nil {
		t.Fatal("expected LLM step")
	}
	if llm.Tier == nil || *llm.Tier != 1 {
		t.Errorf("tier = %v, want 1", llm.Tier)
	}
	if llm.Format != "json" {
		t.Errorf("format = %q, want json", llm.Format)
	}
}

func TestLoadBytes_Sexpr_LLMNoTier(t *testing.T) {
	src := []byte(`
(workflow "test-no-tier"
  :description "no tier set"
  (step "ask"
    (llm
      :prompt "hello")))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	llm := w.Steps[0].LLM
	if llm.Tier != nil {
		t.Errorf("tier should be nil when not set, got %v", *llm.Tier)
	}
}
