package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// readYAMLFile reads a YAML file at the given path
// and unmarshals it into a generic map[string]interface{}.
func readYAMLFile(path string) (any, error) {
	// Read file into memory
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// if there is nothing to unmarshal, return an error
	if len(data) == 0 {
		return nil, errors.New("empty file")
	}

	// Unmarshal into a generic structure
	var out any
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	// if the file is empty but syntactically valid, return an error
	if out == nil {
		return nil, errors.New("failed to parse yaml")
	}

	return out, nil
}
