package gui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

var paramRe = regexp.MustCompile(`\{\{\.param\.(\w+)\}\}`)

type workflowEntry struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	Description string `json:"description"`
}

func (s *Server) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	dir := s.workflowsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var workflows []workflowEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".glitch" && ext != ".yaml" && ext != ".yml" {
			continue
		}
		wf, err := pipeline.LoadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		workflows = append(workflows, workflowEntry{
			Name:        wf.Name,
			File:        e.Name(),
			Description: wf.Description,
		})
	}
	if workflows == nil {
		workflows = []workflowEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflows)
}

func (s *Server) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if strings.Contains(name, "..") {
		http.Error(w, "invalid name", 400)
		return
	}

	path := filepath.Join(s.workflowsDir(), name)
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	params := extractParams(string(data))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"name":   name,
		"source": string(data),
		"params": params,
	})
}

func (s *Server) handlePutWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if strings.Contains(name, "..") {
		http.Error(w, "invalid name", 400)
		return
	}

	var body struct {
		Source string `json:"source"`
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := json.Unmarshal(data, &body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	path := filepath.Join(s.workflowsDir(), name)
	if err := os.WriteFile(path, []byte(body.Source), 0o644); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (s *Server) handleRunWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if strings.Contains(name, "..") {
		http.Error(w, "invalid name", 400)
		return
	}

	var body struct {
		Params map[string]string `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err := pipeline.LoadFile(filepath.Join(s.workflowsDir(), name))
	if err != nil {
		http.Error(w, fmt.Sprintf("load workflow: %v", err), 400)
		return
	}

	// For now, just acknowledge the run request.
	// Full execution with pipeline.Run will be wired in Task 8.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func extractParams(source string) []string {
	var params []string
	seen := make(map[string]bool)
	for _, match := range paramRe.FindAllStringSubmatch(source, -1) {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			params = append(params, match[1])
		}
	}
	if params == nil {
		params = []string{}
	}
	return params
}
