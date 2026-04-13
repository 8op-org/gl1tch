package research

import (
	"context"
	"fmt"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
)

// YAMLResearcher wraps a pipeline Workflow as a Researcher.
type YAMLResearcher struct {
	workflow *pipeline.Workflow
	reg      *provider.ProviderRegistry
}

// NewYAMLResearcher creates a new YAMLResearcher from a workflow and provider registry.
func NewYAMLResearcher(w *pipeline.Workflow, reg *provider.ProviderRegistry) *YAMLResearcher {
	return &YAMLResearcher{workflow: w, reg: reg}
}

func (y *YAMLResearcher) Name() string    { return y.workflow.Name }
func (y *YAMLResearcher) Describe() string { return y.workflow.Description }

func (y *YAMLResearcher) Gather(ctx context.Context, q ResearchQuery, _ EvidenceBundle) (Evidence, error) {
	result, err := pipeline.Run(y.workflow, q.Question, "", nil, y.reg)
	if err != nil {
		return Evidence{}, fmt.Errorf("yaml researcher %s: %w", y.workflow.Name, err)
	}
	return Evidence{
		Source: y.workflow.Name,
		Title:  y.workflow.Name,
		Body:   result.Output,
	}, nil
}

// LoadResearchers loads YAML researcher files from a directory and registers them.
// Returns nil if directory doesn't exist.
func LoadResearchers(dir string, reg *Registry, providerReg *provider.ProviderRegistry) error {
	workflows, err := pipeline.LoadDir(dir)
	if err != nil {
		return fmt.Errorf("load researchers dir %s: %w", dir, err)
	}
	for _, w := range workflows {
		r := NewYAMLResearcher(w, providerReg)
		if err := reg.Register(r); err != nil {
			if err == ErrDuplicateResearcher {
				continue // skip duplicates silently
			}
			return fmt.Errorf("register researcher %s: %w", w.Name, err)
		}
	}
	return nil
}
