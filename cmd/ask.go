package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/batch"
	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/router"
)

var (
	askCompare    bool
	askIterations int
	askVariant    string
	askResultsDir string
)

func init() {
	askCmd.Flags().StringVarP(&targetPath, "path", "C", "", "run against this directory instead of cwd")
	askCmd.Flags().BoolVar(&askCompare, "compare", false, "run all variants and cross-review")
	askCmd.Flags().IntVarP(&askIterations, "iterations", "n", 1, "number of iterations for learning loop")
	askCmd.Flags().StringVarP(&askVariant, "variant", "v", "", "specific variant (default: use issue-to-pr workflow)")
	askCmd.Flags().StringVar(&askResultsDir, "results-dir", "", "directory for results (default: CWD/.glitch/results)")
	rootCmd.AddCommand(askCmd)
}

var askCmd = &cobra.Command{
	Use:   "ask [input]",
	Short: "route a question or issue to the best workflow",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetPath != "" {
			if err := os.Chdir(targetPath); err != nil {
				return fmt.Errorf("chdir %s: %w", targetPath, err)
			}
		}
		input := strings.Join(args, " ")

		cfg, _ := loadConfig()
		workflows, err := loadWorkflows()
		if err != nil {
			return err
		}

		// Wire ES telemetry
		var tel *esearch.Telemetry
		esClient := esearch.NewClient("http://localhost:9200")
		if err := esClient.Ping(context.Background()); err == nil {
			tel = esearch.NewTelemetry(esClient)
			tel.EnsureIndices(context.Background())
		}

		// Try multi-issue parse first
		issues, repo, isMultiIssue := router.ParseMultiIssue(input)
		if isMultiIssue {
			resolved, err := router.ResolveRepo(repo)
			if err != nil {
				return fmt.Errorf("could not resolve repo: %w", err)
			}
			repoPath := resolveRepoPath(resolved)

			if askCompare || len(issues) > 1 {
				// Batch mode
				variants := batch.DefaultVariants
				if askVariant != "" {
					variants = []string{askVariant}
				}
				iterations := askIterations
				if iterations < 1 {
					iterations = 1
				}
				if !askCompare && len(issues) > 1 {
					// Multiple issues without --compare: run sequentially with default workflow
					iterations = 1
					variants = []string{"local"}
					if askVariant != "" {
						variants = []string{askVariant}
					}
				}

				fmt.Printf("Batch: %d issues × %d variants × %d iterations\n", len(issues), len(variants), iterations)
				fmt.Printf("Repo: %s (%s)\n\n", resolved, repoPath)

				err := batch.Run(context.Background(), batch.RunOpts{
					Issues:     issues,
					Repo:       resolved,
					RepoPath:   repoPath,
					ResultsDir: askResultsDir,
					Variants:   variants,
					Iterations: iterations,
					Workflows:  workflows,
					Config: batch.BatchConfig{
						DefaultModel:     cfg.DefaultModel,
						ProviderRegistry: providerReg,
						ProviderResolver: cfg.BuildProviderResolver(),
						Telemetry:        tel,
						Tiers:            cfg.Tiers,
						EvalThreshold:    cfg.EvalThreshold,
					},
				})
				if err != nil {
					return err
				}

				// Print handoff
				rdir := askResultsDir
				if rdir == "" {
					cwd, _ := os.Getwd()
					rdir = filepath.Join(cwd, ".glitch", "results")
				}
				fmt.Printf("\nResults ready:\n")
				for _, issue := range issues {
					fmt.Printf("  #%s: %s/%s/manifest.md\n", issue, rdir, issue)
				}
				fmt.Printf("\nDashboard: http://localhost:5601/app/dashboards#/view/glitch-llm-dashboard\n")
				return nil
			}

			// Single issue — run the default workflow
			issue := issues[0]
			return runSingleIssue(issue, resolved, repoPath, workflows, cfg, tel)
		}

		// Not an issue ref — try workflow routing
		w, resolved, params := router.Match(input, workflows, cfg.DefaultModel)
		if w != nil {
			fmt.Printf(">> %s\n", w.Name)
			result, err := pipeline.Run(w, resolved, cfg.DefaultModel, params, providerReg, pipeline.RunOpts{
				Telemetry:        tel,
				ProviderResolver: cfg.BuildProviderResolver(),
				Tiers:            cfg.Tiers,
				EvalThreshold:    cfg.EvalThreshold,
			})
			if err != nil {
				return err
			}
			fmt.Println(result.Output)
			return nil
		}

		// No match — list available workflows
		fmt.Println("No matching workflow found. Available workflows:")
		fmt.Println()
		for name, w := range workflows {
			desc := w.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("  %-30s %s\n", name, desc)
		}
		fmt.Println()
		fmt.Println("Tip: glitch ask <issue-number> to work on an issue")
		return nil
	},
}

// runSingleIssue runs the issue-to-pr workflow for one issue.
func runSingleIssue(issue, repo, repoPath string, workflows map[string]*pipeline.Workflow, cfg *Config, tel *esearch.Telemetry) error {
	// Pick the workflow
	wfName := "issue-to-pr"
	if askVariant != "" {
		wfName = fmt.Sprintf("issue-to-pr-%s", askVariant)
	}
	w, ok := workflows[wfName]
	if !ok {
		return fmt.Errorf("workflow %q not found", wfName)
	}

	fmt.Printf(">> %s (#%s in %s)\n", w.Name, issue, repo)

	params := map[string]string{
		"issue":     issue,
		"repo":      repo,
		"iteration": "1",
	}
	result, err := pipeline.Run(w, "", cfg.DefaultModel, params, providerReg, pipeline.RunOpts{
		Telemetry:        tel,
		ProviderResolver: cfg.BuildProviderResolver(),
		Tiers:            cfg.Tiers,
		EvalThreshold:    cfg.EvalThreshold,
	})
	if err != nil {
		return err
	}
	fmt.Println(result.Output)

	// Print handoff
	rdir := askResultsDir
	if rdir == "" {
		cwd, _ := os.Getwd()
		rdir = filepath.Join(cwd, ".glitch", "results")
	}
	singleResultsDir := filepath.Join(rdir, issue)
	if _, err := os.Stat(singleResultsDir); err == nil {
		fmt.Printf("\nResults: %s/\n", singleResultsDir)
		fmt.Printf("\nTo create the PR:\n")
		fmt.Printf("  claude \"Create a PR for %s#%s using the plan and PR body in %s/\"\n", repo, issue, singleResultsDir)
	}

	return nil
}

// resolveRepoPath finds the local filesystem path for an org/repo.
func resolveRepoPath(repo string) string {
	parts := strings.Split(repo, "/")
	repoName := parts[len(parts)-1]
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Projects", repoName)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	// Fallback to cwd
	cwd, _ := os.Getwd()
	return cwd
}

func loadWorkflows() (map[string]*pipeline.Workflow, error) {
	workflows := make(map[string]*pipeline.Workflow)
	if home, err := os.UserHomeDir(); err == nil {
		globalDir := home + "/.config/glitch/workflows"
		if m, err := pipeline.LoadDir(globalDir); err == nil {
			for k, v := range m {
				workflows[k] = v
			}
		}
	}
	if m, err := pipeline.LoadDir(".glitch/workflows"); err == nil {
		for k, v := range m {
			workflows[k] = v
		}
	}
	return workflows, nil
}
