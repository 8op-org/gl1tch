package pipeline

import (
	"bytes"
	"context"
	"fmt"

	"github.com/adam-stokes/orcai/internal/plugin"
)

// Run executes a pipeline against the given plugin manager.
// userInput is the initial value injected for the first input step.
// Returns the final output string.
func Run(ctx context.Context, p *Pipeline, mgr *plugin.Manager, userInput string) (string, error) {
	vars := make(map[string]string)

	// Index steps by ID for branch lookups.
	byID := make(map[string]*Step, len(p.Steps))
	order := make([]string, 0, len(p.Steps))
	for i := range p.Steps {
		byID[p.Steps[i].ID] = &p.Steps[i]
		order = append(order, p.Steps[i].ID)
	}

	visited := make(map[string]bool)
	queue := append([]string(nil), order...)

	// lastOutput tracks the most recent output, seeded by userInput so that
	// the first plugin step receives user input when no explicit prompt is set.
	lastOutput := userInput
	var lastPluginOutput string

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		if visited[id] {
			continue
		}
		visited[id] = true

		step, ok := byID[id]
		if !ok {
			return "", fmt.Errorf("pipeline: unknown step id %q", id)
		}

		switch step.Type {
		case "input":
			vars[step.ID+".out"] = userInput
			lastOutput = userInput

		case "output":
			// Return whatever the last plugin step produced.
			return lastPluginOutput, nil

		default:
			// Plugin step.
			pl, ok := mgr.Get(step.Plugin)
			if !ok {
				return "", fmt.Errorf("pipeline: plugin %q not found", step.Plugin)
			}

			raw := step.Prompt + step.Input
			if raw == "" {
				// No explicit prompt/input: pass the most recent output as the input.
				raw = lastOutput
			}
			promptOrInput := Interpolate(raw, vars)
			stepVars := make(map[string]string, len(vars)+1)
			for k, v := range vars {
				stepVars[k] = v
			}
			stepVars["model"] = step.Model

			var buf bytes.Buffer
			if err := pl.Execute(ctx, promptOrInput, stepVars, &buf); err != nil {
				return "", fmt.Errorf("pipeline: step %q: %w", step.ID, err)
			}
			output := buf.String()
			vars[step.ID+".out"] = output
			lastPluginOutput = output
			lastOutput = output

			// Evaluate branch condition if present.
			if step.Condition.If != "" {
				if EvalCondition(step.Condition.If, output) {
					if step.Condition.Then != "" {
						// Jump: prepend branch target, skip remaining queue.
						queue = append([]string{step.Condition.Then}, filterOut(queue, step.Condition.Else)...)
					}
				} else {
					if step.Condition.Else != "" {
						queue = append([]string{step.Condition.Else}, filterOut(queue, step.Condition.Then)...)
					}
				}
			}
		}
	}

	return lastPluginOutput, nil
}

// filterOut removes a single value from a slice.
func filterOut(ss []string, remove string) []string {
	if remove == "" {
		return ss
	}
	out := ss[:0:len(ss)]
	for _, s := range ss {
		if s != remove {
			out = append(out, s)
		}
	}
	return out
}
