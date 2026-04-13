package pipeline

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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
func Run(w *Workflow, input string, defaultModel string, params map[string]string, reg *provider.ProviderRegistry) (*Result, error) {
	steps := make(map[string]string) // step ID → output

	for i, step := range w.Steps {
		data := map[string]any{
			"input": input,
			"param": params,
		}

		if step.Save != "" {
			rendered, err := render(step.Save, data, steps)
			if err != nil {
				return nil, fmt.Errorf("step %s: template: %w", step.ID, err)
			}
			fmt.Printf("  > %s (save → %s)\n", step.ID, rendered)
			// Determine which step's output to save
			sourceStep := step.SaveStep
			if sourceStep == "" && i > 0 {
				sourceStep = w.Steps[i-1].ID
			}
			content := steps[sourceStep]
			if err := os.MkdirAll(filepath.Dir(rendered), 0o755); err != nil {
				return nil, fmt.Errorf("step %s: mkdir: %w", step.ID, err)
			}
			if err := os.WriteFile(rendered, []byte(content), 0o644); err != nil {
				return nil, fmt.Errorf("step %s: write: %w", step.ID, err)
			}
			steps[step.ID] = fmt.Sprintf("saved %s to %s", sourceStep, rendered)

		} else if step.Run != "" {
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
			case "ollama", "":
				if model == "" {
					model = "qwen2.5:7b"
				}
				out, err = provider.RunOllama(model, rendered)
			default:
				out, err = reg.RunProvider(prov, rendered)
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
		// stepfile writes step output to a temp file and returns the path.
		// Use in shell steps where inline content would break escaping:
		//   cat "{{stepfile "fetch-issue"}}"
		"stepfile": func(id string) string {
			content, ok := steps[id]
			if !ok {
				return ""
			}
			f, err := os.CreateTemp("", "glitch-step-*")
			if err != nil {
				return ""
			}
			f.WriteString(content)
			f.Close()
			return f.Name()
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
