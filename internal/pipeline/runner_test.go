package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	result, err := render(`gh issue view ~param.issue --repo ~param.repo`, scopeFromData(data), steps)
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
	result, err := render(`Issue: ~(step fetch)`, scopeFromData(data), steps)
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
				Run: `echo "issue=~param.issue repo=~param.repo"`,
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

func TestRun_SmartRouting_NoProvider(t *testing.T) {
	callLog := []string{}
	resolver := func(name string) (provider.ProviderFunc, bool) {
		return func(model, prompt string) (provider.LLMResult, error) {
			callLog = append(callLog, name)
			return provider.LLMResult{
				Provider: name,
				Model:    model,
				Response: "smart-routed response",
				TokensIn: 10, TokensOut: 5,
			}, nil
		}, true
	}

	w := &Workflow{
		Name: "test-smart",
		Steps: []Step{
			{
				ID: "classify",
				LLM: &LLMStep{
					Prompt: "classify this issue",
				},
			},
		},
	}

	tiers := []provider.TierConfig{
		{Providers: []string{"fake-local"}, Model: "local-model"},
		{Providers: []string{"fake-cloud"}, Model: "cloud-model"},
	}

	reg, _ := provider.LoadProviders(t.TempDir())
	result, err := Run(w, "", "qwen3:8b", nil, reg, RunOpts{
		ProviderResolver: resolver,
		Tiers:            tiers,
		EvalThreshold:    4,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Output != "smart-routed response" {
		t.Errorf("output = %q, want smart-routed response", result.Output)
	}
	if len(callLog) < 1 {
		t.Fatal("expected at least one provider call")
	}
}

func TestParseCrossReview(t *testing.T) {
	output := `--- LOCAL ---
1. Specificity — PASS — good paths
2. Completeness — FAIL — missing IDE
3. Feasibility — PASS — clear steps
4. Testing — PASS — has tests
5. PR Quality — PASS — clean
SCORE: 4/5

--- CLAUDE ---
1. Specificity — PASS — real paths
2. Completeness — PASS — all covered
3. Feasibility — PASS — detailed
4. Testing — PASS — thorough
5. PR Quality — PASS — excellent
SCORE: 5/5

--- COPILOT ---
1. Specificity — PASS — searched repo
2. Completeness — PASS — complete
3. Feasibility — PASS — actionable
4. Testing — FAIL — weak
5. PR Quality — PASS — good
SCORE: 4/5

WINNER: CLAUDE
OVERALL: PASS`

	scores := ParseCrossReview(output)
	if len(scores) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(scores))
	}

	// Check LOCAL
	if scores[0].Variant != "local" || scores[0].Passed != 4 || scores[0].Total != 5 {
		t.Fatalf("local: expected 4/5, got %d/%d", scores[0].Passed, scores[0].Total)
	}
	if scores[0].Winner {
		t.Fatal("local should not be winner")
	}

	// Check CLAUDE
	if scores[1].Variant != "claude" || scores[1].Passed != 5 || scores[1].Total != 5 {
		t.Fatalf("claude: expected 5/5, got %d/%d", scores[1].Passed, scores[1].Total)
	}
	if !scores[1].Winner {
		t.Fatal("claude should be winner")
	}

	// Check COPILOT
	if scores[2].Variant != "copilot" || scores[2].Passed != 4 || scores[2].Total != 5 {
		t.Fatalf("copilot: expected 4/5, got %d/%d", scores[2].Passed, scores[2].Total)
	}
}

func TestParseCrossReview_TwoVariants(t *testing.T) {
	output := `--- LOCAL ---
1. Specificity — PASS — ok
2. Completeness — PASS — ok
SCORE: 2/2

--- CLAUDE ---
1. Specificity — FAIL — vague
2. Completeness — PASS — ok
SCORE: 1/2

WINNER: LOCAL
OVERALL: PASS`

	scores := ParseCrossReview(output)
	if len(scores) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(scores))
	}
	if scores[0].Variant != "local" || scores[0].Passed != 2 || scores[0].Total != 2 {
		t.Fatalf("local: expected 2/2, got %d/%d", scores[0].Passed, scores[0].Total)
	}
	if !scores[0].Winner {
		t.Fatal("local should be winner")
	}
	if scores[1].Winner {
		t.Fatal("claude should not be winner")
	}
}

// --- SDK form unit tests ---

func TestRunSingleStep_Lines(t *testing.T) {
	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  map[string]string{"list": "alpha\nbeta\ngamma"},
		params: map[string]string{},
	}
	step := Step{ID: "split", Lines: "list"}
	out, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	expected := `["alpha","beta","gamma"]`
	if out.output != expected {
		t.Fatalf("expected %q, got %q", expected, out.output)
	}
}

func TestRunSingleStep_Merge(t *testing.T) {
	rctx := &runCtx{
		ctx:    context.Background(),
		steps: map[string]string{
			"a": `{"x":1}`,
			"b": `{"y":2}`,
		},
		params: map[string]string{},
	}
	step := Step{ID: "combined", Merge: []string{"a", "b"}}
	out, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.output, `"x"`) || !strings.Contains(out.output, `"y"`) {
		t.Fatalf("expected merged JSON with x and y, got %q", out.output)
	}
}

func TestRunSingleStep_ReadFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "hello.txt")
	if err := os.WriteFile(tmp, []byte("file contents here"), 0o644); err != nil {
		t.Fatal(err)
	}
	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  map[string]string{},
		params: map[string]string{},
	}
	step := Step{ID: "read", ReadFile: tmp}
	out, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	if out.output != "file contents here" {
		t.Fatalf("expected %q, got %q", "file contents here", out.output)
	}
}

func TestRunSingleStep_WriteFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "sub", "out.txt")
	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  map[string]string{"gen": "hello world"},
		params: map[string]string{},
	}
	step := Step{
		ID:        "write",
		WriteFile: &WriteFileStep{Path: tmp, From: "gen"},
	}
	out, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	if out.output != tmp {
		t.Fatalf("expected path %q, got %q", tmp, out.output)
	}
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Fatalf("file contents = %q, want %q", string(data), "hello world")
	}
}

func TestRunSingleStep_Glob(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.yaml", "b.yaml", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  map[string]string{},
		params: map[string]string{},
	}
	step := Step{
		ID:      "find",
		GlobPat: &GlobStep{Pattern: "*.yaml", Dir: dir},
	}
	out, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	matches := strings.Split(strings.TrimSpace(out.output), "\n")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d: %v", len(matches), matches)
	}
}

func TestRunSingleStep_JsonPick(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH")
	}
	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  map[string]string{"data": `{"name":"alice","age":30}`},
		params: map[string]string{},
	}
	step := Step{
		ID:       "pick",
		JsonPick: &JsonPickStep{Expr: ".name", From: "data"},
	}
	out, err := runSingleStep(context.Background(), rctx, step)
	if err != nil {
		t.Fatal(err)
	}
	if out.output != `"alice"` {
		t.Fatalf("expected %q, got %q", `"alice"`, out.output)
	}
}

func TestParseCrossReview_Empty(t *testing.T) {
	scores := ParseCrossReview("no structured output here")
	if len(scores) != 0 {
		t.Fatalf("expected 0 variants, got %d", len(scores))
	}
}

func TestParseCrossReview_NumericFormat(t *testing.T) {
	output := `VARIANT: local
plan_completeness: 9/10
plan_specificity: 9/10
pr_quality: 9/10
review_accuracy: 9/10
total: 36/40
notes: Structured plan with clear implementation steps

VARIANT: copilot
plan_completeness: 8/10
plan_specificity: 9/10
pr_quality: 9/10
review_accuracy: 9/10
total: 35/40
notes: Detailed self-review

WINNER: local
REASON: The local variant provides a more structured plan`

	scores := ParseCrossReview(output)
	if len(scores) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(scores))
	}

	// Check local: all scores >= 7, so passed=4, total=4
	if scores[0].Variant != "local" {
		t.Fatalf("expected variant 'local', got %q", scores[0].Variant)
	}
	if scores[0].Passed != 4 || scores[0].Total != 4 {
		t.Fatalf("local: expected 4/4, got %d/%d", scores[0].Passed, scores[0].Total)
	}
	if !scores[0].Winner {
		t.Fatal("local should be winner")
	}

	// Check copilot
	if scores[1].Variant != "copilot" {
		t.Fatalf("expected variant 'copilot', got %q", scores[1].Variant)
	}
	if scores[1].Passed != 4 || scores[1].Total != 4 {
		t.Fatalf("copilot: expected 4/4, got %d/%d", scores[1].Passed, scores[1].Total)
	}
	if scores[1].Winner {
		t.Fatal("copilot should not be winner")
	}
}

func TestParseCrossReview_NumericWithLowScores(t *testing.T) {
	output := `VARIANT: local
plan_completeness: 9/10
plan_specificity: 6/10
pr_quality: 5/10
review_accuracy: 8/10
total: 28/40
notes: Mixed results

VARIANT: claude
plan_completeness: 10/10
plan_specificity: 9/10
pr_quality: 9/10
review_accuracy: 9/10
total: 37/40
notes: Strong across the board

WINNER: claude
REASON: Higher overall quality`

	scores := ParseCrossReview(output)
	if len(scores) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(scores))
	}

	// local: 9 >= 7 (pass), 6 < 7 (fail), 5 < 7 (fail), 8 >= 7 (pass) = 2/4
	if scores[0].Variant != "local" || scores[0].Passed != 2 || scores[0].Total != 4 {
		t.Fatalf("local: expected 2/4, got %d/%d", scores[0].Passed, scores[0].Total)
	}
	if scores[0].Winner {
		t.Fatal("local should not be winner")
	}

	// claude: all >= 7 = 4/4
	if scores[1].Variant != "claude" || scores[1].Passed != 4 || scores[1].Total != 4 {
		t.Fatalf("claude: expected 4/4, got %d/%d", scores[1].Passed, scores[1].Total)
	}
	if !scores[1].Winner {
		t.Fatal("claude should be winner")
	}
}

func TestEvaluateGate_ShellPass(t *testing.T) {
	outcome := &stepOutcome{output: "all good", isLLM: false}
	pass, detail := evaluateGate(Step{ID: "check", Run: "true", IsGate: true}, outcome, nil)
	if !pass {
		t.Fatalf("expected shell gate to pass, got fail: %s", detail)
	}
}

func TestEvaluateGate_LLMPass(t *testing.T) {
	outcome := &stepOutcome{output: "Criterion 1: PASS\nCriterion 2: PASS\nOVERALL: PASS", isLLM: true}
	pass, _ := evaluateGate(Step{ID: "review", IsGate: true, LLM: &LLMStep{}}, outcome, nil)
	if !pass {
		t.Fatal("expected LLM gate with OVERALL: PASS to pass")
	}
}

func TestEvaluateGate_LLMFail(t *testing.T) {
	outcome := &stepOutcome{output: "Criterion 1: PASS\nCriterion 2: FAIL\nOVERALL: FAIL", isLLM: true}
	pass, _ := evaluateGate(Step{ID: "review", IsGate: true, LLM: &LLMStep{}}, outcome, nil)
	if pass {
		t.Fatal("expected LLM gate with OVERALL: FAIL to fail")
	}
}

func TestExecutePhase_AllGatesPass(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")

	p := Phase{
		ID:      "test-phase",
		Retries: 0,
		Steps: []Step{
			{ID: "create", Run: fmt.Sprintf("touch %s", marker)},
		},
		Gates: []Step{
			{ID: "verify", Run: fmt.Sprintf("test -f %s", marker), IsGate: true},
		},
	}

	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  make(map[string]string),
		params: map[string]string{},
	}

	report, err := executePhase(rctx, p)
	if err != nil {
		t.Fatalf("expected phase to pass, got error: %v", err)
	}
	if report != nil {
		t.Fatalf("expected nil report on success, got: %v", report)
	}
}

func TestExecutePhase_GateFailsNoRetry(t *testing.T) {
	p := Phase{
		ID:      "test-phase",
		Retries: 0,
		Steps: []Step{
			{ID: "work", Run: "echo done"},
		},
		Gates: []Step{
			{ID: "check", Run: "false", IsGate: true},
		},
	}

	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  make(map[string]string),
		params: map[string]string{},
	}

	report, err := executePhase(rctx, p)
	if err == nil {
		t.Fatal("expected error when gate fails with no retries")
	}
	if report == nil {
		t.Fatal("expected verification report")
	}
	if report.Phase != "test-phase" {
		t.Fatalf("expected phase %q, got %q", "test-phase", report.Phase)
	}
	if report.Gates[0].Passed {
		t.Fatal("expected gate 'check' to fail")
	}
}

func TestExecutePhase_GateFailsThenPassesOnRetry(t *testing.T) {
	dir := t.TempDir()
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	counter := filepath.Join(dir, "counter.txt")

	stepCmd := fmt.Sprintf("if [ ! -f %s ]; then touch %s && touch %s; else touch %s; fi", counter, counter, file1, file2)
	gateCmd := fmt.Sprintf("test -f %s && test -f %s", file1, file2)

	p := Phase{
		ID:      "retry-phase",
		Retries: 1,
		Steps: []Step{
			{ID: "create", Run: stepCmd},
		},
		Gates: []Step{
			{ID: "verify", Run: gateCmd, IsGate: true},
		},
	}

	rctx := &runCtx{
		ctx:    context.Background(),
		steps:  make(map[string]string),
		params: map[string]string{},
	}

	report, err := executePhase(rctx, p)
	if err != nil {
		t.Fatalf("expected phase to pass on retry, got error: %v", err)
	}
	if report != nil {
		t.Fatalf("expected nil report on success, got: %v", report)
	}
}

func TestRun_WithPhases(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "output.txt")

	w := &Workflow{
		Name: "phase-test",
		Items: []WorkflowItem{
			{Phase: &Phase{
				ID: "gather",
				Steps: []Step{
					{ID: "fetch", Run: fmt.Sprintf("echo hello > %s && echo hello", outFile)},
				},
			}},
			{Phase: &Phase{
				ID: "verify",
				Steps: []Step{
					{ID: "process", Run: "echo processed"},
				},
				Gates: []Step{
					{ID: "check", Run: fmt.Sprintf("test -f %s", outFile), IsGate: true},
				},
			}},
		},
	}

	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Steps["fetch"] == "" {
		t.Fatal("expected fetch step output")
	}
	if result.Steps["process"] == "" {
		t.Fatal("expected process step output")
	}
}

func TestRun_WithPhases_GateFails(t *testing.T) {
	w := &Workflow{
		Name: "fail-test",
		Items: []WorkflowItem{
			{Phase: &Phase{
				ID: "work",
				Steps: []Step{
					{ID: "do-thing", Run: "echo working"},
				},
				Gates: []Step{
					{ID: "check", Run: "false", IsGate: true},
				},
			}},
		},
	}

	_, err := Run(w, "", "", nil, nil)
	if err == nil {
		t.Fatal("expected error when gate fails")
	}
	if !strings.Contains(err.Error(), "exhausted") {
		t.Fatalf("expected exhaustion error, got: %v", err)
	}
}

func TestRun_EndToEnd_PhaseGateWorkflow(t *testing.T) {
	w, err := LoadFile("testdata/phase-gate.glitch")
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "phase-gate-test" {
		t.Fatalf("expected name %q, got %q", "phase-gate-test", w.Name)
	}

	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Steps["data"], "hello world") {
		t.Fatalf("expected data to contain 'hello world', got %q", result.Steps["data"])
	}
	if !strings.Contains(result.Steps["transform"], "TRANSFORMED") {
		t.Fatalf("expected transform output, got %q", result.Steps["transform"])
	}
	if strings.TrimSpace(result.Steps["done"]) != "finished" {
		t.Fatalf("expected done='finished', got %q", result.Steps["done"])
	}
}

func TestRun_BareStepsStillWork(t *testing.T) {
	w := &Workflow{
		Name: "bare-test",
		Steps: []Step{
			{ID: "a", Run: "echo alpha"},
			{ID: "b", Run: "echo beta"},
		},
	}

	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(result.Steps["a"]) != "alpha" {
		t.Fatalf("expected 'alpha', got %q", result.Steps["a"])
	}
	if strings.TrimSpace(result.Steps["b"]) != "beta" {
		t.Fatalf("expected 'beta', got %q", result.Steps["b"])
	}
}

func TestRun_PinnedTier(t *testing.T) {
	callLog := []string{}
	resolver := func(name string) (provider.ProviderFunc, bool) {
		return func(model, prompt string) (provider.LLMResult, error) {
			callLog = append(callLog, name)
			return provider.LLMResult{
				Provider: name,
				Model:    model,
				Response: "tier-2 response",
				TokensIn: 10, TokensOut: 5,
			}, nil
		}, true
	}

	tier := 2
	w := &Workflow{
		Name: "test-pinned",
		Steps: []Step{
			{
				ID: "analyze",
				LLM: &LLMStep{
					Tier:   &tier,
					Prompt: "deep analysis",
				},
			},
		},
	}

	tiers := []provider.TierConfig{
		{Providers: []string{"fake-local"}, Model: "local-model"},
		{Providers: []string{"fake-cloud"}, Model: "cloud-model"},
		{Providers: []string{"fake-premium"}, Model: "premium-model"},
	}

	reg, _ := provider.LoadProviders(t.TempDir())
	result, err := Run(w, "", "qwen3:8b", nil, reg, RunOpts{
		ProviderResolver: resolver,
		Tiers:            tiers,
		EvalThreshold:    4,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Output != "tier-2 response" {
		t.Errorf("output = %q, want tier-2 response", result.Output)
	}
	if len(callLog) != 1 || callLog[0] != "fake-premium" {
		t.Errorf("callLog = %v, want [fake-premium]", callLog)
	}
}

func TestRun_Par_Basic(t *testing.T) {
	src := []byte(`
(workflow "test-par"
  (step "setup" (run "echo setup-value"))
  (par
    (step "a" (run "echo alpha"))
    (step "b" (run "echo bravo")))
  (step "final" (run "echo a=~(step a) b=~(step b)")))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	reg, _ := provider.LoadProviders(t.TempDir())
	result, err := Run(w, "", "", nil, reg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Steps["a"], "alpha") {
		t.Fatalf("expected step a to contain 'alpha', got %q", result.Steps["a"])
	}
	if !strings.Contains(result.Steps["b"], "bravo") {
		t.Fatalf("expected step b to contain 'bravo', got %q", result.Steps["b"])
	}
	if !strings.Contains(result.Output, "a=alpha") {
		t.Fatalf("expected final output to contain 'a=alpha', got %q", result.Output)
	}
	if !strings.Contains(result.Output, "b=bravo") {
		t.Fatalf("expected final output to contain 'b=bravo', got %q", result.Output)
	}
}

func TestRun_Par_FailFast(t *testing.T) {
	src := []byte(`
(workflow "test-par-fail"
  (par
    (step "ok" (run "echo fine"))
    (step "bad" (run "exit 1"))))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	reg, _ := provider.LoadProviders(t.TempDir())
	_, err = Run(w, "", "", nil, reg)
	if err == nil {
		t.Fatal("expected error from failing par step")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Fatalf("expected error to mention step 'bad', got: %v", err)
	}
}

func TestRun_Par_WithTimeout(t *testing.T) {
	src := []byte(`
(workflow "test-par-timeout"
  (timeout "1s"
    (par
      (step "fast" (run "echo quick"))
      (step "slow" (run "sleep 10")))))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	reg, _ := provider.LoadProviders(t.TempDir())
	_, err = Run(w, "", "", nil, reg)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRun_Par_InPhase(t *testing.T) {
	src := []byte(`
(workflow "test-par-phase"
  (phase "check" :retries 0
    (step "work" (run "echo done"))
    (par
      (gate "g1" (run "echo PASS"))
      (gate "g2" (run "echo PASS")))))
`)
	w, err := LoadBytes(src, "test.glitch")
	if err != nil {
		t.Fatal(err)
	}
	reg, _ := provider.LoadProviders(t.TempDir())
	result, err := Run(w, "", "", nil, reg)
	if err != nil {
		t.Fatal(err)
	}
	if result.Steps["g1"] == "" {
		t.Fatal("expected gate g1 to have output")
	}
	if result.Steps["g2"] == "" {
		t.Fatal("expected gate g2 to have output")
	}
}

func TestRender_Pick(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"subject":"help me","from":"alice@example.com"}`,
		},
	}

	result, err := render(`~(pick "subject" param.item)`, scopeFromData(data), steps)
	if err != nil {
		t.Fatal(err)
	}
	if result != "help me" {
		t.Fatalf("expected %q, got %q", "help me", result)
	}
}

func TestRender_PickNested(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"email":{"subject":"nested"}}`,
		},
	}

	result, err := render(`~(pick "email.subject" param.item)`, scopeFromData(data), steps)
	if err != nil {
		t.Fatal(err)
	}
	if result != "nested" {
		t.Fatalf("expected %q, got %q", "nested", result)
	}
}

func TestRender_Assoc(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"subject":"help me","from":"alice@example.com"}`,
		},
	}

	result, err := render(`~(assoc "status" "triaged" param.item)`, scopeFromData(data), steps)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if obj["status"] != "triaged" {
		t.Fatalf("expected status %q, got %v", "triaged", obj["status"])
	}
	if obj["subject"] != "help me" {
		t.Fatalf("expected subject preserved, got %v", obj["subject"])
	}
}

func TestRender_AssocOverwrite(t *testing.T) {
	steps := map[string]string{}
	data := map[string]any{
		"param": map[string]string{
			"item": `{"status":"new"}`,
		},
	}

	result, err := render(`~(assoc "status" "closed" param.item)`, scopeFromData(data), steps)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if obj["status"] != "closed" {
		t.Fatalf("expected status %q, got %v", "closed", obj["status"])
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		err   bool
	}{
		{
			name:  "clean json",
			input: `{"category": "billing"}`,
			want:  `{"category": "billing"}`,
		},
		{
			name:  "with think tags",
			input: "<think>\nlet me analyze this\n</think>\n{\"category\": \"billing\"}",
			want:  `{"category": "billing"}`,
		},
		{
			name:  "with markdown fences",
			input: "```json\n{\"category\": \"billing\"}\n```",
			want:  `{"category": "billing"}`,
		},
		{
			name:  "with think and fences",
			input: "<think>\nhmm\n</think>\nHere is the result:\n```json\n{\"category\": \"billing\"}\n```\nDone.",
			want:  `{"category": "billing"}`,
		},
		{
			name:  "no json",
			input: "I cannot help with that",
			err:   true,
		},
		{
			name:  "json array",
			input: "[1,2,3]",
			want:  "[1,2,3]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractJSON(tt.input)
			if tt.err {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestRun_Par_FromFile(t *testing.T) {
	w, err := LoadFile("testdata/par-demo.glitch")
	if err != nil {
		t.Fatal(err)
	}
	reg, _ := provider.LoadProviders(t.TempDir())
	result, err := Run(w, "", "", nil, reg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Output, "left=left-branch") {
		t.Fatalf("expected merged output, got %q", result.Output)
	}
	if !strings.Contains(result.Output, "right=right-branch") {
		t.Fatalf("expected merged output, got %q", result.Output)
	}
}

func TestRun_Compare_Basic(t *testing.T) {
	w, err := LoadFile("testdata/compare-basic.glitch")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Output == "" {
		t.Error("expected non-empty output")
	}
	if result.Steps["setup"] != "shared-data" {
		t.Errorf("setup = %q, want shared-data", result.Steps["setup"])
	}
	if result.Steps["pick"] == "" {
		t.Error("pick step should have winner output")
	}
}

func TestRun_Compare_MultiStep(t *testing.T) {
	w, err := LoadFile("testdata/compare-multi.glitch")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Steps["impl"] == "" {
		t.Error("impl step should have winner output")
	}
}

func TestRun_Compare_BranchFailure(t *testing.T) {
	src := `(workflow "fail-test"
	  :description "one branch fails"
	  (step "pick"
	    (compare
	      :objective "test branch failure handling"
	      (branch "good" (run "echo works"))
	      (branch "bad" (run "exit 1")))))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v (should succeed with one good branch)", err)
	}
	if result.Steps["pick"] != "works" {
		t.Errorf("pick = %q, want 'works'", result.Steps["pick"])
	}
}

func TestRun_Compare_AllBranchesFail(t *testing.T) {
	src := `(workflow "all-fail"
	  :description "both fail"
	  (step "pick"
	    (compare
	      :objective "test all branches failing"
	      (branch "a" (run "exit 1"))
	      (branch "b" (run "exit 2")))))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, err = Run(w, "", "", nil, nil)
	if err == nil {
		t.Fatal("expected error when all branches fail")
	}
}

func TestRun_Compare_StepAccessVariant(t *testing.T) {
	src := `(workflow "variant-access"
	  :description "access specific variant output"
	  (compare
	    :id "impl"
	    :objective "test variant access"
	    (branch "fast"
	      (step "out" (run "echo fast-output")))
	    (branch "slow"
	      (step "out" (run "echo slow-output"))))
	  (step "check" (run "echo winner=~(step impl)")))
	`
	w, err := parseSexprWorkflow([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result, err := Run(w, "", "", nil, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	check := result.Steps["check"]
	if !strings.Contains(check, "winner=") {
		t.Errorf("check = %q, expected winner= prefix", check)
	}
	if check != "winner=fast-output" && check != "winner=slow-output" {
		t.Errorf("check = %q, expected one of the branch outputs", check)
	}
}
