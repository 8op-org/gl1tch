package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// RunShell executes a shell command and returns its stdout.
func RunShell(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("shell: %w\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunOllama sends a prompt to an Ollama model via the HTTP API.
func RunOllama(model, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: read: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama: %s\n%s", resp.Status, data)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("ollama: parse: %w", err)
	}
	return strings.TrimSpace(result.Response), nil
}

// RunClaude sends a prompt to Claude via the CLI and returns the response.
func RunClaude(model, prompt string) (string, error) {
	args := []string{"-p", "--output-format", "text"}
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, prompt)
	cmd := exec.Command("claude", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude: %w\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}
