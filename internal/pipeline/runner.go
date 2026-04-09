package pipeline

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/8op-org/gl1tch/internal/provider"
)

// Result holds the output of a completed workflow run.
type Result struct {
	Workflow string
	Output   string // output of the last step
}

// Run executes a workflow with the given input string.
// Templates use {{.input}} for user input and {{step "id"}} for prior step output.
func Run(w *Workflow, input string, defaultModel string) (*Result, error) {
	steps := make(map[string]string) // step ID → output

	for _, step := range w.Steps {
		data := map[string]any{
			"input": input,
		}

		if step.Run != "" {
			rendered, err := render(step.Run, data, steps)
			if err != nil {
				return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
			}
			fmt.Printf("  > %s\n", step.ID)
			out, err := provider.RunShell(rendered)
			if err != nil {
				return nil, fmt.Errorf("step %s: %w", step.ID, err)
			}
			steps[step.ID] = out

		} else if step.LLM != nil {
			rendered, err := render(step.LLM.Prompt, data, steps)
			if err != nil {
				return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
			}
			fmt.Printf("  > %s\n", step.ID)

			prov := strings.ToLower(step.LLM.Provider)
			model := step.LLM.Model
			if model == "" {
				model = defaultModel
			}

			var out string
			switch prov {
			case "claude":
				out, err = provider.RunClaude(model, rendered)
			default: // ollama
				if model == "" {
					model = "qwen2.5:7b"
				}
				out, err = provider.RunOllama(model, rendered)
			}
			if err != nil {
				return nil, fmt.Errorf("step %s: %w", step.ID, err)
			}
			steps[step.ID] = out

		} else {
			return nil, fmt.Errorf("step %s: must have either 'run' or 'llm'", step.ID)
		}
	}

	// Find the last step's output.
	last := w.Steps[len(w.Steps)-1]
	return &Result{
		Workflow: w.Name,
		Output:   steps[last.ID],
	}, nil
}

func render(tmpl string, data map[string]any, steps map[string]string) (string, error) {
	funcMap := template.FuncMap{
		"step": func(id string) string {
			return steps[id]
		},
	}
	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
