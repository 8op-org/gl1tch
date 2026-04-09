package provider

import (
	"fmt"
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

// RunOllama sends a prompt to an Ollama model and returns the response.
func RunOllama(model, prompt string) (string, error) {
	cmd := exec.Command("ollama", "run", model)
	cmd.Stdin = strings.NewReader(prompt)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ollama: %w\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
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
