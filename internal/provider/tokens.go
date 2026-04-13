package provider

import "time"

// LLMResult captures the output and cost metadata of a single LLM call.
type LLMResult struct {
	Provider  string        `json:"provider"`
	Model     string        `json:"model"`
	Response  string        `json:"response"`
	TokensIn  int           `json:"tokens_in"`
	TokensOut int           `json:"tokens_out"`
	Latency   time.Duration `json:"latency"`
	CostUSD   float64       `json:"cost_usd"`
}

// EstimateTokens returns a rough token count (~4 chars per token).
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}

// pricing per 1M tokens (input, output) in USD.
var pricing = map[string][2]float64{
	"claude":  {3.00, 15.00},
	"copilot": {0.00, 0.00},
	"codex":   {0.00, 0.00},
	"gemini":  {0.00, 0.00},
	"ollama":  {0.00, 0.00},
}

// EstimateCost returns the estimated USD cost for the given provider and token counts.
func EstimateCost(providerName string, tokensIn, tokensOut int) float64 {
	p, ok := pricing[providerName]
	if !ok {
		return 0
	}
	return (float64(tokensIn) * p[0] / 1_000_000) + (float64(tokensOut) * p[1] / 1_000_000)
}
