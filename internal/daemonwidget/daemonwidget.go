// Package daemonwidget launches installed plugins that declare daemon:true in
// their sidecar YAML as background processes on gl1tch session start.
//
// Each daemon binary is started with cmd.Start() (non-blocking) immediately
// after the BUSD event bus is ready. Processes are tracked and killed on
// Stop(). If a daemon exits early it is left dead — no restart logic for MVP.
package daemonwidget

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// sidecar is the minimal subset of executor.SidecarSchema we need here.
// Duplicating it avoids an import cycle with internal/executor.
type sidecar struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args,omitempty"`
	Daemon  bool     `yaml:"daemon,omitempty"`
}

// Manager tracks running daemon processes.
type Manager struct {
	procs []*exec.Cmd
}

// StartAll scans wrappersDir for sidecar YAMLs with daemon:true and starts
// each one as a background process. Errors for individual daemons are printed
// to stderr and skipped — a single bad entry will not prevent the others from
// launching.
func StartAll(wrappersDir string) *Manager {
	m := &Manager{}

	entries, err := os.ReadDir(wrappersDir)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "glitch: daemonwidget: read wrappers dir: %v\n", err)
		}
		return m
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(wrappersDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "glitch: daemonwidget: read %s: %v\n", e.Name(), err)
			continue
		}

		var sc sidecar
		if err := yaml.Unmarshal(data, &sc); err != nil {
			fmt.Fprintf(os.Stderr, "glitch: daemonwidget: parse %s: %v\n", e.Name(), err)
			continue
		}

		if !sc.Daemon {
			continue
		}

		if sc.Command == "" {
			fmt.Fprintf(os.Stderr, "glitch: daemonwidget: %s has daemon:true but no command\n", e.Name())
			continue
		}

		cmd := exec.Command(sc.Command, sc.Args...)
		cmd.Stdout = nil
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "glitch: daemonwidget: start %q: %v\n", sc.Command, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "glitch: daemonwidget: started %s (pid %d)\n", sc.Name, cmd.Process.Pid)
		m.procs = append(m.procs, cmd)
	}

	return m
}

// Stop signals all running daemon processes to exit and reaps them.
// Errors are ignored — best-effort cleanup on session shutdown.
func (m *Manager) Stop() {
	for _, cmd := range m.procs {
		if cmd.Process != nil {
			cmd.Process.Kill() //nolint:errcheck
			cmd.Wait()         //nolint:errcheck
		}
	}
}
