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
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/store"
	"gopkg.in/yaml.v3"
)

var paramRe = regexp.MustCompile(`\{\{\.param\.(\w+)\}\}`)

type workflowEntry struct {
	Name        string   `json:"name"`
	File        string   `json:"file"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Author      string   `json:"author,omitempty"`
	Version     string   `json:"version,omitempty"`
	Created     string   `json:"created,omitempty"`
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
			Tags:        wf.Tags,
			Author:      wf.Author,
			Version:     wf.Version,
			Created:     wf.Created,
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

	wf, err := pipeline.LoadFile(filepath.Join(s.workflowsDir(), name))
	if err != nil {
		http.Error(w, fmt.Sprintf("load workflow: %v", err), 400)
		return
	}

	// Record the run in the store so the GUI can track it
	var runID int64
	if s.store != nil {
		runID, _ = s.store.RecordRun(store.RunRecord{
			Kind:         "workflow",
			Name:         wf.Name,
			WorkflowFile: name,
		})
	}

	// Load config for model/provider/tiers
	cfg := loadGUIConfig()

	// Run in background goroutine
	go func() {
		tel := newTelemetry()
		result, err := pipeline.Run(wf, "", cfg.DefaultModel, body.Params, s.providerReg, pipeline.RunOpts{
			Telemetry:        tel,
			ProviderResolver: cfg.ProviderResolver,
			Tiers:            cfg.Tiers,
			EvalThreshold:    cfg.EvalThreshold,
		})
		if s.store != nil {
			exitStatus := 0
			output := ""
			if err != nil {
				exitStatus = 1
				output = err.Error()
			} else if result != nil {
				output = result.Output
			}
			s.store.FinishRun(runID, output, exitStatus)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "started", "run_id": runID})
}

// guiConfig holds the minimal config needed to run workflows from the GUI.
type guiConfig struct {
	DefaultModel     string
	DefaultProvider  string
	ProviderResolver provider.ResolverFunc
	Tiers            []provider.TierConfig
	EvalThreshold    int
}

// loadGUIConfig reads ~/.config/glitch/config.yaml for the GUI.
func loadGUIConfig() guiConfig {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "glitch", "config.yaml")

	cfg := guiConfig{
		DefaultModel:    "qwen3:8b",
		DefaultProvider: "ollama",
		Tiers:           provider.DefaultTiers(),
		EvalThreshold:   4,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var raw struct {
		DefaultModel  string                `yaml:"default_model"`
		EvalThreshold int                   `yaml:"eval_threshold"`
		Tiers         []provider.TierConfig `yaml:"tiers"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return cfg
	}
	if raw.DefaultModel != "" {
		cfg.DefaultModel = raw.DefaultModel
	}
	if len(raw.Tiers) > 0 {
		cfg.Tiers = raw.Tiers
	}
	if raw.EvalThreshold > 0 {
		cfg.EvalThreshold = raw.EvalThreshold
	}
	return cfg
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
