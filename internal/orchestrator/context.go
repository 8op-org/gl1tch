package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// maxContextValueBytes is the maximum size of a single context value in bytes.
const maxContextValueBytes = 16 * 1024

// WorkflowContext is a thread-safe key-value store threaded through all steps.
// Keys follow ADK-style scoping: "temp." prefix for per-run ephemeral state,
// "<step_id>.output" for step outputs.
type WorkflowContext struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewWorkflowContext creates a new empty WorkflowContext.
func NewWorkflowContext() *WorkflowContext {
	return &WorkflowContext{
		data: make(map[string]string),
	}
}

// Set stores value under key. Values exceeding 16 KB are truncated with a
// warning printed to stderr.
func (c *WorkflowContext) Set(key, value string) {
	if len(value) > maxContextValueBytes {
		fmt.Fprintf(os.Stderr, "orchestrator: context value for key %q truncated from %d to %d bytes\n",
			key, len(value), maxContextValueBytes)
		value = value[:maxContextValueBytes]
	}
	c.mu.Lock()
	c.data[key] = value
	c.mu.Unlock()
}

// Get retrieves a value by key. Returns "" if the key is not found.
func (c *WorkflowContext) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

// Marshal serializes the full context map to JSON.
func (c *WorkflowContext) Marshal() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	b, err := json.Marshal(c.data)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: marshal context: %w", err)
	}
	return b, nil
}

// Unmarshal restores the context from JSON, replacing the current state.
func (c *WorkflowContext) Unmarshal(data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("orchestrator: unmarshal context: %w", err)
	}
	c.mu.Lock()
	c.data = m
	c.mu.Unlock()
	return nil
}

// ExpandTemplate replaces {{ctx.<key>}} references in s with values from c.
// Unknown keys expand to the empty string.
func ExpandTemplate(s string, c *WorkflowContext) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Walk through all known keys and substitute.
	result := s
	for k, v := range c.data {
		placeholder := "{{ctx." + k + "}}"
		result = strings.ReplaceAll(result, placeholder, v)
	}
	// Remove any remaining {{ctx.*}} placeholders (unknown keys expand to "").
	for {
		start := strings.Index(result, "{{ctx.")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+2:]
	}
	return result
}
