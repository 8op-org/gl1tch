package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EmbedOllama calls the Ollama /api/embeddings endpoint and returns the embedding vector.
// baseURL is the Ollama API base (e.g. "http://localhost:11434").
func EmbedOllama(ctx context.Context, baseURL, model, text string) ([]float64, error) {
	body, _ := json.Marshal(map[string]string{
		"model":  model,
		"prompt": text,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama embed: %s\n%s", resp.Status, data)
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("ollama embed: parse: %w", err)
	}
	return result.Embedding, nil
}
