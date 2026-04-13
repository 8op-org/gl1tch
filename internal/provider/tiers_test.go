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
