package router

import (
	"context"

	"github.com/8op-org/gl1tch/internal/brainrag"
)

// OllamaEmbedder implements Embedder using Ollama's /api/embeddings endpoint.
// It wraps brainrag.Embed and defaults to nomic-embed-text if no model is set.
type OllamaEmbedder struct {
	BaseURL string
	Model   string
}

// Embed computes an embedding for text using the configured Ollama model.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return brainrag.Embed(ctx, e.BaseURL, e.Model, text)
}
