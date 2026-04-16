package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/resource"
	"github.com/8op-org/gl1tch/internal/workspace"
)

var syncForce bool

var workspaceSyncCmd = &cobra.Command{
	Use:   "sync [name...]",
	Short: "update resources to their declared refs (all if no names given)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := activeWorkspacePath()
		if err != nil {
			return err
		}
		return runWorkspaceSync(ws, args, syncForce)
	},
}

func init() {
	workspaceSyncCmd.Flags().BoolVar(&syncForce, "force", false, "re-clone git resources from scratch")
	workspaceCmd.AddCommand(workspaceSyncCmd)
}

func runWorkspaceSync(ws string, names []string, force bool) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	w, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}

	st, _ := workspace.LoadResourceState(ws)
	if st.Entries == nil {
		st.Entries = map[string]time.Time{}
	}

	for _, r := range w.Resources {
		if len(want) > 0 && !want[r.Name] {
			continue
		}
		res, err := resource.Sync(ws, resource.Resource{
			Name: r.Name, Kind: resource.Kind(r.Type),
			URL: r.URL, Ref: r.Ref, Path: r.Path, Repo: r.Repo,
		}, resource.SyncOpts{Force: force})
		if err != nil {
			fmt.Fprintf(os.Stderr, "sync %s: %v\n", r.Name, err)
			continue
		}
		if res.Pin != "" && res.Pin != r.Pin {
			if newSrc, err := workspace.UpdatePin(data, r.Name, res.Pin); err == nil {
				data = newSrc
				_ = os.WriteFile(wsFile, data, 0o644)
			}
		}
		st.Entries[r.Name] = time.Unix(res.FetchedAt, 0).UTC()
	}
	return workspace.SaveResourceState(ws, st)
}
