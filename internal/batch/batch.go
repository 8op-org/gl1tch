package batch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	resultsBase := filepath.Join(opts.RepoPath, ".glitch", "results")

	// Ensure .glitch/results is gitignored
	gitignorePath := filepath.Join(opts.RepoPath, ".gitignore")
	if data, err := os.ReadFile(gitignorePath); err == nil {
		if !strings.Contains(string(data), ".glitch/results") {
			f, _ := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0o644)
			if f != nil {
				f.WriteString("\n.glitch/results/\n")
				f.Close()
			}
		}
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
				_, err := pipeline.Run(w, "", opts.Config.DefaultModel, params, opts.Config.ProviderRegistry, pipeline.RunOpts{
					Telemetry:        opts.Config.Telemetry,
					ProviderResolver: opts.Config.ProviderResolver,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: #%s (%s, iter %d) failed: %v\n", issue, variant, iter, err)
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
				_, err := pipeline.Run(crW, "", opts.Config.DefaultModel, params, opts.Config.ProviderRegistry, pipeline.RunOpts{
					Telemetry:        opts.Config.Telemetry,
					ProviderResolver: opts.Config.ProviderResolver,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARN: #%s cross-review iter %d failed: %v\n", issue, iter, err)
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
