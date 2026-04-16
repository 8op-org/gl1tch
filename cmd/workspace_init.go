package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

var workspaceInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "scaffold a new workspace at path (default CWD)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) == 1 {
			path = args[0]
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		return runWorkspaceInit(abs, filepath.Base(abs))
	},
}

func init() { workspaceCmd.AddCommand(workspaceInitCmd) }

func runWorkspaceInit(path, name string) error {
	wsFile := filepath.Join(path, "workspace.glitch")
	if _, err := os.Stat(wsFile); err == nil {
		return fmt.Errorf("workspace.glitch already exists at %s", wsFile)
	}
	if err := os.MkdirAll(filepath.Join(path, "workflows"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(path, ".glitch"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(path, ".glitch", ".gitignore"),
		[]byte("*\n!.gitignore\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(wsFile, []byte(fmt.Sprintf("(workspace %q)\n", name)), 0o644); err != nil {
		return err
	}
	if err := registry.Add(registry.Entry{Name: name, Path: path}); err != nil {
		return err
	}
	fmt.Fprintf(cmdStderr(), "initialized workspace %q at %s\n", name, path)
	return nil
}
