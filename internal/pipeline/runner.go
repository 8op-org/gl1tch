package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/provider"
)

// Result holds the output of a completed workflow run.
type Result struct {
	Workflow string
	Output   string // output of the last step
}

// RunOpts holds optional dependencies for a workflow run.
type RunOpts struct {
	Telemetry       *esearch.Telemetry
	Issue           string
	ComparisonGroup string
}

// parseWorkflowName extracts issue number and comparison group from a workflow name.
// Convention: "3918-wrapper-curl-local" → issue="3918", group="local"
func parseWorkflowName(name string) (issue, compGroup string) {
	wname := name
	if strings.HasSuffix(wname, "-local") {
		compGroup = "local"
		wname = strings.TrimSuffix(wname, "-local")
	} else if strings.HasSuffix(wname, "-claude") {
		compGroup = "claude"
		wname = strings.TrimSuffix(wname, "-claude")
	} else if strings.HasSuffix(wname, "-copilot") {
		compGroup = "copilot"
		wname = strings.TrimSuffix(wname, "-copilot")
	}
	for i, c := range wname {
		if c < '0' || c > '9' {
			if i > 0 {
				issue = wname[:i]
			}
			break
		}
	}
	return
}

// Run executes a workflow with the given input string.
// Templates use {{.input}} for user input and {{step "id"}} for prior step output.
func Run(w *Workflow, input string, defaultModel string, params map[string]string, reg *provider.ProviderRegistry, opts ...RunOpts) (*Result, error) {
	steps := make(map[string]string) // step ID → output

	var tel *esearch.Telemetry
	var issue, compGroup string
	if len(opts) > 0 {
		tel = opts[0].Telemetry
		issue = opts[0].Issue
		compGroup = opts[0].ComparisonGroup
	}
	if issue == "" || compGroup == "" {
		parsedIssue, parsedGroup := parseWorkflowName(w.Name)
		if issue == "" {
			issue = parsedIssue
		}
		if compGroup == "" {
			compGroup = parsedGroup
		}
	}

	runID := esearch.NewRunID()

	// Accumulators for workflow run summary
	var totalTokensIn, totalTokensOut int64
	var totalCostUSD float64
	var totalLatencyMS int64
	var llmSteps int
	var lastLLMOutput string

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
					model = "qwen3:8b"
				}
				result, llmErr := provider.RunOllamaWithResult(model, rendered)
				if llmErr != nil {
					err = llmErr
				} else {
					out = result.Response
					tokIn := int64(result.TokensIn)
					tokOut := int64(result.TokensOut)
					totalTokensIn += tokIn
					totalTokensOut += tokOut
					totalLatencyMS += result.Latency.Milliseconds()
					llmSteps++
					lastLLMOutput = out
					if tel != nil {
						tel.IndexLLMCall(context.Background(), esearch.LLMCallDoc{
							RunID:           runID,
							Step:            fmt.Sprintf("workflow:%s/%s", w.Name, step.ID),
							Tier:            0,
							Provider:        "ollama",
							Model:           model,
							TokensIn:        tokIn,
							TokensOut:       tokOut,
							TokensTotal:     tokIn + tokOut,
							CostUSD:         0,
							LatencyMS:       result.Latency.Milliseconds(),
							WorkflowName:    w.Name,
							Issue:           issue,
							ComparisonGroup: compGroup,
							Timestamp:       time.Now().UTC().Format(time.RFC3339),
						})
					}
				}
			default:
				result, provErr := reg.RunProviderWithResult(prov, model, rendered)
				if provErr != nil {
					err = provErr
				} else {
					out = result.Response
					tokIn := int64(result.TokensIn)
					tokOut := int64(result.TokensOut)
					totalTokensIn += tokIn
					totalTokensOut += tokOut
					totalCostUSD += result.CostUSD
					totalLatencyMS += result.Latency.Milliseconds()
					llmSteps++
					lastLLMOutput = out
					if tel != nil {
						tel.IndexLLMCall(context.Background(), esearch.LLMCallDoc{
							RunID:           runID,
							Step:            fmt.Sprintf("workflow:%s/%s", w.Name, step.ID),
							Tier:            2,
							Provider:        prov,
							Model:           model,
							TokensIn:        tokIn,
							TokensOut:       tokOut,
							TokensTotal:     tokIn + tokOut,
							CostUSD:         result.CostUSD,
							LatencyMS:       result.Latency.Milliseconds(),
							WorkflowName:    w.Name,
							Issue:           issue,
							ComparisonGroup: compGroup,
							Timestamp:       time.Now().UTC().Format(time.RFC3339),
						})
					}
				}
			}
			if err != nil {
				return nil, fmt.Errorf("step %s: %w", step.ID, err)
			}
			steps[step.ID] = out

		} else {
			return nil, fmt.Errorf("step %s: must have either 'run' or 'llm'", step.ID)
		}
	}

	// Index workflow run summary
	if tel != nil {
		// Strip markdown bold/italic markers before checking review verdict
		stripped := strings.ReplaceAll(strings.ToUpper(lastLLMOutput), "*", "")
		reviewPass := strings.Contains(stripped, "OVERALL: PASS")
		tel.IndexWorkflowRun(context.Background(), esearch.WorkflowRunDoc{
			RunID:           runID,
			WorkflowName:    w.Name,
			Issue:           issue,
			ComparisonGroup: compGroup,
			TotalSteps:      len(w.Steps),
			LLMSteps:        llmSteps,
			TotalTokensIn:   totalTokensIn,
			TotalTokensOut:  totalTokensOut,
			TotalCostUSD:    totalCostUSD,
			TotalLatencyMS:  totalLatencyMS,
			ReviewPass:      reviewPass,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
		})
	}

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
