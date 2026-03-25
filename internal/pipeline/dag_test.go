package pipeline

import (
	"strings"
	"testing"
)

func TestDAGCycleDetection(t *testing.T) {
	t.Run("no cycle returns nil", func(t *testing.T) {
		steps := []Step{
			{ID: "a"},
			{ID: "b", Needs: []string{"a"}},
			{ID: "c", Needs: []string{"b"}},
		}
		_, err := buildDAG(steps)
		if err != nil {
			t.Errorf("expected no error for valid DAG, got: %v", err)
		}
	})

	t.Run("direct cycle returns error", func(t *testing.T) {
		steps := []Step{
			{ID: "a", Needs: []string{"b"}},
			{ID: "b", Needs: []string{"a"}},
		}
		_, err := buildDAG(steps)
		if err == nil {
			t.Error("expected cycle error, got nil")
		}
		if !strings.Contains(err.Error(), "cycle") {
			t.Errorf("expected 'cycle' in error message, got: %v", err)
		}
	})

	t.Run("indirect cycle returns error", func(t *testing.T) {
		steps := []Step{
			{ID: "a", Needs: []string{"c"}},
			{ID: "b", Needs: []string{"a"}},
			{ID: "c", Needs: []string{"b"}},
		}
		_, err := buildDAG(steps)
		if err == nil {
			t.Error("expected cycle error, got nil")
		}
	})

	t.Run("unknown needs reference returns error", func(t *testing.T) {
		steps := []Step{
			{ID: "a", Needs: []string{"nonexistent"}},
		}
		_, err := buildDAG(steps)
		if err == nil {
			t.Error("expected error for unknown needs reference")
		}
		if !strings.Contains(err.Error(), "nonexistent") {
			t.Errorf("expected 'nonexistent' in error message, got: %v", err)
		}
	})

	t.Run("parallel diamond shape has no cycle", func(t *testing.T) {
		// a → b, a → c, b → d, c → d (diamond)
		steps := []Step{
			{ID: "a"},
			{ID: "b", Needs: []string{"a"}},
			{ID: "c", Needs: []string{"a"}},
			{ID: "d", Needs: []string{"b", "c"}},
		}
		_, err := buildDAG(steps)
		if err != nil {
			t.Errorf("diamond DAG should have no cycle, got: %v", err)
		}
	})

	t.Run("self-cycle returns error", func(t *testing.T) {
		steps := []Step{
			{ID: "a", Needs: []string{"a"}},
		}
		_, err := buildDAG(steps)
		if err == nil {
			t.Error("expected self-cycle error")
		}
	})

	t.Run("valid pipeline load returns no error", func(t *testing.T) {
		yaml := `
name: test
steps:
  - id: first
  - id: second
    needs: [first]
  - id: third
    needs: [second]
`
		p, err := Load(strings.NewReader(yaml))
		if err != nil {
			t.Fatalf("expected no error loading valid pipeline, got: %v", err)
		}
		if p.Name != "test" {
			t.Errorf("expected name 'test', got %q", p.Name)
		}
	})

	t.Run("cycle in pipeline load returns error", func(t *testing.T) {
		yaml := `
name: cyclic
steps:
  - id: a
    needs: [b]
  - id: b
    needs: [a]
`
		_, err := Load(strings.NewReader(yaml))
		if err == nil {
			t.Error("expected error for cyclic pipeline")
		}
	})
}
