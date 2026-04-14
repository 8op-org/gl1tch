# Smart Provider Routing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add cost-conscious auto-routing to gl1tch's TieredRunner — LLM steps without an explicit provider/tier try local first, self-evaluate, and escalate through cheap→premium tiers only when quality is low.

**Architecture:** Extend the existing `TieredRunner.Run()` with a new `RunSmart()` method that wraps each tier attempt with structural validation and LLM self-evaluation. The pipeline runner dispatches to `RunSmart()` when no provider or tier is pinned. New telemetry fields on `LLMCallDoc` track escalation chains. Three new Kibana Vega visualizations are added to the embedded dashboard.

**Tech Stack:** Go, Ollama HTTP API, existing `provider` and `pipeline` packages, Kibana Vega-Lite NDJSON

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/provider/eval.go` | Self-eval prompt construction + score parsing |
| Create | `internal/provider/eval_test.go` | Unit tests for eval logic |
| Create | `internal/provider/structural.go` | JSON/YAML/refusal structural checks |
| Create | `internal/provider/structural_test.go` | Unit tests for structural checks |
| Modify | `internal/provider/tiers.go` | Add `RunSmart()` with eval+escalation loop |
| Modify | `internal/provider/tiers_test.go` | Tests for `RunSmart()` escalation behavior |
| Modify | `internal/provider/tokens.go` | Add escalation fields to `LLMResult` and `RunResult` |
| Modify | `internal/pipeline/types.go` | Add `Tier` and `Format` fields to `LLMStep` |
| Modify | `internal/pipeline/sexpr.go` | Parse `:tier` and `:format` keywords in LLM steps |
| Modify | `internal/pipeline/sexpr_test.go` | Test sexpr parsing of new keywords |
| Modify | `internal/pipeline/runner.go` | Wire `RunSmart()` for auto-routing path |
| Modify | `internal/pipeline/runner_test.go` | Test auto-routing integration |
| Modify | `internal/esearch/telemetry.go` | Add escalation fields to `LLMCallDoc` |
| Modify | `cmd/config.go` | Add `EvalThreshold` to `Config` |
| Modify | `internal/dashboard/default.ndjson` | Add 3 escalation visualizations |

---

### Task 1: Structural Validation

**Files:**
- Create: `internal/provider/structural.go`
- Create: `internal/provider/structural_test.go`

- [ ] **Step 1: Write failing tests for structural validation**

```go
// internal/provider/structural_test.go
package provider

import "testing"

func TestCheckStructure_JSON(t *testing.T) {
	tests := []struct {
		name   string
		format string
		input  string
		ok     bool
	}{
		{"valid json", "json", `{"key": "value"}`, true},
		{"invalid json", "json", `not json at all`, false},
		{"empty json", "json", "", false},
		{"json array", "json", `[1, 2, 3]`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckStructure(tt.format, tt.input); got != tt.ok {
				t.Errorf("CheckStructure(%q, %q) = %v, want %v", tt.format, tt.input, got, tt.ok)
			}
		})
	}
}

func TestCheckStructure_YAML(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		ok     bool
	}{
		{"valid yaml", "key: value\nlist:\n  - a\n  - b", true},
		{"empty", "", false},
		{"bare string", "just a string", true}, // valid YAML scalar
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckStructure("yaml", tt.input); got != tt.ok {
				t.Errorf("CheckStructure(yaml, %q) = %v, want %v", tt.input, got, tt.ok)
			}
		})
	}
}

func TestCheckStructure_NoFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ok    bool
	}{
		{"normal text", "Here is my analysis of the issue.", true},
		{"empty", "", false},
		{"whitespace only", "   \n  ", false},
		{"refusal I cannot", "I cannot help with that request.", false},
		{"refusal I'm sorry", "I'm sorry, I can't assist with that.", false},
		{"refusal as AI", "As an AI language model, I cannot", false},
		{"contains cannot but not refusal", "The system cannot connect to the database.", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckStructure("", tt.input); got != tt.ok {
				t.Errorf("CheckStructure('', %q) = %v, want %v", tt.input, got, tt.ok)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestCheckStructure -v`
Expected: FAIL — `CheckStructure` not defined

- [ ] **Step 3: Implement structural validation**

```go
// internal/provider/structural.go
package provider

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// CheckStructure validates an LLM response based on the expected format.
// Returns true if the response passes structural checks.
// Format values: "json", "yaml", or "" (plain text).
func CheckStructure(format, response string) bool {
	trimmed := strings.TrimSpace(response)
	if trimmed == "" {
		return false
	}

	switch strings.ToLower(format) {
	case "json":
		var v any
		return json.Unmarshal([]byte(trimmed), &v) == nil
	case "yaml":
		var v any
		return yaml.Unmarshal([]byte(trimmed), &v) == nil
	default:
		return !isRefusal(trimmed)
	}
}

// isRefusal checks if a response looks like an LLM refusal.
func isRefusal(s string) bool {
	lower := strings.ToLower(s)
	prefixes := []string{
		"i cannot",
		"i can't",
		"i'm sorry",
		"i am sorry",
		"as an ai",
		"i'm not able",
		"i am not able",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run TestCheckStructure -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/provider/structural.go internal/provider/structural_test.go
git commit -m "feat: add structural validation for LLM responses"
```

---

### Task 2: Self-Evaluation

**Files:**
- Create: `internal/provider/eval.go`
- Create: `internal/provider/eval_test.go`

- [ ] **Step 1: Write failing tests for self-eval**

```go
// internal/provider/eval_test.go
package provider

import "testing"

func TestBuildEvalPrompt(t *testing.T) {
	prompt := BuildEvalPrompt("classify this bug", "It's a UI rendering issue")
	if prompt == "" {
		t.Fatal("expected non-empty eval prompt")
	}
	// Should contain both the task and response
	if !contains(prompt, "classify this bug") {
		t.Error("eval prompt missing original task")
	}
	if !contains(prompt, "UI rendering issue") {
		t.Error("eval prompt missing response")
	}
}

func TestParseEvalScore(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"bare number", "4", 4},
		{"with newline", "5\n", 5},
		{"with text", "I'd rate this a 3 because...", 3},
		{"markdown", "**4**", 4},
		{"no number", "this is great", 0},
		{"out of range high", "7", 0},
		{"out of range low", "0", 0},
		{"with whitespace", "  3  ", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseEvalScore(tt.input); got != tt.expected {
				t.Errorf("ParseEvalScore(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run "TestBuildEvalPrompt|TestParseEvalScore" -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement self-eval**

```go
// internal/provider/eval.go
package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BuildEvalPrompt constructs a self-evaluation prompt for the local model.
func BuildEvalPrompt(task, response string) string {
	return fmt.Sprintf(`Rate the quality of this response 1-5:
- 5: Complete, accurate, well-structured
- 3: Partially correct but missing key details
- 1: Wrong, irrelevant, or incoherent

Task: %s

Response: %s

Reply with only a number.`, task, response)
}

var scoreRe = regexp.MustCompile(`[1-5]`)

// ParseEvalScore extracts a 1-5 score from an LLM self-eval response.
// Returns 0 if no valid score is found.
func ParseEvalScore(response string) int {
	trimmed := strings.TrimSpace(response)

	// Try first character for bare number responses
	if len(trimmed) >= 1 {
		if n, err := strconv.Atoi(trimmed[:1]); err == nil && n >= 1 && n <= 5 {
			return n
		}
	}

	// Scan for first 1-5 digit
	match := scoreRe.FindString(trimmed)
	if match != "" {
		n, _ := strconv.Atoi(match)
		return n
	}

	return 0
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run "TestBuildEvalPrompt|TestParseEvalScore" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/provider/eval.go internal/provider/eval_test.go
git commit -m "feat: add LLM self-evaluation prompt and score parser"
```

---

### Task 3: Extend Types for Smart Routing

**Files:**
- Modify: `internal/provider/tokens.go` (add escalation fields to `RunResult`)
- Modify: `internal/pipeline/types.go` (add `Tier` and `Format` to `LLMStep`)
- Modify: `internal/esearch/telemetry.go` (add escalation fields to `LLMCallDoc`)
- Modify: `cmd/config.go` (add `EvalThreshold`)

- [ ] **Step 1: Add escalation metadata to RunResult**

In `internal/provider/tiers.go`, update `RunResult` to track the full escalation chain:

```go
// Replace the existing RunResult struct (lines 40-45) with:
// RunResult wraps an LLMResult with escalation metadata.
type RunResult struct {
	LLMResult
	Tier              int                `json:"tier"`
	Escalated         bool               `json:"escalated"`
	EscalationReason  EscalationReason   `json:"escalation_reason,omitempty"`
	EscalationChain   []int              `json:"escalation_chain,omitempty"`
	EvalScores        []int              `json:"eval_scores,omitempty"`
}
```

- [ ] **Step 2: Add Tier and Format to LLMStep**

In `internal/pipeline/types.go`, update the `LLMStep` struct:

```go
// Replace the existing LLMStep struct (lines 29-33) with:
// LLMStep configures an LLM invocation.
type LLMStep struct {
	Provider string `yaml:"provider,omitempty"` // "ollama" or "claude" (default: config)
	Model    string `yaml:"model,omitempty"`
	Prompt   string `yaml:"prompt"`
	Tier     *int   `yaml:"tier,omitempty"`   // pin to specific tier (0=local, 1=cheap, 2=premium)
	Format   string `yaml:"format,omitempty"` // expected output format: "json", "yaml", or "" (plain text)
}
```

- [ ] **Step 3: Add escalation fields to LLMCallDoc**

In `internal/esearch/telemetry.go`, add fields to `LLMCallDoc` (after the existing `EscalationReason` field on line 55):

```go
// Add these fields to LLMCallDoc after EscalationReason:
	EscalationChain []int  `json:"escalation_chain,omitempty"`
	EvalScores      []int  `json:"eval_scores,omitempty"`
	FinalTier       int    `json:"final_tier"`
```

- [ ] **Step 4: Add EvalThreshold to Config**

In `cmd/config.go`, add `EvalThreshold` to the `Config` struct:

```go
// Replace the existing Config struct (lines 19-24) with:
type Config struct {
	DefaultModel    string                    `yaml:"default_model"`
	DefaultProvider string                    `yaml:"default_provider"`
	EvalThreshold   int                       `yaml:"eval_threshold,omitempty"`
	Tiers           []provider.TierConfig     `yaml:"tiers,omitempty"`
	Providers       map[string]ProviderConfig `yaml:"providers,omitempty"`
}
```

Add default in `loadConfigFrom` after the existing tier default (after line 108):

```go
	if cfg.EvalThreshold == 0 {
		cfg.EvalThreshold = 4
	}
```

- [ ] **Step 5: Run all existing tests to verify nothing broke**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ ./internal/pipeline/ ./cmd/ -v -count=1 2>&1 | tail -30`
Expected: All existing tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/provider/tiers.go internal/pipeline/types.go internal/esearch/telemetry.go cmd/config.go
git commit -m "feat: add type scaffolding for smart routing (Tier, Format, EvalThreshold, escalation metadata)"
```

---

### Task 4: Parse Tier and Format in S-Expression Workflows

**Files:**
- Modify: `internal/pipeline/sexpr.go`
- Modify: `internal/pipeline/sexpr_test.go`

- [ ] **Step 1: Write failing test for tier/format parsing**

Add to `internal/pipeline/sexpr_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestLoadBytes_Sexpr_LLMTier|TestLoadBytes_Sexpr_LLMNoTier" -v`
Expected: FAIL — `Tier` field exists but sexpr parser doesn't handle `:tier` or `:format` keywords (returns error "unknown llm keyword")

- [ ] **Step 3: Add tier and format parsing to convertLLM**

In `internal/pipeline/sexpr.go`, add cases to the `switch key` block in `convertLLM` (after line 169, the `:model` case):

```go
			case "tier":
				n := 0
				valStr := resolveVal(val, defs)
				fmt.Sscanf(valStr, "%d", &n)
				llm.Tier = &n
			case "format":
				llm.Format = resolveVal(val, defs)
```

Also add `"fmt"` to the imports at the top of `sexpr.go` if not already present.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestLoadBytes_Sexpr_LLMTier|TestLoadBytes_Sexpr_LLMNoTier" -v`
Expected: PASS

- [ ] **Step 5: Run all pipeline tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add internal/pipeline/sexpr.go internal/pipeline/sexpr_test.go
git commit -m "feat: parse :tier and :format keywords in sexpr LLM steps"
```

---

### Task 5: RunSmart — The Core Escalation Loop

**Files:**
- Modify: `internal/provider/tiers.go`
- Modify: `internal/provider/tiers_test.go`

- [ ] **Step 1: Write failing tests for RunSmart**

Add to `internal/provider/tiers_test.go`:

```go
func TestRunSmart_LocalSucceeds(t *testing.T) {
	callLog := []string{}
	resolver := func(name string) (ProviderFunc, bool) {
		return func(model, prompt string) (LLMResult, error) {
			callLog = append(callLog, name)
			return LLMResult{Provider: name, Model: model, Response: "good answer"}, nil
		}, true
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"fake-local"}, Model: "local-model"},
		{Providers: []string{"fake-cloud"}, Model: "cloud-model"},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	// evalFunc always returns high score — local should succeed
	result, err := runner.RunSmart(context.Background(), "test prompt", "", 4, func(model, prompt string) (LLMResult, error) {
		return LLMResult{Response: "5"}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tier != 0 {
		t.Errorf("tier = %d, want 0 (local succeeded)", result.Tier)
	}
	if result.Escalated {
		t.Error("should not be escalated")
	}
	if len(callLog) != 1 {
		t.Errorf("expected 1 provider call, got %d: %v", len(callLog), callLog)
	}
}

func TestRunSmart_EscalatesToTier1(t *testing.T) {
	callLog := []string{}
	resolver := func(name string) (ProviderFunc, bool) {
		return func(model, prompt string) (LLMResult, error) {
			callLog = append(callLog, name)
			return LLMResult{Provider: name, Model: model, Response: "answer from " + name}, nil
		}, true
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"fake-local"}, Model: "local-model"},
		{Providers: []string{"fake-cloud"}, Model: "cloud-model"},
		{Providers: []string{"fake-premium"}, Model: "premium-model"},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	evalCall := 0
	evalFunc := func(model, prompt string) (LLMResult, error) {
		evalCall++
		// First eval (tier 0): low score. Second eval (tier 1): high score.
		if evalCall == 1 {
			return LLMResult{Response: "2"}, nil
		}
		return LLMResult{Response: "5"}, nil
	}

	result, err := runner.RunSmart(context.Background(), "test prompt", "", 4, evalFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tier != 1 {
		t.Errorf("tier = %d, want 1", result.Tier)
	}
	if !result.Escalated {
		t.Error("should be escalated")
	}
	if len(result.EscalationChain) != 2 {
		t.Errorf("escalation chain = %v, want [0, 1]", result.EscalationChain)
	}
	if len(result.EvalScores) != 2 {
		t.Errorf("eval scores = %v, want [2, 5]", result.EvalScores)
	}
}

func TestRunSmart_StructuralFailureSkipsEval(t *testing.T) {
	callLog := []string{}
	resolver := func(name string) (ProviderFunc, bool) {
		return func(model, prompt string) (LLMResult, error) {
			callLog = append(callLog, name)
			if name == "fake-local" {
				return LLMResult{Provider: name, Response: "not valid json"}, nil
			}
			return LLMResult{Provider: name, Response: `{"valid": true}`}, nil
		}, true
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"fake-local"}, Model: "m"},
		{Providers: []string{"fake-cloud"}, Model: "m"},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	evalCalled := false
	evalFunc := func(model, prompt string) (LLMResult, error) {
		evalCalled = true
		return LLMResult{Response: "5"}, nil
	}

	result, err := runner.RunSmart(context.Background(), "give me json", "json", 4, evalFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Tier 0 fails structural, escalates to tier 1 which passes structural + eval
	if result.Tier != 1 {
		t.Errorf("tier = %d, want 1", result.Tier)
	}
}

func TestRunSmart_FinalTierNoEval(t *testing.T) {
	resolver := func(name string) (ProviderFunc, bool) {
		return func(model, prompt string) (LLMResult, error) {
			return LLMResult{Provider: name, Model: model, Response: "premium answer"}, nil
		}, true
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"fake-premium"}, Model: "m"},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	evalCalled := false
	evalFunc := func(model, prompt string) (LLMResult, error) {
		evalCalled = true
		return LLMResult{Response: "1"}, nil // would fail but shouldn't be called
	}

	result, err := runner.RunSmart(context.Background(), "test", "", 4, evalFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evalCalled {
		t.Error("eval should not be called on final tier")
	}
	if result.Response != "premium answer" {
		t.Errorf("response = %q, want premium answer", result.Response)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run "TestRunSmart" -v`
Expected: FAIL — `RunSmart` not defined

- [ ] **Step 3: Implement RunSmart**

Add to `internal/provider/tiers.go`:

```go
// Add new escalation reasons after existing constants (line 37):
const (
	ReasonStructural EscalationReason = "structural"
	ReasonEval       EscalationReason = "eval"
)

// EvalFunc calls the local LLM to evaluate a response. Used for self-eval.
type EvalFunc func(model, prompt string) (LLMResult, error)

// RunSmart tries providers in tier order with structural validation and self-evaluation.
// format is the expected output format ("json", "yaml", or "").
// threshold is the minimum eval score (1-5) to accept a response.
// evalFn is called to run the self-evaluation (typically a local Ollama call).
// At the final tier, no eval is performed — the response is accepted as-is.
func (tr *TieredRunner) RunSmart(ctx context.Context, prompt, format string, threshold int, evalFn EvalFunc) (RunResult, error) {
	var chain []int
	var scores []int
	var lastReason EscalationReason
	lastTier := len(tr.tiers) - 1

	for tierIdx, tier := range tr.tiers {
		for _, name := range tier.Providers {
			select {
			case <-ctx.Done():
				return RunResult{}, ctx.Err()
			default:
			}

			model := tier.Model
			fmt.Fprintf(os.Stderr, ">> smart tier %d: trying %s\n", tierIdx, name)
			result, err := tr.callProvider(name, model, prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, ">> smart tier %d: %s error: %v\n", tierIdx, name, err)
				lastReason = ReasonProviderError
				continue // next provider in same tier
			}

			chain = append(chain, tierIdx)

			// Structural check
			if !CheckStructure(format, result.Response) {
				fmt.Fprintf(os.Stderr, ">> smart tier %d: %s structural fail, escalating\n", tierIdx, name)
				scores = append(scores, 0)
				lastReason = ReasonStructural
				break // escalate to next tier
			}

			// Final tier: accept without eval
			if tierIdx == lastTier {
				fmt.Fprintf(os.Stderr, ">> smart tier %d: %s accepted (final tier)\n", tierIdx, name)
				return RunResult{
					LLMResult:       result,
					Tier:            tierIdx,
					Escalated:       tierIdx > 0,
					EscalationChain: chain,
					EvalScores:      scores,
				}, nil
			}

			// Self-eval via local model
			evalPrompt := BuildEvalPrompt(prompt, result.Response)
			evalResult, evalErr := evalFn(model, evalPrompt)
			score := 0
			if evalErr == nil {
				score = ParseEvalScore(evalResult.Response)
			}
			scores = append(scores, score)

			if score >= threshold {
				fmt.Fprintf(os.Stderr, ">> smart tier %d: %s accepted (score %d >= %d)\n", tierIdx, name, score, threshold)
				return RunResult{
					LLMResult:       result,
					Tier:            tierIdx,
					Escalated:       tierIdx > 0,
					EscalationChain: chain,
					EvalScores:      scores,
				}, nil
			}

			fmt.Fprintf(os.Stderr, ">> smart tier %d: %s rejected (score %d < %d), escalating\n", tierIdx, name, score, threshold)
			lastReason = ReasonEval
			break // escalate to next tier
		}
	}
	return RunResult{}, fmt.Errorf("all tiers exhausted (last reason: %s)", lastReason)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -run "TestRunSmart" -v`
Expected: All PASS

- [ ] **Step 5: Run all provider tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/provider/ -v`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add internal/provider/tiers.go internal/provider/tiers_test.go
git commit -m "feat: add RunSmart escalation loop with structural validation and self-eval"
```

---

### Task 6: Wire RunSmart into Pipeline Runner

**Files:**
- Modify: `internal/pipeline/runner.go`
- Modify: `internal/pipeline/runner_test.go`

- [ ] **Step 1: Write failing test for auto-routing**

Add to `internal/pipeline/runner_test.go`:

```go
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
					// No provider, no tier — should auto-route
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
	// Should have called only the premium tier provider
	if len(callLog) != 1 || callLog[0] != "fake-premium" {
		t.Errorf("callLog = %v, want [fake-premium]", callLog)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRun_SmartRouting|TestRun_PinnedTier" -v`
Expected: FAIL — `Tiers` and `EvalThreshold` not in `RunOpts`

- [ ] **Step 3: Add Tiers and EvalThreshold to RunOpts**

In `internal/pipeline/runner.go`, update the `RunOpts` struct:

```go
// Replace the existing RunOpts struct (lines 25-30) with:
// RunOpts holds optional dependencies for a workflow run.
type RunOpts struct {
	Telemetry        *esearch.Telemetry
	Issue            string
	ComparisonGroup  string
	ProviderResolver provider.ResolverFunc
	Tiers            []provider.TierConfig
	EvalThreshold    int
}
```

- [ ] **Step 4: Wire smart routing into the LLM execution path**

In `internal/pipeline/runner.go`, replace the entire LLM dispatch block (the `} else if step.LLM != nil {` block, lines 131-261) with:

```go
		} else if step.LLM != nil {
			rendered, err := render(step.LLM.Prompt, data, steps)
			if err != nil {
				return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
			}
			fmt.Printf("  > %s\n", step.ID)

			prov := strings.ToLower(step.LLM.Provider)
			model := step.LLM.Model
			if model == "" {
				model = defaultModel
			}

			var out string
			var stepTier int
			var stepEscalated bool
			var stepEscalationReason string
			var stepEscalationChain []int
			var stepEvalScores []int
			var stepCost float64
			var stepTokensIn, stepTokensOut int

			// Smart routing: no provider AND no pinned tier AND tiers available
			useSmart := prov == "" && step.LLM.Tier == nil && len(tiers) > 0
			// Pinned tier: explicit tier set, use subset of tiers
			usePinned := step.LLM.Tier != nil && len(tiers) > 0

			if useSmart || usePinned {
				activeTiers := tiers
				if usePinned {
					tierIdx := *step.LLM.Tier
					if tierIdx >= 0 && tierIdx < len(tiers) {
						// Single-tier slice — RunSmart treats it as final tier (no eval)
						activeTiers = tiers[tierIdx : tierIdx+1]
					}
				}

				runner := provider.NewTieredRunner(activeTiers, reg)
				runner.Resolver = providerResolver

				threshold := evalThreshold
				evalFn := func(evalModel, evalPrompt string) (provider.LLMResult, error) {
					return provider.RunOllamaWithResult(defaultModel, evalPrompt)
				}

				format := step.LLM.Format
				rr, llmErr := runner.RunSmart(context.Background(), rendered, format, threshold, evalFn)
				if llmErr != nil {
					err = llmErr
				} else {
					out = rr.Response
					stepTier = rr.Tier
					if usePinned {
						stepTier = *step.LLM.Tier
					}
					stepEscalated = rr.Escalated
					stepEscalationChain = rr.EscalationChain
					stepEvalScores = rr.EvalScores
					stepCost = rr.CostUSD
					stepTokensIn = rr.TokensIn
					stepTokensOut = rr.TokensOut
				}
			} else {
				// Original dispatch: explicit provider or no tiers configured
				switch prov {
				case "ollama", "":
					if model == "" {
						model = "qwen3:8b"
					}
					result, llmErr := provider.RunOllamaWithResult(model, rendered)
					if llmErr != nil {
						err = llmErr
					} else {
						out = result.Response
						stepTokensIn = result.TokensIn
						stepTokensOut = result.TokensOut
					}
				default:
					var resolved bool
					if providerResolver != nil {
						if fn, ok := providerResolver(prov); ok {
							resolved = true
							result, llmErr := fn(step.LLM.Model, rendered)
							if llmErr != nil {
								err = llmErr
							} else {
								out = result.Response
								stepTokensIn = result.TokensIn
								stepTokensOut = result.TokensOut
								stepCost = result.CostUSD
							}
						}
					}
					if !resolved {
						result, provErr := reg.RunProviderWithResult(prov, model, rendered)
						if provErr != nil {
							err = provErr
						} else {
							out = result.Response
							stepTokensIn = result.TokensIn
							stepTokensOut = result.TokensOut
							stepCost = result.CostUSD
						}
					}
				}
			}

			if err != nil {
				return nil, fmt.Errorf("step %s: %w", step.ID, err)
			}

			tokIn := int64(stepTokensIn)
			tokOut := int64(stepTokensOut)
			totalTokensIn += tokIn
			totalTokensOut += tokOut
			totalCostUSD += stepCost
			totalLatencyMS += 0 // latency tracked in provider
			llmSteps++
			lastLLMOutput = out

			if tel != nil {
				reason := ""
				if stepEscalationReason != "" {
					reason = stepEscalationReason
				} else if stepEscalated {
					reason = "eval"
				}
				tel.IndexLLMCall(context.Background(), esearch.LLMCallDoc{
					RunID:           runID,
					Step:            fmt.Sprintf("workflow:%s/%s", w.Name, step.ID),
					Tier:            stepTier,
					Provider:        prov,
					Model:           model,
					TokensIn:        tokIn,
					TokensOut:       tokOut,
					TokensTotal:     tokIn + tokOut,
					CostUSD:         stepCost,
					LatencyMS:       0,
					Escalated:       stepEscalated,
					EscalationReason: reason,
					EscalationChain: stepEscalationChain,
					EvalScores:      stepEvalScores,
					FinalTier:       stepTier,
					WorkflowName:    w.Name,
					Issue:           issue,
					ComparisonGroup: compGroup,
					Timestamp:       time.Now().UTC().Format(time.RFC3339),
				})
			}

			steps[step.ID] = out
```

Also update the variable extraction from opts at the top of `Run()`. After the existing `providerResolver` extraction (line 69), add:

```go
	var tiers []provider.TierConfig
	var evalThreshold int
	if len(opts) > 0 {
		tel = opts[0].Telemetry
		issue = opts[0].Issue
		compGroup = opts[0].ComparisonGroup
		providerResolver = opts[0].ProviderResolver
		tiers = opts[0].Tiers
		evalThreshold = opts[0].EvalThreshold
	}
	if evalThreshold == 0 {
		evalThreshold = 4
	}
```

(This replaces the existing opts extraction block at lines 65-70.)

- [ ] **Step 5: Run new tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -run "TestRun_SmartRouting|TestRun_PinnedTier" -v`
Expected: PASS

- [ ] **Step 6: Run all pipeline tests**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./internal/pipeline/ -v`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "feat: wire smart routing into pipeline runner with auto-escalation"
```

---

### Task 7: Wire Config Into Workflow Commands

**Files:**
- Modify: `cmd/workflow.go` (or wherever `Run()` is called with `RunOpts`)

- [ ] **Step 1: Find all call sites of pipeline.Run**

Run: `cd /Users/stokes/Projects/gl1tch && grep -rn "pipeline.Run(" cmd/ internal/ --include="*.go"`

- [ ] **Step 2: Update call sites to pass Tiers and EvalThreshold**

At each call site that constructs `RunOpts`, add the tiers and threshold from config:

```go
pipeline.RunOpts{
	Telemetry:        tel,
	ProviderResolver: cfg.BuildProviderResolver(),
	Tiers:            cfg.Tiers,
	EvalThreshold:    cfg.EvalThreshold,
	// ... existing fields
}
```

- [ ] **Step 3: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... 2>&1 | tail -20`
Expected: All PASS (or only unrelated failures)

- [ ] **Step 4: Commit**

```bash
git add cmd/
git commit -m "feat: pass tiers and eval_threshold from config to pipeline runner"
```

---

### Task 8: Kibana Dashboard Visualizations

**Files:**
- Modify: `internal/dashboard/default.ndjson`

- [ ] **Step 1: Add tier distribution visualization**

Append to `internal/dashboard/default.ndjson` — a Vega-Lite stacked bar chart showing LLM call count by final tier over time:

```json
{"id":"glitch-vega-tier-distribution","type":"visualization","attributes":{"title":"Tier Distribution Over Time","visState":"{\"title\": \"Tier Distribution Over Time\", \"type\": \"vega\", \"params\": {\"spec\": \"{\\\"$schema\\\": \\\"https://vega.github.io/schema/vega-lite/v5.json\\\", \\\"title\\\": {\\\"text\\\": \\\"Where Are Your LLM Calls Landing?\\\", \\\"subtitle\\\": \\\"Green = local (free). Yellow = cheap cloud. Red = premium. Less red = more savings.\\\", \\\"fontSize\\\": 16, \\\"subtitleFontSize\\\": 12, \\\"subtitleColor\\\": \\\"#888\\\"}, \\\"autosize\\\": \\\"fit\\\", \\\"data\\\": {\\\"url\\\": {\\\"%context%\\\": true, \\\"index\\\": \\\"glitch-llm-calls\\\", \\\"body\\\": {\\\"size\\\": 500}}, \\\"format\\\": {\\\"property\\\": \\\"hits.hits\\\"}}, \\\"transform\\\": [{\\\"calculate\\\": \\\"datum._source.final_tier == 0 ? 'Local (Free)' : datum._source.final_tier == 1 ? 'Cloud (Cheap)' : 'Premium'\\\", \\\"as\\\": \\\"tier_label\\\"}, {\\\"calculate\\\": \\\"toDate(datum._source.timestamp)\\\", \\\"as\\\": \\\"time\\\"}], \\\"mark\\\": {\\\"type\\\": \\\"bar\\\", \\\"tooltip\\\": true}, \\\"encoding\\\": {\\\"x\\\": {\\\"field\\\": \\\"time\\\", \\\"type\\\": \\\"temporal\\\", \\\"timeUnit\\\": \\\"hoursminutes\\\", \\\"title\\\": \\\"Time\\\"}, \\\"y\\\": {\\\"aggregate\\\": \\\"count\\\", \\\"title\\\": \\\"LLM Calls\\\"}, \\\"color\\\": {\\\"field\\\": \\\"tier_label\\\", \\\"type\\\": \\\"nominal\\\", \\\"title\\\": \\\"Tier\\\", \\\"scale\\\": {\\\"domain\\\": [\\\"Local (Free)\\\", \\\"Cloud (Cheap)\\\", \\\"Premium\\\"], \\\"range\\\": [\\\"#54B399\\\", \\\"#DA8B45\\\", \\\"#D36086\\\"]}}}, \\\"config\\\": {\\\"view\\\": {\\\"stroke\\\": null}}}\"}, \"aggs\": []}","kibanaSavedObjectMeta":{"searchSourceJSON":"{}"}},"references":[]}
```

- [ ] **Step 2: Add cost savings visualization**

Append to `internal/dashboard/default.ndjson` — a line chart comparing actual cost vs hypothetical all-premium cost:

```json
{"id":"glitch-vega-cost-savings","type":"visualization","attributes":{"title":"Smart Routing Cost Savings","visState":"{\"title\": \"Smart Routing Cost Savings\", \"type\": \"vega\", \"params\": {\"spec\": \"{\\\"$schema\\\": \\\"https://vega.github.io/schema/vega-lite/v5.json\\\", \\\"title\\\": {\\\"text\\\": \\\"Money Saved by Smart Routing\\\", \\\"subtitle\\\": \\\"Blue = what you paid. Red = what premium would have cost. Gap = savings.\\\", \\\"fontSize\\\": 16, \\\"subtitleFontSize\\\": 12, \\\"subtitleColor\\\": \\\"#888\\\"}, \\\"autosize\\\": \\\"fit\\\", \\\"data\\\": {\\\"url\\\": {\\\"%context%\\\": true, \\\"index\\\": \\\"glitch-llm-calls\\\", \\\"body\\\": {\\\"size\\\": 500}}, \\\"format\\\": {\\\"property\\\": \\\"hits.hits\\\"}}, \\\"transform\\\": [{\\\"calculate\\\": \\\"datum._source.cost_usd\\\", \\\"as\\\": \\\"actual_cost\\\"}, {\\\"calculate\\\": \\\"(datum._source.tokens_in * 3.0 / 1000000) + (datum._source.tokens_out * 15.0 / 1000000)\\\", \\\"as\\\": \\\"premium_cost\\\"}, {\\\"calculate\\\": \\\"toDate(datum._source.timestamp)\\\", \\\"as\\\": \\\"time\\\"}, {\\\"fold\\\": [\\\"actual_cost\\\", \\\"premium_cost\\\"], \\\"as\\\": [\\\"cost_type\\\", \\\"cost\\\"]}, {\\\"calculate\\\": \\\"datum.cost_type == 'actual_cost' ? 'Actual (Smart Routed)' : 'If Everything Was Premium'\\\", \\\"as\\\": \\\"label\\\"}], \\\"mark\\\": {\\\"type\\\": \\\"line\\\", \\\"interpolate\\\": \\\"monotone\\\", \\\"strokeWidth\\\": 3, \\\"point\\\": true}, \\\"encoding\\\": {\\\"x\\\": {\\\"field\\\": \\\"time\\\", \\\"type\\\": \\\"temporal\\\", \\\"title\\\": \\\"Time\\\"}, \\\"y\\\": {\\\"field\\\": \\\"cost\\\", \\\"type\\\": \\\"quantitative\\\", \\\"title\\\": \\\"Cumulative Cost (USD)\\\", \\\"axis\\\": {\\\"format\\\": \\\"$.4f\\\"}}, \\\"color\\\": {\\\"field\\\": \\\"label\\\", \\\"type\\\": \\\"nominal\\\", \\\"title\\\": \\\"Cost Type\\\", \\\"scale\\\": {\\\"domain\\\": [\\\"Actual (Smart Routed)\\\", \\\"If Everything Was Premium\\\"], \\\"range\\\": [\\\"#6092C0\\\", \\\"#D36086\\\"]}}}, \\\"config\\\": {\\\"view\\\": {\\\"stroke\\\": null}}}\"}, \"aggs\": []}","kibanaSavedObjectMeta":{"searchSourceJSON":"{}"}},"references":[]}
```

- [ ] **Step 3: Add escalation hotspots visualization**

Append to `internal/dashboard/default.ndjson` — a table showing which workflow steps escalate most:

```json
{"id":"glitch-vega-escalation-hotspots","type":"visualization","attributes":{"title":"Escalation Hotspots","visState":"{\"title\": \"Escalation Hotspots\", \"type\": \"vega\", \"params\": {\"spec\": \"{\\\"$schema\\\": \\\"https://vega.github.io/schema/vega-lite/v5.json\\\", \\\"title\\\": {\\\"text\\\": \\\"Which Steps Escalate Most?\\\", \\\"subtitle\\\": \\\"Taller bars = steps that can't be handled locally. Consider fixing the prompt or pinning the tier.\\\", \\\"fontSize\\\": 16, \\\"subtitleFontSize\\\": 12, \\\"subtitleColor\\\": \\\"#888\\\"}, \\\"autosize\\\": \\\"fit\\\", \\\"data\\\": {\\\"url\\\": {\\\"%context%\\\": true, \\\"index\\\": \\\"glitch-llm-calls\\\", \\\"body\\\": {\\\"size\\\": 500}}, \\\"format\\\": {\\\"property\\\": \\\"hits.hits\\\"}}, \\\"transform\\\": [{\\\"filter\\\": \\\"datum._source.escalated == true\\\"}, {\\\"calculate\\\": \\\"datum._source.step\\\", \\\"as\\\": \\\"step\\\"}, {\\\"calculate\\\": \\\"datum._source.escalation_reason\\\", \\\"as\\\": \\\"reason\\\"}], \\\"mark\\\": {\\\"type\\\": \\\"bar\\\", \\\"cornerRadiusEnd\\\": 6, \\\"tooltip\\\": true}, \\\"encoding\\\": {\\\"y\\\": {\\\"field\\\": \\\"step\\\", \\\"type\\\": \\\"nominal\\\", \\\"title\\\": \\\"Workflow Step\\\", \\\"sort\\\": \\\"-x\\\", \\\"axis\\\": {\\\"labelFontSize\\\": 12}}, \\\"x\\\": {\\\"aggregate\\\": \\\"count\\\", \\\"title\\\": \\\"Escalation Count\\\"}, \\\"color\\\": {\\\"field\\\": \\\"reason\\\", \\\"type\\\": \\\"nominal\\\", \\\"title\\\": \\\"Reason\\\", \\\"scale\\\": {\\\"domain\\\": [\\\"structural\\\", \\\"eval\\\", \\\"provider_error\\\"], \\\"range\\\": [\\\"#D36086\\\", \\\"#DA8B45\\\", \\\"#6092C0\\\"]}}}, \\\"config\\\": {\\\"view\\\": {\\\"stroke\\\": null}}}\"}, \"aggs\": []}","kibanaSavedObjectMeta":{"searchSourceJSON":"{}"}},"references":[]}
```

- [ ] **Step 4: Update the dashboard panel list to include new visualizations**

Find the dashboard object line in `default.ndjson` (the line with `"type":"dashboard"`) and add the three new visualization references to its panels array. The exact edit depends on the current dashboard line — read it and add panel entries for `glitch-vega-tier-distribution`, `glitch-vega-cost-savings`, and `glitch-vega-escalation-hotspots`.

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/default.ndjson
git commit -m "feat: add tier distribution, cost savings, and escalation hotspot Kibana visualizations"
```

---

### Task 9: End-to-End Smoke Test

- [ ] **Step 1: Build and verify**

Run: `cd /Users/stokes/Projects/gl1tch && go build ./...`
Expected: Clean build, no errors

- [ ] **Step 2: Run full test suite**

Run: `cd /Users/stokes/Projects/gl1tch && go test ./... -count=1 2>&1 | tail -30`
Expected: All PASS

- [ ] **Step 3: Manual smoke test with a simple workflow**

Create a test workflow that exercises smart routing:

```bash
cat > /tmp/test-smart-routing.yaml << 'EOF'
name: test-smart-routing
description: smoke test for smart routing
steps:
  - id: classify
    llm:
      format: json
      prompt: |
        Classify this GitHub issue title into one category.
        Title: "Button click handler not working in Safari"
        Reply with JSON: {"category": "bug|feature|docs|chore"}
EOF

cd /Users/stokes/Projects/gl1tch && go run . workflow run /tmp/test-smart-routing.yaml
```

Expected: Output shows `>> smart tier 0: trying ollama` and produces valid JSON. If local model fails, should escalate to tier 1.

- [ ] **Step 4: Test with pinned tier**

```bash
cat > /tmp/test-pinned-tier.yaml << 'EOF'
name: test-pinned-tier
description: smoke test for pinned tier
steps:
  - id: analyze
    llm:
      tier: 1
      prompt: "Summarize: smart routing saves money by trying cheap models first"
EOF

cd /Users/stokes/Projects/gl1tch && go run . workflow run /tmp/test-pinned-tier.yaml
```

Expected: Output shows tier 1 provider being used directly, no tier 0 attempt.

- [ ] **Step 5: Clean up test files**

```bash
rm /tmp/test-smart-routing.yaml /tmp/test-pinned-tier.yaml
```

- [ ] **Step 6: Final commit if any fixes were needed**

```bash
# Only if fixes were needed during smoke testing
git add -A && git commit -m "fix: address smoke test findings in smart routing"
```
