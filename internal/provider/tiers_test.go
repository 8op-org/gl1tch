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
