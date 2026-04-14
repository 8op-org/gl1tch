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
	Output   string            // output of the last step
	Steps    map[string]string // all step outputs keyed by step ID
}

// RunOpts holds optional dependencies for a workflow run.
type RunOpts struct {
	Telemetry        *esearch.Telemetry
	Issue            string
	ComparisonGroup  string
	ProviderResolver provider.ResolverFunc
	Tiers            []provider.TierConfig
	EvalThreshold    int
}

// parseWorkflowName extracts issue number and comparison group from a workflow name.
// Convention: "issue-to-pr-local" → group="local", "3918-wrapper-curl-copilot" → issue="3918", group="copilot"
// The last hyphen-separated segment is the comparison group.
func parseWorkflowName(name string) (issue, compGroup string) {
	if idx := strings.LastIndex(name, "-"); idx > 0 {
		compGroup = name[idx+1:]
		name = name[:idx]
	}
	for i, c := range name {
		if c < '0' || c > '9' {
			if i > 0 {
				issue = name[:i]
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
	var providerResolver provider.ResolverFunc
	var tiers []provider.TierConfig
	var evalThreshold int
	if len(opts) > 0 {
		tel = opts[0].Telemetry
		issue = opts[0].Issue
		compGroup = opts[0].ComparisonGroup
		providerResolver = opts[0].ProviderResolver
		tiers = opts[0].Tiers
		evalThreshold = opts[0].EvalThreshold
	}
	if evalThreshold == 0 {
		evalThreshold = 4
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
	// Fall back to params for issue if not derived from workflow name
	if issue == "" && params != nil {
		issue = params["issue"]
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

			// Prepend skill content to prompt if specified
			if step.LLM.Skill != "" {
				skillContent, skillErr := loadSkill(step.LLM.Skill)
				if skillErr != nil {
					return nil, fmt.Errorf("step %s: skill %q: %w", step.ID, step.LLM.Skill, skillErr)
				}
				rendered = skillContent + "\n\n---\n\n" + rendered
			}

			fmt.Printf("  > %s\n", step.ID)

			prov := strings.ToLower(step.LLM.Provider)
			model := step.LLM.Model
			if model == "" {
				model = defaultModel
			}

			var out string
			var stepTier int
			var stepEscalated bool
			var stepEscalationChain []int
			var stepEvalScores []int
			var stepCost float64
			var stepTokensIn, stepTokensOut int

			// Smart routing: no provider AND no pinned tier AND tiers available
			useSmart := prov == "" && step.LLM.Tier == nil && len(tiers) > 0
			// Pinned tier: explicit tier set
			usePinned := step.LLM.Tier != nil && len(tiers) > 0

			if useSmart || usePinned {
				activeTiers := tiers
				if usePinned {
					tierIdx := *step.LLM.Tier
					if tierIdx >= 0 && tierIdx < len(tiers) {
						// Single-tier slice — RunSmart treats it as final tier (no eval)
						activeTiers = tiers[tierIdx : tierIdx+1]
					}
				}

				runner := provider.NewTieredRunner(activeTiers, reg)
				runner.Resolver = providerResolver

				format := step.LLM.Format
				evalFn := func(evalModel, evalPrompt string) (provider.LLMResult, error) {
					return provider.RunOllamaWithResult(defaultModel, evalPrompt)
				}

				rr, llmErr := runner.RunSmart(context.Background(), rendered, format, evalThreshold, evalFn)
				if llmErr != nil {
					err = llmErr
				} else {
					out = rr.Response
					stepTier = rr.Tier
					if usePinned {
						stepTier = *step.LLM.Tier
					}
					stepEscalated = rr.Escalated
					stepEscalationChain = rr.EscalationChain
					stepEvalScores = rr.EvalScores
					stepCost = rr.CostUSD
					stepTokensIn = rr.TokensIn
					stepTokensOut = rr.TokensOut
				}
			} else {
				// Original dispatch: explicit provider or no tiers configured
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
						stepTokensIn = result.TokensIn
						stepTokensOut = result.TokensOut
					}
				default:
					var resolved bool

					// Agent providers (claude, copilot, gemini) — full tool-using agents
					if provider.IsAgent(prov) {
						agent := provider.KnownAgents[prov]
						result, llmErr := agent.Run(step.LLM.Model, rendered)
						if llmErr != nil {
							err = llmErr
						} else {
							resolved = true
							out = result.Response
							stepTokensIn = result.TokensIn
							stepTokensOut = result.TokensOut
							stepCost = result.CostUSD
						}
					}

					// Config-defined providers (openai-compatible, etc.)
					if !resolved && providerResolver != nil {
						if fn, ok := providerResolver(prov); ok {
							resolved = true
							result, llmErr := fn(step.LLM.Model, rendered)
							if llmErr != nil {
								err = llmErr
							} else {
								out = result.Response
								stepTokensIn = result.TokensIn
								stepTokensOut = result.TokensOut
								stepCost = result.CostUSD
							}
						}
					}

					// Shell-template providers from ~/.config/glitch/providers/
					if !resolved {
						result, provErr := reg.RunProviderWithResult(prov, model, rendered)
						if provErr != nil {
							err = provErr
						} else {
							out = result.Response
							stepTokensIn = result.TokensIn
							stepTokensOut = result.TokensOut
							stepCost = result.CostUSD
						}
					}
				}
			}

			if err != nil {
				return nil, fmt.Errorf("step %s: %w", step.ID, err)
			}

			tokIn := int64(stepTokensIn)
			tokOut := int64(stepTokensOut)
			totalTokensIn += tokIn
			totalTokensOut += tokOut
			totalCostUSD += stepCost
			llmSteps++
			lastLLMOutput = out

			if tel != nil {
				reason := ""
				if stepEscalated {
					reason = "eval"
				}
				tel.IndexLLMCall(context.Background(), esearch.LLMCallDoc{
					RunID:            runID,
					Step:             fmt.Sprintf("workflow:%s/%s", w.Name, step.ID),
					Tier:             stepTier,
					Provider:         prov,
					Model:            model,
					TokensIn:         tokIn,
					TokensOut:        tokOut,
					TokensTotal:      tokIn + tokOut,
					CostUSD:          stepCost,
					Escalated:        stepEscalated,
					EscalationReason: reason,
					EscalationChain:  stepEscalationChain,
					EvalScores:       stepEvalScores,
					FinalTier:        stepTier,
					WorkflowName:     w.Name,
					Issue:            issue,
					ComparisonGroup:  compGroup,
					Timestamp:        time.Now().UTC().Format(time.RFC3339),
				})
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

		// Parse per-criterion PASS/FAIL for confidence score
		passed, total := 0, 0
		for _, line := range strings.Split(stripped, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "OVERALL") {
				continue // skip the summary line
			}
			hasPass := strings.Contains(line, "PASS")
			hasFail := strings.Contains(line, "FAIL")
			if hasPass && !hasFail {
				passed++
				total++
			} else if hasFail {
				total++
			}
		}
		confidence := 0.0
		if total > 0 {
			confidence = float64(passed) / float64(total)
		}

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
			Confidence:      confidence,
			CriteriaPassed:  passed,
			CriteriaTotal:   total,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
		})

		// If this is a cross-review workflow, parse and index per-variant scores
		if strings.Contains(w.Name, "cross-review") {
			crossScores := ParseCrossReview(lastLLMOutput)
			iteration := ""
			if params != nil {
				iteration = params["iteration"]
			}
			for _, cs := range crossScores {
				conf := 0.0
				if cs.Total > 0 {
					conf = float64(cs.Passed) / float64(cs.Total)
				}
				tel.IndexCrossReview(context.Background(), esearch.CrossReviewDoc{
					RunID:        runID,
					Issue:        issue,
					Iteration:    iteration,
					Variant:      cs.Variant,
					Passed:       cs.Passed,
					Total:        cs.Total,
					Confidence:   conf,
					Winner:       cs.Winner,
					WorkflowName: w.Name,
					Timestamp:    time.Now().UTC().Format(time.RFC3339),
				})
			}
		}
	}

	last := w.Steps[len(w.Steps)-1]
	return &Result{
		Workflow: w.Name,
		Output:   steps[last.ID],
		Steps:    steps,
	}, nil
}

// CrossReviewScore holds a parsed per-variant score from a cross-review.
type CrossReviewScore struct {
	Variant string
	Passed  int
	Total   int
	Winner  bool
}

// ParseCrossReview extracts per-variant scores from cross-review LLM output.
func ParseCrossReview(output string) []CrossReviewScore {
	var scores []CrossReviewScore
	upper := strings.ToUpper(strings.ReplaceAll(output, "*", ""))
	lines := strings.Split(upper, "\n")

	// Find WINNER line
	winner := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "WINNER:") {
			winner = strings.TrimSpace(strings.TrimPrefix(line, "WINNER:"))
			break
		}
	}

	// Parse --- VARIANT --- sections
	currentVariant := ""
	passed, total := 0, 0
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect variant header: --- LOCAL --- or --- CLAUDE --- etc
		if strings.HasPrefix(line, "---") && strings.HasSuffix(line, "---") {
			// Save previous variant if any
			if currentVariant != "" && total > 0 {
				isWinner := strings.Contains(strings.ToUpper(winner), strings.ToUpper(currentVariant))
				scores = append(scores, CrossReviewScore{
					Variant: strings.ToLower(currentVariant),
					Passed:  passed,
					Total:   total,
					Winner:  isWinner,
				})
			}
			// Extract variant name
			name := strings.Trim(line, "- ")
			if name != "" {
				currentVariant = name
				passed = 0
				total = 0
			}
			continue
		}

		// Skip SCORE: lines (we count from PASS/FAIL)
		if strings.HasPrefix(line, "SCORE:") {
			continue
		}

		// Skip OVERALL and WINNER lines
		if strings.HasPrefix(line, "OVERALL") || strings.HasPrefix(line, "WINNER") {
			continue
		}

		// Count PASS/FAIL within current variant section
		if currentVariant != "" {
			hasPass := strings.Contains(line, "PASS")
			hasFail := strings.Contains(line, "FAIL")
			if hasPass && !hasFail {
				passed++
				total++
			} else if hasFail {
				total++
			}
		}
	}

	// Don't forget the last variant
	if currentVariant != "" && total > 0 {
		isWinner := strings.Contains(strings.ToUpper(winner), strings.ToUpper(currentVariant))
		scores = append(scores, CrossReviewScore{
			Variant: strings.ToLower(currentVariant),
			Passed:  passed,
			Total:   total,
			Winner:  isWinner,
		})
	}

	return scores
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

// loadSkill resolves a skill name or path to its content.
//
// Resolution order:
//  1. Absolute or relative file path (if the string contains "/" or ends in ".md")
//  2. Skill name looked up in standard locations:
//     - .claude/skills/<name>/SKILL.md (project-local)
//     - ~/.config/glitch/skills/<name>/SKILL.md (user-global)
//     - skills/<name>/SKILL.md (gl1tch built-in)
func loadSkill(nameOrPath string) (string, error) {
	// Direct file path
	if strings.Contains(nameOrPath, "/") || strings.HasSuffix(nameOrPath, ".md") {
		data, err := os.ReadFile(nameOrPath)
		if err != nil {
			return "", fmt.Errorf("read skill file: %w", err)
		}
		return string(data), nil
	}

	// Search standard skill locations
	searchPaths := []string{
		filepath.Join(".claude", "skills", nameOrPath, "SKILL.md"),
		filepath.Join(".cursor", "skills", nameOrPath, "SKILL.md"),
	}

	// User-global location
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(home, ".config", "glitch", "skills", nameOrPath, "SKILL.md"),
		)
	}

	for _, p := range searchPaths {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("skill %q not found in: %s", nameOrPath, strings.Join(searchPaths, ", "))
}
