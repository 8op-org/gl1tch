package router

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/8op-org/gl1tch/internal/pipeline"
)

// pipelineEmbedding is the on-disk (and in-memory) cache entry for one pipeline.
type pipelineEmbedding struct {
	// DescHash is the SHA256 of (name + description); used to detect staleness.
	DescHash string    `json:"desc_hash"`
	Vector   []float32 `json:"vector"`
}

// EmbeddingRouter computes cosine similarity between the prompt embedding and
// cached pipeline description embeddings. It maintains an in-memory cache
// (invalidated by description hash) and optionally persists to disk.
type EmbeddingRouter struct {
	embedder Embedder
	cfg      Config

	mu    sync.Mutex
	cache map[string]pipelineEmbedding // keyed by pipeline name
}

// newEmbeddingRouter creates an EmbeddingRouter. It loads any existing disk
// cache from cfg.CacheDir on construction.
func newEmbeddingRouter(embedder Embedder, cfg Config) *EmbeddingRouter {
	r := &EmbeddingRouter{
		embedder: embedder,
		cfg:      cfg,
		cache:    make(map[string]pipelineEmbedding),
	}
	r.loadDiskCache()
	return r
}

// Route computes similarity between the prompt and each pipeline description,
// returning the best match above AmbiguousThreshold (or nil pipeline if none qualify).
func (r *EmbeddingRouter) Route(ctx context.Context, prompt string, pipelines []pipeline.PipelineRef) (*RouteResult, error) {
	if len(pipelines) == 0 {
		return &RouteResult{Method: "none"}, nil
	}

	// Ensure all descriptions are embedded.
	if err := r.ensureEmbedded(ctx, pipelines); err != nil {
		return nil, err
	}

	// Embed the prompt.
	promptVec, err := r.embedder.Embed(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Find best match.
	var best *pipeline.PipelineRef
	var bestScore float64

	r.mu.Lock()
	for i := range pipelines {
		entry, ok := r.cache[pipelines[i].Name]
		if !ok {
			continue
		}
		score := cosineSimilarity(promptVec, entry.Vector)
		if score > bestScore {
			bestScore = score
			ref := pipelines[i]
			best = &ref
		}
	}
	r.mu.Unlock()

	if best == nil || bestScore < r.cfg.AmbiguousThreshold {
		return &RouteResult{Method: "none", Confidence: bestScore}, nil
	}

	return &RouteResult{
		Pipeline:   best,
		Confidence: bestScore,
		Method:     "embedding",
	}, nil
}

// ensureEmbedded checks the in-memory cache for each pipeline and (re-)embeds
// descriptions whose hash has changed.
func (r *EmbeddingRouter) ensureEmbedded(ctx context.Context, pipelines []pipeline.PipelineRef) error {
	r.mu.Lock()
	var toEmbed []pipeline.PipelineRef
	for _, p := range pipelines {
		h := hashDescription(p.Name + p.Description)
		if entry, ok := r.cache[p.Name]; ok && entry.DescHash == h {
			continue // cache hit
		}
		toEmbed = append(toEmbed, p)
	}
	r.mu.Unlock()

	if len(toEmbed) == 0 {
		return nil
	}

	var dirty bool
	for _, p := range toEmbed {
		vec, err := r.embedder.Embed(ctx, p.Description)
		if err != nil {
			return err
		}
		h := hashDescription(p.Name + p.Description)
		r.mu.Lock()
		r.cache[p.Name] = pipelineEmbedding{DescHash: h, Vector: vec}
		r.mu.Unlock()
		dirty = true
	}

	if dirty {
		r.saveDiskCache()
	}
	return nil
}

// ── disk cache ────────────────────────────────────────────────────────────────

func (r *EmbeddingRouter) cacheFilePath() string {
	if r.cfg.CacheDir == "" {
		return ""
	}
	return filepath.Join(r.cfg.CacheDir, "routing-index.json")
}

func (r *EmbeddingRouter) loadDiskCache() {
	path := r.cacheFilePath()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return // no cache yet — normal
	}
	var cache map[string]pipelineEmbedding
	if err := json.Unmarshal(data, &cache); err != nil {
		return // corrupted cache — ignore
	}
	r.mu.Lock()
	for k, v := range cache {
		r.cache[k] = v
	}
	r.mu.Unlock()
}

func (r *EmbeddingRouter) saveDiskCache() {
	path := r.cacheFilePath()
	if path == "" {
		return
	}
	r.mu.Lock()
	data, err := json.MarshalIndent(r.cache, "", "  ")
	r.mu.Unlock()
	if err != nil {
		return
	}
	// Write atomically via temp file.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// cosineSimilarity returns the cosine similarity between two float32 vectors.
// Returns 0 if lengths differ or either is the zero vector.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

// hashDescription returns the SHA256 hex digest of s.
func hashDescription(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
