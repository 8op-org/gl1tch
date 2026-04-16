package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

var workspaceUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "set the active workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, ok, err := registry.Find(args[0]); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("workspace %q not registered", args[0])
		}
		if err := registry.SetActive(args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmdStderr(), "active workspace: %s\n", args[0])
		return nil
	},
}

func init() { workspaceCmd.AddCommand(workspaceUseCmd) }
