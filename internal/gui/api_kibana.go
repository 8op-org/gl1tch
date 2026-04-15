package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const defaultKibanaURL = "http://localhost:5601"

func (s *Server) handleKibanaWorkflow(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	filter := fmt.Sprintf(`(query:(match_phrase:(workflow_name:'%s')))`, name)
	iframeURL := fmt.Sprintf(
		"%s/app/discover#/?_g=(time:(from:now-24h,to:now))&_a=(dataView:glitch-llm-calls,filters:!(%s),columns:!(step_id,model,tokens_in,tokens_out,latency_ms,cost_usd))",
		defaultKibanaURL, url.PathEscape(filter),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  iframeURL,
		"type": "workflow",
		"name": name,
	})
}

func (s *Server) handleKibanaRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	filter := fmt.Sprintf(`(query:(match_phrase:(run_id:'%s')))`, id)
	iframeURL := fmt.Sprintf(
		"%s/app/discover#/?_g=(time:(from:now-24h,to:now))&_a=(dataView:glitch-llm-calls,filters:!(%s),columns:!(step_id,model,tokens_in,tokens_out,latency_ms,cost_usd,escalated))",
		defaultKibanaURL, url.PathEscape(filter),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":  iframeURL,
		"type": "run",
		"id":   id,
	})
}
