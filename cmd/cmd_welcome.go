package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/adam-stokes/orcai/internal/welcome"
)

func init() {
	rootCmd.AddCommand(welcomeCmd)
}

var welcomeCmd = &cobra.Command{
	Use:   "welcome",
	Short: "First-run onboarding guide (GLITCH)",
	Long:  "Opens the GLITCH onboarding TUI for first-time ORCAI users.",
	RunE:  runWelcomeTUI,
}

func runWelcomeTUI(cmd *cobra.Command, args []string) error {
	// If inside tmux and not already in the welcome window, open a new window.
	if os.Getenv("TMUX") != "" {
		out, _ := exec.Command("tmux", "display-message", "-p", "#W").Output()
		if strings.TrimSpace(string(out)) != "orcai-welcome" {
			self, err := os.Executable()
			if err != nil {
				return fmt.Errorf("welcome: resolve executable: %w", err)
			}
			return exec.Command("tmux", "new-window", "-n", "orcai-welcome",
				filepath.Clean(self)+" welcome").Run()
		}
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "welcome: panic: %v\n", r)
			os.Exit(2)
		}
	}()

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("welcome: home dir: %w", err)
	}
	cfgDir := filepath.Join(home, ".config", "orcai")

	m := welcome.New(cfgDir)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
