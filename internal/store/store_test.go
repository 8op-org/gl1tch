package store

import (
	"path/filepath"
	"testing"
)

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRecordRun(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	id, err := s.RecordRun("pipeline", "test-run", "some input")
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	if err := s.FinishRun(id, "some output", 0); err != nil {
		t.Fatalf("FinishRun: %v", err)
	}
}

func TestRecordStep(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	runID, err := s.RecordRun("pipeline", "step-test", "")
	if err != nil {
		t.Fatalf("RecordRun: %v", err)
	}

	if err := s.RecordStep(runID, "step-1", "my prompt", "model output", "qwen2.5:7b", 123); err != nil {
		t.Fatalf("RecordStep: %v", err)
	}

	// Insert OR REPLACE — should not error on duplicate step_id
	if err := s.RecordStep(runID, "step-1", "updated prompt", "new output", "qwen2.5:7b", 456); err != nil {
		t.Fatalf("RecordStep (replace): %v", err)
	}
}

func TestRecordResearchEvent(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	evt := ResearchEvent{
		QueryID:        "q1",
		Question:       "What is the deployment frequency?",
		Researchers:    "git,es",
		CompositeScore: 0.85,
		Reason:         "high confidence from git data",
	}
	if err := s.RecordResearchEvent(evt); err != nil {
		t.Fatalf("RecordResearchEvent: %v", err)
	}
}

func TestSimilarResearchEvents(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenAt(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	defer s.Close()

	evtA := ResearchEvent{
		QueryID:        "qa",
		Question:       "What is the deployment frequency for this service?",
		Researchers:    "git",
		CompositeScore: 0.9,
		Reason:         "based on git tags",
	}
	evtB := ResearchEvent{
		QueryID:        "qb",
		Question:       "How many open pull requests exist in the repository?",
		Researchers:    "es",
		CompositeScore: 0.7,
		Reason:         "es index data",
	}
	if err := s.RecordResearchEvent(evtA); err != nil {
		t.Fatalf("seed evtA: %v", err)
	}
	if err := s.RecordResearchEvent(evtB); err != nil {
		t.Fatalf("seed evtB: %v", err)
	}

	results, err := s.SimilarResearchEvents("What is the release deployment frequency?", 2)
	if err != nil {
		t.Fatalf("SimilarResearchEvents: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	// The deployment-frequency event (evtA) should be ranked first.
	if results[0].QueryID != "qa" {
		t.Errorf("expected first result to be 'qa' (deployment frequency), got %q", results[0].QueryID)
	}
}
