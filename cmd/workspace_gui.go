package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/gui"
)

var (
	workspaceGUIPort int
	workspaceGUIDev  bool
)

var workspaceGUICmd = &cobra.Command{
	Use:   "gui",
	Short: "launch the workspace GUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWorkspaceGUI()
	},
}

func init() {
	workspaceGUICmd.Flags().IntVar(&workspaceGUIPort, "port", 8374, "port to listen on")
	workspaceGUICmd.Flags().BoolVar(&workspaceGUIDev, "dev", false, "dev mode (proxy to Vite on :5173)")
	workspaceCmd.AddCommand(workspaceGUICmd)
}

func runWorkspaceGUI() error {
	ws := workspacePath
	if ws == "" {
		var err error
		ws, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	ws, _ = filepath.Abs(ws)

	addr := fmt.Sprintf("127.0.0.1:%d", workspaceGUIPort)
	srv, err := gui.New(addr, ws, workspaceGUIDev)
	if err != nil {
		return err
	}
	defer srv.Close()

	fmt.Printf(">> gl1tch gui: http://%s\n", addr)
	if workspaceGUIDev {
		fmt.Println(">> dev mode: frontend at http://127.0.0.1:5173")
	}
	return srv.ListenAndServe()
}
