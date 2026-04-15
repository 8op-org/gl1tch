package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/provider"
)

var (
	targetPath    string
	workspacePath string
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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
