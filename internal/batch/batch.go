package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/store"
)

// DefaultVariants is the list of variant suffixes for comparison mode.
var DefaultVariants = []string{"local", "claude", "copilot", "gemma", "grok"}

// RunOpts configures a batch run.
type RunOpts struct {
	Items        []string          // items to iterate (issues, prompts, test IDs, etc.)
	Params       map[string]string // base params passed to every workflow run
	ResultsDir   string            // explicit results dir; empty = CWD/.glitch/results
	Variants     []string
	Iterations   int
	Workflows    map[string]*pipeline.Workflow
	WorkflowsDir string // directory for resolving call-workflow targets in nested invocations
	Config       BatchConfig
	Name         string       // optional human-readable name for the parent batch run row
	Store        *store.Store // when non-nil, records parent + child run rows with parent_run_id linkage
}

// BatchConfig holds dependencies for running workflows.
type BatchConfig struct {
	DefaultModel     string
	ProviderRegistry *provider.ProviderRegistry
	ProviderResolver provider.ResolverFunc
	Telemetry        *esearch.Telemetry
	Tiers            []provider.TierConfig
	EvalThreshold    int
	Workspace        string                       // resolved workspace name
	Resources        map[string]map[string]string // resource bindings for .resource.<name>.<field>
}

// variantWorkflows groups loaded workflows into variant sets for a given item.
// Matches workflows where the last hyphen-separated segment is a known variant
// and the prefix contains the item identifier (or is a generic name like "issue-to-pr").
func variantWorkflows(workflows map[string]*pipeline.Workflow, item string, variants []string) map[string]*pipeline.Workflow {
	variantSet := make(map[string]bool, len(variants))
	for _, v := range variants {
		variantSet[v] = true
	}

	result := make(map[string]*pipeline.Workflow)
	for name, wf := range workflows {
		idx := strings.LastIndex(name, "-")
		if idx < 0 {
			continue
		}
		suffix := name[idx+1:]
		if !variantSet[suffix] {
			continue
		}
		prefix := name[:idx]
		if strings.Contains(prefix, item) || prefix == "issue-to-pr" {
			result[suffix] = wf
		}
	}
	return result
}

// Run executes the full batch: items × variants × iterations + cross-reviews + manifests.
func Run(ctx context.Context, opts RunOpts) error {
	resultsBase := opts.ResultsDir
	if resultsBase == "" {
		cwd, _ := os.Getwd()
		resultsBase = filepath.Join(cwd, ".glitch", "results")
	}

	// Create the parent batch run row. Children (per variant × iteration) will
	// link to this row via parent_run_id. If Store is nil (tests), skip and use
	// 0 as the parent id — child rows will be skipped too.
	var parentID int64
	if opts.Store != nil {
		name := opts.Name
		if name == "" {
			name = "batch"
		}
		id, err := opts.Store.RecordRun(store.RunRecord{
			Kind:         "batch",
			Name:         name,
			WorkflowName: "batch",
			Workspace:    opts.Config.Workspace,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: record batch parent run: %v\n", err)
		} else {
			parentID = id
			defer func() { _ = opts.Store.FinishRun(parentID, "", 0) }()
		}
	}

	for iter := 1; iter <= opts.Iterations; iter++ {
		fmt.Printf("\n=========================================\n")
		fmt.Printf("ITERATION %d of %d\n", iter, opts.Iterations)
		fmt.Printf("=========================================\n")

		for _, item := range opts.Items {
			// Build per-run params: base params + item + iteration
			params := make(map[string]string, len(opts.Params)+2)
			for k, v := range opts.Params {
				params[k] = v
			}
			params["item"] = item
			params["iteration"] = fmt.Sprintf("%d", iter)
			// Backwards compat: also set "issue" so existing workflows work
			if _, hasIssue := params["issue"]; !hasIssue {
				params["issue"] = item
			}

			// Discover variant workflows for this item
			vws := variantWorkflows(opts.Workflows, item, opts.Variants)
			if len(vws) == 0 {
				fmt.Fprintf(os.Stderr, "WARN: no variant workflows found for %q, skipping\n", item)
				continue
			}

			// Pre-run shared (run ...) steps once using the first available variant.
			// These are data-gathering shell steps identical across all variants.
			var seedSteps map[string]string
			for _, wf := range vws {
				fmt.Printf("\n>>> %s (pre-run shared steps, iter %d)\n", item, iter)
				var err error
				seedSteps, err = pipeline.PreRunSharedSteps(wf, params)
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: %s shared steps failed: %v\n", item, err)
				}
				break
			}

			// Run all variants in parallel, seeded with shared step results
			var wg sync.WaitGroup
			for variant, wf := range vws {
				wg.Add(1)
				go func(v string, wf *pipeline.Workflow) {
					defer wg.Done()
					fmt.Printf("\n>>> %s (%s, iter %d)\n", item, v, iter)

					// Pre-create the child run row so the on-disk path can include
					// its DB row id. If Store is nil or RecordRun fails, fall back
					// to childID=0 (layout will embed 0 for the runid segment).
					var childID int64
					var childCreator func(int64, string) (int64, error)
					if opts.Store != nil {
						id, err := opts.Store.RecordRun(store.RunRecord{
							Kind:         "workflow",
							Name:         item,
							Variant:      v,
							WorkflowName: wf.Name,
							ParentRunID:  parentID,
							Workspace:    opts.Config.Workspace,
						})
						if err != nil {
							fmt.Fprintf(os.Stderr, "WARN: record child run %s (%s, iter %d): %v\n", item, v, iter, err)
						} else {
							childID = id
							defer func() { _ = opts.Store.FinishRun(childID, "", 0) }()
						}
						ws := opts.Config.Workspace
						childCreator = func(parent int64, wfName string) (int64, error) {
							return opts.Store.RecordRun(store.RunRecord{
								Kind:         "workflow",
								Name:         wfName,
								WorkflowName: wfName,
								ParentRunID:  parent,
								Workspace:    ws,
							})
						}
					}

					result, err := pipeline.Run(wf, "", opts.Config.DefaultModel, params, opts.Config.ProviderRegistry, pipeline.RunOpts{
						Telemetry:        opts.Config.Telemetry,
						ProviderResolver: opts.Config.ProviderResolver,
						Tiers:            opts.Config.Tiers,
						EvalThreshold:    opts.Config.EvalThreshold,
						Issue:            item,
						ComparisonGroup:  v,
						SeedSteps:        seedSteps,
						Workspace:        opts.Config.Workspace,
						Resources:        opts.Config.Resources,
						WorkflowsDir:     opts.WorkflowsDir,
						ParentRunID:      childID,
						ChildRunCreator:  childCreator,
						StepRecorder:     newStepRecorder(opts.Store, childID),
					})
					if err != nil {
						fmt.Fprintf(os.Stderr, "WARN: %s (%s, iter %d) failed: %v\n", item, v, iter, err)
					} else {
						saveResults(resultsBase, item, v, iter, childID, result)
					}
					fmt.Printf(">>> %s (%s, iter %d) complete\n", item, v, iter)
				}(variant, wf)
			}
			wg.Wait()

			// Cross-review
			if crW, ok := opts.Workflows["cross-review"]; ok {
				fmt.Printf("\n>>> %s (cross-review, iter %d)\n", item, iter)
				var crChildID int64
				var crCreator func(int64, string) (int64, error)
				if opts.Store != nil {
					id, err := opts.Store.RecordRun(store.RunRecord{
						Kind:         "workflow",
						Name:         item,
						Variant:      "cross-review",
						WorkflowName: crW.Name,
						ParentRunID:  parentID,
						Workspace:    opts.Config.Workspace,
					})
					if err != nil {
						fmt.Fprintf(os.Stderr, "WARN: record cross-review run %s iter %d: %v\n", item, iter, err)
					} else {
						crChildID = id
						defer func() { _ = opts.Store.FinishRun(crChildID, "", 0) }()
					}
					ws := opts.Config.Workspace
					crCreator = func(parent int64, wfName string) (int64, error) {
						return opts.Store.RecordRun(store.RunRecord{
							Kind:         "workflow",
							Name:         wfName,
							WorkflowName: wfName,
							ParentRunID:  parent,
							Workspace:    ws,
						})
					}
				}
				result, err := pipeline.Run(crW, "", opts.Config.DefaultModel, params, opts.Config.ProviderRegistry, pipeline.RunOpts{
					Telemetry:        opts.Config.Telemetry,
					ProviderResolver: opts.Config.ProviderResolver,
					Tiers:            opts.Config.Tiers,
					EvalThreshold:    opts.Config.EvalThreshold,
					Workspace:        opts.Config.Workspace,
					Resources:        opts.Config.Resources,
					WorkflowsDir:     opts.WorkflowsDir,
					ParentRunID:      crChildID,
					ChildRunCreator:  crCreator,
					StepRecorder:     newStepRecorder(opts.Store, crChildID),
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: %s cross-review iter %d failed: %v\n", item, iter, err)
				} else if cr, ok := result.Steps["cross-review"]; ok {
					itemDir := filepath.Join(resultsBase, item)
					crDir := resultPath(itemDir, "cross-review", iter, crChildID)
					os.MkdirAll(crDir, 0o755)
					os.WriteFile(filepath.Join(crDir, "cross-review.md"), []byte(cr), 0o644)
				}
			}
		}
	}

	// Generate manifests
	fmt.Printf("\n=========================================\n")
	fmt.Printf("GENERATING MANIFESTS\n")
	fmt.Printf("=========================================\n")

	for _, item := range opts.Items {
		itemDir := filepath.Join(resultsBase, item)
		m, err := GenerateManifest(itemDir, item, opts.Variants, opts.Iterations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: manifest for %s: %v\n", item, err)
			continue
		}
		if err := WriteManifest(itemDir, m, opts.Variants, opts.Iterations); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: write manifest %s: %v\n", item, err)
			continue
		}
		fmt.Printf("  %s: %s/manifest.md (winner: %s iter %d)\n", item, itemDir, m.BestVariant, m.BestIteration)
	}

	// CLEAR Review: batch-level reflection
	if opts.Config.ProviderRegistry != nil && opts.Config.DefaultModel != "" {
		fmt.Printf("\n=========================================\n")
		fmt.Printf("BATCH REFLECTION\n")
		fmt.Printf("=========================================\n")

		var summaries strings.Builder
		for _, item := range opts.Items {
			itemDir := filepath.Join(resultsBase, item)
			m, err := GenerateManifest(itemDir, item, opts.Variants, opts.Iterations)
			if err != nil || m.BestVariant == "" {
				continue
			}
			fmt.Fprintf(&summaries, "Item %s: winner=%s (%d/%d)\n", item, m.BestVariant, m.BestScore, m.BestTotal)
		}

		if summaries.Len() > 0 {
			reflectionPrompt := fmt.Sprintf(`You are reviewing the results of a batch comparison across multiple items and variants.

Batch summary:
%s
Variants tested: %s

Produce a structured batch learning with EXACTLY this format:

FINDING: <one sentence — what pattern emerged across all items?>
MODEL_INSIGHT: <for each variant, one sentence about consistency/reliability>
CONFIDENCE: <high|medium|low — how consistent were results across items?>
RECOMMENDATION: <one sentence — what should future batch runs do differently?>`, summaries.String(), strings.Join(opts.Variants, ", "))

			reflectWf := &pipeline.Workflow{
				Name: "batch-reflection",
				Steps: []pipeline.Step{{
					ID: "reflect",
					LLM: &pipeline.LLMStep{
						Prompt: reflectionPrompt,
						Model:  opts.Config.DefaultModel,
					},
				}},
			}
			for i := range reflectWf.Steps {
				reflectWf.Items = append(reflectWf.Items, pipeline.WorkflowItem{Step: &reflectWf.Steps[i]})
			}

			result, err := pipeline.Run(reflectWf, "", opts.Config.DefaultModel, nil, opts.Config.ProviderRegistry, pipeline.RunOpts{
				Telemetry:        opts.Config.Telemetry,
				ProviderResolver: opts.Config.ProviderResolver,
				Workspace:        opts.Config.Workspace,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARN: batch reflection failed: %v\n", err)
			} else {
				fmt.Printf("\n%s\n", result.Output)

				if opts.Config.Telemetry != nil {
					finding, confidence, recommendation := "", "", ""
					for _, line := range strings.Split(result.Output, "\n") {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, "FINDING:") {
							finding = strings.TrimSpace(strings.TrimPrefix(line, "FINDING:"))
						} else if strings.HasPrefix(line, "CONFIDENCE:") {
							confidence = strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
						} else if strings.HasPrefix(line, "RECOMMENDATION:") {
							recommendation = strings.TrimSpace(strings.TrimPrefix(line, "RECOMMENDATION:"))
						}
					}

					opts.Config.Telemetry.IndexLearning(context.Background(), esearch.LearningDoc{
						RunID:          fmt.Sprintf("batch-%d", parentID),
						Objective:      "batch aggregate",
						Scope:          "batch",
						Finding:        finding,
						Confidence:     confidence,
						Recommendation: recommendation,
						ModelsTested:   opts.Variants,
						WorkflowName:   "batch",
						Workspace:      opts.Config.Workspace,
						Timestamp:      time.Now().UTC().Format(time.RFC3339),
					})
				}
			}
		}
	}

	return nil
}

// resultPath computes the result directory for one child run inside a batch.
// Pattern: baseDir/children/<variant>-<iter>-<runid>
// Callers pass the per-item directory (e.g. resultsBase/<item>) as baseDir so
// each item keeps its own children/ pool; runID is the DB row id of the child
// run (0 when Store is not wired).
func resultPath(baseDir string, variant string, iter int, runID int64) string {
	return filepath.Join(baseDir, "children", fmt.Sprintf("%s-%d-%d", variant, iter, runID))
}

// saveResults writes workflow step outputs to the results directory.
func saveResults(resultsBase, item, variant string, iter int, runID int64, result *pipeline.Result) {
	itemDir := filepath.Join(resultsBase, item)
	dir := resultPath(itemDir, variant, iter, runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: mkdir %s: %v\n", dir, err)
		return
	}

	// Save all step outputs by step ID
	for stepID, content := range result.Steps {
		if strings.TrimSpace(content) == "" {
			continue
		}
		os.WriteFile(filepath.Join(dir, stepID+".md"), []byte(content), 0o644)
	}

	// Write run metadata
	meta := map[string]interface{}{
		"item":      item,
		"variant":   variant,
		"iteration": iter,
		"run_id":    runID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	metaJSON, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(dir, "run.json"), metaJSON, 0o644)

	fmt.Printf("  Results saved: %s/\n", dir)
}

// newStepRecorder returns a pipeline.StepRecorder that writes each completed
// step into the given store under runID. Returns nil when the store is absent
// or runID is zero so callers can pass the result directly into RunOpts.
func newStepRecorder(s *store.Store, runID int64) func(pipeline.StepRecord) {
	if s == nil || runID == 0 {
		return nil
	}
	return func(rec pipeline.StepRecord) {
		_ = s.RecordStep(store.StepRecord{
			RunID:      runID,
			StepID:     rec.StepID,
			Prompt:     rec.Prompt,
			Output:     rec.Output,
			Model:      rec.Model,
			DurationMs: rec.DurationMs,
			Kind:       rec.Kind,
			ExitStatus: rec.ExitStatus,
			TokensIn:   rec.TokensIn,
			TokensOut:  rec.TokensOut,
		})
	}
}
