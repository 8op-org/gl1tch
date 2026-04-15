package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// LMStudioProvider calls LM Studio's OpenAI-compatible API at localhost:1234
// with model management via native REST endpoints.
type LMStudioProvider struct {
	BaseURL      string        // default "http://localhost:1234"
	DefaultModel string        // default "qwen3-8b"
	PollInterval time.Duration // polling interval for download status; default 2s
}

// checkModels queries LM Studio for model availability and load state.
func (p *LMStudioProvider) checkModels(model string) (exists, loaded bool, err error) {
	url := strings.TrimRight(p.BaseURL, "/") + "/api/v0/models"
	resp, err := http.Get(url)
	if err != nil {
		return false, false, fmt.Errorf("lm-studio: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, false, fmt.Errorf("lm-studio: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return false, false, fmt.Errorf("lm-studio: %s\n%s", resp.Status, data)
	}

	var result struct {
		Data []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return false, false, fmt.Errorf("lm-studio: parse: %w", err)
	}

	for _, m := range result.Data {
		if m.ID == model || strings.Contains(m.ID, model) {
			return true, m.State == "loaded", nil
		}
	}
	return false, false, nil
}

// pollInterval returns the configured poll interval or the default of 2s.
func (p *LMStudioProvider) pollInterval() time.Duration {
	if p.PollInterval > 0 {
		return p.PollInterval
	}
	return 2 * time.Second
}

// downloadModel triggers a model download and polls until completed or failed.
func (p *LMStudioProvider) downloadModel(model string) error {
	base := strings.TrimRight(p.BaseURL, "/")

	// POST to start download
	reqBody, _ := json.Marshal(map[string]string{"model": model})
	resp, err := http.Post(base+"/api/v1/models/download", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("lm-studio: download request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("lm-studio: download read: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("lm-studio: download: %s\n%s", resp.Status, data)
	}

	var dlResp struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(data, &dlResp); err != nil {
		return fmt.Errorf("lm-studio: download parse: %w", err)
	}

	// Poll for status
	for {
		time.Sleep(p.pollInterval())

		statusURL := fmt.Sprintf("%s/api/v1/models/download/status/%s", base, dlResp.JobID)
		sResp, err := http.Get(statusURL)
		if err != nil {
			return fmt.Errorf("lm-studio: download status: %w", err)
		}

		sData, err := io.ReadAll(sResp.Body)
		sResp.Body.Close()
		if err != nil {
			return fmt.Errorf("lm-studio: download status read: %w", err)
		}

		var status struct {
			Status   string  `json:"status"`
			Progress float64 `json:"progress"`
		}
		if err := json.Unmarshal(sData, &status); err != nil {
			return fmt.Errorf("lm-studio: download status parse: %w", err)
		}

		switch status.Status {
		case "completed":
			return nil
		case "failed":
			return fmt.Errorf("lm-studio: download failed for %s", model)
		default:
			fmt.Fprintf(os.Stderr, ">> lm-studio: downloading %s (%.0f%%)...\n", model, status.Progress)
		}
	}
}
