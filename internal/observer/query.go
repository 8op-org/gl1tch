package observer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
)

const defaultModel = "qwen2.5:7b"

// QueryEngine bridges natural language questions to Elasticsearch and
// synthesizes answers via a local Ollama model.
type QueryEngine struct {
	es    *esearch.Client
	model string
}

// NewQueryEngine returns a new QueryEngine. If model is empty, defaultModel is used.
func NewQueryEngine(es *esearch.Client, model string) *QueryEngine {
	if model == "" {
		model = defaultModel
	}
	return &QueryEngine{es: es, model: model}
}

func allIndices() []string {
	return []string{
		esearch.IndexEvents,
		esearch.IndexSummaries,
		esearch.IndexPipelines,
		esearch.IndexInsights,
	}
}

// Answer takes a natural language question, searches Elasticsearch, and
// returns a synthesized answer from the local LLM.
func (q *QueryEngine) Answer(ctx context.Context, question string) (string, error) {
	results, err := q.searchWithFallback(ctx, question)
	if err != nil {
		return "", fmt.Errorf("observer: search: %w", err)
	}
	if results.Total == 0 && len(results.Results) == 0 {
		return "I don't have any indexed data matching that question yet. Try running a pipeline or indexing some data first.", nil
	}
	return q.synthesize(ctx, question, results)
}

// searchWithFallback tries to generate an ES query via LLM; falls back to
// defaultQuery on any error or if the LLM-generated query fails to search.
func (q *QueryEngine) searchWithFallback(ctx context.Context, question string) (*esearch.SearchResponse, error) {
	esQuery, err := q.generateQuery(ctx, question)
	if err != nil {
		esQuery = defaultQuery(question)
	}

	raw, err := json.Marshal(esQuery)
	if err != nil {
		raw, _ = json.Marshal(defaultQuery(question))
	}

	resp, err := q.es.Search(ctx, allIndices(), json.RawMessage(raw))
	if err != nil {
		// Retry with the plain default query
		fallbackRaw, _ := json.Marshal(defaultQuery(question))
		resp, err = q.es.Search(ctx, allIndices(), json.RawMessage(fallbackRaw))
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// defaultQuery returns a safe multi-match query that works against all indices.
func defaultQuery(question string) map[string]any {
	return map[string]any{
		"size": 20,
		"sort": []any{
			map[string]any{
				"timestamp": map[string]any{
					"order":        "desc",
					"unmapped_type": "date",
				},
			},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"should": []any{
					map[string]any{
						"multi_match": map[string]any{
							"query":  question,
							"fields": []string{"message", "body", "summary", "pattern", "stdout", "name"},
							"type":   "best_fields",
						},
					},
				},
				"minimum_should_match": 1,
			},
		},
	}
}

// generateQuery asks the LLM to produce an Elasticsearch query JSON for the
// given question. Falls back to defaultQuery on any failure.
func (q *QueryEngine) generateQuery(ctx context.Context, question string) (map[string]any, error) {
	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)

	prompt := fmt.Sprintf(`You are an Elasticsearch query builder. Produce ONLY valid JSON — no explanation, no markdown fences.

Available indices and their key fields:
- glitch-events:     type, source, repo, author, message, body, timestamp
- glitch-summaries:  scope, date, summary, timestamp
- glitch-pipelines:  name, status, stdout, model, provider, timestamp
- glitch-insights:   type, pattern, recommendation, timestamp

Today: %s
One week ago: %s

Rules:
1. Return a single valid ES query JSON object (size, sort, query).
2. Use range filter on timestamp when the question implies recency.
3. Use multi_match or match_phrase for text fields.
4. Always include unmapped_type: "date" on timestamp sort.
5. Limit size to 20.

Question: %s

JSON:`, now.Format(time.RFC3339), weekAgo.Format(time.RFC3339), question)

	raw, err := ollamaGenerate(ctx, q.model, prompt)
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("generateQuery: no JSON found in LLM response")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("generateQuery: parse JSON: %w", err)
	}
	return result, nil
}

// synthesize formats search results and asks the LLM to answer the question.
func (q *QueryEngine) synthesize(ctx context.Context, question string, results *esearch.SearchResponse) (string, error) {
	formatted := formatResults(results)

	prompt := fmt.Sprintf(`You are an observability assistant. Answer the question based ONLY on the data below.

Rules:
- Be direct and concise.
- Cite specific repos, timestamps, or pipeline names from the data.
- Never fabricate information not present in the data.
- Only reference data shown below.

Observed data:
%s

Question: %s

Answer:`, formatted, question)

	answer, err := ollamaGenerate(ctx, q.model, prompt)
	if err != nil {
		return "", fmt.Errorf("synthesize: %w", err)
	}
	return strings.TrimSpace(answer), nil
}

// formatResults formats up to 15 search hits as readable text.
func formatResults(results *esearch.SearchResponse) string {
	if results == nil || len(results.Results) == 0 {
		return "(no results found)"
	}

	hits := results.Results
	if len(hits) > 15 {
		hits = hits[:15]
	}

	var sb strings.Builder
	for _, h := range hits {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", h.Index, string(h.Source)))
	}
	return sb.String()
}

// extractJSON strips markdown fences and extracts the first complete JSON object.
func extractJSON(s string) string {
	// Strip markdown code fences
	s = strings.ReplaceAll(s, "```json", "")
	s = strings.ReplaceAll(s, "```", "")

	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end < start {
		return ""
	}
	return s[start : end+1]
}

// ollamaGenerate calls the local Ollama API with stream:false and returns the
// response field.
func ollamaGenerate(ctx context.Context, model, prompt string) (string, error) {
	payload, err := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})
	if err != nil {
		return "", fmt.Errorf("ollamaGenerate: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:11434/api/generate", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("ollamaGenerate: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollamaGenerate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollamaGenerate: status %s — %s", resp.Status, string(body))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollamaGenerate: decode: %w", err)
	}
	return result.Response, nil
}
