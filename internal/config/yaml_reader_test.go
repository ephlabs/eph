package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadYAMLFile(t *testing.T) {
	t.Run("read valid YAML file", func(t *testing.T) {
		data, err := readYAMLFile("testdata/valid.yaml")
		require.NoError(t, err)
		require.NotNil(t, data)

		assert.Equal(t, "1.0", data["version"])
		assert.Equal(t, "test-project", data["name"])

		envs, ok := data["environments"].([]interface{})
		require.True(t, ok, "environments should be a slice")
		require.Len(t, envs, 1)

		env0, ok := envs[0].(map[string]interface{})
		require.True(t, ok, "environment should be a map")
		assert.Equal(t, "dev", env0["name"])
		assert.Equal(t, "kubernetes", env0["provider"])
		assert.Equal(t, "dev-namespace", env0["namespace"])
	})

	t.Run("file not found", func(t *testing.T) {
		data, err := readYAMLFile("testdata/nonexistent.yaml")
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.True(t, os.IsNotExist(err), "error should be file not found")
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		data, err := readYAMLFile("testdata/invalid.yaml")
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "yaml:")
	})

	t.Run("empty file path", func(t *testing.T) {
		data, err := readYAMLFile("")
		assert.Error(t, err)
		assert.Nil(t, data)
	})
}
