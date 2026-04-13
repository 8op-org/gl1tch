package esearch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// ResearchRunDoc represents a single research run for ES indexing.
type ResearchRunDoc struct {
	RunID          string  `json:"run_id"`
	InputSource    string  `json:"input_source"`
	SourceURL      string  `json:"source_url"`
	Goal           string  `json:"goal"`
	TotalToolCalls int     `json:"total_tool_calls"`
	TotalLLMCalls  int     `json:"total_llm_calls"`
	TotalTokensIn  int64   `json:"total_tokens_in"`
	TotalTokensOut int64   `json:"total_tokens_out"`
	TotalCostUSD   float64 `json:"total_cost_usd"`
	DurationMS     int64   `json:"duration_ms"`
	FinalTierUsed  int     `json:"final_tier_used"`
	EscalationCnt  int     `json:"escalation_count"`
	ConfidencePass bool    `json:"confidence_pass"`
	Timestamp      string  `json:"timestamp"`
}

// ToolCallDoc represents a single tool call for ES indexing.
type ToolCallDoc struct {
	RunID           string `json:"run_id"`
	ToolName        string `json:"tool_name"`
	InputSummary    string `json:"input_summary"`
	OutputSizeBytes int    `json:"output_size_bytes"`
	LatencyMS       int64  `json:"latency_ms"`
	Success         bool   `json:"success"`
	Timestamp       string `json:"timestamp"`
}

// LLMCallDoc represents a single LLM call for ES indexing.
type LLMCallDoc struct {
	RunID            string  `json:"run_id"`
	Step             string  `json:"step"`
	Tier             int     `json:"tier"`
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	TokensIn         int64   `json:"tokens_in"`
	TokensOut        int64   `json:"tokens_out"`
	CostUSD          float64 `json:"cost_usd"`
	LatencyMS        int64   `json:"latency_ms"`
	Escalated        bool    `json:"escalated"`
	EscalationReason string  `json:"escalation_reason"`
	Timestamp        string  `json:"timestamp"`
}

// Telemetry provides nil-safe methods for indexing research telemetry into ES.
type Telemetry struct {
	client *Client
}

// NewTelemetry returns a Telemetry instance, or nil if client is nil.
func NewTelemetry(client *Client) *Telemetry {
	if client == nil {
		return nil
	}
	return &Telemetry{client: client}
}

// NewRunID generates a unique run identifier: run-<unix_nano>-<8_random_hex_bytes>.
func NewRunID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("run-%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

// EnsureIndices creates all telemetry indices if they don't exist.
func (t *Telemetry) EnsureIndices(ctx context.Context) error {
	if t == nil {
		return nil
	}
	for idx, mapping := range AllIndices() {
		if err := t.client.EnsureIndex(ctx, idx, mapping); err != nil {
			return fmt.Errorf("telemetry: ensure %s: %w", idx, err)
		}
	}
	return nil
}

// IndexResearchRun indexes a research run document.
func (t *Telemetry) IndexResearchRun(ctx context.Context, doc ResearchRunDoc) error {
	if t == nil {
		return nil
	}
	return t.indexDoc(ctx, IndexResearchRuns, doc.RunID, doc)
}

// IndexToolCall indexes a tool call document.
func (t *Telemetry) IndexToolCall(ctx context.Context, doc ToolCallDoc) error {
	if t == nil {
		return nil
	}
	return t.indexDoc(ctx, IndexToolCalls, "", doc)
}

// IndexLLMCall indexes an LLM call document.
func (t *Telemetry) IndexLLMCall(ctx context.Context, doc LLMCallDoc) error {
	if t == nil {
		return nil
	}
	return t.indexDoc(ctx, IndexLLMCalls, "", doc)
}

func (t *Telemetry) indexDoc(ctx context.Context, index, id string, doc any) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("telemetry: marshal: %w", err)
	}
	if id == "" {
		id = NewRunID() // use as a unique doc ID
	}
	return t.client.BulkIndex(ctx, index, []BulkDoc{{ID: id, Body: body}})
}
