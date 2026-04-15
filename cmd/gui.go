package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/8op-org/gl1tch/internal/gui"
)

var (
	guiPort int
	guiDev  bool
)

func init() {
	guiCmd.Flags().IntVar(&guiPort, "port", 8374, "port to listen on")
	guiCmd.Flags().BoolVar(&guiDev, "dev", false, "dev mode (proxy to Vite on :5173)")
	workflowCmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "start the workflow management web UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws := workspacePath
		if ws == "" {
			var err error
			ws, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		ws, _ = filepath.Abs(ws)

		addr := fmt.Sprintf("127.0.0.1:%d", guiPort)
		srv, err := gui.New(addr, ws, guiDev)
		if err != nil {
			return err
		}
		defer srv.Close()

		fmt.Printf(">> gl1tch gui: http://%s\n", addr)
		if guiDev {
			fmt.Println(">> dev mode: frontend at http://127.0.0.1:5173")
		}
		return srv.ListenAndServe()
	},
}
