package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace"
)

var workspaceRmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "remove a workspace resource and its files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspaceRm(ws, args[0])
	},
}

func init() { workspaceCmd.AddCommand(workspaceRmCmd) }

func runWorkspaceRm(ws, name string) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	found := false
	out := w.Resources[:0]
	for _, r := range w.Resources {
		if r.Name == name {
			found = true
			continue
		}
		out = append(out, r)
	}
	if !found {
		return fmt.Errorf("resource %q not found", name)
	}
	w.Resources = out
	if err := os.WriteFile(wsFile, workspace.Serialize(w), 0o644); err != nil {
		return err
	}
	_ = os.RemoveAll(filepath.Join(ws, "resources", name))
	return nil
}
