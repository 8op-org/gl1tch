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

// TierLogFunc is called for tier escalation events. If nil, logs to stderr.
type TierLogFunc func(format string, args ...any)

// TieredRunner tries providers in tier order, escalating on failure or validation rejection.
type TieredRunner struct {
	tiers    []TierConfig
	reg      *ProviderRegistry
	Resolver ResolverFunc
	Log      TierLogFunc
}

func (tr *TieredRunner) log(format string, args ...any) {
	if tr.Log != nil {
		tr.Log(format, args...)
	} else {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// EscalationReason describes why a tier was skipped.
type EscalationReason string

const (
	ReasonMalformed     EscalationReason = "malformed_output"
	ReasonEmpty         EscalationReason = "empty_response"
	ReasonHallucinated  EscalationReason = "hallucinated_tool"
	ReasonProviderError EscalationReason = "provider_error"
	ReasonStructural    EscalationReason = "structural"
	ReasonEval          EscalationReason = "eval"
)

// EvalFunc calls the local LLM to evaluate a response.
type EvalFunc func(model, prompt string) (LLMResult, error)

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
			tr.log("tier %d: trying %s", tierIdx, name)
			result, err := tr.callProvider(name, model, prompt)
			if err != nil {
				tr.log("tier %d: %s error: %v", tierIdx, name, err)
				lastReason = ReasonProviderError
				continue // next provider in same tier
			}

			if reason := validate(result.Response); reason != "" {
				preview := result.Response
				if len(preview) > 500 {
					preview = preview[:500]
				}
				tr.log("tier %d: %s rejected (%s), escalating", tierIdx, name, reason)
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

	if name == "lm-studio" {
		p := &LMStudioProvider{
			BaseURL:      "http://localhost:1234",
			DefaultModel: "qwen3-8b",
		}
		if model == "" {
			model = p.DefaultModel
		}
		return p.Chat(model, prompt)
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
			tr.log("tier %d: trying %s", tierIdx, name)
			result, err := tr.callProvider(name, model, prompt)
			if err != nil {
				tr.log("tier %d: %s error: %v", tierIdx, name, err)
				lastReason = ReasonProviderError
				continue // next provider in same tier
			}

			chain = append(chain, tierIdx)

			// Structural check
			if !CheckStructure(format, result.Response) {
				tr.log("tier %d: %s structural fail, escalating", tierIdx, name)
				scores = append(scores, 0)
				lastReason = ReasonStructural
				break // escalate to next tier
			}

			// Final tier: accept without eval
			if tierIdx == lastTier {
				tr.log("tier %d: %s accepted (final tier)", tierIdx, name)
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
				tr.log("tier %d: %s accepted (score %d/%d)", tierIdx, name, score, threshold)
				return RunResult{
					LLMResult:       result,
					Tier:            tierIdx,
					Escalated:       tierIdx > 0,
					EscalationChain: chain,
					EvalScores:      scores,
				}, nil
			}

			tr.log("tier %d: %s rejected (score %d/%d), escalating", tierIdx, name, score, threshold)
			lastReason = ReasonEval
			break // escalate to next tier
		}
	}
	return RunResult{}, fmt.Errorf("all tiers exhausted (last reason: %s)", lastReason)
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
