package gui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/workspace"
	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

// resolvedWorkspacePath returns the active workspace path for HTTP handlers.
// Precedence: Server.workspace (set from CLI --workspace at construction) →
// GLITCH_WORKSPACE env → walk-up from CWD → registry active name.
// Returns ok=false when no workspace can be resolved (global mode).
func (s *Server) resolvedWorkspacePath() (string, bool) {
	if s.workspace != "" {
		return s.workspace, true
	}
	cwd, _ := os.Getwd()
	active, _ := registry.GetActive()
	r := workspace.Resolve(workspace.ResolveOpts{
		EnvPath:    os.Getenv("GLITCH_WORKSPACE"),
		StartDir:   cwd,
		ActiveName: active,
	})
	if r.Path == "" {
		return "", false
	}
	return r.Path, true
}

// resourceBindings reads workspace.glitch from the server's workspace dir
// and returns a map suitable for pipeline.RunOpts.Resources. Returns an
// empty map if no workspace.glitch is present or cannot be parsed.
func (s *Server) resourceBindings() map[string]map[string]string {
	out := map[string]map[string]string{}
	if s.workspace == "" {
		return out
	}
	data, err := os.ReadFile(filepath.Join(s.workspace, "workspace.glitch"))
	if err != nil {
		return out
	}
	ws, err := workspace.ParseFile(data)
	if err != nil {
		return out
	}
	for _, r := range ws.Resources {
		m := map[string]string{"url": r.URL, "ref": r.Ref, "pin": r.Pin, "repo": r.Repo}
		switch r.Type {
		case "git":
			m["path"] = filepath.Join(s.workspace, "resources", r.Name)
			if m["repo"] == "" {
				m["repo"] = inferRepoFromURL(r.URL)
			}
		case "local":
			m["path"] = filepath.Join(s.workspace, "resources", r.Name)
		}
		out[r.Name] = m
	}
	return out
}

func inferRepoFromURL(url string) string {
	if !strings.Contains(url, "github.com") {
		return ""
	}
	after := url[strings.Index(url, "github.com")+len("github.com"):]
	after = strings.TrimPrefix(after, "/")
	after = strings.TrimPrefix(after, ":")
	after = strings.TrimSuffix(after, ".git")
	return after
}
