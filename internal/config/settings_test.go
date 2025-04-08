package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSettings(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "fastagent.config.yaml")

	configData := `
default_model: haiku

logger:
  type: console
  level: info
  progress_display: true
  path: test.jsonl
  batch_size: 50

mcp:
  servers:
    test_server:
      name: "Test Server"
      description: "A test server"
      transport: stdio
      command: npx
      args: ["-y", "@modelcontextprotocol/server-test"]
      env:
        TEST_KEY: test_value
`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Test loading the config
	settings, err := LoadSettings(configPath)
	require.NoError(t, err)

	// Verify the loaded settings
	assert.Equal(t, "haiku", settings.DefaultModel)

	// Verify logger settings
	assert.Equal(t, "console", settings.Logger.Type)
	assert.Equal(t, "info", settings.Logger.Level)
	assert.True(t, settings.Logger.ProgressDisplay)
	assert.Equal(t, "test.jsonl", settings.Logger.Path)
	assert.Equal(t, 50, settings.Logger.BatchSize)

	// Verify MCP server settings
	server, ok := settings.MCP.Servers["test_server"]
	require.True(t, ok)
	assert.Equal(t, "Test Server", server.Name)
	assert.Equal(t, "A test server", server.Description)
	assert.Equal(t, "stdio", server.Transport)
	assert.Equal(t, "npx", server.Command)
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-test"}, server.Args)
	assert.Equal(t, "test_value", server.Env["TEST_KEY"])
}

func TestLoadSettings_FileNotFound(t *testing.T) {
	// Test with nonexistent file - should return default settings
	settings, err := LoadSettings("nonexistent.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, settings)

	// Verify default values
	assert.Equal(t, "file", settings.Logger.Type)
	assert.Equal(t, "warning", settings.Logger.Level)
	assert.True(t, settings.Logger.ProgressDisplay)
	assert.Equal(t, "fastagent.jsonl", settings.Logger.Path)
	assert.Equal(t, 100, settings.Logger.BatchSize)
}

func TestLoadSettings_InvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `
default_model: haiku
logger:
  invalid yaml content
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	_, err = LoadSettings(configPath)
	assert.Error(t, err)
}

func TestLoadSettings_EnvironmentOverrides(t *testing.T) {
	// Set up environment variables
	envVars := map[string]string{
		"FASTAGENT_DEFAULT_MODEL":           "sonnet",
		"FASTAGENT_LOGGER_TYPE":             "console",
		"FASTAGENT_LOGGER_LEVEL":            "debug",
		"FASTAGENT_LOGGER_PROGRESS_DISPLAY": "false",
		"FASTAGENT_LOGGER_PATH":             "env.jsonl",
		"FASTAGENT_LOGGER_BATCH_SIZE":       "200",
		"FASTAGENT_MCP_SERVER_TEST_NAME":    "Env Test Server",
		"FASTAGENT_MCP_SERVER_TEST_COMMAND": "test-cmd",
		"FASTAGENT_MCP_SERVER_TEST_ARGS":    "arg1,arg2",
		"FASTAGENT_MCP_SERVER_TEST_ENV":     "KEY1=value1,KEY2=value2",
	}

	// Set environment variables
	for k, v := range envVars {
		t.Setenv(k, v)
	}

	// Load settings without a config file
	settings, err := LoadSettings("")
	require.NoError(t, err)

	// Verify environment overrides
	assert.Equal(t, "sonnet", settings.DefaultModel)
	assert.Equal(t, "console", settings.Logger.Type)
	assert.Equal(t, "debug", settings.Logger.Level)
	assert.False(t, settings.Logger.ProgressDisplay)
	assert.Equal(t, "env.jsonl", settings.Logger.Path)
	assert.Equal(t, 200, settings.Logger.BatchSize)

	// Verify MCP server settings from environment
	server, ok := settings.MCP.Servers["test"]
	require.True(t, ok)
	assert.Equal(t, "Env Test Server", server.Name)
	assert.Equal(t, "test-cmd", server.Command)
	assert.Equal(t, []string{"arg1", "arg2"}, server.Args)
	assert.Equal(t, "value1", server.Env["KEY1"])
	assert.Equal(t, "value2", server.Env["KEY2"])
}

func TestLoadSettings_EnvironmentOverridesWithFile(t *testing.T) {
	// Create a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "fastagent.config.yaml")

	configData := `
default_model: haiku
logger:
  type: file
  level: info
  progress_display: true
  path: file.jsonl
  batch_size: 50
`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Set environment variables that should override the file
	t.Setenv("FASTAGENT_DEFAULT_MODEL", "sonnet")
	t.Setenv("FASTAGENT_LOGGER_TYPE", "console")
	t.Setenv("FASTAGENT_LOGGER_BATCH_SIZE", "200")

	// Load settings
	settings, err := LoadSettings(configPath)
	require.NoError(t, err)

	// Verify that environment variables override file settings
	assert.Equal(t, "sonnet", settings.DefaultModel)
	assert.Equal(t, "console", settings.Logger.Type)
	assert.Equal(t, 200, settings.Logger.BatchSize)

	// Verify that unset environment variables retain file values
	assert.Equal(t, "info", settings.Logger.Level)
	assert.True(t, settings.Logger.ProgressDisplay)
	assert.Equal(t, "file.jsonl", settings.Logger.Path)
}
