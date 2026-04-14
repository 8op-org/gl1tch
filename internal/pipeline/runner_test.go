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

func TestParseCrossReview_Empty(t *testing.T) {
	scores := ParseCrossReview("no structured output here")
	if len(scores) != 0 {
		t.Fatalf("expected 0 variants, got %d", len(scores))
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
