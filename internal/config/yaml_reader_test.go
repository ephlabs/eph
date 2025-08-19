package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// start with "throw an error"
func TestCanHandleEmptyFile(t *testing.T) {
	path := filepath.Join("testdata", "empty.yaml")
	_, err := readYAMLFile(path)
	assert.Error(t, err, "expected error when reading an empty file")
}

// also handles files with only comments in them
func TestCanHandleCommentsOnlyFile(t *testing.T) {
	path := filepath.Join("testdata", "comments_only.yaml")
	_, err := readYAMLFile(path)
	assert.Error(t, err, "expected error when reading a comments-only file")
}

// also handles:
// - different types in the yaml
// - nested yaml
func TestCanReadYAMLFile(t *testing.T) {
	path := filepath.Join("testdata", "valid.yaml")
	data, err := readYAMLFile(path)
	assert.NoError(t, err)

	yamlMap := data.(map[string]any)

	assert.Equal(t, "TestData", yamlMap["name"])
	assert.Equal(t, 1.0, yamlMap["version"])
	assert.Equal(t, true, yamlMap["enabled"])
	assert.Equal(t, int(100), yamlMap["max_users"])

	features := yamlMap["features"].(map[string]any)
	assert.Equal(t, true, features["authentication"])
	assert.Equal(t, "info", features["logging"].(map[string]any)["level"])

	logging := features["logging"].(map[string]any)
	assert.ElementsMatch(t, []any{"file", "console"}, logging["destinations"])

	database := yamlMap["database"].(map[string]any)
	assert.Equal(t, "localhost", database["host"])
	assert.Equal(t, int(1234), database["port"])
}

// also handles when given a path to a dir rather than a file
func TestThrowsWhenFileDoesNotExist(t *testing.T) {
	path := filepath.Join("testdata", "nonexistent.yaml")
	_, err := readYAMLFile(path)
	assert.Error(t, err)
}

func TestThrowsWhenFileIsBadYAML(t *testing.T) {
	path := filepath.Join("testdata", "invalid.yaml")
	_, err := readYAMLFile(path)
	assert.Error(t, err)
}
