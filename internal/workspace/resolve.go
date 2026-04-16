package workspace

import (
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

// Resolved is the fully-resolved workspace (name + path) or empty if global mode.
type Resolved struct {
	Name string
	Path string // absolute directory containing workspace.glitch, or empty for global mode
}

// ResolveOpts controls the precedence chain. First non-empty wins:
// ExplicitPath → EnvPath → walk-up from StartDir → ActiveName (looked up via registry).
type ResolveOpts struct {
	ExplicitPath string
	EnvPath      string
	StartDir     string
	ActiveName   string // usually populated from registry.GetActive() by the caller
}

// Resolve returns the effective workspace or an empty Resolved when in global mode.
func Resolve(opts ResolveOpts) Resolved {
	for _, p := range []string{opts.ExplicitPath, opts.EnvPath} {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if r, ok := loadAt(abs); ok {
			return r
		}
	}
	if opts.StartDir != "" {
		if r, ok := walkUp(opts.StartDir); ok {
			return r
		}
	}
	if opts.ActiveName != "" {
		if e, ok, _ := registry.Find(opts.ActiveName); ok {
			abs, err := filepath.Abs(expandHome(e.Path))
			if err == nil {
				if r, ok := loadAt(abs); ok {
					return r
				}
			}
		}
	}
	return Resolved{}
}

// ResolveWorkspace (legacy) returns just the name for callers that haven't migrated.
func ResolveWorkspace(startDir string) string {
	r := Resolve(ResolveOpts{StartDir: startDir})
	if r.Name != "" {
		return r.Name
	}
	return filepath.Base(startDir)
}

func loadAt(dir string) (Resolved, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "workspace.glitch"))
	if err != nil {
		return Resolved{}, false
	}
	ws, err := ParseFile(data)
	if err != nil || ws.Name == "" {
		return Resolved{}, false
	}
	return Resolved{Name: ws.Name, Path: dir}, true
}

func walkUp(start string) (Resolved, bool) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return Resolved{}, false
	}
	dir := abs
	for {
		if r, ok := loadAt(dir); ok {
			return r, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Resolved{}, false
		}
		dir = parent
	}
}

func expandHome(p string) string {
	if len(p) > 1 && p[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
