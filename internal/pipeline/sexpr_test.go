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

func TestSexprWorkflow_Retry(t *testing.T) {
	src := []byte(`
(workflow "test"
  (retry 3
    (step "flaky"
      (run "curl http://example.com"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	s := w.Steps[0]
	if s.ID != "flaky" {
		t.Fatalf("expected id %q, got %q", "flaky", s.ID)
	}
	if s.Retry != 3 {
		t.Fatalf("expected retry 3, got %d", s.Retry)
	}
	if s.Run != "curl http://example.com" {
		t.Fatalf("expected run command, got %q", s.Run)
	}
}

func TestSexprWorkflow_Timeout(t *testing.T) {
	src := []byte(`
(workflow "test"
  (timeout "30s"
    (step "slow"
      (llm :prompt "think hard"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.ID != "slow" {
		t.Fatalf("expected id %q, got %q", "slow", s.ID)
	}
	if s.Timeout != "30s" {
		t.Fatalf("expected timeout %q, got %q", "30s", s.Timeout)
	}
	if s.LLM == nil || s.LLM.Prompt != "think hard" {
		t.Fatal("expected LLM step with prompt")
	}
}

func TestSexprWorkflow_RetryAndTimeout(t *testing.T) {
	src := []byte(`
(workflow "test"
  (retry 2
    (timeout "10s"
      (step "both"
        (run "flaky-thing")))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Retry != 2 {
		t.Fatalf("expected retry 2, got %d", s.Retry)
	}
	if s.Timeout != "10s" {
		t.Fatalf("expected timeout %q, got %q", "10s", s.Timeout)
	}
}

func TestSexprWorkflow_Let(t *testing.T) {
	src := []byte(`
(workflow "test"
  (let ((api-url "https://api.example.com")
        (token "abc123"))
    (step "call"
      (run api-url))
    (step "auth"
      (run token))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}
	if w.Steps[0].Run != "https://api.example.com" {
		t.Fatalf("expected resolved api-url, got %q", w.Steps[0].Run)
	}
	if w.Steps[1].Run != "abc123" {
		t.Fatalf("expected resolved token, got %q", w.Steps[1].Run)
	}
}

func TestSexprWorkflow_LetScoped(t *testing.T) {
	src := []byte(`
(def x "outer")

(workflow "test"
  (let ((x "inner"))
    (step "inside"
      (run x)))
  (step "outside"
    (run x)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].Run != "inner" {
		t.Fatalf("expected inner binding, got %q", w.Steps[0].Run)
	}
	if w.Steps[1].Run != "outer" {
		t.Fatalf("expected outer binding, got %q", w.Steps[1].Run)
	}
}

func TestSexprWorkflow_Catch(t *testing.T) {
	src := []byte(`
(workflow "test"
  (catch
    (step "try"
      (run "risky-command"))
    (step "fallback"
      (run "safe-command"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	s := w.Steps[0]
	if s.ID != "try" {
		t.Fatalf("expected id %q, got %q", "try", s.ID)
	}
	if s.Form != "catch" {
		t.Fatalf("expected form %q, got %q", "catch", s.Form)
	}
	if s.Fallback == nil {
		t.Fatal("expected fallback step")
	}
	if s.Fallback.ID != "fallback" {
		t.Fatalf("expected fallback id %q, got %q", "fallback", s.Fallback.ID)
	}
	if s.Fallback.Run != "safe-command" {
		t.Fatalf("expected fallback run, got %q", s.Fallback.Run)
	}
}

func TestSexprWorkflow_Cond(t *testing.T) {
	src := []byte(`
(workflow "test"
  (cond
    ("test -f critical.log"
      (step "critical"
        (run "alert critical")))
    ("test -f warning.log"
      (step "warn"
        (run "alert warning")))
    (else
      (step "ok"
        (run "echo all-clear")))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	s := w.Steps[0]
	if s.Form != "cond" {
		t.Fatalf("expected form %q, got %q", "cond", s.Form)
	}
	if len(s.Branches) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(s.Branches))
	}
	if s.Branches[0].Pred != "test -f critical.log" {
		t.Fatalf("branch 0: expected pred, got %q", s.Branches[0].Pred)
	}
	if s.Branches[0].Step.ID != "critical" {
		t.Fatalf("branch 0: expected step id %q, got %q", "critical", s.Branches[0].Step.ID)
	}
	if s.Branches[2].Pred != "else" {
		t.Fatalf("branch 2: expected %q, got %q", "else", s.Branches[2].Pred)
	}
}

func TestSexprWorkflow_Map(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "list"
    (run "echo -e 'a\nb\nc'"))
  (map "list"
    (step "process"
      (run "echo handling item"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}
	s := w.Steps[1]
	if s.Form != "map" {
		t.Fatalf("expected form %q, got %q", "map", s.Form)
	}
	if s.MapOver != "list" {
		t.Fatalf("expected map-over %q, got %q", "list", s.MapOver)
	}
	if s.MapBody == nil {
		t.Fatal("expected map body step")
	}
	if s.MapBody.ID != "process" {
		t.Fatalf("expected map body id %q, got %q", "process", s.MapBody.ID)
	}
}

func TestSexprWorkflow_LetWithRetry(t *testing.T) {
	src := []byte(`
(workflow "test"
  (let ((endpoint "https://api.example.com"))
    (retry 5
      (step "call"
        (run endpoint)))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Run != "https://api.example.com" {
		t.Fatalf("expected resolved endpoint, got %q", s.Run)
	}
	if s.Retry != 5 {
		t.Fatalf("expected retry 5, got %d", s.Retry)
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
