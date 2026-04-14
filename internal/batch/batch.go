package batch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
)

// DefaultVariants is the list of variant suffixes for comparison mode.
var DefaultVariants = []string{"local", "claude", "copilot", "gemma", "llama"}

// RunOpts configures a batch run.
type RunOpts struct {
	Issues     []string
	Repo       string
	RepoPath   string
	ResultsDir string // explicit results dir; empty = CWD/.glitch/results
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
}

// Run executes the full batch: issues x variants x iterations + cross-reviews + manifests.
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

		for _, issue := range opts.Issues {
			for _, variant := range opts.Variants {
				wfName := fmt.Sprintf("issue-to-pr-%s", variant)
				w, ok := opts.Workflows[wfName]
				if !ok {
					fmt.Fprintf(os.Stderr, "WARN: workflow %q not found, skipping\n", wfName)
					continue
				}

				fmt.Printf("\n>>> #%s (%s, iter %d)\n", issue, variant, iter)
				params := map[string]string{
					"issue":     issue,
					"repo":      opts.Repo,
					"iteration": fmt.Sprintf("%d", iter),
				}
				result, err := pipeline.Run(w, "", opts.Config.DefaultModel, params, opts.Config.ProviderRegistry, pipeline.RunOpts{
					Telemetry:        opts.Config.Telemetry,
					ProviderResolver: opts.Config.ProviderResolver,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: #%s (%s, iter %d) failed: %v\n", issue, variant, iter, err)
				} else {
					saveResults(resultsBase, issue, variant, iter, opts.Repo, result)
				}
				fmt.Printf(">>> #%s (%s, iter %d) complete\n", issue, variant, iter)
			}

			// Cross-review
			if crW, ok := opts.Workflows["cross-review"]; ok {
				fmt.Printf("\n>>> #%s (cross-review, iter %d)\n", issue, iter)
				params := map[string]string{
					"issue":     issue,
					"repo":      opts.Repo,
					"iteration": fmt.Sprintf("%d", iter),
				}
				result, err := pipeline.Run(crW, "", opts.Config.DefaultModel, params, opts.Config.ProviderRegistry, pipeline.RunOpts{
					Telemetry:        opts.Config.Telemetry,
					ProviderResolver: opts.Config.ProviderResolver,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: #%s cross-review iter %d failed: %v\n", issue, iter, err)
				} else if cr, ok := result.Steps["cross-review"]; ok {
					crDir := filepath.Join(resultsBase, issue, fmt.Sprintf("iteration-%d", iter))
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

	for _, issue := range opts.Issues {
		issueDir := filepath.Join(resultsBase, issue)
		m, err := GenerateManifest(issueDir, issue, opts.Variants, opts.Iterations)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: manifest for #%s: %v\n", issue, err)
			continue
		}
		if err := WriteManifest(issueDir, m, opts.Variants, opts.Iterations); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: write manifest #%s: %v\n", issue, err)
			continue
		}
		fmt.Printf("  #%s: %s/manifest.md (winner: %s iter %d)\n", issue, issueDir, m.BestVariant, m.BestIteration)
	}

	return nil
}

// saveResults writes workflow step outputs to the results directory.
func saveResults(resultsBase, issue, variant string, iter int, repo string, result *pipeline.Result) {
	dir := filepath.Join(resultsBase, issue, fmt.Sprintf("iteration-%d", iter), variant)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: mkdir %s: %v\n", dir, err)
		return
	}

	// Save known step outputs
	stepFiles := map[string]string{
		"classify": "classification.json",
		"research": "plan.md",
		"review":   "review.md",
	}
	for stepID, filename := range stepFiles {
		if content, ok := result.Steps[stepID]; ok {
			os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644)
		}
	}

	// Extract PR sections from build-pr step
	if prContent, ok := result.Steps["build-pr"]; ok {
		extractSection(prContent, "---PR_TITLE---", "---END_PR_TITLE---", filepath.Join(dir, "pr-title.txt"))
		extractSection(prContent, "---PR_BODY---", "---END_PR_BODY---", filepath.Join(dir, "pr-body.md"))
		extractSection(prContent, "---NEXT_STEPS---", "---END_NEXT_STEPS---", filepath.Join(dir, "next-steps.md"))
		// Fallback: save full PR output if title extraction got nothing
		if data, err := os.ReadFile(filepath.Join(dir, "pr-title.txt")); err != nil || len(strings.TrimSpace(string(data))) == 0 {
			os.WriteFile(filepath.Join(dir, "pr-artifacts.md"), []byte(prContent), 0o644)
		}
	}

	// Write run metadata
	runJSON := fmt.Sprintf(`{"repo":"%s","issue":"%s","iteration":%d,"variant":"%s","timestamp":"%s"}`,
		repo, issue, iter, variant, time.Now().UTC().Format(time.RFC3339))
	os.WriteFile(filepath.Join(dir, "run.json"), []byte(runJSON), 0o644)

	fmt.Printf("  Results saved: %s/\n", dir)
}

// extractSection pulls text between start/end delimiters and writes to path.
func extractSection(content, startDelim, endDelim, path string) {
	start := strings.Index(content, startDelim)
	end := strings.Index(content, endDelim)
	if start >= 0 && end > start {
		section := strings.TrimSpace(content[start+len(startDelim) : end])
		os.WriteFile(path, []byte(section), 0o644)
	}
}
