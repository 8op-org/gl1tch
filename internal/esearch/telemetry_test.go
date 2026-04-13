package esearch

import (
	"context"
	"testing"
)

func TestNewTelemetryNilClient(t *testing.T) {
	tel := NewTelemetry(nil)
	if tel != nil {
		t.Fatal("NewTelemetry(nil) should return nil")
	}
}

func TestNilTelemetrySafe(t *testing.T) {
	var tel *Telemetry
	ctx := context.Background()

	// None of these should panic.
	if err := tel.EnsureIndices(ctx); err != nil {
		t.Errorf("EnsureIndices on nil: %v", err)
	}
	if err := tel.IndexResearchRun(ctx, ResearchRunDoc{}); err != nil {
		t.Errorf("IndexResearchRun on nil: %v", err)
	}
	if err := tel.IndexToolCall(ctx, ToolCallDoc{}); err != nil {
		t.Errorf("IndexToolCall on nil: %v", err)
	}
	if err := tel.IndexLLMCall(ctx, LLMCallDoc{}); err != nil {
		t.Errorf("IndexLLMCall on nil: %v", err)
	}
}

func TestNewRunIDUnique(t *testing.T) {
	a := NewRunID()
	b := NewRunID()

	if a == b {
		t.Errorf("NewRunID returned duplicate values: %s", a)
	}
	if len(a) < 10 {
		t.Errorf("NewRunID too short: %q (len %d)", a, len(a))
	}
	if len(b) < 10 {
		t.Errorf("NewRunID too short: %q (len %d)", b, len(b))
	}
}
