package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/workspace"
	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

var (
	targetPath    string
	workspacePath string
	mergedConfig  *Config // set by PersistentPreRunE when --workspace is active
)

var providerReg *provider.ProviderRegistry

func init() {
	rootCmd.PersistentFlags().StringVar(&workspacePath, "workspace", "", "workspace directory for workflows and results")

	if home, err := os.UserHomeDir(); err == nil {
		providerReg, _ = provider.LoadProviders(filepath.Join(home, ".config", "glitch", "providers"))
	}
	if providerReg == nil {
		providerReg, _ = provider.LoadProviders("")
	}
}

var rootCmd = &cobra.Command{
	Use:   "glitch",
	Short: "your GitHub co-pilot",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		resolved := resolveWorkspaceForCommand()
		workspacePath = resolved.Path
		if resolved.Path == "" {
			return nil
		}
		if err := ensureWorkspaceDir(resolved.Path); err != nil {
			return err
		}
		cfg, _ := loadConfig()
		wsFile := filepath.Join(resolved.Path, "workspace.glitch")
		if data, err := os.ReadFile(wsFile); err == nil {
			if ws, err := workspace.ParseFile(data); err == nil {
				ApplyWorkspace(ws, cfg)
			}
		}
		mergedConfig = cfg
		return nil
	},
}

func resolveWorkspaceForCommand() workspace.Resolved {
	cwd, _ := os.Getwd()
	active, _ := registry.GetActive()
	return workspace.Resolve(workspace.ResolveOpts{
		ExplicitPath: workspacePath,
		EnvPath:      os.Getenv("GLITCH_WORKSPACE"),
		StartDir:     cwd,
		ActiveName:   active,
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
