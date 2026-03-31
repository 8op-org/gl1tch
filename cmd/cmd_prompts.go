package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/adam-stokes/orcai/internal/picker"
	"github.com/adam-stokes/orcai/internal/plugin"
	"github.com/adam-stokes/orcai/internal/promptbuilder"
	"github.com/adam-stokes/orcai/internal/promptmgr"
	"github.com/adam-stokes/orcai/internal/store"
	"github.com/adam-stokes/orcai/internal/styles"
	"github.com/adam-stokes/orcai/internal/themes"
)

func init() {
	rootCmd.AddCommand(promptsCmd)
	promptsCmd.AddCommand(promptsTuiCmd)
	promptsCmd.AddCommand(promptsStartCmd)

	rootCmd.AddCommand(promptBuilderCmd)
	promptBuilderCmd.AddCommand(promptBuilderStartCmd)
}

var promptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Manage AI prompts",
}

var promptsTuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive prompt manager TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "prompts tui: panic: %v\n", r)
				os.Exit(2)
			}
		}()

		var bundle *themes.Bundle
		home, _ := os.UserHomeDir()
		userThemesDir := filepath.Join(home, ".config", "orcai", "themes")
		if reg, err := themes.NewRegistry(userThemesDir); err == nil {
			bundle = reg.Active()
			themes.SetGlobalRegistry(reg)
		}

		st, err := store.Open()
		if err != nil {
			return fmt.Errorf("prompts tui: open store: %w", err)
		}
		defer st.Close()

		pluginMgr := plugin.NewManager()
		home2, _ := os.UserHomeDir()
		pluginDir := filepath.Join(home2, ".config", "orcai", "plugins")
		pluginMgr.LoadWrappersFromDir(pluginDir)
		// Register provider CLI adapters so the test runner can execute them.
		for _, prov := range picker.BuildProviders() {
			binary := prov.Command
			if binary == "" {
				binary = prov.ID
			}
			_ = pluginMgr.Register(plugin.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...))
		}

		m := promptmgr.New(st, pluginMgr, bundle)
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

var promptsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Open the prompt manager in a tmux window",
	RunE: func(cmd *cobra.Command, args []string) error {
		self, err := os.Executable()
		if err != nil {
			return fmt.Errorf("prompts: resolve executable: %w", err)
		}
		self = filepath.Clean(self)
		newArgs := []string{
			"new-window", "-n", "orcai-prompts",
			self + " prompts tui",
		}
		if err := exec.Command("tmux", newArgs...).Run(); err != nil {
			return fmt.Errorf("prompts: open window: %w", err)
		}
		return nil
	},
}

// promptBuilderCmd opens the two-column interactive prompt builder TUI.
var promptBuilderCmd = &cobra.Command{
	Use:   "prompt-builder",
	Short: "Open the interactive prompt builder (two-column layout)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If inside tmux and not already in the prompt-builder window, open a new window.
		if os.Getenv("TMUX") != "" {
			out, _ := exec.Command("tmux", "display-message", "-p", "#W").Output()
			if strings.TrimSpace(string(out)) != "orcai-prompt-builder" {
				self, err := os.Executable()
				if err != nil {
					return fmt.Errorf("prompt-builder: resolve executable: %w", err)
				}
				return exec.Command("tmux", "new-window", "-n", "orcai-prompt-builder",
					filepath.Clean(self)+" prompt-builder").Run()
			}
		}

		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "prompt-builder: panic: %v\n", r)
				os.Exit(2)
			}
		}()

		home, _ := os.UserHomeDir()
		userThemesDir := filepath.Join(home, ".config", "orcai", "themes")

		var pal styles.ANSIPalette
		if reg, err := themes.NewRegistry(userThemesDir); err == nil {
			themes.SetGlobalRegistry(reg)
			if bundle := reg.Active(); bundle != nil {
				pal = styles.BundleANSI(bundle)
			}
		}

		promptsDir := filepath.Join(home, ".config", "orcai", "prompt-builder")
		if err := os.MkdirAll(promptsDir, 0o755); err != nil {
			return fmt.Errorf("prompt-builder: create prompts dir: %w", err)
		}

		providers := picker.BuildProviders()
		pluginMgr := plugin.NewManager()
		pluginDir := filepath.Join(home, ".config", "orcai", "plugins")
		pluginMgr.LoadWrappersFromDir(pluginDir)
		for _, prov := range providers {
			binary := prov.Command
			if binary == "" {
				binary = prov.ID
			}
			_ = pluginMgr.Register(plugin.NewCliAdapter(prov.ID, prov.Label+" CLI adapter", binary, prov.PipelineArgs...))
		}

		m := promptbuilder.NewTwoColumn(promptsDir, providers, pluginMgr)
		if pal.Accent != "" {
			m.SetPalette(pal)
		}
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}


// promptBuilderStartCmd opens prompt-builder in a new tmux window.
var promptBuilderStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Open prompt-builder in a new tmux window",
	RunE: func(cmd *cobra.Command, args []string) error {
		self, err := os.Executable()
		if err != nil {
			return fmt.Errorf("prompt-builder: resolve executable: %w", err)
		}
		self = filepath.Clean(self)
		newArgs := []string{
			"new-window", "-n", "orcai-prompt-builder",
			self + " prompt-builder",
		}
		if err := exec.Command("tmux", newArgs...).Run(); err != nil {
			return fmt.Errorf("prompt-builder: open window: %w", err)
		}
		return nil
	},
}
