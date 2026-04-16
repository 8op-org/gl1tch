package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/batch"
	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/ui"
	"github.com/8op-org/gl1tch/internal/workspace"
)

var (
	workflowParams         []string
	workflowTagFilter      string
	workflowVariants       []string
	workflowCompare        bool
	workflowReviewCriteria string
)

func init() {
	workflowRunCmd.Flags().StringVarP(&targetPath, "path", "C", "", "run against this directory instead of cwd")
	workflowRunCmd.Flags().StringArrayVar(&workflowParams, "set", nil, "set workflow param (key=value), repeatable")
	workflowRunCmd.Flags().StringArrayVar(&workflowVariants, "variant", nil, "variant provider:model for comparison (repeatable)")
	workflowRunCmd.Flags().BoolVar(&workflowCompare, "compare", false, "discover variant workflows and cross-review")
	workflowRunCmd.Flags().StringVar(&workflowReviewCriteria, "review-criteria", "", "comma-separated review criteria for comparison")
	workflowListCmd.Flags().StringVar(&workflowTagFilter, "tag", "", "filter workflows by tag")
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowRunCmd)
}

var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Aliases: []string{"wf"},
	Short:   "manage and run workflows",
}

var workflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "list available workflows",
	RunE: func(cmd *cobra.Command, args []string) error {
		workflows, err := loadWorkflows()
		if err != nil {
			return err
		}

		names := make([]string, 0, len(workflows))
		for name := range workflows {
			names = append(names, name)
		}
		sort.Strings(names)

		if workflowTagFilter != "" {
			var filtered []string
			for _, name := range names {
				w := workflows[name]
				for _, tag := range w.Tags {
					if tag == workflowTagFilter {
						filtered = append(filtered, name)
						break
					}
				}
			}
			names = filtered
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		for _, name := range names {
			w := workflows[name]
			desc := strings.TrimSpace(w.Description)
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Fprintf(tw, "%s\t%s\n", name, desc)
		}
		return tw.Flush()
	},
}

var workflowRunCmd = &cobra.Command{
	Use:   "run <name> [input]",
	Short: "run a workflow by name",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetPath != "" {
			if err := os.Chdir(targetPath); err != nil {
				return fmt.Errorf("chdir %s: %w", targetPath, err)
			}
		}

		name := args[0]
		input := ""
		if len(args) > 1 {
			input = strings.Join(args[1:], " ")
		}

		workflows, err := loadWorkflows()
		if err != nil {
			return err
		}

		w, ok := workflows[name]
		if !ok {
			return fmt.Errorf("workflow %q not found", name)
		}

		ui.WorkflowStart(w.Name)

		// Wire ES telemetry for workflow LLM calls
		var tel *esearch.Telemetry
		esClient := esearch.NewClient("http://localhost:9200")
		if err := esClient.Ping(context.Background()); err == nil {
			tel = esearch.NewTelemetry(esClient)
			tel.EnsureIndices(context.Background())
		}

		// Parse --set params
		params := make(map[string]string)
		for _, p := range workflowParams {
			parts := strings.SplitN(p, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --set value %q (expected key=value)", p)
			}
			params[parts[0]] = parts[1]
		}

		cfg, _ := loadConfig()

		// Try to load workspace config for ES URL and resolve workspace name
		var wsESURL string
		wsDir := workspacePath
		if wsDir == "" {
			wsDir, _ = os.Getwd()
		}
		wsName := workspace.ResolveWorkspace(wsDir)

		wsFile := filepath.Join(wsDir, "workspace.glitch")
		if wsData, err := os.ReadFile(wsFile); err == nil {
			if ws, err := workspace.ParseFile(wsData); err == nil {
				if ws.Defaults.Elasticsearch != "" {
					wsESURL = ws.Defaults.Elasticsearch
				}
			}
		}

		// Handle --compare: discover sibling variant workflows (before --variant injection)
		if workflowCompare {
			variants := batch.DefaultVariants
			if len(workflowVariants) > 0 {
				variants = workflowVariants
			}
			return runCompareWorkflows(name, workflows, variants, params, cfg, tel, wsESURL, wsName)
		}

		// Handle --variant: inject implicit compare blocks around LLM steps
		if len(workflowVariants) == 1 {
			fmt.Fprintf(os.Stderr, "WARN: --variant needs at least 2 variants to compare, ignoring\n")
		}
		if len(workflowVariants) > 1 {
			injectImplicitCompare(w, workflowVariants, workflowReviewCriteria)
		}

		result, err := pipeline.Run(w, input, cfg.DefaultModel, params, providerReg, pipeline.RunOpts{
			Telemetry:        tel,
			ProviderResolver: cfg.BuildProviderResolver(),
			Tiers:            cfg.Tiers,
			EvalThreshold:    cfg.EvalThreshold,
			ESURL:            wsESURL,
			Workspace:        wsName,
		})
		if err != nil {
			return err
		}
		fmt.Println(result.Output)
		return nil
	},
}

// injectImplicitCompare wraps every LLM step (not inside an existing compare) in
// an implicit compare block with one branch per variant.
func injectImplicitCompare(w *pipeline.Workflow, variants []string, reviewCriteria string) {
	var criteria []string
	if reviewCriteria != "" {
		criteria = strings.Split(reviewCriteria, ",")
		for i := range criteria {
			criteria[i] = strings.TrimSpace(criteria[i])
		}
	}

	for i, item := range w.Items {
		if item.Step != nil && item.Step.Form == "" && item.Step.LLM != nil {
			original := *item.Step
			var branches []pipeline.CompareBranch
			for _, v := range variants {
				branchStep := original
				parts := strings.SplitN(v, ":", 2)
				branchLLM := *original.LLM
				branchLLM.Provider = parts[0]
				if len(parts) == 2 {
					branchLLM.Model = parts[1]
				}
				branchStep.LLM = &branchLLM
				branches = append(branches, pipeline.CompareBranch{
					Name:  parts[0],
					Steps: []pipeline.Step{branchStep},
				})
			}
			original.Form = "compare"
			original.CompareBranches = branches
			original.LLM = nil
			if len(criteria) > 0 {
				original.CompareReview = &pipeline.ReviewConfig{Criteria: criteria}
			}
			w.Items[i].Step = &original
		}
	}

	for i, s := range w.Steps {
		if s.Form == "" && s.LLM != nil {
			original := s
			var branches []pipeline.CompareBranch
			for _, v := range variants {
				branchStep := original
				parts := strings.SplitN(v, ":", 2)
				branchLLM := *original.LLM
				branchLLM.Provider = parts[0]
				if len(parts) == 2 {
					branchLLM.Model = parts[1]
				}
				branchStep.LLM = &branchLLM
				branches = append(branches, pipeline.CompareBranch{
					Name:  parts[0],
					Steps: []pipeline.Step{branchStep},
				})
			}
			original.Form = "compare"
			original.CompareBranches = branches
			original.LLM = nil
			if len(criteria) > 0 {
				original.CompareReview = &pipeline.ReviewConfig{Criteria: criteria}
			}
			w.Steps[i] = original
		}
	}
}

// runCompareWorkflows discovers sibling variant workflows and runs them as a batch compare.
func runCompareWorkflows(name string, workflows map[string]*pipeline.Workflow, variants []string, params map[string]string, cfg *Config, tel *esearch.Telemetry, esURL, wsName string) error {
	found := make(map[string]*pipeline.Workflow)
	for _, v := range variants {
		variantName := name + "-" + v
		if w, ok := workflows[variantName]; ok {
			found[v] = w
		}
	}

	if len(found) < 2 {
		return fmt.Errorf("--compare: found %d variant workflows for %q (need at least 2). Expected: %s-<variant>", len(found), name, name)
	}

	fmt.Printf("Compare: %d variants for %s\n", len(found), name)
	for v := range found {
		fmt.Printf("  - %s-%s\n", name, v)
	}

	return batch.Run(context.Background(), batch.RunOpts{
		Items:      []string{name},
		Params:     params,
		ResultsDir: resolveResultsDir(),
		Variants:   variants,
		Iterations: 1,
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
}
