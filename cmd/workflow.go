package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	workflowTagFilter string
)

func init() {
	workflowListCmd.Flags().StringVar(&workflowTagFilter, "tag", "", "filter workflows by tag")
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowListCmd)
}

var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Aliases: []string{"wf"},
	Short:   "manage workflows (deprecated — use `glitch run` to invoke)",
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
