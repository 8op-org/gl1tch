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
	Model    string `json:"model,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func (s *Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.workspace, "workspace.glitch")
	data, err := os.ReadFile(path)
	if err != nil {
		// No workspace.glitch — return empty response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaceResponse{Repos: []string{}})
		return
	}

	ws, err := workspace.ParseFile(data)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaceResponse{
		Name:        ws.Name,
		Description: ws.Description,
		Owner:       ws.Owner,
		Repos:       ws.Repos,
		Defaults: workspaceDefaults{
			Model:    ws.Defaults.Model,
			Provider: ws.Defaults.Provider,
		},
	})
}
