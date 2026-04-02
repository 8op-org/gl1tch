package orchestrator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadWorkflow parses a workflow definition from a YAML reader.
func LoadWorkflow(r io.Reader) (*WorkflowDef, error) {
	var def WorkflowDef
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&def); err != nil {
		return nil, fmt.Errorf("orchestrator: parse workflow yaml: %w", err)
	}
	return &def, nil
}

// FindWorkflow resolves a workflow name to its full path under
// ~/.config/glitch/workflows/<name>.workflow.yaml.
// Returns an error if the file does not exist.
func FindWorkflow(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("orchestrator: resolve home dir: %w", err)
	}
	path := filepath.Join(home, ".config", "glitch", "workflows", name+".workflow.yaml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("orchestrator: workflow %q not found at %s", name, path)
		}
		return "", fmt.Errorf("orchestrator: stat workflow %q: %w", name, err)
	}
	return path, nil
}
