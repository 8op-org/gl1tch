package observer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
)

// LLMFunc generates text from a prompt. The observer doesn't care how — it
// could be Ollama, Copilot, Claude, or any other provider.
type LLMFunc func(prompt string) (string, error)

// EdgeHit represents a single edge in the code graph.
type EdgeHit struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Kind     string `json:"kind"`
}

// graphBFS performs breadth-first traversal on edges from a starting symbol.
func graphBFS(edgeFetcher func(symbolID, edgeKind string) []EdgeHit, startID, edgeKind string, depth int) []string {
	if edgeFetcher == nil || depth < 1 {
		return nil
	}
	visited := map[string]bool{startID: true}
	frontier := []string{startID}
	var collected []string
	for d := 0; d < depth && len(frontier) > 0; d++ {
		var next []string
		for _, id := range frontier {
			edges := edgeFetcher(id, edgeKind)
			for _, e := range edges {
				targetID := e.TargetID
				if e.TargetID == id {
					targetID = e.SourceID
				}
				if !visited[targetID] {
					visited[targetID] = true
					next = append(next, targetID)
					collected = append(collected, targetID)
				}
			}
		}
		frontier = next
	}
	return collected
}

// QueryEngine bridges natural language questions to Elasticsearch and
// synthesizes answers via an LLM.
type QueryEngine struct {
	es    *esearch.Client
	llm   LLMFunc
	repo  string
	depth int
}

// WithRepo returns a copy of the engine scoped to a specific repository.
func (q *QueryEngine) WithRepo(repo string) *QueryEngine {
	return &QueryEngine{es: q.es, llm: q.llm, repo: repo, depth: q.depth}
}

// WithDepth sets the BFS traversal depth for graph queries.
func (q *QueryEngine) WithDepth(d int) *QueryEngine {
	q.depth = d
	return q
}

// NewQueryEngine returns a new QueryEngine using the given LLM function.
func NewQueryEngine(es *esearch.Client, llm LLMFunc) *QueryEngine {
	return &QueryEngine{es: es, llm: llm}
}

func allIndices() []string {
	return []string{
		esearch.IndexEvents,
		esearch.IndexResearchRuns,
		esearch.IndexToolCalls,
		esearch.IndexLLMCalls,
	}
}

func allIndicesForRepo(repo string) []string {
	base := allIndices()
	if repo != "" {
		base = append(base,
			esearch.IndexSymbolsPrefix+repo,
			esearch.IndexEdgesPrefix+repo,
		)
	}
	return base
}

// knowledgeIndex returns the knowledge index name for a repo (e.g. "elastic/oblt-cli" → "glitch-knowledge-oblt-cli").
func knowledgeIndex(repo string) string {
	parts := strings.SplitN(repo, "/", 2)
	name := repo
	if len(parts) == 2 {
		name = parts[1]
	}
	return esearch.IndexKnowledgePrefix + name
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
		esQuery = defaultQueryWithRepo(question, q.repo)
	}

	raw, err := json.Marshal(esQuery)
	if err != nil {
		raw, _ = json.Marshal(defaultQueryWithRepo(question, q.repo))
	}

	indices := allIndicesForRepo(q.repo)
	resp, err := q.es.Search(ctx, indices, json.RawMessage(raw))
	if err != nil {
		// Retry with the plain default query
		fallbackRaw, _ := json.Marshal(defaultQueryWithRepo(question, q.repo))
		resp, err = q.es.Search(ctx, indices, json.RawMessage(fallbackRaw))
		if err != nil {
			// Index-not-found or similar — return empty results
			resp = &esearch.SearchResponse{}
		}
	}

	// Also search knowledge indices when a repo is scoped.
	if q.repo != "" {
		kResp, kErr := q.searchKnowledge(ctx, question)
		if kErr == nil && kResp != nil {
			resp = mergeResponses(resp, kResp)
		}
	}

	return resp, nil
}

// searchKnowledge queries the knowledge index for the scoped repo using
// hybrid BM25 + summary boosting.
func (q *QueryEngine) searchKnowledge(ctx context.Context, question string) (*esearch.SearchResponse, error) {
	idx := knowledgeIndex(q.repo)
	query := map[string]any{
		"size": 10,
		"query": map[string]any{
			"bool": map[string]any{
				"should": []any{
					map[string]any{
						"term": map[string]any{
							"type": map[string]any{"value": "summary", "boost": 3},
						},
					},
					map[string]any{
						"multi_match": map[string]any{
							"query":  question,
							"fields": []string{"content", "title", "path"},
						},
					},
				},
			},
		},
	}

	raw, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	return q.es.Search(ctx, []string{idx}, json.RawMessage(raw))
}

// mergeResponses combines two search responses. Results from different indices
// so no dedup needed.
func mergeResponses(a, b *esearch.SearchResponse) *esearch.SearchResponse {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	a.Results = append(a.Results, b.Results...)
	a.Total += b.Total
	return a
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

// defaultQueryWithRepo returns a default query scoped to a repo when non-empty.
func defaultQueryWithRepo(question, repo string) map[string]any {
	q := defaultQuery(question)
	if repo == "" {
		return q
	}

	// Add a bool filter for repo/source fields.
	boolClause := q["query"].(map[string]any)["bool"].(map[string]any)
	boolClause["filter"] = []any{
		map[string]any{
			"bool": map[string]any{
				"should": []any{
					map[string]any{"term": map[string]any{"repo": repo}},
					map[string]any{"term": map[string]any{"source": repo}},
				},
				"minimum_should_match": 1,
			},
		},
	}
	return q
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
- glitch-knowledge-*: type (summary/doc/architecture/pattern/pr_insight/decision), title, content, path, repo, embedding (dense_vector), timestamp

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

	raw, err := q.llmGenerate(prompt)
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

	// Build optional context sections.
	var contextLines []string

	// Index stats (best-effort — don't fail the whole answer if ES is flaky).
	if stats, err := q.es.IndexStats(ctx); err == nil && len(stats) > 0 {
		var sb strings.Builder
		sb.WriteString("Available indices:\n")
		for _, s := range stats {
			sb.WriteString(fmt.Sprintf("  %s  docs=%s  size=%s\n", s.Index, s.DocCount, s.StoreSize))
		}
		contextLines = append(contextLines, sb.String())
	}

	if q.repo != "" {
		contextLines = append(contextLines, fmt.Sprintf("Repo context: %s", q.repo))
	}

	contextBlock := ""
	if len(contextLines) > 0 {
		contextBlock = strings.Join(contextLines, "\n") + "\n\n"
	}

	prompt := fmt.Sprintf(`You are an observability assistant. Answer the question based ONLY on the data below.

Rules:
- Be direct and concise.
- Cite specific repos, timestamps, or pipeline names from the data.
- Never fabricate information not present in the data.
- Only reference data shown below.
- If results come from glitch-knowledge-* indices, treat them as curated knowledge (architecture, patterns, guides) — prioritize these over raw event data.
- If the data doesn't contain what was asked about, say so clearly and suggest what to index.

%sObserved data:
%s

Question: %s

Answer:`, contextBlock, formatted, question)

	answer, err := q.llmGenerate(prompt)
	if err != nil {
		return "", fmt.Errorf("synthesize: %w", err)
	}
	return strings.TrimSpace(answer), nil
}

// llmGenerate sends a prompt to the configured LLM and returns the response.
func (q *QueryEngine) llmGenerate(prompt string) (string, error) {
	return q.llm(prompt)
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

