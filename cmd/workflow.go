package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/executor"
	"github.com/8op-org/gl1tch/internal/orchestrator"
	"github.com/8op-org/gl1tch/internal/picker"
	"github.com/8op-org/gl1tch/internal/store"
)

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowRunCmd)
	workflowRunCmd.Flags().StringVar(&workflowRunInput, "input", "", "input string passed to the workflow as temp.input")
	workflowCmd.AddCommand(workflowResumeCmd)
	workflowResumeCmd.Flags().Int64Var(&workflowResumeRunID, "run-id", 0, "workflow run ID to resume")
	_ = workflowResumeCmd.MarkFlagRequired("run-id")
}

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Run and manage multi-step workflows",
}

var workflowRunInput string

var workflowRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a workflow by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		configDir, err := glitchConfigDir()
		if err != nil {
			return err
		}

		// Resolve workflow file.
		workflowPath, err := orchestrator.FindWorkflow(name)
		if err != nil {
			return err
		}

		f, err := os.Open(workflowPath)
		if err != nil {
			return fmt.Errorf("workflow: open %q: %w", workflowPath, err)
		}
		defer f.Close()

		def, err := orchestrator.LoadWorkflow(f)
		if err != nil {
			return err
		}
		if err := def.Validate(); err != nil {
			return err
		}

		fmt.Printf("[workflow] starting: %s\n", def.Name)

		mgr := buildExecutorManager()

		// Load sidecar plugins.
		wrappersDir := filepath.Join(configDir, "wrappers")
		if errs := mgr.LoadWrappersFromDir(wrappersDir); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "workflow: sidecar load warning: %v\n", e)
			}
		}

		var conductorOpts []orchestrator.ConductorOption
		conductorOpts = append(conductorOpts, orchestrator.WithExecutorManager(mgr))

		// Wire store.
		if s, serr := store.Open(); serr == nil {
			defer s.Close()
			conductorOpts = append(conductorOpts, orchestrator.WithStore(s))
		} else {
			fmt.Fprintf(os.Stderr, "workflow: store unavailable: %v\n", serr)
		}

		// Wire busd publisher.
		if pub := newBusPublisher(); pub != nil {
			conductorOpts = append(conductorOpts, orchestrator.WithBusPublisher(pub))
		}

		runner := orchestrator.NewConductorRunner(configDir, conductorOpts...)

		result, err := runner.Run(cmd.Context(), def, workflowRunInput)
		if err != nil {
			return err
		}
		fmt.Printf("\n[workflow] completed: %s\n\n", def.Name)
		fmt.Println(result)
		return nil
	},
}

var workflowResumeRunID int64

var workflowResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a workflow run from its last checkpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir, err := glitchConfigDir()
		if err != nil {
			return err
		}

		s, err := store.Open()
		if err != nil {
			return fmt.Errorf("workflow resume: open store: %w", err)
		}
		defer s.Close()

		// Load the run to get the workflow name.
		wr, err := s.GetWorkflowRun(cmd.Context(), workflowResumeRunID)
		if err != nil {
			return fmt.Errorf("workflow resume: load run: %w", err)
		}

		// Resolve workflow definition.
		workflowPath, err := orchestrator.FindWorkflow(wr.Name)
		if err != nil {
			return err
		}

		f, err := os.Open(workflowPath)
		if err != nil {
			return fmt.Errorf("workflow resume: open %q: %w", workflowPath, err)
		}
		defer f.Close()

		def, err := orchestrator.LoadWorkflow(f)
		if err != nil {
			return err
		}

		mgr := buildExecutorManager()
		wrappersDir := filepath.Join(configDir, "wrappers")
		if errs := mgr.LoadWrappersFromDir(wrappersDir); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "workflow: sidecar load warning: %v\n", e)
			}
		}

		var conductorOpts []orchestrator.ConductorOption
		conductorOpts = append(conductorOpts, orchestrator.WithExecutorManager(mgr))
		conductorOpts = append(conductorOpts, orchestrator.WithStore(s))

		if pub := newBusPublisher(); pub != nil {
			conductorOpts = append(conductorOpts, orchestrator.WithBusPublisher(pub))
		}

		runner := orchestrator.NewConductorRunner(configDir, conductorOpts...)

		fmt.Printf("[workflow] resuming run %d (%s)\n", workflowResumeRunID, wr.Name)
		result, err := runner.Resume(cmd.Context(), def, workflowResumeRunID)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

// buildExecutorManager constructs an executor manager using picker.BuildProviders,
// mirroring the pattern in cmd/pipeline.go.
func buildExecutorManager() *executor.Manager {
	runProviders := picker.BuildProviders()
	mgr := executor.NewManager()
	for _, prov := range runProviders {
		if prov.SidecarPath != "" {
			continue
		}
		binary := prov.Command
		if binary == "" {
			binary = prov.ID
		}
		if err := mgr.Register(executor.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...)); err != nil {
			fmt.Fprintf(os.Stderr, "workflow: register provider %q: %v\n", prov.ID, err)
		}
	}
	return mgr
}
