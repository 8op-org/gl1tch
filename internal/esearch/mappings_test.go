package esearch

import (
	"encoding/json"
	"testing"
)

func TestSymbolsMappingIsValidJSON(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(SymbolsMapping), &m); err != nil {
		t.Fatalf("SymbolsMapping is not valid JSON: %v", err)
	}

	props := extractProperties(t, m)

	expected := map[string]string{
		"id":         "keyword",
		"file":       "keyword",
		"kind":       "keyword",
		"language":   "keyword",
		"start_line": "integer",
		"end_line":   "integer",
		"parent_id":  "keyword",
		"file_hash":  "keyword",
		"repo":       "keyword",
		"indexed_at": "date",
		"signature":  "text",
		"docstring":  "text",
	}
	for field, typ := range expected {
		assertFieldType(t, props, field, typ)
	}

	// name should be text with a keyword sub-field "raw"
	nameRaw, ok := props["name"]
	if !ok {
		t.Fatal("SymbolsMapping missing field: name")
	}
	nameMap := nameRaw.(map[string]any)
	if nameMap["type"] != "text" {
		t.Fatalf("name: expected type text, got %v", nameMap["type"])
	}
	fields, ok := nameMap["fields"].(map[string]any)
	if !ok {
		t.Fatal("name: missing fields sub-object")
	}
	rawField, ok := fields["raw"].(map[string]any)
	if !ok {
		t.Fatal("name: missing fields.raw sub-object")
	}
	if rawField["type"] != "keyword" {
		t.Fatalf("name.raw: expected type keyword, got %v", rawField["type"])
	}
}

func TestEdgesMappingIsValidJSON(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(EdgesMapping), &m); err != nil {
		t.Fatalf("EdgesMapping is not valid JSON: %v", err)
	}

	props := extractProperties(t, m)

	expected := map[string]string{
		"source_id": "keyword",
		"target_id": "keyword",
		"kind":      "keyword",
		"file":      "keyword",
		"repo":      "keyword",
	}
	for field, typ := range expected {
		assertFieldType(t, props, field, typ)
	}
}

func TestIndexPrefixConstants(t *testing.T) {
	if IndexSymbolsPrefix != "glitch-symbols-" {
		t.Fatalf("IndexSymbolsPrefix = %q, want %q", IndexSymbolsPrefix, "glitch-symbols-")
	}
	if IndexEdgesPrefix != "glitch-edges-" {
		t.Fatalf("IndexEdgesPrefix = %q, want %q", IndexEdgesPrefix, "glitch-edges-")
	}
}

func TestAllIndicesExcludesPrefixes(t *testing.T) {
	all := AllIndices()
	// Per-repo graph indices are created dynamically; prefix keys should
	// not appear in AllIndices (they aren't valid index names).
	if _, ok := all[IndexSymbolsPrefix]; ok {
		t.Fatal("AllIndices should not contain IndexSymbolsPrefix")
	}
	if _, ok := all[IndexEdgesPrefix]; ok {
		t.Fatal("AllIndices should not contain IndexEdgesPrefix")
	}
	// Verify global indices are present.
	for _, idx := range []string{IndexEvents, IndexResearchRuns, IndexToolCalls, IndexLLMCalls, IndexWorkflowRuns, IndexCrossReviews, IndexRuns} {
		if _, ok := all[idx]; !ok {
			t.Errorf("AllIndices missing %s", idx)
		}
	}
}

// helpers

func extractProperties(t *testing.T, m map[string]any) map[string]any {
	t.Helper()
	mappings, ok := m["mappings"].(map[string]any)
	if !ok {
		t.Fatal("missing top-level mappings key")
	}
	props, ok := mappings["properties"].(map[string]any)
	if !ok {
		t.Fatal("missing mappings.properties key")
	}
	return props
}

func assertFieldType(t *testing.T, props map[string]any, field, expectedType string) {
	t.Helper()
	raw, ok := props[field]
	if !ok {
		t.Fatalf("missing field: %s", field)
	}
	fm := raw.(map[string]any)
	if fm["type"] != expectedType {
		t.Fatalf("%s: expected type %s, got %v", field, expectedType, fm["type"])
	}
}
