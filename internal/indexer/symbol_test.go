package indexer

import (
	"encoding/json"
	"testing"
)

func TestSymbolID(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		a := SymbolID("main.go", "function", "Foo", 10)
		b := SymbolID("main.go", "function", "Foo", 10)
		if a != b {
			t.Fatalf("same inputs produced different IDs: %q vs %q", a, b)
		}
	})

	t.Run("not empty", func(t *testing.T) {
		id := SymbolID("main.go", "function", "Foo", 10)
		if id == "" {
			t.Fatal("SymbolID returned empty string")
		}
	})

	t.Run("different inputs differ", func(t *testing.T) {
		a := SymbolID("main.go", "function", "Foo", 10)
		b := SymbolID("main.go", "function", "Bar", 10)
		c := SymbolID("other.go", "function", "Foo", 10)
		d := SymbolID("main.go", "method", "Foo", 10)
		e := SymbolID("main.go", "function", "Foo", 20)

		ids := []string{a, b, c, d, e}
		seen := make(map[string]bool)
		for _, id := range ids {
			if seen[id] {
				t.Fatalf("collision found: %q appears more than once in %v", id, ids)
			}
			seen[id] = true
		}
	})

	t.Run("length is 24 hex chars", func(t *testing.T) {
		id := SymbolID("main.go", "function", "Foo", 10)
		// 12 bytes hex-encoded = 24 characters
		if len(id) != 24 {
			t.Fatalf("expected 24 chars, got %d: %q", len(id), id)
		}
	})
}

func TestSymbolDocJSON(t *testing.T) {
	doc := SymbolDoc{
		ID:        SymbolID("main.go", KindFunction, "Foo", 10),
		File:      "main.go",
		Kind:      KindFunction,
		Name:      "Foo",
		Signature: "func Foo(x int) string",
		Language:  "go",
		StartLine: 10,
		EndLine:   20,
		ParentID:  "",
		Docstring: "Foo does stuff.",
		FileHash:  "abc123",
		Repo:      "myrepo",
		IndexedAt: "2026-04-17T00:00:00Z",
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Verify required fields present
	for _, key := range []string{"id", "file", "kind", "name", "signature", "language", "start_line", "end_line", "file_hash", "repo", "indexed_at"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}

	// ParentID empty => omitted
	if _, ok := m["parent_id"]; ok {
		t.Error("parent_id should be omitted when empty")
	}

	// Docstring present because it has a value
	if _, ok := m["docstring"]; !ok {
		t.Error("docstring should be present when non-empty")
	}

	// Set ParentID and verify it appears
	doc.ParentID = "parent123"
	data2, _ := json.Marshal(doc)
	var m2 map[string]interface{}
	json.Unmarshal(data2, &m2)
	if _, ok := m2["parent_id"]; !ok {
		t.Error("parent_id should be present when non-empty")
	}
}

func TestEdgeID(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		a := EdgeID("src1", "tgt1", "calls", "main.go")
		b := EdgeID("src1", "tgt1", "calls", "main.go")
		if a != b {
			t.Fatalf("same inputs produced different IDs: %q vs %q", a, b)
		}
	})

	t.Run("length is 24 hex chars", func(t *testing.T) {
		id := EdgeID("src1", "tgt1", "calls", "main.go")
		if len(id) != 24 {
			t.Fatalf("expected 24 chars, got %d: %q", len(id), id)
		}
	})

	t.Run("different inputs differ", func(t *testing.T) {
		a := EdgeID("src1", "tgt1", "calls", "main.go")
		b := EdgeID("src2", "tgt1", "calls", "main.go")
		c := EdgeID("src1", "tgt2", "calls", "main.go")
		d := EdgeID("src1", "tgt1", "imports", "main.go")
		e := EdgeID("src1", "tgt1", "calls", "other.go")

		ids := []string{a, b, c, d, e}
		seen := make(map[string]bool)
		for _, id := range ids {
			if seen[id] {
				t.Fatalf("collision found: %q", id)
			}
			seen[id] = true
		}
	})
}

func TestEdgeDocFields(t *testing.T) {
	edge := EdgeDoc{
		SourceID: "src123",
		TargetID: "tgt456",
		Kind:     EdgeCalls,
		File:     "main.go",
		Repo:     "myrepo",
	}

	data, err := json.Marshal(edge)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"source_id", "target_id", "kind", "file", "repo"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}

	if m["kind"] != EdgeCalls {
		t.Errorf("kind = %q, want %q", m["kind"], EdgeCalls)
	}
}
