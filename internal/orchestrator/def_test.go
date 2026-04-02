package orchestrator

import (
	"testing"
)

func TestWorkflowDefValidate(t *testing.T) {
	tests := []struct {
		name    string
		def     WorkflowDef
		wantErr bool
		errContains string
	}{
		{
			name: "valid pipeline-ref",
			def: WorkflowDef{
				Steps: []WorkflowStep{
					{ID: "step1", Type: StepTypePipelineRef, Pipeline: "my-pipeline"},
				},
			},
		},
		{
			name: "valid agent-ref",
			def: WorkflowDef{
				Steps: []WorkflowStep{
					{ID: "step1", Type: StepTypeAgentRef, Agent: "my-agent"},
				},
			},
		},
		{
			name: "valid decision",
			def: WorkflowDef{
				Steps: []WorkflowStep{
					{
						ID: "decide", Type: StepTypeDecision, Model: "llama3",
						Prompt: "choose", On: map[string]string{"yes": "step2", "no": "step3"},
					},
					{ID: "step2", Type: StepTypePipelineRef, Pipeline: "p2"},
					{ID: "step3", Type: StepTypePipelineRef, Pipeline: "p3"},
				},
			},
		},
		{
			name: "valid parallel",
			def: WorkflowDef{
				Steps: []WorkflowStep{
					{
						ID: "par", Type: StepTypeParallel,
						Branches: []ParallelBranch{
							{Steps: []WorkflowStep{{ID: "b1", Type: StepTypePipelineRef, Pipeline: "p1"}}},
							{Steps: []WorkflowStep{{ID: "b2", Type: StepTypePipelineRef, Pipeline: "p2"}}},
						},
					},
				},
			},
		},
		{
			name:    "empty step id",
			def:     WorkflowDef{Steps: []WorkflowStep{{Type: StepTypePipelineRef, Pipeline: "p"}}},
			wantErr: true,
			errContains: "empty id",
		},
		{
			name: "duplicate step id",
			def: WorkflowDef{
				Steps: []WorkflowStep{
					{ID: "dup", Type: StepTypePipelineRef, Pipeline: "p"},
					{ID: "dup", Type: StepTypePipelineRef, Pipeline: "p2"},
				},
			},
			wantErr:     true,
			errContains: "duplicate step id",
		},
		{
			name:        "unknown step type",
			def:         WorkflowDef{Steps: []WorkflowStep{{ID: "s", Type: "unknown"}}},
			wantErr:     true,
			errContains: "unknown type",
		},
		{
			name:        "pipeline-ref missing pipeline field",
			def:         WorkflowDef{Steps: []WorkflowStep{{ID: "s", Type: StepTypePipelineRef}}},
			wantErr:     true,
			errContains: "empty pipeline field",
		},
		{
			name:        "agent-ref missing agent field",
			def:         WorkflowDef{Steps: []WorkflowStep{{ID: "s", Type: StepTypeAgentRef}}},
			wantErr:     true,
			errContains: "empty agent field",
		},
		{
			name:        "decision empty on map",
			def:         WorkflowDef{Steps: []WorkflowStep{{ID: "s", Type: StepTypeDecision, Model: "m", Prompt: "p"}}},
			wantErr:     true,
			errContains: "empty on map",
		},
		{
			name: "parallel no branches",
			def: WorkflowDef{
				Steps: []WorkflowStep{{ID: "s", Type: StepTypeParallel}},
			},
			wantErr:     true,
			errContains: "no branches",
		},
		{
			name: "parallel empty branch",
			def: WorkflowDef{
				Steps: []WorkflowStep{
					{
						ID: "par", Type: StepTypeParallel,
						Branches: []ParallelBranch{{Steps: nil}},
					},
				},
			},
			wantErr:     true,
			errContains: "no steps",
		},
		{
			name: "duplicate id in parallel branch",
			def: WorkflowDef{
				Steps: []WorkflowStep{
					{ID: "top", Type: StepTypePipelineRef, Pipeline: "p"},
					{
						ID: "par", Type: StepTypeParallel,
						Branches: []ParallelBranch{
							{Steps: []WorkflowStep{{ID: "top", Type: StepTypePipelineRef, Pipeline: "q"}}},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "duplicate step id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.def.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" {
					if got := err.Error(); !contains(got, tt.errContains) {
						t.Errorf("error %q does not contain %q", got, tt.errContains)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}
