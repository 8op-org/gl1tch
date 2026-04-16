package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/8op-org/gl1tch/internal/workspace"
)

// loadResourceBindings reads workspace.glitch from wsDir (if present) and
// returns the resource bindings map for RunOpts.Resources. Returns an empty
// map when the workspace file is missing or cannot be parsed (global mode).
func loadResourceBindings(wsDir string) map[string]map[string]string {
	if wsDir == "" {
		return map[string]map[string]string{}
	}
	data, err := os.ReadFile(filepath.Join(wsDir, "workspace.glitch"))
	if err != nil {
		return map[string]map[string]string{}
	}
	ws, err := workspace.ParseFile(data)
	if err != nil {
		return map[string]map[string]string{}
	}
	return ResourceBindings(ws, wsDir)
}

// ResourceBindings returns the map used to populate RunOpts.Resources.
// Empty map if ws is nil (global mode).
func ResourceBindings(ws *workspace.Workspace, wsPath string) map[string]map[string]string {
	out := map[string]map[string]string{}
	if ws == nil {
		return out
	}
	for _, r := range ws.Resources {
		m := map[string]string{"url": r.URL, "ref": r.Ref, "pin": r.Pin, "repo": r.Repo}
		switch r.Type {
		case "git":
			m["path"] = filepath.Join(wsPath, "resources", r.Name)
			if m["repo"] == "" {
				m["repo"] = inferRepoFromURL(r.URL)
			}
		case "local":
			m["path"] = filepath.Join(wsPath, "resources", r.Name)
		case "tracker":
			// no path — tracker alias only
		}
		out[r.Name] = m
	}
	return out
}

// inferRepoFromURL pulls "org/name" out of a GitHub clone URL.
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
