package pipeline

import (
	"strings"
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

func TestSexprWorkflow_Flatten(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "fetch"
    (run "echo '[{\"a\":1},{\"b\":2}]'"))
  (step "flat"
    (flatten "fetch")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(w.Steps))
	}
	s := w.Steps[1]
	if s.Flatten != "fetch" {
		t.Fatalf("expected flatten %q, got %q", "fetch", s.Flatten)
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

func TestSexprWorkflow_JsonPick(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "fetch"
    (run "curl http://example.com/api"))
  (step "pick"
    (json-pick ".a" :from "fetch")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.JsonPick == nil {
		t.Fatal("expected JsonPick step")
	}
	if s.JsonPick.Expr != ".a" {
		t.Fatalf("expected expr %q, got %q", ".a", s.JsonPick.Expr)
	}
	if s.JsonPick.From != "fetch" {
		t.Fatalf("expected from %q, got %q", "fetch", s.JsonPick.From)
	}
}

func TestSexprWorkflow_Lines(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "list"
    (run "echo -e 'a\nb\nc'"))
  (step "split"
    (lines "list")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.Lines != "list" {
		t.Fatalf("expected lines %q, got %q", "list", s.Lines)
	}
}

func TestSexprWorkflow_Merge(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "a"
    (run "echo a"))
  (step "b"
    (run "echo b"))
  (step "combined"
    (merge "a" "b")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[2]
	if len(s.Merge) != 2 {
		t.Fatalf("expected 2 merge IDs, got %d", len(s.Merge))
	}
	if s.Merge[0] != "a" || s.Merge[1] != "b" {
		t.Fatalf("expected merge [a b], got %v", s.Merge)
	}
}

func TestSexprWorkflow_HttpGet(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "fetch"
    (http-get "https://api.example.com" :headers {"Auth" "Bearer token"})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected HttpCall step")
	}
	if s.HttpCall.Method != "GET" {
		t.Fatalf("expected method %q, got %q", "GET", s.HttpCall.Method)
	}
	if s.HttpCall.URL != "https://api.example.com" {
		t.Fatalf("expected URL, got %q", s.HttpCall.URL)
	}
	if s.HttpCall.Headers["Auth"] != "Bearer token" {
		t.Fatalf("expected header Auth=%q, got %q", "Bearer token", s.HttpCall.Headers["Auth"])
	}
}

func TestSexprWorkflow_HttpPost(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "post"
    (http-post "https://api.example.com" :body "{}" :headers {"Content-Type" "application/json"})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected HttpCall step")
	}
	if s.HttpCall.Method != "POST" {
		t.Fatalf("expected method %q, got %q", "POST", s.HttpCall.Method)
	}
	if s.HttpCall.Body != "{}" {
		t.Fatalf("expected body %q, got %q", "{}", s.HttpCall.Body)
	}
	if s.HttpCall.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected header Content-Type=%q, got %q", "application/json", s.HttpCall.Headers["Content-Type"])
	}
}

func TestSexprWorkflow_ReadFile(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "cfg"
    (read-file "config.json")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.ReadFile != "config.json" {
		t.Fatalf("expected read-file %q, got %q", "config.json", s.ReadFile)
	}
}

func TestSexprWorkflow_WriteFile(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "gen"
    (llm :prompt "generate"))
  (step "save"
    (write-file "output.json" :from "gen")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.WriteFile == nil {
		t.Fatal("expected WriteFile step")
	}
	if s.WriteFile.Path != "output.json" {
		t.Fatalf("expected path %q, got %q", "output.json", s.WriteFile.Path)
	}
	if s.WriteFile.From != "gen" {
		t.Fatalf("expected from %q, got %q", "gen", s.WriteFile.From)
	}
}

func TestSexprWorkflow_Glob(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "find"
    (glob "*.yaml" :dir "configs/")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.GlobPat == nil {
		t.Fatal("expected GlobPat step")
	}
	if s.GlobPat.Pattern != "*.yaml" {
		t.Fatalf("expected pattern %q, got %q", "*.yaml", s.GlobPat.Pattern)
	}
	if s.GlobPat.Dir != "configs/" {
		t.Fatalf("expected dir %q, got %q", "configs/", s.GlobPat.Dir)
	}
}

func TestSexprWorkflow_PluginCall(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "prs"
    (plugin "github" "prs" :since "yesterday" :authored)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.PluginCall == nil {
		t.Fatal("expected PluginCall step")
	}
	if s.PluginCall.Plugin != "github" {
		t.Fatalf("expected plugin %q, got %q", "github", s.PluginCall.Plugin)
	}
	if s.PluginCall.Subcommand != "prs" {
		t.Fatalf("expected subcommand %q, got %q", "prs", s.PluginCall.Subcommand)
	}
	if s.PluginCall.Args["since"] != "yesterday" {
		t.Fatalf("expected since=%q, got %q", "yesterday", s.PluginCall.Args["since"])
	}
	if s.PluginCall.Args["authored"] != "true" {
		t.Fatalf("expected authored=%q, got %q", "true", s.PluginCall.Args["authored"])
	}
}

func TestSexprWorkflow_PluginCallNamespaced(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "prs"
    (github/prs :since "yesterday" :authored)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.PluginCall == nil {
		t.Fatal("expected PluginCall step")
	}
	if s.PluginCall.Plugin != "github" {
		t.Fatalf("expected plugin %q, got %q", "github", s.PluginCall.Plugin)
	}
	if s.PluginCall.Subcommand != "prs" {
		t.Fatalf("expected subcommand %q, got %q", "prs", s.PluginCall.Subcommand)
	}
	if s.PluginCall.Args["since"] != "yesterday" {
		t.Fatalf("expected since=%q, got %q", "yesterday", s.PluginCall.Args["since"])
	}
	if s.PluginCall.Args["authored"] != "true" {
		t.Fatalf("expected authored=%q, got %q", "true", s.PluginCall.Args["authored"])
	}
}

func TestSexprWorkflow_Phase(t *testing.T) {
	src := []byte(`
(workflow "test"
  (phase "gather"
    (step "fetch" (run "echo data")))

  (phase "verify" :retries 2
    (step "process" (run "echo processed"))
    (gate "check" (run "test -f output.txt"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(w.Items))
	}

	p0 := w.Items[0].Phase
	if p0 == nil {
		t.Fatal("expected item 0 to be a phase")
	}
	if p0.ID != "gather" {
		t.Fatalf("expected phase id %q, got %q", "gather", p0.ID)
	}
	if p0.Retries != 0 {
		t.Fatalf("expected retries 0, got %d", p0.Retries)
	}
	if len(p0.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(p0.Steps))
	}
	if p0.Steps[0].ID != "fetch" {
		t.Fatalf("expected step id %q, got %q", "fetch", p0.Steps[0].ID)
	}
	if len(p0.Gates) != 0 {
		t.Fatalf("expected 0 gates, got %d", len(p0.Gates))
	}

	p1 := w.Items[1].Phase
	if p1 == nil {
		t.Fatal("expected item 1 to be a phase")
	}
	if p1.ID != "verify" {
		t.Fatalf("expected phase id %q, got %q", "verify", p1.ID)
	}
	if p1.Retries != 2 {
		t.Fatalf("expected retries 2, got %d", p1.Retries)
	}
	if len(p1.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(p1.Steps))
	}
	if len(p1.Gates) != 1 {
		t.Fatalf("expected 1 gate, got %d", len(p1.Gates))
	}
	if p1.Gates[0].ID != "check" {
		t.Fatalf("expected gate id %q, got %q", "check", p1.Gates[0].ID)
	}
	if !p1.Gates[0].IsGate {
		t.Fatal("expected gate to have IsGate=true")
	}
}

func TestSexprWorkflow_MixedStepsAndPhases(t *testing.T) {
	src := []byte(`
(workflow "mixed"
  (step "setup" (run "echo start"))
  (phase "core" :retries 1
    (step "work" (run "echo working"))
    (gate "check" (run "test -f done.txt")))
  (step "cleanup" (run "echo done")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(w.Items))
	}
	if w.Items[0].Step == nil || w.Items[0].Step.ID != "setup" {
		t.Fatal("expected item 0 to be bare step 'setup'")
	}
	if w.Items[1].Phase == nil || w.Items[1].Phase.ID != "core" {
		t.Fatal("expected item 1 to be phase 'core'")
	}
	if w.Items[2].Step == nil || w.Items[2].Step.ID != "cleanup" {
		t.Fatal("expected item 2 to be bare step 'cleanup'")
	}
}

func TestSexprWorkflow_GateOutsidePhase(t *testing.T) {
	src := []byte(`
(workflow "bad"
  (gate "orphan" (run "echo fail")))
`)
	_, err := parseSexprWorkflow(src)
	if err == nil {
		t.Fatal("expected error for gate outside phase")
	}
	if !strings.Contains(err.Error(), "must be inside") {
		t.Fatalf("expected 'must be inside' error, got: %v", err)
	}
}

func TestSexprWorkflow_PhaseWithLLMGate(t *testing.T) {
	src := []byte(`
(workflow "test"
  (phase "analyze" :retries 1
    (step "gen"
      (llm :prompt "generate something"))
    (gate "review"
      (llm :tier 2 :prompt "review: {{step \"gen\"}}"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	p := w.Items[0].Phase
	if p == nil {
		t.Fatal("expected phase")
	}
	if len(p.Gates) != 1 {
		t.Fatalf("expected 1 gate, got %d", len(p.Gates))
	}
	g := p.Gates[0]
	if g.LLM == nil {
		t.Fatal("expected LLM gate")
	}
	if g.LLM.Tier == nil || *g.LLM.Tier != 2 {
		t.Fatalf("expected tier 2, got %v", g.LLM.Tier)
	}
	if !g.IsGate {
		t.Fatal("expected IsGate=true")
	}
}

func TestSexprWorkflow_PhaseNoName(t *testing.T) {
	src := []byte(`
(workflow "test"
  (phase
    (step "s" (run "echo"))))
`)
	_, err := parseSexprWorkflow(src)
	if err == nil {
		t.Fatal("expected error for phase without name")
	}
}

func TestSexprWorkflow_Metadata(t *testing.T) {
	src := []byte(`
(workflow "pr-review"
  :description "Review PRs"
  :tags ("review" "ci" "code-quality")
  :author "adam"
  :version "1.0"
  :created "2026-04-01"
  (step "s1" (run "echo hi")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "pr-review" {
		t.Errorf("name: %q", w.Name)
	}
	if len(w.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(w.Tags))
	}
	if w.Tags[0] != "review" || w.Tags[1] != "ci" || w.Tags[2] != "code-quality" {
		t.Errorf("tags: %v", w.Tags)
	}
	if w.Author != "adam" {
		t.Errorf("author: %q", w.Author)
	}
	if w.Version != "1.0" {
		t.Errorf("version: %q", w.Version)
	}
	if w.Created != "2026-04-01" {
		t.Errorf("created: %q", w.Created)
	}
}

func TestSexprWorkflow_NoMetadata(t *testing.T) {
	src := []byte(`(workflow "bare" (step "s1" (run "echo hi")))`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Tags) != 0 {
		t.Errorf("expected no tags, got %v", w.Tags)
	}
	if w.Author != "" {
		t.Errorf("expected empty author, got %q", w.Author)
	}
}

func TestSexprWorkflow_Par(t *testing.T) {
	src := []byte(`
(workflow "test-par"
  (step "setup" (run "echo setup"))
  (par
    (step "a" (run "echo alpha"))
    (step "b" (run "echo bravo")))
  (step "final" (run "echo done")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(w.Items))
	}
	parItem := w.Items[1]
	if parItem.Step == nil {
		t.Fatal("expected par item to be a step")
	}
	if parItem.Step.Form != "par" {
		t.Fatalf("expected form %q, got %q", "par", parItem.Step.Form)
	}
	if len(parItem.Step.ParSteps) != 2 {
		t.Fatalf("expected 2 par steps, got %d", len(parItem.Step.ParSteps))
	}
	if parItem.Step.ParSteps[0].ID != "a" {
		t.Fatalf("expected par step 0 ID %q, got %q", "a", parItem.Step.ParSteps[0].ID)
	}
	if parItem.Step.ParSteps[1].ID != "b" {
		t.Fatalf("expected par step 1 ID %q, got %q", "b", parItem.Step.ParSteps[1].ID)
	}
}

func TestSexprWorkflow_ParSingleChild(t *testing.T) {
	src := []byte(`
(workflow "test"
  (par
    (step "a" (run "echo one"))))
`)
	_, err := parseSexprWorkflow(src)
	if err == nil {
		t.Fatal("expected error for par with single child")
	}
	if !strings.Contains(err.Error(), "at least 2") {
		t.Fatalf("expected 'at least 2' error, got: %v", err)
	}
}

func TestSexprWorkflow_ParWithRetry(t *testing.T) {
	src := []byte(`
(workflow "test"
  (par
    (retry 2 (step "a" (run "echo alpha")))
    (step "b" (run "echo bravo"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	parItem := w.Items[0]
	if parItem.Step.Form != "par" {
		t.Fatalf("expected par form, got %q", parItem.Step.Form)
	}
	if len(parItem.Step.ParSteps) != 2 {
		t.Fatalf("expected 2 par steps, got %d", len(parItem.Step.ParSteps))
	}
	if parItem.Step.ParSteps[0].Retry != 2 {
		t.Fatalf("expected retry 2, got %d", parItem.Step.ParSteps[0].Retry)
	}
}

func TestSexprWorkflow_ParInPhase(t *testing.T) {
	src := []byte(`
(workflow "test"
  (phase "verify" :retries 0
    (step "work" (run "echo work"))
    (par
      (gate "g1" (run "echo pass1"))
      (gate "g2" (run "echo pass2")))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 1 {
		t.Fatalf("expected 1 item (phase), got %d", len(w.Items))
	}
	p := w.Items[0].Phase
	if p == nil {
		t.Fatal("expected phase item")
	}
	if len(p.Steps) != 1 {
		t.Fatalf("expected 1 work step, got %d", len(p.Steps))
	}
	if len(p.Gates) != 1 {
		t.Fatalf("expected 1 gate (par wrapper), got %d", len(p.Gates))
	}
	parGate := p.Gates[0]
	if parGate.Form != "par" {
		t.Fatalf("expected gate form %q, got %q", "par", parGate.Form)
	}
	if len(parGate.ParSteps) != 2 {
		t.Fatalf("expected 2 par gates, got %d", len(parGate.ParSteps))
	}
	if !parGate.ParSteps[0].IsGate {
		t.Fatal("expected par child 0 to be a gate")
	}
	if !parGate.ParSteps[1].IsGate {
		t.Fatal("expected par child 1 to be a gate")
	}
}

func TestSexprWorkflow_AliasEach(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "list"
    (run "echo -e 'a\nb\nc'"))
  (each "list"
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
}

func TestSexprWorkflow_AliasPick(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (run "curl http://example.com/api"))
  (step "s2"
    (pick ".a" :from "s1")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.JsonPick == nil {
		t.Fatal("expected JsonPick step")
	}
	if s.JsonPick.Expr != ".a" {
		t.Fatalf("expected expr %q, got %q", ".a", s.JsonPick.Expr)
	}
	if s.JsonPick.From != "s1" {
		t.Fatalf("expected from %q, got %q", "s1", s.JsonPick.From)
	}
}

func TestSexprWorkflow_AliasFetch(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (fetch "http://example.com")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected HttpCall step")
	}
	if s.HttpCall.Method != "GET" {
		t.Fatalf("expected method %q, got %q", "GET", s.HttpCall.Method)
	}
	if s.HttpCall.URL != "http://example.com" {
		t.Fatalf("expected URL %q, got %q", "http://example.com", s.HttpCall.URL)
	}
}

func TestSexprWorkflow_AliasSend(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (send "http://example.com" :body "{\"key\":\"val\"}")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.HttpCall == nil {
		t.Fatal("expected HttpCall step")
	}
	if s.HttpCall.Method != "POST" {
		t.Fatalf("expected method %q, got %q", "POST", s.HttpCall.Method)
	}
	if s.HttpCall.URL != "http://example.com" {
		t.Fatalf("expected URL %q, got %q", "http://example.com", s.HttpCall.URL)
	}
}

func TestSexprWorkflow_AliasRead(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (read "path/to/file.txt")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.ReadFile != "path/to/file.txt" {
		t.Fatalf("expected read-file %q, got %q", "path/to/file.txt", s.ReadFile)
	}
}

func TestSexprWorkflow_AliasWrite(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "gen"
    (llm :prompt "generate"))
  (step "s1"
    (write "out.txt" :from "gen")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[1]
	if s.WriteFile == nil {
		t.Fatal("expected WriteFile step")
	}
	if s.WriteFile.Path != "out.txt" {
		t.Fatalf("expected path %q, got %q", "out.txt", s.WriteFile.Path)
	}
	if s.WriteFile.From != "gen" {
		t.Fatalf("expected from %q, got %q", "gen", s.WriteFile.From)
	}
}

func TestSexprWorkflow_NoPhases_ItemsPopulated(t *testing.T) {
	src := []byte(`
(workflow "old-style"
  (step "a" (run "echo a"))
  (step "b" (run "echo b")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Items) != 2 {
		t.Fatalf("expected 2 items for bare-step workflow, got %d", len(w.Items))
	}
	if w.Items[0].Step == nil || w.Items[0].Step.ID != "a" {
		t.Fatal("expected item 0 to be step 'a'")
	}
	if w.Items[1].Step == nil || w.Items[1].Step.ID != "b" {
		t.Fatal("expected item 1 to be step 'b'")
	}
}

func TestSexprWorkflow_Search(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "q" (search :index "my-index" :query {"term" {"type" "doc"}} :size 50 :fields ("title" "content"))))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if s.Search.IndexName != "my-index" {
		t.Fatalf("index = %q, want my-index", s.Search.IndexName)
	}
	if s.Search.Size != 50 {
		t.Fatalf("size = %d, want 50", s.Search.Size)
	}
	if len(s.Search.Fields) != 2 || s.Search.Fields[0] != "title" || s.Search.Fields[1] != "content" {
		t.Fatalf("fields = %v, want [title content]", s.Search.Fields)
	}
}

func TestSexprWorkflow_SearchDefaultSize(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "q" (search :index "my-index" :query {"term" {"type" "doc"}})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].Search.Size != 10 {
		t.Fatalf("default size = %d, want 10", w.Steps[0].Search.Size)
	}
}

func TestSexprWorkflow_SearchWithES(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "q" (search :index "my-index" :query {"match_all" {}} :es "http://remote:9200")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	if w.Steps[0].Search.ESURL != "http://remote:9200" {
		t.Fatalf("es url = %q, want http://remote:9200", w.Steps[0].Search.ESURL)
	}
}

func TestSexprWorkflow_Index(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "idx" (index :index "my-index" :doc "{\"title\":\"hello\"}" :id "doc-1")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Index == nil {
		t.Fatal("expected index step")
	}
	if s.Index.IndexName != "my-index" {
		t.Fatalf("index = %q", s.Index.IndexName)
	}
	if s.Index.DocID != "doc-1" {
		t.Fatalf("doc id = %q", s.Index.DocID)
	}
}

func TestSexprWorkflow_IndexWithEmbed(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "idx" (index :index "my-index" :doc "{}" :embed :field "content" :model "nomic-embed-text")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Index.EmbedField != "content" {
		t.Fatalf("embed field = %q, want content", s.Index.EmbedField)
	}
	if s.Index.EmbedModel != "nomic-embed-text" {
		t.Fatalf("embed model = %q, want nomic-embed-text", s.Index.EmbedModel)
	}
}

func TestSexprWorkflow_Delete(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "del" (delete :index "my-index" :query {"term" {"type" "old"}})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Delete == nil {
		t.Fatal("expected delete step")
	}
	if s.Delete.IndexName != "my-index" {
		t.Fatalf("index = %q", s.Delete.IndexName)
	}
}

func TestSexprWorkflow_Embed(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "vec" (embed :input "hello world" :model "nomic-embed-text")))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatal(err)
	}
	s := w.Steps[0]
	if s.Embed == nil {
		t.Fatal("expected embed step")
	}
	if s.Embed.Input != "hello world" {
		t.Fatalf("input = %q", s.Embed.Input)
	}
	if s.Embed.Model != "nomic-embed-text" {
		t.Fatalf("model = %q", s.Embed.Model)
	}
}

func TestSexprWorkflow_SearchSort(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (search :index "my-index" :size 10 :sort {"indexed_at" "desc"})))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if s.Search.Sort == "" {
		t.Fatal("expected sort to be set")
	}
}

func TestSexprWorkflow_SearchNDJSON(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (search :index "my-index" :ndjson)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if !s.Search.NDJSON {
		t.Fatal("expected ndjson to be true")
	}
}

func TestSexprWorkflow_SearchQueryOptional(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (search :index "my-index" :size 5)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(w.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(w.Steps))
	}
	s := w.Steps[0]
	if s.Search == nil {
		t.Fatal("expected search step")
	}
	if s.Search.IndexName != "my-index" {
		t.Fatalf("expected index %q, got %q", "my-index", s.Search.IndexName)
	}
	if s.Search.Size != 5 {
		t.Fatalf("expected size 5, got %d", s.Search.Size)
	}
	if s.Search.Query != "" {
		t.Fatalf("expected empty query, got %q", s.Search.Query)
	}
}

func TestSexprWorkflow_IndexUpsertFalse(t *testing.T) {
	src := []byte(`
(workflow "test"
  (step "s1"
    (index :index "my-index" :doc "{}" :id "doc1" :upsert false)))
`)
	w, err := parseSexprWorkflow(src)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	s := w.Steps[0]
	if s.Index == nil {
		t.Fatal("expected index step")
	}
	if s.Index.Upsert == nil || *s.Index.Upsert {
		t.Fatal("expected upsert to be false")
	}
}
