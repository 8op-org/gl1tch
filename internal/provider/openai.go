package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OpenAICompatibleProvider calls any OpenAI-compatible chat completions API.
type OpenAICompatibleProvider struct {
	Name         string
	BaseURL      string // e.g. "https://openrouter.ai/api/v1"
	APIKey       string
	DefaultModel string
}

// Chat sends a prompt to the chat completions endpoint and returns a structured LLMResult.
func (p *OpenAICompatibleProvider) Chat(model, prompt string) (LLMResult, error) {
	if model == "" {
		model = p.DefaultModel
	}

	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: marshal: %w", err)
	}

	start := time.Now()

	url := strings.TrimRight(p.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResult{}, fmt.Errorf("openai: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return LLMResult{}, fmt.Errorf("openai: %s\n%s", resp.Status, data)
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &chatResp); err != nil {
		return LLMResult{}, fmt.Errorf("openai: parse: %w", err)
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = strings.TrimSpace(chatResp.Choices[0].Message.Content)
	}

	cost := 0.0
	if h := resp.Header.Get("x-openrouter-cost"); h != "" {
		cost, _ = strconv.ParseFloat(h, 64)
	}

	return LLMResult{
		Provider:  p.Name,
		Model:     model,
		Response:  content,
		TokensIn:  chatResp.Usage.PromptTokens,
		TokensOut: chatResp.Usage.CompletionTokens,
		Latency:   time.Since(start),
		CostUSD:   cost,
	}, nil
}
