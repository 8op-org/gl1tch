package provider

import (
	"context"
	"testing"
)

func TestDefaultTiers(t *testing.T) {
	tiers := DefaultTiers()
	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}
	if tiers[0].Providers[0] != "ollama" {
		t.Errorf("tier 0 first provider = %q, want ollama", tiers[0].Providers[0])
	}
	if tiers[0].Model != "qwen3:8b" {
		t.Errorf("tier 0 model = %q, want qwen3:8b", tiers[0].Model)
	}
}

func TestTieredRunner_WithResolver(t *testing.T) {
	called := false
	fakeProvider := func(model, prompt string) (LLMResult, error) {
		called = true
		return LLMResult{
			Provider: "openrouter",
			Model:    model,
			Response: "resolved-response",
			TokensIn: 10, TokensOut: 5,
		}, nil
	}

	resolver := func(name string) (ProviderFunc, bool) {
		if name == "openrouter" {
			return fakeProvider, true
		}
		return nil, false
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"openrouter"}, Model: "test-model"},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	result, err := runner.Run(context.Background(), "hello", func(s string) EscalationReason {
		return ""
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("resolver provider was not called")
	}
	if result.Response != "resolved-response" {
		t.Errorf("response = %q, want resolved-response", result.Response)
	}
	if result.Provider != "openrouter" {
		t.Errorf("provider = %q, want openrouter", result.Provider)
	}
}

func TestTieredRunner_ResolverFallsBackToRegistry(t *testing.T) {
	resolver := func(name string) (ProviderFunc, bool) {
		return nil, false
	}

	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"unknown"}},
	}
	runner := NewTieredRunner(tiers, reg)
	runner.Resolver = resolver

	_, err := runner.Run(context.Background(), "hello", func(s string) EscalationReason {
		return ""
	})
	if err == nil {
		t.Fatal("expected error when resolver and registry both miss")
	}
}

func TestTieredRunner_AllFail(t *testing.T) {
	reg := &ProviderRegistry{providers: make(map[string]*Provider)}
	tiers := []TierConfig{
		{Providers: []string{"nonexistent"}},
	}
	runner := NewTieredRunner(tiers, reg)
	_, err := runner.Run(context.Background(), "hello", func(s string) EscalationReason {
		return "" // accept anything
	})
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

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

	result, err := runner.RunSmart(context.Background(), "test prompt", "", 4, func(model, prompt string) (LLMResult, error) {
		return LLMResult{Response: "5"}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tier != 0 {
		t.Errorf("tier = %d, want 0", result.Tier)
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
		t.Errorf("escalation chain = %v, want len 2", result.EscalationChain)
	}
	if len(result.EvalScores) != 2 {
		t.Errorf("eval scores = %v, want len 2", result.EvalScores)
	}
}

func TestRunSmart_StructuralFailureSkipsEval(t *testing.T) {
	resolver := func(name string) (ProviderFunc, bool) {
		return func(model, prompt string) (LLMResult, error) {
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
	if result.Tier != 1 {
		t.Errorf("tier = %d, want 1", result.Tier)
	}
	if evalCalled {
		t.Error("eval should not be called when structural check fails")
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
		return LLMResult{Response: "1"}, nil
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
