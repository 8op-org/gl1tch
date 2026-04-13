package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/8op-org/gl1tch/internal/esearch"
)

// ESActivityResearcher queries glitch event, pipeline, and summary indices.
type ESActivityResearcher struct {
	client *esearch.Client
}

// NewESActivityResearcher creates a new ESActivityResearcher.
func NewESActivityResearcher(client *esearch.Client) *ESActivityResearcher {
	return &ESActivityResearcher{client: client}
}

func (e *ESActivityResearcher) Name() string { return "es-activity" }
func (e *ESActivityResearcher) Describe() string {
	return "indexed activity from git commits, GitHub PRs, issues, and pipeline runs"
}

func (e *ESActivityResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	queryJSON := json.RawMessage(fmt.Sprintf(`{
		"size": 20,
		"sort": [{"timestamp": {"order": "desc"}}],
		"query": {
			"multi_match": {
				"query": %s,
				"fields": ["message", "body", "summary", "name"]
			}
		}
	}`, jsonString(q.Question)))

	indices := []string{esearch.IndexEvents, esearch.IndexPipelines, esearch.IndexSummaries}
	resp, err := e.client.Search(ctx, indices, queryJSON)
	if err != nil {
		return Evidence{}, fmt.Errorf("es-activity gather: %w", err)
	}

	var sb strings.Builder
	for _, hit := range resp.Results {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", hit.Index, string(hit.Source)))
	}

	return Evidence{
		Source: "es-activity",
		Title:  "es-activity",
		Body:   sb.String(),
	}, nil
}

// ESCodeResearcher queries glitch-code-* indices for source code.
type ESCodeResearcher struct {
	client *esearch.Client
}

// NewESCodeResearcher creates a new ESCodeResearcher.
func NewESCodeResearcher(client *esearch.Client) *ESCodeResearcher {
	return &ESCodeResearcher{client: client}
}

func (e *ESCodeResearcher) Name() string { return "es-code" }
func (e *ESCodeResearcher) Describe() string {
	return "indexed source code from repositories — functions, types, and file contents"
}

func (e *ESCodeResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	queryJSON := json.RawMessage(fmt.Sprintf(`{
		"size": 10,
		"query": {
			"bool": {
				"should": [
					{"match": {"symbols": %s}},
					{"match": {"content": %s}}
				]
			}
		}
	}`, jsonString(q.Question), jsonString(q.Question)))

	resp, err := e.client.Search(ctx, []string{"glitch-code-*"}, queryJSON)
	if err != nil {
		return Evidence{}, fmt.Errorf("es-code gather: %w", err)
	}

	var sb strings.Builder
	for _, hit := range resp.Results {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", hit.Index, string(hit.Source)))
	}

	return Evidence{
		Source: "es-code",
		Title:  "es-code",
		Body:   sb.String(),
	}, nil
}

// jsonString returns s as a JSON string literal.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
