package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/esearch"
	"github.com/8op-org/gl1tch/internal/pipeline"
)

var workflowParams []string

func init() {
	workflowRunCmd.Flags().StringVarP(&targetPath, "path", "C", "", "run against this directory instead of cwd")
	workflowRunCmd.Flags().StringArrayVar(&workflowParams, "set", nil, "set workflow param (key=value), repeatable")
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

		fmt.Printf(">> %s\n", w.Name)

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
		result, err := pipeline.Run(w, input, cfg.DefaultModel, params, providerReg, pipeline.RunOpts{
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
	},
}
