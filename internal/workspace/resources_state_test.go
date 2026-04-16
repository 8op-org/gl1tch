package workspace

import (
	"testing"
	"time"
)

func TestResourceStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	st := ResourceState{
		Entries: map[string]time.Time{
			"ensemble": time.Unix(1700000000, 0).UTC(),
			"kibana":   time.Unix(1700000500, 0).UTC(),
		},
	}
	if err := SaveResourceState(dir, st); err != nil {
		t.Fatal(err)
	}
	got, err := LoadResourceState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 2 || !got.Entries["ensemble"].Equal(st.Entries["ensemble"]) {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestResourceStateMissingReturnsEmpty(t *testing.T) {
	got, err := LoadResourceState(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 0 {
		t.Fatalf("expected empty, got %+v", got)
	}
}
