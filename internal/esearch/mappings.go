package esearch

const (
	IndexEvents          = "glitch-events"
	IndexResearchRuns    = "glitch-research-runs"
	IndexToolCalls       = "glitch-tool-calls"
	IndexLLMCalls        = "glitch-llm-calls"
	IndexWorkflowRuns    = "glitch-workflow-runs"
	IndexCrossReviews    = "glitch-cross-reviews"
	IndexRuns            = "glitch-runs"
	IndexKnowledgePrefix = "glitch-knowledge-"
)

const EventsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "type":      { "type": "keyword" },
      "source":    { "type": "keyword" },
      "repo":      { "type": "keyword" },
      "author":    { "type": "keyword" },
      "message":   { "type": "text" },
      "body":      { "type": "text" },
      "metadata":  { "type": "object", "enabled": false },
      "timestamp": { "type": "date" }
    }
  }
}`

const ResearchRunsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":           { "type": "keyword" },
      "input_source":     { "type": "keyword" },
      "source_url":       { "type": "keyword" },
      "goal":             { "type": "keyword" },
      "total_tool_calls": { "type": "integer" },
      "total_llm_calls":  { "type": "integer" },
      "total_tokens_in":  { "type": "long" },
      "total_tokens_out": { "type": "long" },
      "total_cost_usd":   { "type": "float" },
      "duration_ms":      { "type": "long" },
      "final_tier_used":  { "type": "integer" },
      "escalation_count": { "type": "integer" },
      "confidence_pass":  { "type": "boolean" },
      "timestamp":        { "type": "date" }
    }
  }
}`

const ToolCallsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":            { "type": "keyword" },
      "tool_name":         { "type": "keyword" },
      "input_summary":     { "type": "text" },
      "output_size_bytes": { "type": "integer" },
      "latency_ms":        { "type": "long" },
      "success":           { "type": "boolean" },
      "timestamp":         { "type": "date" }
    }
  }
}`

const LLMCallsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":             { "type": "keyword" },
      "step":               { "type": "keyword" },
      "tier":               { "type": "integer" },
      "provider":           { "type": "keyword" },
      "model":              { "type": "keyword" },
      "tokens_in":          { "type": "long" },
      "tokens_out":         { "type": "long" },
      "tokens_total":       { "type": "long" },
      "cost_usd":           { "type": "float" },
      "latency_ms":         { "type": "long" },
      "escalated":          { "type": "boolean" },
      "escalation_reason":  { "type": "keyword" },
      "workflow_name":      { "type": "keyword" },
      "issue":              { "type": "keyword" },
      "comparison_group":   { "type": "keyword" },
      "timestamp":          { "type": "date" }
    }
  }
}`

const WorkflowRunsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":           { "type": "keyword" },
      "workflow_name":    { "type": "keyword" },
      "issue":            { "type": "keyword" },
      "comparison_group": { "type": "keyword" },
      "total_steps":      { "type": "integer" },
      "llm_steps":        { "type": "integer" },
      "total_tokens_in":  { "type": "long" },
      "total_tokens_out": { "type": "long" },
      "total_cost_usd":   { "type": "float" },
      "total_latency_ms": { "type": "long" },
      "review_pass":      { "type": "boolean" },
      "confidence":       { "type": "float" },
      "criteria_passed":  { "type": "integer" },
      "criteria_total":   { "type": "integer" },
      "timestamp":        { "type": "date" }
    }
  }
}`

const CrossReviewsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":          { "type": "keyword" },
      "issue":           { "type": "keyword" },
      "iteration":       { "type": "keyword" },
      "variant":         { "type": "keyword" },
      "passed":          { "type": "integer" },
      "total":           { "type": "integer" },
      "confidence":      { "type": "float" },
      "winner":          { "type": "boolean" },
      "workflow_name":   { "type": "keyword" },
      "timestamp":       { "type": "date" },
      "compare_id":      { "type": "keyword" },
      "scope":           { "type": "keyword" },
      "criteria_name":   { "type": "keyword" },
      "criteria_score":  { "type": "integer" },
      "workspace":       { "type": "keyword" }
    }
  }
}`

const RunsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "run_id":        { "type": "keyword" },
      "workflow_name": { "type": "keyword" },
      "workspace":     { "type": "keyword" },
      "source":        { "type": "keyword" },
      "status":        { "type": "keyword" },
      "has_compare":   { "type": "boolean" },
      "duration_ms":   { "type": "long" },
      "timestamp":     { "type": "date" }
    }
  }
}`

// AllIndices returns a map of index name → mapping JSON for all managed indices.
func AllIndices() map[string]string {
	return map[string]string{
		IndexEvents:       EventsMapping,
		IndexResearchRuns: ResearchRunsMapping,
		IndexToolCalls:    ToolCallsMapping,
		IndexLLMCalls:     LLMCallsMapping,
		IndexWorkflowRuns:  WorkflowRunsMapping,
		IndexCrossReviews:  CrossReviewsMapping,
		IndexRuns:          RunsMapping,
	}
}
