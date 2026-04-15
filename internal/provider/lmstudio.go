package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
