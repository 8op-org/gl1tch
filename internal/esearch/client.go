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

// IndexStat holds basic metadata about an ES index.
type IndexStat struct {
	Index     string `json:"index"`
	DocCount  string `json:"docs.count"`
	StoreSize string `json:"store.size"`
}

// IndexStats returns doc counts for all glitch-* indices.
func (c *Client) IndexStats(ctx context.Context) ([]IndexStat, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/_cat/indices/glitch-*?format=json&h=index,docs.count,store.size", nil)
	if err != nil {
		return nil, fmt.Errorf("index stats: build request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("index stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("index stats: status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("index stats: read body: %w", err)
	}

	return parseIndexStats(body), nil
}

// parseIndexStats parses /_cat/indices JSON, filtering to glitch-* indices.
func parseIndexStats(data []byte) []IndexStat {
	var raw []IndexStat
	json.Unmarshal(data, &raw)

	var filtered []IndexStat
	for _, s := range raw {
		if strings.HasPrefix(s.Index, "glitch-") {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// IndexDocResponse holds the parsed result of indexing a single document.
type IndexDocResponse struct {
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

// IndexDoc indexes a single document. If docID is empty, ES auto-generates an ID (POST);
// otherwise it upserts the document at the given ID (PUT).
func (c *Client) IndexDoc(ctx context.Context, index, docID string, doc json.RawMessage) (*IndexDocResponse, error) {
	method := http.MethodPut
	path := "/" + index + "/_doc/" + docID
	if docID == "" {
		method = http.MethodPost
		path = "/" + index + "/_doc"
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(doc))
	if err != nil {
		return nil, fmt.Errorf("index doc: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("index doc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("index doc: status %s — %s", resp.Status, truncate(string(body), 256))
	}

	var result IndexDocResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("index doc: decode response: %w", err)
	}
	return &result, nil
}

// IndexDocCreate indexes a document only if it doesn't already exist (op_type=create).
// Returns (response, existed, error). If the doc already exists, existed=true and response is nil.
func (c *Client) IndexDocCreate(ctx context.Context, index, docID string, doc json.RawMessage) (*IndexDocResponse, bool, error) {
	path := "/" + index + "/_doc/" + docID + "?op_type=create"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, bytes.NewReader(doc))
	if err != nil {
		return nil, false, fmt.Errorf("index doc create: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("index doc create: %w", err)
	}
	defer resp.Body.Close()

	// 409 Conflict means doc already exists — not an error for dedup
	if resp.StatusCode == 409 {
		return nil, true, nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("index doc create: status %s — %s", resp.Status, truncate(string(body), 256))
	}

	var result IndexDocResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, fmt.Errorf("index doc create: decode response: %w", err)
	}
	return &result, false, nil
}

// DeleteByQueryResponse holds the parsed result of a delete-by-query operation.
type DeleteByQueryResponse struct {
	Deleted int `json:"deleted"`
}

// DeleteByQuery deletes documents matching query from the given index.
func (c *Client) DeleteByQuery(ctx context.Context, index string, query json.RawMessage) (*DeleteByQueryResponse, error) {
	path := "/" + index + "/_delete_by_query"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("delete by query: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("delete by query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("delete by query: status %s — %s", resp.Status, truncate(string(body), 256))
	}

	var result DeleteByQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("delete by query: decode response: %w", err)
	}
	return &result, nil
}

// TermsAgg runs a terms aggregation on keyField with a nested top_hits to
// extract one valueField per bucket. Returns map[key]→value.
// A 404 (index not found) returns an empty map, not an error.
func (c *Client) TermsAgg(ctx context.Context, index, keyField, valueField string, size int) (map[string]string, error) {
	body := fmt.Sprintf(`{
		"size": 0,
		"aggs": {
			"keys": {
				"terms": { "field": %q, "size": %d },
				"aggs": {
					"latest": {
						"top_hits": { "size": 1, "_source": [%q] }
					}
				}
			}
		}
	}`, keyField, size, valueField)

	path := "/" + index + "/_search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("terms agg: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("terms agg: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return map[string]string{}, nil
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("terms agg: status %s — %s", resp.Status, truncate(string(b), 256))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("terms agg: read body: %w", err)
	}
	return parseTermsAggResponse(data, valueField), nil
}

// parseTermsAggResponse extracts key→value pairs from an ES terms aggregation
// response with nested top_hits.
func parseTermsAggResponse(data []byte, valueField string) map[string]string {
	var raw struct {
		Aggregations struct {
			Keys struct {
				Buckets []struct {
					Key    string `json:"key"`
					Latest struct {
						Hits struct {
							Hits []struct {
								Source map[string]any `json:"_source"`
							} `json:"hits"`
						} `json:"hits"`
					} `json:"latest"`
				} `json:"buckets"`
			} `json:"keys"`
		} `json:"aggregations"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return map[string]string{}
	}

	result := make(map[string]string, len(raw.Aggregations.Keys.Buckets))
	for _, b := range raw.Aggregations.Keys.Buckets {
		if len(b.Latest.Hits.Hits) == 0 {
			continue
		}
		if v, ok := b.Latest.Hits.Hits[0].Source[valueField]; ok {
			result[b.Key] = fmt.Sprint(v)
		}
	}
	return result
}

// truncate returns at most n bytes of s, appending "…" if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
