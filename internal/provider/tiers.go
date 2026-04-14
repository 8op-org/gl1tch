package provider

import (
	"context"
	"fmt"
	"os"
	"time"
)

// TierConfig defines a group of providers at the same escalation level.
type TierConfig struct {
	Providers []string `yaml:"providers"`
	Model     string   `yaml:"model,omitempty"`
}

// ProviderFunc calls an LLM provider and returns a result.
type ProviderFunc func(model, prompt string) (LLMResult, error)

// ResolverFunc looks up a provider by name. Returns the call function and true if found.
type ResolverFunc func(name string) (ProviderFunc, bool)

// TieredRunner tries providers in tier order, escalating on failure or validation rejection.
type TieredRunner struct {
	tiers    []TierConfig
	reg      *ProviderRegistry
	Resolver ResolverFunc
}

// EscalationReason describes why a tier was skipped.
type EscalationReason string

const (
	ReasonMalformed     EscalationReason = "malformed_output"
	ReasonEmpty         EscalationReason = "empty_response"
	ReasonHallucinated  EscalationReason = "hallucinated_tool"
	ReasonProviderError EscalationReason = "provider_error"
)

// RunResult wraps an LLMResult with escalation metadata.
type RunResult struct {
	LLMResult
	Tier              int                `json:"tier"`
	Escalated         bool               `json:"escalated"`
	EscalationReason  EscalationReason   `json:"escalation_reason,omitempty"`
	EscalationChain   []int              `json:"escalation_chain,omitempty"`
	EvalScores        []int              `json:"eval_scores,omitempty"`
}

// NewTieredRunner creates a runner that walks through tiers in order.
func NewTieredRunner(tiers []TierConfig, reg *ProviderRegistry) *TieredRunner {
	return &TieredRunner{tiers: tiers, reg: reg}
}

// Run tries providers in tier order. For each tier it tries each provider; if the
// provider errors it moves to the next provider in the same tier. If validate
// returns a non-empty reason the runner escalates to the next tier. If all tiers
// are exhausted it returns an error.
func (tr *TieredRunner) Run(ctx context.Context, prompt string, validate func(string) EscalationReason) (RunResult, error) {
	var lastReason EscalationReason
	for tierIdx, tier := range tr.tiers {
		for _, name := range tier.Providers {
			select {
			case <-ctx.Done():
				return RunResult{}, ctx.Err()
			default:
			}

			model := tier.Model
			fmt.Fprintf(os.Stderr, ">> tier %d: trying %s\n", tierIdx, name)
			result, err := tr.callProvider(name, model, prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, ">> tier %d: %s error: %v\n", tierIdx, name, err)
				lastReason = ReasonProviderError
				continue // next provider in same tier
			}

			if reason := validate(result.Response); reason != "" {
				preview := result.Response
				if len(preview) > 500 {
					preview = preview[:500]
				}
				fmt.Fprintf(os.Stderr, ">> tier %d: %s rejected (%s), escalating\n>> response preview: [%s]\n", tierIdx, name, reason, preview)
				lastReason = reason
				break // escalate to next tier
			}

			return RunResult{
				LLMResult: result,
				Tier:      tierIdx,
				Escalated: tierIdx > 0,
				EscalationReason: lastReason,
			}, nil
		}
	}
	return RunResult{}, fmt.Errorf("all tiers exhausted (last reason: %s)", lastReason)
}

func (tr *TieredRunner) callProvider(name, model, prompt string) (LLMResult, error) {
	start := time.Now()

	if name == "ollama" {
		if model == "" {
			model = "qwen3:8b"
		}
		return RunOllamaWithResult(model, prompt)
	}

	// Check resolver (openai-compatible providers from config)
	if tr.Resolver != nil {
		if fn, ok := tr.Resolver(name); ok {
			return fn(model, prompt)
		}
	}

	raw, err := tr.reg.RunProvider(name, model, prompt)
	if err != nil {
		return LLMResult{}, err
	}

	tokIn := EstimateTokens(prompt)
	tokOut := EstimateTokens(raw)
	return LLMResult{
		Provider:  name,
		Model:     model,
		Response:  raw,
		TokensIn:  tokIn,
		TokensOut: tokOut,
		Latency:   time.Since(start),
		CostUSD:   EstimateCost(name, tokIn, tokOut),
	}, nil
}

// DefaultTiers returns the standard 3-tier escalation chain:
// local free -> free paid -> paid.
func DefaultTiers() []TierConfig {
	return []TierConfig{
		{Providers: []string{"ollama"}, Model: "qwen3:8b"},
		{Providers: []string{"codex", "gemini"}},
		{Providers: []string{"copilot", "claude"}},
	}
}
