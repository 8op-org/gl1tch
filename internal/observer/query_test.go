package observer

import (
	"encoding/json"
	"testing"

	"github.com/8op-org/gl1tch/internal/esearch"
)

func TestDefaultQuery(t *testing.T) {
	q := defaultQuery("what broke today")
	raw, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("defaultQuery: marshal error: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("defaultQuery: marshaled to empty JSON")
	}

	// Validate it round-trips back to a map
	var back map[string]any
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("defaultQuery: round-trip unmarshal: %v", err)
	}
	if _, ok := back["query"]; !ok {
		t.Error("defaultQuery: missing 'query' key")
	}
}

func TestExtractJSON(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`{"query":{}}`, `{"query":{}}`},
		{"```json\n{\"query\":{}}\n```", `{"query":{}}`},
		{`some text {"query":{}} more text`, `{"query":{}}`},
	}

	for _, c := range cases {
		got := extractJSON(c.input)
		if got != c.want {
			t.Errorf("extractJSON(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestFormatResultsEmpty(t *testing.T) {
	got := formatResults(nil)
	if got != "(no results found)" {
		t.Errorf("formatResults(nil) = %q, want \"(no results found)\"", got)
	}

	got = formatResults(&esearch.SearchResponse{})
	if got != "(no results found)" {
		t.Errorf("formatResults(empty) = %q, want \"(no results found)\"", got)
	}
}

func TestFormatResultsCap(t *testing.T) {
	// Build 20 dummy hits — only 15 should appear in the output
	hits := make([]esearch.SearchResult, 20)
	for i := range hits {
		hits[i] = esearch.SearchResult{
			Index:  "glitch-events",
			Source: json.RawMessage(`{"message":"test"}`),
		}
	}
	resp := &esearch.SearchResponse{Total: 20, Results: hits}
	out := formatResults(resp)

	// Count occurrences of "[glitch-events]"
	count := 0
	for i := 0; i < len(out); i++ {
		if i+len("[glitch-events]") <= len(out) && out[i:i+len("[glitch-events]")] == "[glitch-events]" {
			count++
		}
	}
	if count != 15 {
		t.Errorf("formatResults capped at %d hits, want 15", count)
	}
}

func TestNewQueryEngine(t *testing.T) {
	es := esearch.NewClient("http://localhost:9200")
	qe := NewQueryEngine(es, "")
	if qe.model != defaultModel {
		t.Errorf("NewQueryEngine with empty model: got %q, want %q", qe.model, defaultModel)
	}

	qe2 := NewQueryEngine(es, "llama3:8b")
	if qe2.model != "llama3:8b" {
		t.Errorf("NewQueryEngine with explicit model: got %q, want %q", qe2.model, "llama3:8b")
	}
}

func TestAllIndices(t *testing.T) {
	indices := allIndices()
	want := []string{
		esearch.IndexEvents,
		esearch.IndexResearchRuns,
		esearch.IndexToolCalls,
		esearch.IndexLLMCalls,
	}
	if len(indices) != len(want) {
		t.Fatalf("allIndices: got %d indices, want %d", len(indices), len(want))
	}
	for i, idx := range indices {
		if idx != want[i] {
			t.Errorf("allIndices[%d] = %q, want %q", i, idx, want[i])
		}
	}
}
