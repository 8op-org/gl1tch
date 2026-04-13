package research

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
)

// Goal describes what the research loop is trying to produce.
type Goal string

const (
	GoalSummarize Goal = "summarize"
	GoalImplement Goal = "implement"
)

// ToolCall is the JSON structure the LLM emits to invoke a tool.
type ToolCall struct {
	Tool      string         `json:"tool"`
	RawParams map[string]any `json:"params"`
}

// Params returns all param values as strings (LLMs may emit ints, bools, etc.).
func (tc ToolCall) Params() map[string]string {
	out := make(map[string]string, len(tc.RawParams))
	for k, v := range tc.RawParams {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

// LoopResult captures everything that happened during a tool-use research run.
type LoopResult struct {
	RunID       string           `json:"run_id"`
	Document    ResearchDocument `json:"document"`
	Goal        Goal             `json:"goal"`
	Output      string           `json:"output"`
	ToolCalls   []ToolResult     `json:"tool_calls"`
	LLMCalls    int              `json:"llm_calls"`
	TokensIn    int              `json:"tokens_in"`
	TokensOut   int              `json:"tokens_out"`
	CostUSD     float64          `json:"cost_usd"`
	MaxTier     int              `json:"max_tier"`
	Escalations int              `json:"escalations"`
	Duration    time.Duration    `json:"duration"`
}

// ToolLoop is the v2 research engine: LLM calls tools iteratively with tiered escalation.
type ToolLoop struct {
	tools     *ToolSet
	runner    *provider.TieredRunner
	telemetry *esearch.Telemetry
}

const maxToolCalls = 15

// stripThinkTags removes qwen3-style <think>...</think> blocks from LLM output.
var reThinkBlock = regexp.MustCompile(`(?s)<think>.*?</think>\s*`)

// stripCLINoise removes CLI tool preamble/footer from provider output.
// Handles codex ("OpenAI Codex v..."), gemini headers, and "tokens used\nN" footers.
var reCLIPreamble = regexp.MustCompile(`(?s)^.*?--------\n`)
var reTokensFooter = regexp.MustCompile(`(?s)\ntokens used\n[\d,]+\n.*$`)

func stripCLINoise(s string) string {
	// Strip codex/gemini-style preamble (everything up to last "--------\n")
	if idx := strings.LastIndex(s, "--------\n"); idx >= 0 {
		s = s[idx+len("--------\n"):]
	}
	// Strip "tokens used\nN\n..." footer
	s = reTokensFooter.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func cleanLLMResponse(s string) string {
	s = reThinkBlock.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	// Strip markdown code fences around JSON
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 3 && strings.HasPrefix(lines[len(lines)-1], "```") {
			s = strings.Join(lines[1:len(lines)-1], "\n")
			s = strings.TrimSpace(s)
		}
	}
	return s
}

// NewToolLoop creates a ToolLoop with the given tools, tiered runner, and telemetry sink.
func NewToolLoop(tools *ToolSet, runner *provider.TieredRunner, tel *esearch.Telemetry) *ToolLoop {
	return &ToolLoop{
		tools:     tools,
		runner:    runner,
		telemetry: tel,
	}
}

// buildSystemPrompt constructs the system prompt for the tool-use loop.
func buildSystemPrompt(doc ResearchDocument, goal Goal, tools []Tool) string {
	var sb strings.Builder

	sb.WriteString("## Input\n")
	sb.WriteString(fmt.Sprintf("Source: %s\n", doc.Source))
	sb.WriteString(fmt.Sprintf("Title: %s\n", doc.Title))
	if doc.Repo != "" {
		sb.WriteString(fmt.Sprintf("Repo: %s\n", doc.Repo))
	}
	sb.WriteString(fmt.Sprintf("\n%s\n", doc.Body))

	sb.WriteString("\n## Goal\n")
	switch goal {
	case GoalImplement:
		sb.WriteString("Produce an implementation plan with concrete file paths, code changes, and a patch if possible.\n")
	default:
		sb.WriteString("Produce a thorough summary of the issue, root cause analysis, and affected components.\n")
	}

	sb.WriteString("\n## Tools\n")
	sb.WriteString("You may call tools by responding with a single JSON object:\n")
	sb.WriteString("```\n{\"tool\": \"tool_name\", \"params\": {\"key\": \"value\"}}\n```\n\n")
	sb.WriteString("Available tools:\n")
	for _, t := range tools {
		sb.WriteString(fmt.Sprintf("- **%s**: %s  \n  Params: %s\n", t.Name, t.Description, t.Params))
	}

	sb.WriteString("\n## Rules\n")
	sb.WriteString("- Call tools first to gather evidence before answering.\n")
	sb.WriteString("- Cite sources as file:line where possible.\n")
	sb.WriteString("- Do not guess — use tools to verify.\n")
	sb.WriteString(fmt.Sprintf("- You have a budget of %d tool calls.\n", maxToolCalls))
	sb.WriteString("- When done, respond with your final output (not a JSON tool call).\n")

	return sb.String()
}

// Run executes the tool-use research loop: LLM calls tools iteratively until it
// produces a final text answer or the budget is exhausted.
func (tl *ToolLoop) Run(ctx context.Context, doc ResearchDocument, goal Goal) (LoopResult, error) {
	start := time.Now()
	runID := esearch.NewRunID()

	result := LoopResult{
		RunID:    runID,
		Document: doc,
		Goal:     goal,
	}

	systemPrompt := buildSystemPrompt(doc, goal, tl.tools.Definitions())
	conversation := []string{systemPrompt}

	toolCallCount := 0
	maxIterations := maxToolCalls + 5

	for i := 0; i < maxIterations; i++ {
		prompt := strings.Join(conversation, "\n\n")

		rr, err := tl.runner.Run(ctx, prompt, func(response string) provider.EscalationReason {
			cleaned := cleanLLMResponse(response)
			if cleaned == "" {
				return provider.ReasonEmpty
			}
			// If it doesn't look like JSON, accept as final output.
			if !strings.HasPrefix(cleaned, "{") {
				return ""
			}
			// Try to parse as a tool call.
			var tc ToolCall
			if err := json.Unmarshal([]byte(cleaned), &tc); err != nil {
				return provider.ReasonMalformed
			}
			if tc.Tool == "" {
				return provider.ReasonMalformed
			}
			if !tl.tools.ValidTool(tc.Tool) {
				return provider.ReasonHallucinated
			}
			return ""
		})
		if err != nil {
			return result, fmt.Errorf("toolloop: llm run: %w", err)
		}

		// Track metrics.
		result.LLMCalls++
		result.TokensIn += rr.TokensIn
		result.TokensOut += rr.TokensOut
		result.CostUSD += rr.CostUSD
		if rr.Tier > result.MaxTier {
			result.MaxTier = rr.Tier
		}
		if rr.Escalated {
			result.Escalations++
		}

		// Index LLM call telemetry.
		tl.telemetry.IndexLLMCall(ctx, esearch.LLMCallDoc{
			RunID:            runID,
			Step:             fmt.Sprintf("loop-%d", i),
			Tier:             rr.Tier,
			Provider:         rr.Provider,
			Model:            rr.Model,
			TokensIn:         int64(rr.TokensIn),
			TokensOut:        int64(rr.TokensOut),
			CostUSD:          rr.CostUSD,
			LatencyMS:        rr.Latency.Milliseconds(),
			Escalated:        rr.Escalated,
			EscalationReason: string(rr.EscalationReason),
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		})

		// Try to parse as tool call.
		cleaned := cleanLLMResponse(rr.Response)
		var tc ToolCall
		isToolCall := false
		if strings.HasPrefix(cleaned, "{") {
			if err := json.Unmarshal([]byte(cleaned), &tc); err == nil && tc.Tool != "" {
				isToolCall = true
			}
		}

		if isToolCall && toolCallCount < maxToolCalls {
			fmt.Fprintf(os.Stderr, "  > %s\n", tc.Tool)

			toolStart := time.Now()
			tr := tl.tools.Execute(ctx, tc.Tool, tc.Params())
			toolLatency := time.Since(toolStart)

			result.ToolCalls = append(result.ToolCalls, tr)
			toolCallCount++

			// Index tool call telemetry.
			inputSummary, _ := json.Marshal(tc.Params())
			tl.telemetry.IndexToolCall(ctx, esearch.ToolCallDoc{
				RunID:           runID,
				ToolName:        tc.Tool,
				InputSummary:    string(inputSummary),
				OutputSizeBytes: len(tr.Output),
				LatencyMS:       toolLatency.Milliseconds(),
				Success:         tr.Err == "",
				Timestamp:       time.Now().UTC().Format(time.RFC3339),
			})

			// Append tool result to conversation.
			toolOutput := tr.Output
			if tr.Err != "" {
				toolOutput = fmt.Sprintf("ERROR: %s", tr.Err)
			}
			conversation = append(conversation, fmt.Sprintf("[tool: %s]\n%s", tc.Tool, toolOutput))
			continue
		}

		if isToolCall && toolCallCount >= maxToolCalls {
			conversation = append(conversation, "Tool budget exhausted. Produce your final output now.")
			continue
		}

		// Not a tool call — final output. Strip CLI preamble/footer noise.
		result.Output = stripCLINoise(cleaned)
		break
	}

	result.Duration = time.Since(start)

	// Index research run telemetry.
	tl.telemetry.IndexResearchRun(ctx, esearch.ResearchRunDoc{
		RunID:          runID,
		InputSource:    doc.Source,
		SourceURL:      doc.SourceURL,
		Goal:           string(goal),
		TotalToolCalls: toolCallCount,
		TotalLLMCalls:  result.LLMCalls,
		TotalTokensIn:  int64(result.TokensIn),
		TotalTokensOut: int64(result.TokensOut),
		TotalCostUSD:   result.CostUSD,
		DurationMS:     result.Duration.Milliseconds(),
		FinalTierUsed:  result.MaxTier,
		EscalationCnt:  result.Escalations,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	})

	return result, nil
}
