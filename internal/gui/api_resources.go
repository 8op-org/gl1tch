package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/8op-org/gl1tch/internal/resource"
	"github.com/8op-org/gl1tch/internal/workspace"
)

type resourceEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	URL     string `json:"url,omitempty"`
	Ref     string `json:"ref,omitempty"`
	Pin     string `json:"pin,omitempty"`
	Path    string `json:"path,omitempty"`
	Repo    string `json:"repo,omitempty"`
	Fetched string `json:"fetched,omitempty"`
}

func (s *Server) handleListResources(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.resolvedWorkspacePath()
	if !ok {
		respondJSONError(w, http.StatusNotFound, "no active workspace")
		return
	}
	data, err := os.ReadFile(filepath.Join(ws, "workspace.glitch"))
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	wsp, err := workspace.ParseFile(data)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	state, _ := workspace.LoadResourceState(ws)

	out := make([]resourceEntry, 0, len(wsp.Resources))
	for _, res := range wsp.Resources {
		entry := resourceEntry{
			Name: res.Name,
			Type: res.Type,
			URL:  res.URL,
			Ref:  res.Ref,
			Pin:  res.Pin,
			Path: res.Path,
			Repo: res.Repo,
		}
		if t, ok := state.Entries[res.Name]; ok {
			entry.Fetched = t.UTC().Format(time.RFC3339)
		}
		out = append(out, entry)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type addResourceReq struct {
	Input string `json:"input"` // url, path, or org/repo
	Name  string `json:"name"`
	Pin   string `json:"pin"`
	Type  string `json:"type"` // optional; inferred otherwise
}

func (s *Server) handleAddResource(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.resolvedWorkspacePath()
	if !ok {
		respondJSONError(w, http.StatusNotFound, "no active workspace")
		return
	}
	var req addResourceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Input == "" {
		respondJSONError(w, http.StatusBadRequest, "input required")
		return
	}
	name := req.Name
	if name == "" {
		name = inferResourceNameGUI(req.Input)
	}
	if err := resource.ValidateName(name); err != nil {
		respondJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	kind := req.Type
	if kind == "" {
		kind = inferKindGUI(req.Input)
	}
	if kind == "" {
		respondJSONError(w, http.StatusBadRequest, "could not infer resource kind")
		return
	}

	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	wsp, err := workspace.ParseFile(data)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, existing := range wsp.Resources {
		if existing.Name == name {
			respondJSONError(w, http.StatusConflict, "resource "+name+" already exists")
			return
		}
	}

	res := resource.Resource{Name: name, Kind: resource.Kind(kind)}
	switch kind {
	case "git":
		res.URL = req.Input
		if req.Pin != "" {
			res.Ref = req.Pin
		} else {
			res.Ref = "main"
		}
	case "local":
		res.Path = req.Input
	case "tracker":
		res.Repo = req.Input
	}
	result, err := resource.Sync(ws, res)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	wsp.Resources = append(wsp.Resources, workspace.Resource{
		Name: name,
		Type: kind,
		URL:  res.URL,
		Ref:  res.Ref,
		Pin:  result.Pin,
		Path: res.Path,
		Repo: res.Repo,
	})
	if err := os.WriteFile(wsFile, workspace.Serialize(wsp), 0o644); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	state, _ := workspace.LoadResourceState(ws)
	if state.Entries == nil {
		state.Entries = map[string]time.Time{}
	}
	state.Entries[name] = time.Unix(result.FetchedAt, 0).UTC()
	_ = workspace.SaveResourceState(ws, state)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "name": name})
}

func (s *Server) handleRemoveResource(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.resolvedWorkspacePath()
	if !ok {
		respondJSONError(w, http.StatusNotFound, "no active workspace")
		return
	}
	name := r.PathValue("name")
	if err := resource.ValidateName(name); err != nil {
		respondJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	wsp, err := workspace.ParseFile(data)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	found := false
	out := wsp.Resources[:0]
	for _, res := range wsp.Resources {
		if res.Name == name {
			found = true
			continue
		}
		out = append(out, res)
	}
	if !found {
		respondJSONError(w, http.StatusNotFound, "resource "+name+" not found")
		return
	}
	wsp.Resources = out
	if err := os.WriteFile(wsFile, workspace.Serialize(wsp), 0o644); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = os.RemoveAll(filepath.Join(ws, "resources", name))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleSyncResources(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.resolvedWorkspacePath()
	if !ok {
		respondJSONError(w, http.StatusNotFound, "no active workspace")
		return
	}
	name := r.PathValue("name") // empty means sync all
	if err := syncWorkspaceResources(ws, name); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type pinResourceReq struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
}

func (s *Server) handlePinResource(w http.ResponseWriter, r *http.Request) {
	ws, ok := s.resolvedWorkspacePath()
	if !ok {
		respondJSONError(w, http.StatusNotFound, "no active workspace")
		return
	}
	var req pinResourceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Name == "" || req.Ref == "" {
		respondJSONError(w, http.StatusBadRequest, "name and ref required")
		return
	}
	if err := pinWorkspaceResource(ws, req.Name, req.Ref); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// --- helpers (inlined; mirror cmd/workspace_sync.go + workspace_pin.go) ---

func syncWorkspaceResources(ws, onlyName string) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	wsp, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	state, _ := workspace.LoadResourceState(ws)
	if state.Entries == nil {
		state.Entries = map[string]time.Time{}
	}
	for _, res := range wsp.Resources {
		if onlyName != "" && res.Name != onlyName {
			continue
		}
		result, err := resource.Sync(ws, resource.Resource{
			Name: res.Name,
			Kind: resource.Kind(res.Type),
			URL:  res.URL,
			Ref:  res.Ref,
			Path: res.Path,
			Repo: res.Repo,
		})
		if err != nil {
			return err
		}
		if result.Pin != "" && result.Pin != res.Pin {
			if newSrc, uerr := workspace.UpdatePin(data, res.Name, result.Pin); uerr == nil {
				data = newSrc
				_ = os.WriteFile(wsFile, data, 0o644)
			}
		}
		state.Entries[res.Name] = time.Unix(result.FetchedAt, 0).UTC()
	}
	return workspace.SaveResourceState(ws, state)
}

func pinWorkspaceResource(ws, name, ref string) error {
	wsFile := filepath.Join(ws, "workspace.glitch")
	data, err := os.ReadFile(wsFile)
	if err != nil {
		return err
	}
	// Parse-only: confirm the named resource exists. Do not re-serialize —
	// that would strip user comments.
	wsp, err := workspace.ParseFile(data)
	if err != nil {
		return err
	}
	found := false
	for _, r := range wsp.Resources {
		if r.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("resource %q not found", name)
	}
	updated, err := workspace.UpdateRef(data, name, ref)
	if err != nil {
		return err
	}
	if err := os.WriteFile(wsFile, updated, 0o644); err != nil {
		return err
	}
	return syncWorkspaceResources(ws, name)
}

// --- inference helpers (duplicated from cmd/ — package boundary) ---

func inferKindGUI(input string) string {
	switch {
	case strings.HasPrefix(input, "http://"),
		strings.HasPrefix(input, "https://"),
		strings.HasPrefix(input, "git@"),
		strings.HasSuffix(input, ".git"):
		return "git"
	case strings.HasPrefix(input, "~"),
		strings.HasPrefix(input, "/"),
		strings.HasPrefix(input, "."):
		return "local"
	case strings.Contains(input, "/") && !strings.Contains(input, ":"):
		return "tracker"
	}
	return ""
}

func inferResourceNameGUI(input string) string {
	if u, err := url.Parse(input); err == nil && u.Host != "" {
		return strings.TrimSuffix(path.Base(u.Path), ".git")
	}
	if strings.HasPrefix(input, "~") || strings.HasPrefix(input, "/") || strings.HasPrefix(input, ".") {
		return filepath.Base(input)
	}
	if i := strings.Index(input, "/"); i >= 0 {
		return input[i+1:]
	}
	return input
}

func respondJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
