package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace"
)

var workspacePinCmd = &cobra.Command{
	Use:   "pin <name> <ref>",
	Short: "update a resource's :ref, sync, and write the resolved :pin",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspacePin(ws, args[0], args[1])
	},
}

func init() { workspaceCmd.AddCommand(workspacePinCmd) }

func runWorkspacePin(ws, name, ref string) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	for i := range w.Resources {
		if w.Resources[i].Name == name {
			w.Resources[i].Ref = ref
			break
		}
	}
	if err := os.WriteFile(wsFile, workspace.Serialize(w), 0o644); err != nil {
		return err
	}
	return runWorkspaceSync(ws, []string{name}, false)
}
