package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openCapTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("OpenAt: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCapabilityNotes_ReturnsOnlySystemNotes(t *testing.T) {
	s := openCapTestStore(t)
	ctx := context.Background()
	now := time.Now().UnixMilli()

	// Insert a system capability note (run_id=0).
	_, err := s.InsertBrainNote(ctx, BrainNote{
		RunID:     0,
		StepID:    "gl1tch.capability.executor.shell",
		CreatedAt: now,
		Tags:      "type:capability title:Shell",
		Body:      "Shell executor: runs arbitrary shell commands",
	})
	if err != nil {
		t.Fatalf("InsertBrainNote: %v", err)
	}

	// Insert a regular run note — should NOT appear in CapabilityNotes.
	_, err = s.InsertBrainNote(ctx, BrainNote{
		RunID:     42,
		StepID:    "step-a",
		CreatedAt: now,
		Tags:      "type:finding",
		Body:      "regular finding",
	})
	if err != nil {
		t.Fatalf("InsertBrainNote: %v", err)
	}

	notes, err := s.CapabilityNotes(ctx)
	if err != nil {
		t.Fatalf("CapabilityNotes: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("expected 1 capability note, got %d", len(notes))
	}
	if notes[0].RunID != 0 {
		t.Errorf("expected run_id=0, got %d", notes[0].RunID)
	}
	if !strings.Contains(notes[0].Tags, "type:capability") {
		t.Errorf("expected type:capability tag, got %q", notes[0].Tags)
	}
}

func TestCapabilityNotes_EmptyWhenNoneSeeded(t *testing.T) {
	s := openCapTestStore(t)
	ctx := context.Background()

	notes, err := s.CapabilityNotes(ctx)
	if err != nil {
		t.Fatalf("CapabilityNotes: %v", err)
	}
	if notes == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(notes))
	}
}

func TestUpsertCapabilityNote_Idempotent(t *testing.T) {
	s := openCapTestStore(t)
	ctx := context.Background()
	now := time.Now().UnixMilli()

	note := BrainNote{
		RunID:     0,
		StepID:    "gl1tch.capability.executor.shell",
		CreatedAt: now,
		Tags:      "type:capability title:Shell",
		Body:      "first body",
	}

	if err := s.UpsertCapabilityNote(ctx, note); err != nil {
		t.Fatalf("first UpsertCapabilityNote: %v", err)
	}

	note.Body = "updated body"
	if err := s.UpsertCapabilityNote(ctx, note); err != nil {
		t.Fatalf("second UpsertCapabilityNote: %v", err)
	}

	notes, err := s.CapabilityNotes(ctx)
	if err != nil {
		t.Fatalf("CapabilityNotes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note after upsert, got %d", len(notes))
	}
	if notes[0].Body != "updated body" {
		t.Errorf("expected updated body, got %q", notes[0].Body)
	}
}

func TestUpsertCapabilityNote_MultipleEntries(t *testing.T) {
	s := openCapTestStore(t)
	ctx := context.Background()
	now := time.Now().UnixMilli()

	entries := []BrainNote{
		{RunID: 0, StepID: "gl1tch.capability.executor.shell", CreatedAt: now, Tags: "type:capability", Body: "shell"},
		{RunID: 0, StepID: "gl1tch.capability.executor.ollama", CreatedAt: now, Tags: "type:capability", Body: "ollama"},
		{RunID: 0, StepID: "gl1tch.capability.executor.claude", CreatedAt: now, Tags: "type:capability", Body: "claude"},
	}

	for _, e := range entries {
		if err := s.UpsertCapabilityNote(ctx, e); err != nil {
			t.Fatalf("UpsertCapabilityNote %q: %v", e.StepID, err)
		}
	}

	notes, err := s.CapabilityNotes(ctx)
	if err != nil {
		t.Fatalf("CapabilityNotes: %v", err)
	}
	if len(notes) != 3 {
		t.Errorf("expected 3 capability notes, got %d", len(notes))
	}
}

func TestCapabilityNotes_IsolatedFromRecentBrainNotes(t *testing.T) {
	s := openCapTestStore(t)
	ctx := context.Background()
	now := time.Now().UnixMilli()

	// Seed a capability note.
	if err := s.UpsertCapabilityNote(ctx, BrainNote{
		RunID:     0,
		StepID:    "gl1tch.capability.executor.shell",
		CreatedAt: now,
		Tags:      "type:capability title:Shell",
		Body:      "shell capability",
	}); err != nil {
		t.Fatalf("UpsertCapabilityNote: %v", err)
	}

	// RecentBrainNotes for run 0 should NOT return capability notes —
	// capability notes live in a separate query path to avoid count interference.
	runNotes, err := s.RecentBrainNotes(ctx, 0, 10)
	if err != nil {
		t.Fatalf("RecentBrainNotes: %v", err)
	}
	for _, n := range runNotes {
		if strings.Contains(n.Tags, "type:capability") {
			t.Errorf("RecentBrainNotes should not return capability notes; got: %+v", n)
		}
	}
}
