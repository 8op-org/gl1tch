package observer

import (
	"encoding/json"
	"strings"
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
	called := false
	llm := func(prompt string) (string, error) {
		called = true
		return "test response", nil
	}
	qe := NewQueryEngine(es, llm)
	if qe.es == nil {
		t.Error("NewQueryEngine: es should not be nil")
	}
	if qe.llm == nil {
		t.Error("NewQueryEngine: llm should not be nil")
	}
	// Verify the function is wired correctly
	_, _ = qe.llm("test")
	if !called {
		t.Error("NewQueryEngine: llm function was not called")
	}
}

func TestDefaultQueryWithRepo(t *testing.T) {
	q := defaultQueryWithRepo("what broke", "elastic/kibana")
	raw, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	s := string(raw)

	// Must contain the repo filter term.
	if !contains(s, `"repo":"elastic/kibana"`) {
		t.Errorf("expected repo filter, got: %s", s)
	}
	if !contains(s, `"source":"elastic/kibana"`) {
		t.Errorf("expected source filter, got: %s", s)
	}
}

func TestDefaultQueryWithRepoEmpty(t *testing.T) {
	withRepo := defaultQueryWithRepo("what broke", "")
	plain := defaultQuery("what broke")

	a, _ := json.Marshal(withRepo)
	b, _ := json.Marshal(plain)

	if string(a) != string(b) {
		t.Errorf("empty repo should equal plain default query\ngot:  %s\nwant: %s", a, b)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 && strings.Contains(haystack, needle)
}

func TestAllIndicesForRepo(t *testing.T) {
	indices := allIndicesForRepo("gl1tch")
	hasSymbols := false
	hasEdges := false
	for _, idx := range indices {
		if idx == esearch.IndexSymbolsPrefix+"gl1tch" {
			hasSymbols = true
		}
		if idx == esearch.IndexEdgesPrefix+"gl1tch" {
			hasEdges = true
		}
	}
	if !hasSymbols {
		t.Error("missing symbols index")
	}
	if !hasEdges {
		t.Error("missing edges index")
	}
}

func TestAllIndicesForRepoEmpty(t *testing.T) {
	indices := allIndicesForRepo("")
	// Should equal allIndices() — no graph indices added
	base := allIndices()
	if len(indices) != len(base) {
		t.Errorf("empty repo: got %d indices, want %d", len(indices), len(base))
	}
}

func TestGraphBFSEmpty(t *testing.T) {
	results := graphBFS(nil, "sym1", "calls", 2)
	if len(results) != 0 {
		t.Errorf("nil fetcher: got %d, want 0", len(results))
	}
}

func TestGraphBFSTraversal(t *testing.T) {
	edges := map[string][]EdgeHit{
		"a": {{SourceID: "a", TargetID: "b", Kind: "calls"}, {SourceID: "a", TargetID: "c", Kind: "calls"}},
		"b": {{SourceID: "b", TargetID: "d", Kind: "calls"}},
	}
	fetcher := func(id, kind string) []EdgeHit { return edges[id] }

	result := graphBFS(fetcher, "a", "calls", 1)
	if len(result) != 2 {
		t.Errorf("depth 1: got %d, want 2", len(result))
	}

	result2 := graphBFS(fetcher, "a", "calls", 2)
	if len(result2) != 3 {
		t.Errorf("depth 2: got %d, want 3", len(result2))
	}
}

func TestGraphBFSZeroDepth(t *testing.T) {
	fetcher := func(id, kind string) []EdgeHit {
		return []EdgeHit{{SourceID: "a", TargetID: "b", Kind: "calls"}}
	}
	results := graphBFS(fetcher, "a", "calls", 0)
	if len(results) != 0 {
		t.Errorf("zero depth: got %d, want 0", len(results))
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
