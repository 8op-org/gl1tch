package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/8op-org/gl1tch/internal/busd/topics"
	"github.com/8op-org/gl1tch/internal/executor"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/store"
)

// BusPublisher is the interface ConductorRunner uses to emit BUSD events.
// It is satisfied by *busPublisher in cmd/busd_publisher.go or any test double.
type BusPublisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// conductorConfig holds the optional dependencies for a ConductorRunner.
type conductorConfig struct {
	store      *store.Store
	bus        BusPublisher
	mgr        *executor.Manager
	ollamaURL  string
	beforeStep func(ctx context.Context, step WorkflowStep, wctx *WorkflowContext) error
	afterStep  func(ctx context.Context, step WorkflowStep, output string, wctx *WorkflowContext)
}

// ConductorOption configures a ConductorRunner.
type ConductorOption func(*conductorConfig)

// WithStore attaches a result store so workflow runs and checkpoints are persisted.
func WithStore(s *store.Store) ConductorOption {
	return func(c *conductorConfig) { c.store = s }
}

// WithBusPublisher attaches a BUSD publisher for lifecycle events.
func WithBusPublisher(b BusPublisher) ConductorOption {
	return func(c *conductorConfig) { c.bus = b }
}

// WithExecutorManager sets the executor manager used for pipeline dispatch.
func WithExecutorManager(m *executor.Manager) ConductorOption {
	return func(c *conductorConfig) { c.mgr = m }
}

// WithOllamaURL overrides the default Ollama endpoint for decision steps.
func WithOllamaURL(url string) ConductorOption {
	return func(c *conductorConfig) { c.ollamaURL = url }
}

// WithBeforeStep registers an ADK-inspired callback invoked before each step.
// If it returns an error the step (and workflow) is aborted.
func WithBeforeStep(fn func(ctx context.Context, step WorkflowStep, wctx *WorkflowContext) error) ConductorOption {
	return func(c *conductorConfig) { c.beforeStep = fn }
}

// WithAfterStep registers an ADK-inspired callback invoked after each step succeeds.
func WithAfterStep(fn func(ctx context.Context, step WorkflowStep, output string, wctx *WorkflowContext)) ConductorOption {
	return func(c *conductorConfig) { c.afterStep = fn }
}

// ConductorRunner orchestrates multi-step workflows, delegating individual steps
// to StepDispatcher and DecisionNode.
type ConductorRunner struct {
	cfg        conductorConfig
	dispatcher *StepDispatcher
	configDir  string
}

// NewConductorRunner creates a ConductorRunner rooted at configDir.
func NewConductorRunner(configDir string, opts ...ConductorOption) *ConductorRunner {
	c := &ConductorRunner{configDir: configDir}
	for _, o := range opts {
		o(&c.cfg)
	}
	c.dispatcher = NewStepDispatcher(configDir)
	return c
}

// publish emits a BUSD event, silently swallowing errors.
func (c *ConductorRunner) publish(ctx context.Context, topic string, payload map[string]any) {
	if c.cfg.bus == nil {
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = c.cfg.bus.Publish(ctx, topic, b)
}

// saveCheckpoint persists step state; non-fatal.
func (c *ConductorRunner) saveCheckpoint(ctx context.Context, runID int64, stepID, status string, wctx *WorkflowContext) {
	if c.cfg.store == nil {
		return
	}
	b, err := wctx.Marshal()
	if err != nil {
		log.Printf("orchestrator: marshal context for checkpoint: %v", err)
		return
	}
	if err := c.cfg.store.SaveWorkflowCheckpoint(ctx, runID, stepID, status, string(b)); err != nil {
		log.Printf("orchestrator: save checkpoint for step %q: %v", stepID, err)
	}
}

// Run executes the workflow defined by def with the given input string.
// Returns the output of the last step, or an error if any step fails.
func (c *ConductorRunner) Run(ctx context.Context, def *WorkflowDef, input string) (string, error) {
	// 1. Create workflow_runs record.
	var runID int64
	if c.cfg.store != nil {
		id, err := c.cfg.store.CreateWorkflowRun(ctx, def.Name, input)
		if err != nil {
			log.Printf("orchestrator: create workflow run: %v", err)
		} else {
			runID = id
		}
	}

	// 2. Publish run started.
	c.publish(ctx, topics.WorkflowRunStarted, map[string]any{
		"workflow_name": def.Name,
		"run_id":        runID,
	})

	// 3. Create context and seed input.
	wctx := NewWorkflowContext()
	wctx.Set("temp.input", input)

	// 4. Execute steps.
	output, err := c.executeSteps(ctx, def.Steps, wctx, runID, "")
	if err != nil {
		c.publish(ctx, topics.WorkflowRunFailed, map[string]any{
			"workflow_name": def.Name,
			"run_id":        runID,
			"error":         err.Error(),
		})
		if c.cfg.store != nil {
			_ = c.cfg.store.FailWorkflowRun(ctx, runID, err.Error())
		}
		return "", err
	}

	// 7. Publish completed.
	c.publish(ctx, topics.WorkflowRunCompleted, map[string]any{
		"workflow_name": def.Name,
		"run_id":        runID,
	})
	if c.cfg.store != nil {
		_ = c.cfg.store.CompleteWorkflowRun(ctx, runID, output)
	}
	return output, nil
}

// executeSteps runs a slice of steps sequentially, handling decision branching
// and parallel fan-out. branchPrefix tracks the execution path for BUSD events.
func (c *ConductorRunner) executeSteps(
	ctx context.Context,
	steps []WorkflowStep,
	wctx *WorkflowContext,
	runID int64,
	branchPrefix string,
) (string, error) {
	var lastOutput string

	i := 0
	for i < len(steps) {
		step := steps[i]

		// Resolve branch label for BUSD events (ADK pattern).
		branchLabel := step.ID
		if branchPrefix != "" {
			branchLabel = branchPrefix + "/" + step.ID
		}

		// BeforeStep callback.
		if c.cfg.beforeStep != nil {
			if err := c.cfg.beforeStep(ctx, step, wctx); err != nil {
				return "", fmt.Errorf("orchestrator: before-step hook for %q: %w", step.ID, err)
			}
		}

		// Publish step started.
		c.publish(ctx, topics.WorkflowStepStarted, map[string]any{
			"step_id": step.ID,
			"run_id":  runID,
			"branch":  branchLabel,
			"type":    step.Type,
		})

		var (
			stepOutput string
			stepErr    error
		)

		switch step.Type {
		case StepTypeDecision:
			stepOutput, stepErr = c.runDecisionStep(ctx, step, wctx, steps, &i, runID, branchLabel)
			if stepErr != nil {
				c.publish(ctx, topics.WorkflowStepFailed, map[string]any{
					"step_id": step.ID,
					"run_id":  runID,
					"branch":  branchLabel,
					"error":   stepErr.Error(),
				})
				c.saveCheckpoint(ctx, runID, step.ID, "failed", wctx)
				return "", stepErr
			}
			// Decision steps do not produce a direct output — the branch jump is the result.
			c.publish(ctx, topics.WorkflowStepDone, map[string]any{
				"step_id": step.ID,
				"run_id":  runID,
				"branch":  branchLabel,
			})
			c.saveCheckpoint(ctx, runID, step.ID, "done", wctx)
			if c.cfg.afterStep != nil {
				c.cfg.afterStep(ctx, step, stepOutput, wctx)
			}
			// i was already updated inside runDecisionStep; do not increment again.
			continue

		case StepTypeParallel:
			stepOutput, stepErr = c.runParallelStep(ctx, step, wctx, runID, branchLabel)

		default: // pipeline-ref, agent-ref
			var pipelineOpts []pipeline.RunOption
			stepOutput, stepErr = c.dispatcher.Dispatch(ctx, step, wctx, c.cfg.mgr, pipelineOpts...)
		}

		if stepErr != nil {
			c.publish(ctx, topics.WorkflowStepFailed, map[string]any{
				"step_id": step.ID,
				"run_id":  runID,
				"branch":  branchLabel,
				"error":   stepErr.Error(),
			})
			c.saveCheckpoint(ctx, runID, step.ID, "failed", wctx)
			return "", stepErr
		}

		lastOutput = stepOutput
		wctx.Set(step.ID+".output", stepOutput)

		c.publish(ctx, topics.WorkflowStepDone, map[string]any{
			"step_id": step.ID,
			"run_id":  runID,
			"branch":  branchLabel,
		})
		c.saveCheckpoint(ctx, runID, step.ID, "done", wctx)

		if c.cfg.afterStep != nil {
			c.cfg.afterStep(ctx, step, stepOutput, wctx)
		}

		i++
	}

	return lastOutput, nil
}

// runDecisionStep evaluates a decision node and jumps to the target step.
// It modifies *i to point to the resolved target step index in steps.
func (c *ConductorRunner) runDecisionStep(
	ctx context.Context,
	step WorkflowStep,
	wctx *WorkflowContext,
	steps []WorkflowStep,
	i *int,
	runID int64,
	branchLabel string,
) (string, error) {
	node := &DecisionNode{
		Model:       step.Model,
		Prompt:      step.Prompt,
		TimeoutSecs: step.TimeoutSecs,
		OllamaURL:   c.cfg.ollamaURL,
	}

	branch, err := node.Evaluate(ctx, wctx)
	if err != nil {
		if step.DefaultBranch != "" {
			log.Printf("orchestrator: decision step %q failed (%v); using default_branch %q", step.ID, err, step.DefaultBranch)
			branch = step.DefaultBranch
		} else {
			return "", fmt.Errorf("orchestrator: decision step %q failed: %w", step.ID, err)
		}
	}

	// Resolve branch name -> target step ID via the On map.
	targetStepID, ok := step.On[branch]
	if !ok {
		if step.DefaultBranch != "" {
			targetStepID, ok = step.On[step.DefaultBranch]
			if !ok {
				return "", fmt.Errorf("orchestrator: decision step %q: default_branch %q not in on map", step.ID, step.DefaultBranch)
			}
		} else {
			return "", fmt.Errorf("orchestrator: decision step %q: branch %q not in on map", step.ID, branch)
		}
	}

	// Find the target step index.
	for j, s := range steps {
		if s.ID == targetStepID {
			*i = j
			return "", nil
		}
	}
	return "", fmt.Errorf("orchestrator: decision step %q: target step %q not found", step.ID, targetStepID)
}

// runParallelStep launches each branch as a goroutine and waits for all to complete.
// Results are merged back into wctx.
func (c *ConductorRunner) runParallelStep(
	ctx context.Context,
	step WorkflowStep,
	wctx *WorkflowContext,
	runID int64,
	branchLabel string,
) (string, error) {
	type branchResult struct {
		output string
		err    error
	}

	results := make([]branchResult, len(step.Branches))
	var wg sync.WaitGroup

	for idx, branch := range step.Branches {
		wg.Add(1)
		go func(bIdx int, b ParallelBranch) {
			defer wg.Done()
			prefix := fmt.Sprintf("%d", bIdx)
			if branchLabel != "" {
				prefix = branchLabel + "/" + prefix
			}
			out, err := c.executeSteps(ctx, b.Steps, wctx, runID, prefix)
			results[bIdx] = branchResult{output: out, err: err}
		}(idx, branch)
	}

	wg.Wait()

	// Collect errors.
	var lastOutput string
	for _, r := range results {
		if r.err != nil {
			return "", r.err
		}
		lastOutput = r.output
	}
	return lastOutput, nil
}

// Resume loads checkpoints for runID, restores WorkflowContext, and re-executes
// from the last failed (or incomplete) step.
func (c *ConductorRunner) Resume(ctx context.Context, def *WorkflowDef, runID int64) (string, error) {
	if c.cfg.store == nil {
		return "", fmt.Errorf("orchestrator: resume requires a store")
	}

	checkpoints, err := c.cfg.store.LoadWorkflowCheckpoints(ctx, runID)
	if err != nil {
		return "", fmt.Errorf("orchestrator: resume load checkpoints: %w", err)
	}

	// Find the last checkpoint to restore context.
	if len(checkpoints) == 0 {
		return "", fmt.Errorf("orchestrator: no checkpoints found for run %d", runID)
	}

	// Find last failed step, or last checkpoint if no failures.
	var resumeStepID string
	var lastContextJSON string
	for _, cp := range checkpoints {
		lastContextJSON = cp.ContextJSON
		if cp.Status == "failed" {
			resumeStepID = cp.StepID
			break
		}
	}
	if resumeStepID == "" {
		// No failed step: resume from the step after the last checkpoint.
		last := checkpoints[len(checkpoints)-1]
		resumeStepID = last.StepID
		lastContextJSON = last.ContextJSON
	}

	// Restore context.
	wctx := NewWorkflowContext()
	if lastContextJSON != "" {
		if err := wctx.Unmarshal([]byte(lastContextJSON)); err != nil {
			return "", fmt.Errorf("orchestrator: resume unmarshal context: %w", err)
		}
	}

	// Find the resume step index.
	startIdx := -1
	for i, s := range def.Steps {
		if s.ID == resumeStepID {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return "", fmt.Errorf("orchestrator: resume: step %q not found in workflow definition", resumeStepID)
	}

	c.publish(ctx, topics.WorkflowRunStarted, map[string]any{
		"workflow_name": def.Name,
		"run_id":        runID,
		"resumed":       true,
	})

	output, err := c.executeSteps(ctx, def.Steps[startIdx:], wctx, runID, "")
	if err != nil {
		c.publish(ctx, topics.WorkflowRunFailed, map[string]any{
			"workflow_name": def.Name,
			"run_id":        runID,
			"error":         err.Error(),
		})
		_ = c.cfg.store.FailWorkflowRun(ctx, runID, err.Error())
		return "", err
	}

	c.publish(ctx, topics.WorkflowRunCompleted, map[string]any{
		"workflow_name": def.Name,
		"run_id":        runID,
	})
	_ = c.cfg.store.CompleteWorkflowRun(ctx, runID, output)
	return output, nil
}
