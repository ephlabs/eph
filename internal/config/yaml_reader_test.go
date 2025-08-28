package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: Currently only supports the top-level YAML file "eph.yaml"
// In the future, once/if we allow setting of custom configuration files, we can
// remove the feature flag logic and also remove the repetitive code in each of
// these tests.

func TestCanHandleEmptyFile(t *testing.T) {
	// set the feature flag so we can test additional functionality (do this every test)
	AllowCustomConfigPath = true

	path := filepath.Join("testdata", "empty.yaml")
	_, err := readYAMLFile(path)
	assert.Error(t, err, "expected error when reading an empty file")

	// reset the feature flag (do this every test)
	AllowCustomConfigPath = false
}

// also handles files with only comments in them
func TestCanHandleCommentsOnlyFile(t *testing.T) {
	AllowCustomConfigPath = true

	path := filepath.Join("testdata", "comments_only.yaml")
	_, err := readYAMLFile(path)
	assert.Error(t, err, "expected error when reading a comments-only file")

	AllowCustomConfigPath = false
}

// also handles:
// - different types in the yaml
// - nested yaml
func TestCanReadYAMLFile(t *testing.T) {
	AllowCustomConfigPath = true

	path := filepath.Join("testdata", "valid.yaml")
	data, err := readYAMLFile(path)
	assert.NoError(t, err)

	yamlMap := data.(map[string]any)

	assert.Equal(t, "TestData", yamlMap["name"])
	assert.Equal(t, 1.0, yamlMap["version"])
	assert.Equal(t, true, yamlMap["enabled"])
	assert.Equal(t, 100, yamlMap["max_users"])

	features := yamlMap["features"].(map[string]any)
	assert.Equal(t, true, features["authentication"])
	assert.Equal(t, "info", features["logging"].(map[string]any)["level"])

	logging := features["logging"].(map[string]any)
	assert.ElementsMatch(t, []any{"file", "console"}, logging["destinations"])

	database := yamlMap["database"].(map[string]any)
	assert.Equal(t, "localhost", database["host"])
	assert.Equal(t, 1234, database["port"])

	AllowCustomConfigPath = false
}

// also handles when given a path to a dir rather than a file
// also handles using the default config file
func TestThrowsWhenFileDoesNotExist(t *testing.T) {
	// specifically not setting the feature flag here

	_, err := readYAMLFile("blah")
	assert.Error(t, err)
}

func TestThrowsWhenFileIsBadYAML(t *testing.T) {
	AllowCustomConfigPath = true

	path := filepath.Join("testdata", "invalid.yaml")
	_, err := readYAMLFile(path)
	assert.Error(t, err)

	AllowCustomConfigPath = false
}
