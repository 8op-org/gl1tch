package gui

import (
	"encoding/json"
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
		http.Error(w, "not found", 404)
		return
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
