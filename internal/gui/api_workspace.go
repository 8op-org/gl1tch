package gui

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/8op-org/gl1tch/internal/workspace"
)

type workspaceResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Owner       string            `json:"owner"`
	Repos       []string          `json:"repos"`
	Defaults    workspaceDefaults `json:"defaults"`
}

type workspaceDefaults struct {
	Model         string            `json:"model,omitempty"`
	Provider      string            `json:"provider,omitempty"`
	Elasticsearch string            `json:"elasticsearch,omitempty"`
	Params        map[string]string `json:"params"`
}

func (s *Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.workspace, "workspace.glitch")
	data, err := os.ReadFile(path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaceResponse{Repos: []string{}, Defaults: workspaceDefaults{Params: map[string]string{}}})
		return
	}

	ws, err := workspace.ParseFile(data)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	params := ws.Defaults.Params
	if params == nil {
		params = map[string]string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaceResponse{
		Name:        ws.Name,
		Description: ws.Description,
		Owner:       ws.Owner,
		Repos:       ws.Repos,
		Defaults: workspaceDefaults{
			Model:         ws.Defaults.Model,
			Provider:      ws.Defaults.Provider,
			Elasticsearch: ws.Defaults.Elasticsearch,
			Params:        params,
		},
	})
}

func (s *Server) handlePutWorkspace(w http.ResponseWriter, r *http.Request) {
	var req workspaceResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), 400)
		return
	}
	if req.Name == "" {
		http.Error(w, "workspace name is required", 400)
		return
	}

	// Read existing workspace to preserve resources and other fields
	wsPath := filepath.Join(s.workspace, "workspace.glitch")
	existing := &workspace.Workspace{}
	if data, err := os.ReadFile(wsPath); err == nil {
		if parsed, err := workspace.ParseFile(data); err == nil {
			existing = parsed
		}
	}

	repos := req.Repos
	if repos == nil {
		repos = []string{}
	}
	params := req.Defaults.Params
	if params == nil {
		params = map[string]string{}
	}

	// Update only the fields that settings manages, preserve everything else
	existing.Name = req.Name
	existing.Description = req.Description
	existing.Owner = req.Owner
	existing.Repos = repos
	existing.Defaults = workspace.Defaults{
		Model:         req.Defaults.Model,
		Provider:      req.Defaults.Provider,
		Elasticsearch: req.Defaults.Elasticsearch,
		Params:        params,
	}

	data := workspace.Serialize(existing)
	if err := os.WriteFile(wsPath, data, 0o644); err != nil {
		http.Error(w, "write workspace: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetProviders(w http.ResponseWriter, r *http.Request) {
	names := []string{}
	if s.providerReg != nil {
		names = s.providerReg.Names()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}
