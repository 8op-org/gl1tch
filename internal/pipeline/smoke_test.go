//go:build integration

package pipeline_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/adam-stokes/orcai/internal/pipeline"
)

// smokeModel returns the model to use for smoke tests.
// Override with ORCAI_SMOKE_MODEL (e.g. "llama3.2:1b" in CI).
func smokeModel() string {
	if m := os.Getenv("ORCAI_SMOKE_MODEL"); m != "" {
		return m
	}
	return "llama3.2"
}

// smokeModelBase strips the tag so checkModelAvailable can match "llama3.2:1b" → "llama3.2".
func smokeModelBase(full string) string {
	if idx := strings.Index(full, ":"); idx >= 0 {
		return full[:idx]
	}
	return full
}

// TestSmokePipeline_SingleStep verifies a one-step ollama pipeline runs without error
// and produces non-empty output. Intentionally minimal — just confirms the pipeline
// executor is wired correctly and ollama is reachable.
func TestSmokePipeline_SingleStep(t *testing.T) {
	model := smokeModel()
	checkModelAvailable(t, smokeModelBase(model))

	p := &pipeline.Pipeline{
		Name:    "smoke",
		Version: "1",
		Steps: []pipeline.Step{
			{
				ID:       "ping",
				Executor: "ollama",
				Model:    model,
				Prompt:   `Reply with the single word "ok".`,
			},
		},
	}

	mgr := buildManager()
	pub := &collectPublisher{}

	result, err := pipeline.Run(context.Background(), p, mgr, "", pipeline.WithEventPublisher(pub))
	if err != nil {
		t.Fatalf("pipeline.Run: %v", err)
	}
	if strings.TrimSpace(result) == "" {
		t.Error("expected non-empty output")
	}
}
