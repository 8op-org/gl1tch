package provider

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// AgentProvider dispatches prompts to AI coding agents (Claude Code, GitHub
// Copilot CLI, Gemini CLI) running in headless/print mode. Unlike the OpenAI
// provider which hits a stateless API, agent providers run full tool-using
// agents that can read files, execute commands, and reason across steps.
type AgentProvider struct {
	Name    string // "claude", "copilot", "gemini"
	Command string // Base command (e.g., "claude", "gh copilot")
}

// KnownAgents maps provider names to their headless CLI invocations.
var KnownAgents = map[string]AgentProvider{
	"claude": {
		Name:    "claude",
		Command: "claude",
	},
	"copilot": {
		Name:    "copilot",
		Command: "gh copilot",
	},
	"gemini": {
		Name:    "gemini",
		Command: "gemini",
	},
}

// Run executes the agent in headless mode with the given prompt and returns
// a structured LLMResult. The prompt (including any prepended skill context)
// is piped via stdin to avoid shell escaping issues with large prompts.
func (a *AgentProvider) Run(model, prompt string) (LLMResult, error) {
	start := time.Now()

	args := a.buildArgs(model)
	cmd := exec.Command("sh", "-c", strings.Join(args, " "))
	cmd.Stdin = strings.NewReader(prompt)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return LLMResult{}, fmt.Errorf("agent %s: %w\n%s", a.Name, err, out)
	}

	response := strings.TrimSpace(string(out))

	// Try to parse JSON output (claude --output-format json)
	result := LLMResult{
		Provider:  a.Name,
		Model:     model,
		Response:  response,
		TokensIn:  EstimateTokens(prompt),
		TokensOut: EstimateTokens(response),
		Latency:   time.Since(start),
	}

	// If the agent returned JSON (claude -p --output-format json), extract the
	// response text and token counts from the structured output.
	if parsed, ok := parseAgentJSON(response); ok {
		result.Response = parsed.response
		if parsed.tokensIn > 0 {
			result.TokensIn = parsed.tokensIn
		}
		if parsed.tokensOut > 0 {
			result.TokensOut = parsed.tokensOut
		}
		result.CostUSD = parsed.costUSD
	}

	return result, nil
}

// buildArgs constructs the CLI arguments for headless execution.
func (a *AgentProvider) buildArgs(model string) []string {
	switch a.Name {
	case "claude":
		args := []string{a.Command, "-p", "--output-format", "json"}
		if model != "" {
			args = append(args, "--model", model)
		}
		return args
	case "copilot":
		// gh copilot explain reads from stdin
		return []string{"gh", "copilot", "explain"}
	case "gemini":
		return []string{a.Command, "-p"}
	default:
		return []string{a.Command, "-p"}
	}
}

type agentParsed struct {
	response  string
	tokensIn  int
	tokensOut int
	costUSD   float64
}

// parseAgentJSON attempts to parse claude --output-format json output.
func parseAgentJSON(raw string) (agentParsed, bool) {
	var msg struct {
		Result string `json:"result"`
		Usage  struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return agentParsed{}, false
	}
	if msg.Result == "" {
		return agentParsed{}, false
	}
	return agentParsed{
		response:  msg.Result,
		tokensIn:  msg.Usage.InputTokens,
		tokensOut: msg.Usage.OutputTokens,
		costUSD:   msg.CostUSD,
	}, true
}

// IsAgent returns true if the provider name is a known agent CLI.
func IsAgent(name string) bool {
	_, ok := KnownAgents[name]
	return ok
}
