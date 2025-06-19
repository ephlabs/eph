package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func readYAMLFile(path string) (map[string]interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}

	return result, nil
}
