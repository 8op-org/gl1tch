package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "list registered workspaces (active prefixed with *)",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := registry.List()
		if err != nil {
			return err
		}
		active, _ := registry.GetActive()
		for _, e := range entries {
			marker := " "
			if e.Name == active {
				marker = "*"
			}
			fmt.Printf("%s %s\t%s\n", marker, e.Name, e.Path)
		}
		return nil
	},
}

func init() { workspaceCmd.AddCommand(workspaceListCmd) }
