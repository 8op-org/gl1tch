package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

func init() {
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
		result, err := pipeline.Run(w, input, "")
		if err != nil {
			return err
		}
		fmt.Println(result.Output)
		return nil
	},
}
