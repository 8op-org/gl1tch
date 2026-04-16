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
	"github.com/8op-org/gl1tch/internal/store"
	"github.com/8op-org/gl1tch/internal/ui"
	"github.com/8op-org/gl1tch/internal/workspace"
)

var (
	runSetFlags       []string
	runPathFlag       string
	runResultsDir     string
	runVariants       []string
	runCompare        bool
	runReviewCriteria string
)

var runCmd = &cobra.Command{
	Use:   "run <workflow> [input]",
	Short: "run a workflow (resolves in active workspace, falls back to global)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRun,
}

func init() {
	runCmd.Flags().StringVar(&runPathFlag, "path", "", "workflow file path (overrides name resolution)")
	runCmd.Flags().StringArrayVar(&runSetFlags, "set", nil, "set workflow params as key=value (repeatable)")
	runCmd.Flags().StringArrayVar(&runVariants, "variant", nil, "variant provider:model for comparison (repeatable)")
	runCmd.Flags().BoolVar(&runCompare, "compare", false, "discover variant workflows and cross-review")
	runCmd.Flags().StringVar(&runReviewCriteria, "review-criteria", "", "comma-separated review criteria for comparison")
	runCmd.Flags().StringVar(&runResultsDir, "results-dir", "", "output directory for results (defaults to <workspace>/results)")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	name := args[0]
	input := ""
	if len(args) > 1 {
		input = strings.Join(args[1:], " ")
	}

	// Resolve the workflow file path using:
	//   1. --path flag (explicit override)
	//   2. <active workspace>/workflows/<name>.glitch (or .yaml/.yml)
	//   3. ~/.config/glitch/workflows/<name>.glitch (or .yaml/.yml)
	wsDir := workspacePath
	if wsDir == "" {
		wsDir, _ = os.Getwd()
	}

	workflowPath, err := resolveWorkflowPath(name)
	if err != nil {
		return err
	}

	w, err := pipeline.LoadFile(workflowPath)
	if err != nil {
		return fmt.Errorf("load workflow %s: %w", workflowPath, err)
	}

	// Also load the full map for --compare variant discovery.
	workflows, err := loadWorkflows()
	if err != nil {
		return err
	}
	// Ensure the explicitly-resolved workflow is in the map (in case --path
	// points outside the workflows dir).
	if _, ok := workflows[w.Name]; !ok {
		workflows[w.Name] = w
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
	for _, p := range runSetFlags {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set value %q (expected key=value)", p)
		}
		params[parts[0]] = parts[1]
	}

	cfg, _ := loadConfig()

	// Resolve workspace name and resources from workspace.glitch if present.
	wsName := workspace.ResolveWorkspace(wsDir)
	var wsParsed *workspace.Workspace
	var wsESURL string
	wsFile := filepath.Join(wsDir, "workspace.glitch")
	if wsData, err := os.ReadFile(wsFile); err == nil {
		if ws, err := workspace.ParseFile(wsData); err == nil {
			wsParsed = ws
			if ws.Defaults.Elasticsearch != "" {
				wsESURL = ws.Defaults.Elasticsearch
			}
		}
	}
	resources := ResourceBindings(wsParsed, wsDir)

	// Open the store (workspace-scoped when active, else global) and pre-create
	// a parent run row so call-workflow children can link back via parent_run_id.
	var s *store.Store
	if workspacePath != "" {
		s, _ = store.OpenForWorkspace(workspacePath)
	} else {
		s, _ = store.Open()
	}
	if s != nil {
		defer s.Close()
	}

	// Handle --compare: discover sibling variant workflows (before --variant injection)
	if runCompare {
		variants := batch.DefaultVariants
		if len(runVariants) > 0 {
			variants = runVariants
		}
		return runCompareWorkflows(name, workflows, variants, params, cfg, tel, wsESURL, wsName, resources, s, filepath.Dir(workflowPath))
	}

	// Handle --variant: inject implicit compare blocks around LLM steps
	if len(runVariants) == 1 {
		fmt.Fprintf(os.Stderr, "WARN: --variant needs at least 2 variants to compare, ignoring\n")
	}
	if len(runVariants) > 1 {
		injectImplicitCompare(w, runVariants, runReviewCriteria)
	}

	// Pre-create the parent run row; exit status finalized on defer.
	var parentID int64
	exitStatus := 0
	var finalOutput string
	if s != nil {
		id, err := s.RecordRun(store.RunRecord{
			Kind:         "workflow",
			Name:         name,
			Input:        input,
			WorkflowFile: workflowPath,
			WorkflowName: w.Name,
			Workspace:    wsName,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: record parent run: %v\n", err)
		} else {
			parentID = id
			defer func() {
				_ = s.FinishRun(parentID, finalOutput, exitStatus)
			}()
		}
	}

	// Child run creator for nested call-workflow invocations. Mirrors the
	// pattern used by internal/batch/batch.go so multi-level call trees get
	// correct per-level parent linkage.
	var childCreator func(parent int64, workflowName string) (int64, error)
	if s != nil {
		workspaceName := wsName
		childCreator = func(parent int64, workflowName string) (int64, error) {
			return s.RecordRun(store.RunRecord{
				Kind:         "workflow",
				Name:         workflowName,
				WorkflowName: workflowName,
				ParentRunID:  parent,
				Workspace:    workspaceName,
			})
		}
	}

	result, err := pipeline.Run(w, input, cfg.DefaultModel, params, providerReg, pipeline.RunOpts{
		Telemetry:        tel,
		ProviderResolver: cfg.BuildProviderResolver(),
		Tiers:            cfg.Tiers,
		EvalThreshold:    cfg.EvalThreshold,
		ESURL:            wsESURL,
		Workspace:        wsName,
		Resources:        resources,
		WorkflowsDir:     filepath.Dir(workflowPath),
		ParentRunID:      parentID,
		ChildRunCreator:  childCreator,
	})
	if err != nil {
		exitStatus = 1
		return err
	}
	finalOutput = result.Output
	fmt.Println(result.Output)
	return nil
}

// resolveWorkflowPath finds the absolute file path for a workflow by name,
// honoring (in order): --path flag, active workspace's workflows/ dir,
// ~/.config/glitch/workflows/. Returns an error if nothing is found.
func resolveWorkflowPath(name string) (string, error) {
	if runPathFlag != "" {
		if _, err := os.Stat(runPathFlag); err != nil {
			return "", fmt.Errorf("workflow path %q not found: %w", runPathFlag, err)
		}
		abs, err := filepath.Abs(runPathFlag)
		if err != nil {
			return runPathFlag, nil
		}
		return abs, nil
	}

	exts := []string{".glitch", ".yaml", ".yml"}

	// Active workspace first.
	if workspacePath != "" {
		wfDir := filepath.Join(workspacePath, "workflows")
		for _, ext := range exts {
			p := filepath.Join(wfDir, name+ext)
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}

	// Global fallback.
	if home, err := os.UserHomeDir(); err == nil {
		globalDir := filepath.Join(home, ".config", "glitch", "workflows")
		for _, ext := range exts {
			p := filepath.Join(globalDir, name+ext)
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}

	return "", fmt.Errorf("workflow %q not found in workspace or global workflows dir", name)
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
func runCompareWorkflows(name string, workflows map[string]*pipeline.Workflow, variants []string, params map[string]string, cfg *Config, tel *esearch.Telemetry, esURL, wsName string, resources map[string]map[string]string, s *store.Store, workflowsDir string) error {
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
		Items:        []string{name},
		Params:       params,
		ResultsDir:   resolveResultsDir(),
		Variants:     variants,
		Iterations:   1,
		Workflows:    workflows,
		WorkflowsDir: workflowsDir,
		Name:         name,
		Store:        s,
		Config: batch.BatchConfig{
			DefaultModel:     cfg.DefaultModel,
			ProviderRegistry: providerReg,
			ProviderResolver: cfg.BuildProviderResolver(),
			Telemetry:        tel,
			Tiers:            cfg.Tiers,
			EvalThreshold:    cfg.EvalThreshold,
			Workspace:        wsName,
			Resources:        resources,
		},
	})
}
