package promptbuilder

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Save marshals the pipeline to YAML and writes it to path.
func Save(m *Model, path string) error {
	p := m.ToPipeline()
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
