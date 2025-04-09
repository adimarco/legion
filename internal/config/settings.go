package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Settings represents the root configuration structure
type Settings struct {
	// DefaultModel takes format: <provider>.<model_string>.<reasoning_effort?>
	// e.g. anthropic.claude-3-5-sonnet-20241022 or openai.o3-mini.low
	DefaultModel string `yaml:"default_model" env:"DEFAULT_MODEL"`

	// Logger configuration
	Logger LoggerSettings `yaml:"logger"`

	// MCP server configurations
	MCP MCPSettings `yaml:"mcp"`
}

// LoggerSettings configures logging behavior
type LoggerSettings struct {
	// Type of logger to use
	Type string `yaml:"type" env:"LOGGER_TYPE" default:"file"`

	// Minimum logging level
	Level string `yaml:"level" env:"LOGGER_LEVEL" default:"warning"`

	// Enable or disable progress display
	ProgressDisplay bool `yaml:"progress_display" env:"LOGGER_PROGRESS_DISPLAY" default:"true"`

	// Path to log file if Type is "file"
	Path string `yaml:"path" env:"LOGGER_PATH" default:"fastagent.jsonl"`

	// Number of events to accumulate before processing
	BatchSize int `yaml:"batch_size" env:"LOGGER_BATCH_SIZE" default:"100"`
}

// MCPSettings holds MCP server configurations
type MCPSettings struct {
	// Map of server name to server configuration
	Servers map[string]MCPServerSettings `yaml:"servers"`
}

// MCPServerSettings configures an individual MCP server
type MCPServerSettings struct {
	// Name of the server (optional)
	Name string `yaml:"name,omitempty"`

	// Description of the server (optional)
	Description string `yaml:"description,omitempty"`

	// Transport mechanism ("stdio" or "sse")
	Transport string `yaml:"transport" default:"stdio"`

	// Command to execute the server (e.g. npx)
	Command string `yaml:"command,omitempty"`

	// Arguments for the server command
	Args []string `yaml:"args,omitempty"`

	// URL for the server (required for SSE transport)
	URL string `yaml:"url,omitempty"`

	// Environment variables to pass to the server process
	Env map[string]string `yaml:"env,omitempty"`
}

// LoadSettings loads configuration from a YAML file and environment variables
func LoadSettings(configPath string) (*Settings, error) {
	// If no path specified, look in default locations
	if configPath == "" {
		configPath = "fastagent.config.yaml"
	}

	// Create settings with defaults
	settings := &Settings{
		Logger: LoggerSettings{
			Type:            "file",
			Level:           "warning",
			ProgressDisplay: true,
			Path:            "fastagent.jsonl",
			BatchSize:       100,
		},
		MCP: MCPSettings{
			Servers: make(map[string]MCPServerSettings),
		},
	}

	// Load from YAML if file exists
	if configPath != "" {
		// Resolve absolute path
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, err
		}

		// Read the config file if it exists
		if _, err := os.Stat(absPath); err == nil {
			data, err := os.ReadFile(absPath)
			if err != nil {
				return nil, err
			}

			if err := yaml.Unmarshal(data, settings); err != nil {
				return nil, err
			}
		}
	}

	// Override with environment variables
	if err := loadFromEnv(settings); err != nil {
		return nil, err
	}

	// Validate the settings
	if err := settings.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return settings, nil
}

// loadFromEnv overrides settings with environment variables
func loadFromEnv(settings *Settings) error {
	// Load root settings
	if val := os.Getenv(EnvPrefix + "DEFAULT_MODEL"); val != "" {
		settings.DefaultModel = val
	}

	// Load logger settings
	if val := os.Getenv(EnvPrefix + "LOGGER_TYPE"); val != "" {
		settings.Logger.Type = val
	}
	if val := os.Getenv(EnvPrefix + "LOGGER_LEVEL"); val != "" {
		settings.Logger.Level = val
	}
	if val := os.Getenv(EnvPrefix + "LOGGER_PROGRESS_DISPLAY"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			settings.Logger.ProgressDisplay = b
		}
	}
	if val := os.Getenv(EnvPrefix + "LOGGER_PATH"); val != "" {
		settings.Logger.Path = val
	}
	if val := os.Getenv(EnvPrefix + "LOGGER_BATCH_SIZE"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			settings.Logger.BatchSize = i
		}
	}

	// Load MCP server settings from environment
	// Format: FASTAGENT_MCP_SERVER_<name>_<field>=value
	prefix := EnvPrefix + "MCP_SERVER_"
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}

		key, value := splitEnv(env)
		parts := strings.Split(strings.TrimPrefix(key, prefix), "_")
		if len(parts) < 2 {
			continue
		}

		serverName := strings.ToLower(parts[0])
		field := strings.ToLower(strings.Join(parts[1:], "_"))

		// Create server if it doesn't exist
		if _, exists := settings.MCP.Servers[serverName]; !exists {
			settings.MCP.Servers[serverName] = MCPServerSettings{
				Transport: "stdio", // Set default transport
			}
		}

		// Get the server settings
		server := settings.MCP.Servers[serverName]

		// Update the appropriate field
		switch field {
		case "name":
			server.Name = value
		case "description":
			server.Description = value
		case "transport":
			server.Transport = value
		case "command":
			server.Command = value
		case "args":
			server.Args = strings.Split(value, ",")
		case "url":
			server.URL = value
		case "env":
			// Format: key1=value1,key2=value2
			if server.Env == nil {
				server.Env = make(map[string]string)
			}
			for _, pair := range strings.Split(value, ",") {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					server.Env[kv[0]] = kv[1]
				}
			}
		}

		// Update the server in the map
		settings.MCP.Servers[serverName] = server
	}

	return nil
}

// splitEnv splits an environment variable into key and value
func splitEnv(env string) (key, value string) {
	parts := strings.SplitN(env, "=", 2)
	if len(parts) != 2 {
		return env, ""
	}
	return parts[0], parts[1]
}
