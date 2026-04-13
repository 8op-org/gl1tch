package esearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a thin HTTP Elasticsearch client with no SDK dependencies.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient returns a new Client targeting baseURL with a 30s timeout.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchResult is a single hit from an ES search response.
type SearchResult struct {
	Index  string          `json:"_index"`
	Source json.RawMessage `json:"_source"`
}

// SearchResponse holds the parsed results of an ES search.
type SearchResponse struct {
	Total   int
	Results []SearchResult
}

// BulkDoc is a document to be indexed via the bulk API.
type BulkDoc struct {
	ID   string
	Body json.RawMessage
}

// Ping does a GET to the baseURL and returns an error if the status >= 400.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("ping: build request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("ping: unexpected status %s", resp.Status)
	}
	return nil
}

// Search posts query to /{indices}/_search and returns parsed hits.
func (c *Client) Search(ctx context.Context, indices []string, query json.RawMessage) (*SearchResponse, error) {
	path := "/" + strings.Join(indices, ",") + "/_search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("search: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search: status %s — %s", resp.Status, truncate(string(body), 256))
	}

	var raw struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []SearchResult `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("search: decode response: %w", err)
	}

	return &SearchResponse{
		Total:   raw.Hits.Total.Value,
		Results: raw.Hits.Hits,
	}, nil
}

// BulkIndex sends docs to /_bulk as NDJSON.
func (c *Client) BulkIndex(ctx context.Context, index string, docs []BulkDoc) error {
	var buf bytes.Buffer
	for _, d := range docs {
		meta := fmt.Sprintf(`{"index":{"_index":%q,"_id":%q}}`, index, d.ID)
		buf.WriteString(meta)
		buf.WriteByte('\n')
		buf.Write(d.Body)
		buf.WriteByte('\n')
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/_bulk", &buf)
	if err != nil {
		return fmt.Errorf("bulk index: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("bulk index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bulk index: status %s — %s", resp.Status, truncate(string(body), 256))
	}
	return nil
}

// EnsureIndex checks for the index via HEAD; creates it with the given mapping if absent.
func (c *Client) EnsureIndex(ctx context.Context, index, mapping string) error {
	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL+"/"+index, nil)
	if err != nil {
		return fmt.Errorf("ensure index: build head request: %w", err)
	}
	headResp, err := c.http.Do(headReq)
	if err != nil {
		return fmt.Errorf("ensure index: head: %w", err)
	}
	headResp.Body.Close()

	if headResp.StatusCode == http.StatusOK {
		return nil
	}

	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/"+index, strings.NewReader(mapping))
	if err != nil {
		return fmt.Errorf("ensure index: build put request: %w", err)
	}
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := c.http.Do(putReq)
	if err != nil {
		return fmt.Errorf("ensure index: put: %w", err)
	}
	defer putResp.Body.Close()

	if putResp.StatusCode >= 400 {
		body, _ := io.ReadAll(putResp.Body)
		return fmt.Errorf("ensure index: create %s: status %s — %s", index, putResp.Status, truncate(string(body), 256))
	}
	return nil
}

// truncate returns at most n bytes of s, appending "…" if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
