package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var workspaceWorkflowTagFilter string

var workspaceWorkflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "manage workflows in the active workspace",
}

var workspaceWorkflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "list workflows in the active workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWorkspaceWorkflowList()
	},
}

var workspaceWorkflowNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "scaffold a new workflow file in the active workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspaceWorkflowNew(ws, args[0])
	},
}

var workspaceWorkflowEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "open a workflow in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspaceWorkflowEdit(ws, args[0])
	},
}

func init() {
	workspaceWorkflowListCmd.Flags().StringVar(&workspaceWorkflowTagFilter, "tag", "", "filter workflows by tag")
	workspaceWorkflowCmd.AddCommand(workspaceWorkflowListCmd)
	workspaceWorkflowCmd.AddCommand(workspaceWorkflowNewCmd)
	workspaceWorkflowCmd.AddCommand(workspaceWorkflowEditCmd)
	workspaceCmd.AddCommand(workspaceWorkflowCmd)
}

// runWorkspaceWorkflowList prints workflows resolved from the active workspace.
// Body ported from the old cmd/workflow.go:workflowListCmd.RunE.
func runWorkspaceWorkflowList() error {
	workflows, err := loadWorkflows()
	if err != nil {
		return err
	}

	names := make([]string, 0, len(workflows))
	for name := range workflows {
		names = append(names, name)
	}
	sort.Strings(names)

	if workspaceWorkflowTagFilter != "" {
		var filtered []string
		for _, name := range names {
			w := workflows[name]
			for _, tag := range w.Tags {
				if tag == workspaceWorkflowTagFilter {
					filtered = append(filtered, name)
					break
				}
			}
		}
		names = filtered
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	for _, name := range names {
		w := workflows[name]
		desc := strings.TrimSpace(w.Description)
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Fprintf(tw, "%s\t%s\n", name, desc)
	}
	return tw.Flush()
}

// runWorkspaceWorkflowNew writes a minimal workflow skeleton at
// <ws>/workflows/<name>.glitch and refuses to overwrite.
func runWorkspaceWorkflowNew(ws, name string) error {
	dir := filepath.Join(ws, "workflows")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, name+".glitch")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("workflow %s already exists at %s", name, path)
	}
	tmpl := fmt.Sprintf(`(workflow %q
  (step "main"
    (run "echo todo")))
`, name)
	if err := os.WriteFile(path, []byte(tmpl), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(cmdStderr(), "created %s\n", path)
	return nil
}

// runWorkspaceWorkflowEdit opens <ws>/workflows/<name>.glitch in $EDITOR.
func runWorkspaceWorkflowEdit(ws, name string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	path := filepath.Join(ws, "workflows", name+".glitch")
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("workflow %s not found at %s", name, path)
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
