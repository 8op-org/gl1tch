package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "manage workspaces and their resources",
}

func init() { rootCmd.AddCommand(workspaceCmd) }

// activeWorkspacePath returns the resolved workspace path or an error in global mode.
func activeWorkspacePath() (string, error) {
	r := resolveWorkspaceForCommand()
	if r.Path == "" {
		return "", fmt.Errorf("no active workspace — cd into one or run `glitch workspace use <name>`")
	}
	return r.Path, nil
}

func cmdStderr() *os.File { return os.Stderr }

// activeWorkspace loads the workspace.glitch at the resolved path.
func activeWorkspace() (string, *workspace.Workspace, error) {
	p, err := activeWorkspacePath()
	if err != nil {
		return "", nil, err
	}
	data, err := os.ReadFile(p + "/workspace.glitch")
	if err != nil {
		return p, nil, err
	}
	ws, err := workspace.ParseFile(data)
	return p, ws, err
}
