package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/store"
	"github.com/8op-org/gl1tch/internal/workspace"
)

var workspaceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "print the active workspace, its resources, and recent runs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWorkspaceStatus()
	},
}

func init() { workspaceCmd.AddCommand(workspaceStatusCmd) }

func runWorkspaceStatus() error {
	resolved := resolveWorkspaceForCommand()
	if resolved.Path == "" {
		fmt.Println("no active workspace")
		return nil
	}

	wsPath := resolved.Path
	wsFile := filepath.Join(wsPath, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	ws, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}

	fmt.Printf("workspace: %s\n", ws.Name)
	fmt.Printf("path: %s\n", wsPath)

	if len(ws.Resources) > 0 {
		state, _ := workspace.LoadResourceState(wsPath)
		fmt.Println("resources:")
		for _, r := range ws.Resources {
			ref := resourceRef(r)
			line := fmt.Sprintf("  %s\t%s\t%s", r.Name, r.Type, ref)
			if t, ok := state.Entries[r.Name]; ok {
				line += fmt.Sprintf("\tfetched=%s", t.UTC().Format(time.RFC3339))
			}
			fmt.Println(line)
		}
	}

	fmt.Println("recent runs:")
	s, err := store.OpenForWorkspace(wsPath)
	if err != nil {
		fmt.Println("  (no runs)")
		return nil
	}
	defer s.Close()

	rows, err := s.DB().Query(
		`SELECT id, name, kind, exit_status, started_at FROM runs
		 WHERE workspace = ? ORDER BY id DESC LIMIT 5`, ws.Name)
	if err != nil {
		fmt.Println("  (no runs)")
		return nil
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id int64
		var name, kind string
		var exitStatus *int
		var startedAt int64
		if err := rows.Scan(&id, &name, &kind, &exitStatus, &startedAt); err != nil {
			continue
		}
		status := "running"
		if exitStatus != nil {
			if *exitStatus == 0 {
				status = "ok"
			} else {
				status = fmt.Sprintf("exit=%d", *exitStatus)
			}
		}
		ts := time.UnixMilli(startedAt).UTC().Format(time.RFC3339)
		fmt.Printf("  %d\t%s\t%s\t%s\t%s\n", id, name, kind, status, ts)
		count++
	}
	if count == 0 {
		fmt.Println("  (no runs)")
	}
	return nil
}

func resourceRef(r workspace.Resource) string {
	switch r.Type {
	case "git":
		if r.Pin != "" {
			return fmt.Sprintf("%s@%s", r.URL, r.Pin)
		}
		if r.Ref != "" {
			return fmt.Sprintf("%s@%s", r.URL, r.Ref)
		}
		return r.URL
	case "local":
		return r.Path
	case "tracker":
		return r.Repo
	}
	return ""
}
