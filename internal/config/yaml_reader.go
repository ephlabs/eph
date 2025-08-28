package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultConfigFile = "eph.yaml"

var (
	// Feature flag: if false, the function will default to using ./eph.yaml
	AllowCustomConfigPath = false
)

// readYAMLFile reads a YAML file at the given path
// and unmarshals it into a generic map[string]interface{}.
func readYAMLFile(path string) (any, error) {
	var filePath string
	if AllowCustomConfigPath && len(path) > 0 && path != "" {
		filePath = path
	} else {
		// Default to toplevel `eph.yaml`
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		filePath = filepath.Join(cwd, defaultConfigFile)
	}

	// Read file into memory
	data, err := os.ReadFile(filePath) /* #nosec G304 user input for filepath is ok for yaml config */
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
		return nil, errors.New("yaml file contains no data (empty or comments-only)")
	}

	return out, nil
}
