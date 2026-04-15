package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/workspace"
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
		if workspacePath == "" {
			return nil
		}
		if err := ensureWorkspaceDir(workspacePath); err != nil {
			return err
		}
		cfg, _ := loadConfig()
		wsFile := filepath.Join(workspacePath, "workspace.glitch")
		if data, err := os.ReadFile(wsFile); err == nil {
			if ws, err := workspace.ParseFile(data); err == nil {
				ApplyWorkspace(ws, cfg)
			}
		}
		// Store merged config for subcommands
		mergedConfig = cfg
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
