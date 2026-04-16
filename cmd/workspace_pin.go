package cmd

import (
	"fmt"
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
	// Parse-only: confirm the named resource exists in the workspace.
	// Do not re-serialize — that would strip user comments.
	wsp, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	found := false
	for _, r := range wsp.Resources {
		if r.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("resource %q not found", name)
	}
	updated, err := workspace.UpdateRef(data, name, ref)
	if err != nil {
		return err
	}
	if err := os.WriteFile(wsFile, updated, 0o644); err != nil {
		return err
	}
	return runWorkspaceSync(ws, []string{name}, false)
}
