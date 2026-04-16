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
)

// DefaultVariants is the list of variant suffixes for comparison mode.
var DefaultVariants = []string{"local", "claude", "copilot", "gemma", "grok"}

// RunOpts configures a batch run.
type RunOpts struct {
	Items      []string          // items to iterate (issues, prompts, test IDs, etc.)
	Params     map[string]string // base params passed to every workflow run
	ResultsDir string            // explicit results dir; empty = CWD/.glitch/results
	Variants   []string
	Iterations int
	Workflows  map[string]*pipeline.Workflow
	Config     BatchConfig
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
					})
					if err != nil {
						fmt.Fprintf(os.Stderr, "WARN: %s (%s, iter %d) failed: %v\n", item, v, iter, err)
					} else {
						saveResults(resultsBase, item, v, iter, result)
					}
					fmt.Printf(">>> %s (%s, iter %d) complete\n", item, v, iter)
				}(variant, wf)
			}
			wg.Wait()

			// Cross-review
			if crW, ok := opts.Workflows["cross-review"]; ok {
				fmt.Printf("\n>>> %s (cross-review, iter %d)\n", item, iter)
				result, err := pipeline.Run(crW, "", opts.Config.DefaultModel, params, opts.Config.ProviderRegistry, pipeline.RunOpts{
					Telemetry:        opts.Config.Telemetry,
					ProviderResolver: opts.Config.ProviderResolver,
					Tiers:            opts.Config.Tiers,
					EvalThreshold:    opts.Config.EvalThreshold,
					Workspace:        opts.Config.Workspace,
					Resources:        opts.Config.Resources,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: %s cross-review iter %d failed: %v\n", item, iter, err)
				} else if cr, ok := result.Steps["cross-review"]; ok {
					crDir := resultPath(resultsBase, item, "", iter)
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

	return nil
}

// resultPath computes the result directory for a batch run.
// Pattern: baseDir/<item>/iteration-<n>[/<variant>]
func resultPath(baseDir, item, variant string, iter int) string {
	dir := filepath.Join(baseDir, item, fmt.Sprintf("iteration-%d", iter))
	if variant != "" {
		dir = filepath.Join(dir, variant)
	}
	return dir
}

// saveResults writes workflow step outputs to the results directory.
func saveResults(resultsBase, item, variant string, iter int, result *pipeline.Result) {
	dir := resultPath(resultsBase, item, variant, iter)
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
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	metaJSON, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(dir, "run.json"), metaJSON, 0o644)

	fmt.Printf("  Results saved: %s/\n", dir)
}
