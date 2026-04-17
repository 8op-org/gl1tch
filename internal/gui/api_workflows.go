package gui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/provider"
	"github.com/8op-org/gl1tch/internal/sexpr"
	"github.com/8op-org/gl1tch/internal/store"
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
	Actions     []string `json:"actions,omitempty"`
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
			Actions:     wf.Actions,
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

	resp := map[string]any{
		"name":   name,
		"source": string(data),
		"params": params,
	}

	// Parse workflow to extract metadata
	wf, err := pipeline.LoadFile(path)
	if err == nil {
		if wf.Name != "" {
			resp["name"] = wf.Name
		}
		resp["description"] = wf.Description
		resp["tags"] = wf.Tags
		resp["author"] = wf.Author
		resp["version"] = wf.Version
		resp["created"] = wf.Created
		resp["actions"] = wf.Actions
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

	// Build resource bindings once before the goroutine to avoid races on
	// concurrent workspace.glitch writes.
	resources := s.resourceBindings()
	workflowsDir := s.workflowsDir()

	// Run in background goroutine
	go func() {
		tel := newTelemetry()
		var stepRecorder func(pipeline.StepRecord)
		if s.store != nil && runID != 0 {
			stepRecorder = func(rec pipeline.StepRecord) {
				_ = s.store.RecordStep(store.StepRecord{
					RunID:      runID,
					StepID:     rec.StepID,
					Prompt:     rec.Prompt,
					Output:     rec.Output,
					Model:      rec.Model,
					DurationMs: rec.DurationMs,
					Kind:       rec.Kind,
					ExitStatus: rec.ExitStatus,
					TokensIn:   rec.TokensIn,
					TokensOut:  rec.TokensOut,
				})
			}
		}
		result, err := pipeline.Run(wf, "", cfg.DefaultModel, body.Params, s.providerReg, pipeline.RunOpts{
			Telemetry:        tel,
			ProviderResolver: cfg.ProviderResolver,
			Tiers:            cfg.Tiers,
			EvalThreshold:    cfg.EvalThreshold,
			Resources:        resources,
			WorkflowsDir:     workflowsDir,
			StepRecorder:     stepRecorder,
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

// loadGUIConfig reads ~/.config/glitch/config.glitch for the GUI.
// Inlined rather than reusing cmd.LoadConfigForGUI because cmd imports
// internal/gui — creating an import cycle.
func loadGUIConfig() guiConfig {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "glitch", "config.glitch")

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
	nodes, err := sexpr.Parse(data)
	if err != nil {
		return cfg
	}
	for _, n := range nodes {
		if !n.IsList() || len(n.Children) == 0 || n.Children[0].SymbolVal() != "config" {
			continue
		}
		kids := n.Children[1:]
		i := 0
		for i < len(kids) {
			c := kids[i]
			if c.IsAtom() && c.Atom != nil && c.Atom.Type == sexpr.TokenKeyword {
				if i+1 >= len(kids) {
					break
				}
				val := kids[i+1]
				switch c.KeywordVal() {
				case "default-model":
					if v := val.StringVal(); v != "" {
						cfg.DefaultModel = v
					}
				case "default-provider":
					if v := val.StringVal(); v != "" {
						cfg.DefaultProvider = v
					}
				case "eval-threshold":
					if val.IsAtom() && val.Atom != nil {
						if nv, err := strconv.Atoi(val.Atom.Val); err == nil && nv > 0 {
							cfg.EvalThreshold = nv
						}
					}
				}
				i += 2
				continue
			}
			if c.IsList() && len(c.Children) > 0 && c.Children[0].SymbolVal() == "tiers" {
				var tiers []provider.TierConfig
				for _, entry := range c.Children[1:] {
					if !entry.IsList() || len(entry.Children) == 0 || entry.Children[0].SymbolVal() != "tier" {
						continue
					}
					t := provider.TierConfig{}
					rest := entry.Children[1:]
					for j := 0; j < len(rest); j++ {
						k := rest[j]
						if !k.IsAtom() || k.Atom == nil || k.Atom.Type != sexpr.TokenKeyword {
							continue
						}
						if j+1 >= len(rest) {
							break
						}
						v := rest[j+1]
						switch k.KeywordVal() {
						case "providers":
							if v.IsList() {
								names := make([]string, 0, len(v.Children))
								for _, p := range v.Children {
									names = append(names, p.StringVal())
								}
								t.Providers = names
							}
						case "model":
							t.Model = v.StringVal()
						}
						j++
					}
					tiers = append(tiers, t)
				}
				if len(tiers) > 0 {
					cfg.Tiers = tiers
				}
			}
			i++
		}
	}
	return cfg
}

func (s *Server) handleGetWorkflowActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.PathValue("context")
	prefix := ctx + ":"

	dir := s.workflowsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var matches []workflowEntry
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
		for _, action := range wf.Actions {
			if action == ctx || strings.HasPrefix(action, prefix) {
				matches = append(matches, workflowEntry{
					Name:        wf.Name,
					File:        e.Name(),
					Description: wf.Description,
					Tags:        wf.Tags,
					Author:      wf.Author,
					Version:     wf.Version,
					Actions:     wf.Actions,
				})
				break
			}
		}
	}
	if matches == nil {
		matches = []workflowEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
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
