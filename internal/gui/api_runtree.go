package gui

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// handleGetRunTree returns the nested run tree rooted at the given id.
func (s *Server) handleGetRunTree(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "store not available", http.StatusInternalServerError)
		return
	}
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tree, err := s.store.GetRunTree(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tree)
}
