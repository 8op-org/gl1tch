package workspace

import (
	"os"
	"path/filepath"
)

// ResolveWorkspace walks up from startDir looking for workspace.glitch.
// If found, parses and returns the workspace name.
// Falls back to: directory containing .glitch/ → basename of startDir.
func ResolveWorkspace(startDir string) string {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return filepath.Base(startDir)
	}

	// Walk up looking for workspace.glitch
	dir := absDir
	for {
		wsFile := filepath.Join(dir, "workspace.glitch")
		if data, err := os.ReadFile(wsFile); err == nil {
			if ws, err := ParseFile(data); err == nil && ws.Name != "" {
				return ws.Name
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	// Walk up looking for .glitch/ directory
	dir = absDir
	for {
		dotGlitch := filepath.Join(dir, ".glitch")
		if info, err := os.Stat(dotGlitch); err == nil && info.IsDir() {
			return filepath.Base(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return filepath.Base(absDir)
}
