package gui

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleGetResult(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if strings.Contains(path, "..") {
		http.Error(w, "invalid path", 400)
		return
	}

	fullPath := filepath.Join(s.resultsDir(), path)
	info, err := os.Stat(fullPath)
	if err != nil {
		// Fallback: try path as-is (cwd-relative, e.g. "results/demo/file.md")
		fullPath = path
		info, err = os.Stat(fullPath)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
	}

	if info.IsDir() {
		entries, _ := os.ReadDir(fullPath)
		type fileEntry struct {
			Name  string `json:"name"`
			IsDir bool   `json:"is_dir"`
			Size  int64  `json:"size"`
		}
		var files []fileEntry
		for _, e := range entries {
			ei, _ := e.Info()
			size := int64(0)
			if ei != nil {
				size = ei.Size()
			}
			files = append(files, fileEntry{
				Name:  e.Name(),
				IsDir: e.IsDir(),
				Size:  size,
			})
		}
		if files == nil {
			files = []fileEntry{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
		return
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	ext := filepath.Ext(fullPath)
	switch ext {
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".md":
		w.Header().Set("Content-Type", "text/markdown")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}
	w.Write(data)
}

func (s *Server) handlePutResult(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if strings.Contains(path, "..") {
		http.Error(w, "invalid path", 400)
		return
	}

	fullPath := filepath.Join(s.resultsDir(), path)
	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	if info.IsDir() {
		http.Error(w, "cannot write to directory", 400)
		return
	}

	var body struct {
		Content string `json:"content"`
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

	if err := os.WriteFile(fullPath, []byte(body.Content), 0644); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
