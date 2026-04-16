package gui

import (
	"encoding/json"
	"net/http"

	"github.com/8op-org/gl1tch/internal/workspace/registry"
)

type workspaceRegistryEntry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Active bool   `json:"active"`
}

func (s *Server) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	entries, err := registry.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	active, _ := registry.GetActive()
	out := make([]workspaceRegistryEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, workspaceRegistryEntry{Name: e.Name, Path: e.Path, Active: e.Name == active})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Server) handleUseWorkspace(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if _, ok, _ := registry.Find(body.Name); !ok {
		http.Error(w, "workspace not registered", http.StatusNotFound)
		return
	}
	if err := registry.SetActive(body.Name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "active": body.Name})
}
