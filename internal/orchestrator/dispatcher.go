package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/executor"
	"github.com/8op-org/gl1tch/internal/pipeline"
)

// StepDispatcher resolves and executes a single workflow step by delegating to
// the pipeline runner.
type StepDispatcher struct {
	// ConfigDir is the resolved path to ~/.config/glitch.
	ConfigDir string
}

// NewStepDispatcher returns a StepDispatcher rooted at configDir.
func NewStepDispatcher(configDir string) *StepDispatcher {
	return &StepDispatcher{ConfigDir: configDir}
}

// Dispatch executes a single workflow step.
//
//   - For pipeline-ref: resolves ~/.config/glitch/pipelines/<name>.pipeline.yaml
//   - For agent-ref: resolves ~/.config/glitch/pipelines/apm.<name>.pipeline.yaml
//
// The step Input is expanded via ExpandTemplate before being passed to pipeline.Run.
// On success, the output is written into wctx under "<step.ID>.output".
// game:false is injected into pipeline vars to suppress per-pipeline game events.
func (d *StepDispatcher) Dispatch(
	ctx context.Context,
	step WorkflowStep,
	wctx *WorkflowContext,
	mgr *executor.Manager,
	opts ...pipeline.RunOption,
) (string, error) {
	var pipelinePath string
	switch step.Type {
	case StepTypePipelineRef:
		pipelinePath = filepath.Join(d.ConfigDir, "pipelines", step.Pipeline+".pipeline.yaml")
		if _, err := os.Stat(pipelinePath); err != nil {
			return "", fmt.Errorf("pipeline not found: %s", pipelinePath)
		}
	case StepTypeAgentRef:
		pipelinePath = filepath.Join(d.ConfigDir, "pipelines", "apm."+step.Agent+".pipeline.yaml")
		if _, err := os.Stat(pipelinePath); err != nil {
			return "", fmt.Errorf("agent pipeline not found: %s", pipelinePath)
		}
	default:
		return "", fmt.Errorf("orchestrator: dispatcher: unexpected step type %q", step.Type)
	}

	f, err := os.Open(pipelinePath)
	if err != nil {
		return "", fmt.Errorf("orchestrator: dispatcher: open pipeline %q: %w", pipelinePath, err)
	}
	defer f.Close()

	p, err := pipeline.Load(f)
	if err != nil {
		return "", fmt.Errorf("orchestrator: dispatcher: load pipeline %q: %w", pipelinePath, err)
	}

	// Inject game:false into pipeline vars to suppress per-pipeline game events.
	if p.Vars == nil {
		p.Vars = make(map[string]any)
	}
	gameDisabled := false
	p.Game = &gameDisabled

	expandedInput := ExpandTemplate(step.Input, wctx)

	// Capture output in a buffer as well as any configured step writer.
	var buf bytes.Buffer
	runOpts := append([]pipeline.RunOption{pipeline.WithStepWriter(&buf)}, opts...)

	output, err := pipeline.Run(ctx, p, mgr, expandedInput, runOpts...)
	if err != nil {
		return "", fmt.Errorf("orchestrator: step %q pipeline error: %w", step.ID, err)
	}

	// TODO(game): aggregate TokenUsage when token-score-gamification is implemented.
	// Wire ScoreTeeWriter here once the game spec lands.

	wctx.Set(step.ID+".output", output)
	return output, nil
}
