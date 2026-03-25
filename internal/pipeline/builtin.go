package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// BuiltinFunc is the function signature for all builtin step implementations.
// args are the interpolated step Args map. Returns structured output or error.
type BuiltinFunc func(ctx context.Context, args map[string]any, w io.Writer) (map[string]any, error)

// builtinRegistry maps "builtin.<name>" → BuiltinFunc.
var builtinRegistry = map[string]BuiltinFunc{
	"builtin.assert":   builtinAssert,
	"builtin.set_data": builtinSetData,
	"builtin.log":      builtinLog,
	"builtin.sleep":    builtinSleep,
	"builtin.http_get": builtinHTTPGet,
}

// builtinAssert evaluates a condition expression against an optional value arg.
// Args:
//   - "condition": required — same syntax as EvalCondition ("always", "contains:<s>", etc.)
//     Additionally supports "true" and "false" literals.
//   - "value":     optional — the string to match against; defaults to "".
//   - "message":   optional — error message on failure.
func builtinAssert(_ context.Context, args map[string]any, _ io.Writer) (map[string]any, error) {
	condRaw, _ := args["condition"]
	cond := toString(condRaw)

	value := toString(args["value"])
	message := toString(args["message"])

	var passed bool
	switch {
	case cond == "true":
		passed = true
	case cond == "false":
		passed = false
	default:
		vars := map[string]any{"_output": value}
		passed = EvalCondition(cond, vars)
	}

	if !passed {
		if message != "" {
			return nil, fmt.Errorf("assert failed: %s", message)
		}
		return nil, fmt.Errorf("assert failed: condition %q evaluated to false for value %q", cond, value)
	}
	return map[string]any{"passed": true}, nil
}

// builtinSetData merges the "data" arg (map[string]any) into the output map and returns it.
func builtinSetData(_ context.Context, args map[string]any, _ io.Writer) (map[string]any, error) {
	raw, ok := args["data"]
	if !ok {
		return map[string]any{}, nil
	}
	data, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("builtin.set_data: 'data' arg must be a map, got %T", raw)
	}
	out := make(map[string]any, len(data))
	for k, v := range data {
		out[k] = v
	}
	return out, nil
}

// builtinLog writes a message to the pipeline output writer. Returns nil output.
func builtinLog(_ context.Context, args map[string]any, w io.Writer) (map[string]any, error) {
	msg := toString(args["message"])
	if w != nil {
		fmt.Fprintln(w, msg)
	}
	return nil, nil
}

// builtinSleep sleeps for the specified duration, respecting context cancellation.
// Args: "duration" — a duration string like "2s", "500ms".
func builtinSleep(ctx context.Context, args map[string]any, _ io.Writer) (map[string]any, error) {
	durStr := toString(args["duration"])
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return nil, fmt.Errorf("builtin.sleep: invalid duration %q: %w", durStr, err)
	}
	select {
	case <-time.After(dur):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return nil, nil
}

// builtinHTTPGet performs an HTTP GET to the "url" arg.
// Args:
//   - "url":     required — target URL.
//   - "timeout": optional — request timeout string (default "10s").
//
// Returns {"status": <int>, "body": <string>} on 2xx; error on non-2xx or network failure.
func builtinHTTPGet(ctx context.Context, args map[string]any, _ io.Writer) (map[string]any, error) {
	urlStr := toString(args["url"])
	if urlStr == "" {
		return nil, fmt.Errorf("builtin.http_get: 'url' arg is required")
	}

	timeoutStr := toString(args["timeout"])
	if timeoutStr == "" {
		timeoutStr = "10s"
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("builtin.http_get: invalid timeout %q: %w", timeoutStr, err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("builtin.http_get: create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("builtin.http_get: request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("builtin.http_get: read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("builtin.http_get: non-2xx status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return map[string]any{
		"status": resp.StatusCode,
		"body":   string(bodyBytes),
	}, nil
}

// expandForEach resolves the ForEach field of a step to a list of items.
// It first tries to parse the string as a JSON array; if that fails, it splits
// on newlines and trims whitespace, skipping empty lines.
func expandForEach(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// Try JSON array first.
	if strings.HasPrefix(raw, "[") {
		var items []any
		if err := json.Unmarshal([]byte(raw), &items); err == nil {
			out := make([]string, 0, len(items))
			for _, item := range items {
				out = append(out, fmt.Sprint(item))
			}
			return out
		}
	}
	// Fall back to newline-separated list.
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// conditionMatchesError checks whether an error string matches a simple condition
// for retry On field. Currently just checks "always" (always retry) vs "on_failure"
// (only on non-nil errors).
func conditionMatchesError(on string, err error) bool {
	switch strings.TrimSpace(on) {
	case "", "always":
		return true
	case "on_failure":
		return err != nil
	}
	// Try evaluating as a condition expression against the error message.
	if err == nil {
		return false
	}
	vars := map[string]any{"_output": err.Error()}
	re, compErr := regexp.Compile(on)
	if compErr == nil {
		return re.MatchString(err.Error())
	}
	return EvalCondition(on, vars)
}

