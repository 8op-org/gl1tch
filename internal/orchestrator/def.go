package orchestrator

import "fmt"

// StepType constants define valid step kinds in a workflow definition.
const (
	StepTypePipelineRef = "pipeline-ref"
	StepTypeAgentRef    = "agent-ref"
	StepTypeDecision    = "decision"
	StepTypeParallel    = "parallel"
)

// WorkflowDef is the top-level structure loaded from a .workflow.yaml file.
type WorkflowDef struct {
	Name    string         `yaml:"name"`
	Version string         `yaml:"version"`
	Steps   []WorkflowStep `yaml:"steps"`
}

// WorkflowStep is a single unit of work in a workflow definition.
type WorkflowStep struct {
	ID   string `yaml:"id"`
	Type string `yaml:"type"`

	// pipeline-ref / agent-ref fields
	Pipeline string            `yaml:"pipeline,omitempty"`
	Agent    string            `yaml:"agent,omitempty"`
	Input    string            `yaml:"input,omitempty"`
	Vars     map[string]string `yaml:"vars,omitempty"`

	// decision fields
	Model         string            `yaml:"model,omitempty"`
	Prompt        string            `yaml:"prompt,omitempty"`
	On            map[string]string `yaml:"on,omitempty"`
	DefaultBranch string            `yaml:"default_branch,omitempty"`
	TimeoutSecs   int               `yaml:"timeout_secs,omitempty"`

	// parallel fields
	Branches []ParallelBranch `yaml:"branches,omitempty"`
}

// ParallelBranch is a set of steps executed concurrently within a parallel step.
type ParallelBranch struct {
	Steps []WorkflowStep `yaml:"steps"`
}

// Validate checks the workflow definition for structural correctness.
// It ensures:
//   - all step IDs are unique (including nested steps inside parallel branches)
//   - all step types are valid
//   - decision steps have a non-empty On map
//   - pipeline-ref steps have a non-empty Pipeline field
//   - agent-ref steps have a non-empty Agent field
//   - parallel steps have at least one branch with at least one step
func (d *WorkflowDef) Validate() error {
	seen := make(map[string]bool)
	return validateSteps(d.Steps, seen)
}

func validateSteps(steps []WorkflowStep, seen map[string]bool) error {
	for _, s := range steps {
		if s.ID == "" {
			return fmt.Errorf("orchestrator: step has empty id")
		}
		if seen[s.ID] {
			return fmt.Errorf("orchestrator: duplicate step id %q", s.ID)
		}
		seen[s.ID] = true

		switch s.Type {
		case StepTypePipelineRef:
			if s.Pipeline == "" {
				return fmt.Errorf("orchestrator: step %q (pipeline-ref) has empty pipeline field", s.ID)
			}
		case StepTypeAgentRef:
			if s.Agent == "" {
				return fmt.Errorf("orchestrator: step %q (agent-ref) has empty agent field", s.ID)
			}
		case StepTypeDecision:
			if len(s.On) == 0 {
				return fmt.Errorf("orchestrator: step %q (decision) has empty on map", s.ID)
			}
		case StepTypeParallel:
			if len(s.Branches) == 0 {
				return fmt.Errorf("orchestrator: step %q (parallel) has no branches", s.ID)
			}
			for i, b := range s.Branches {
				if len(b.Steps) == 0 {
					return fmt.Errorf("orchestrator: step %q (parallel) branch %d has no steps", s.ID, i)
				}
				if err := validateSteps(b.Steps, seen); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("orchestrator: step %q has unknown type %q", s.ID, s.Type)
		}
	}
	return nil
}
