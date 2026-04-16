package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace"
	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

var registerAs string

var workspaceRegisterCmd = &cobra.Command{
	Use:   "register <path>",
	Short: "add an existing workspace directory to the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWorkspaceRegister(args[0], registerAs)
	},
}

var workspaceUnregisterCmd = &cobra.Command{
	Use:   "unregister <name>",
	Short: "remove a workspace from the registry (files untouched)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWorkspaceUnregister(args[0])
	},
}

func init() {
	workspaceRegisterCmd.Flags().StringVar(&registerAs, "as", "", "override the registered name")
	workspaceCmd.AddCommand(workspaceRegisterCmd)
	workspaceCmd.AddCommand(workspaceUnregisterCmd)
}

func runWorkspaceRegister(path, asName string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(abs, "workspace.glitch"))
	if err != nil {
		return fmt.Errorf("no workspace.glitch at %s: %w", abs, err)
	}
	ws, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	name := asName
	if name == "" {
		name = ws.Name
	}
	if name == "" {
		return fmt.Errorf("workspace at %s has no name; pass --as", abs)
	}
	return registry.Add(registry.Entry{Name: name, Path: abs})
}

func runWorkspaceUnregister(name string) error {
	return registry.Remove(name)
}
