package research

import (
	"context"
	"errors"
	"testing"
)

type stubResearcher struct {
	name     string
	describe string
	evidence Evidence
}

func (s *stubResearcher) Name() string    { return s.name }
func (s *stubResearcher) Describe() string { return s.describe }
func (s *stubResearcher) Gather(_ context.Context, _ ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	return s.evidence, nil
}

func TestRegistryRegisterAndLookup(t *testing.T) {
	reg := NewRegistry()

	stub := &stubResearcher{name: "alpha", describe: "alpha researcher", evidence: Evidence{Source: "alpha-src"}}
	if err := reg.Register(stub); err != nil {
		t.Fatalf("Register: unexpected error: %v", err)
	}

	got, ok := reg.Lookup("alpha")
	if !ok {
		t.Fatal("Lookup: expected to find 'alpha'")
	}
	if got.Name() != "alpha" {
		t.Errorf("Lookup: got name %q, want %q", got.Name(), "alpha")
	}

	_, ok = reg.Lookup("missing")
	if ok {
		t.Error("Lookup: expected false for missing name")
	}
}

func TestRegistryDuplicate(t *testing.T) {
	reg := NewRegistry()

	stub := &stubResearcher{name: "dup"}
	if err := reg.Register(stub); err != nil {
		t.Fatalf("first Register: unexpected error: %v", err)
	}

	err := reg.Register(&stubResearcher{name: "dup"})
	if err == nil {
		t.Fatal("second Register: expected error, got nil")
	}
	if !errors.Is(err, ErrDuplicateResearcher) {
		t.Errorf("second Register: expected ErrDuplicateResearcher, got %v", err)
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()

	_ = reg.Register(&stubResearcher{name: "zebra"})
	_ = reg.Register(&stubResearcher{name: "apple"})

	list := reg.List()
	if len(list) != 2 {
		t.Fatalf("List: expected 2 researchers, got %d", len(list))
	}
	if list[0].Name() != "apple" || list[1].Name() != "zebra" {
		t.Errorf("List: expected [apple, zebra], got [%s, %s]", list[0].Name(), list[1].Name())
	}

	names := reg.Names()
	if len(names) != 2 {
		t.Fatalf("Names: expected 2, got %d", len(names))
	}
	if names[0] != "apple" || names[1] != "zebra" {
		t.Errorf("Names: expected [apple, zebra], got %v", names)
	}
}
