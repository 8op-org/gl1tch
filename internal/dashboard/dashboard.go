package dashboard

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

//go:embed default.ndjson
var defaultDashboard embed.FS

// Seed imports the default dashboard into Kibana.
// It creates data views and imports all visualizations + dashboard.
// Safe to call multiple times — uses overwrite mode.
func Seed(ctx context.Context, kibanaURL string) error {
	kibanaURL = strings.TrimRight(kibanaURL, "/")

	// Wait for Kibana to be ready (up to 60s)
	client := &http.Client{Timeout: 10 * time.Second}
	for i := 0; i < 12; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", kibanaURL+"/api/status", nil)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i == 11 {
			return fmt.Errorf("kibana not ready at %s after 60s", kibanaURL)
		}
		time.Sleep(5 * time.Second)
	}

	// Create data views
	dataViews := []struct {
		ID    string
		Title string
		Name  string
	}{
		{"glitch-llm-calls", "glitch-llm-calls", "glitch LLM Calls"},
		{"glitch-workflow-runs", "glitch-workflow-runs", "glitch Workflow Runs"},
		{"glitch-tool-calls", "glitch-tool-calls", "glitch Tool Calls"},
		{"glitch-research-runs", "glitch-research-runs", "glitch Research Runs"},
	}

	for _, dv := range dataViews {
		// Delete existing data view first (ensures timeFieldName is set correctly)
		delReq, _ := http.NewRequestWithContext(ctx, "DELETE", kibanaURL+"/api/data_views/data_view/"+dv.ID, nil)
		delReq.Header.Set("kbn-xsrf", "true")
		if delResp, err := client.Do(delReq); err == nil && delResp != nil {
			delResp.Body.Close()
		}

		body, _ := json.Marshal(map[string]any{
			"data_view": map[string]any{
				"id":            dv.ID,
				"title":         dv.Title,
				"timeFieldName": "timestamp",
				"name":          dv.Name,
			},
		})
		req, _ := http.NewRequestWithContext(ctx, "POST", kibanaURL+"/api/data_views/data_view", bytes.NewReader(body))
		req.Header.Set("kbn-xsrf", "true")
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
		if err != nil {
			fmt.Printf("  data view %s: %v\n", dv.ID, err)
		}
	}

	// Import dashboard NDJSON
	ndjson, err := defaultDashboard.ReadFile("default.ndjson")
	if err != nil {
		return fmt.Errorf("read embedded dashboard: %w", err)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "default.ndjson")
	if err != nil {
		return fmt.Errorf("create form: %w", err)
	}
	part.Write(ndjson)
	w.Close()

	req, _ := http.NewRequestWithContext(ctx, "POST", kibanaURL+"/api/saved_objects/_import?overwrite=true", &buf)
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("import dashboard: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool `json:"success"`
	}
	json.Unmarshal(respBody, &result)

	if !result.Success {
		return fmt.Errorf("dashboard import failed: %s", respBody)
	}

	fmt.Println("  dashboard seeded: http://localhost:5601/app/dashboards#/view/glitch-llm-dashboard")
	return nil
}
