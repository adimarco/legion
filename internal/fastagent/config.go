package fastagent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds configuration values for the FastAgent system
type Config struct {
	AnthropicAPIKey string        `envconfig:"ANTHROPIC_API_KEY" required:"true"`
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info"`
	MaxTokens       int           `envconfig:"MAX_TOKENS" default:"512"`
	DefaultModel    string        `envconfig:"DEFAULT_MODEL" default:"claude-3-haiku-20240307"`
	Temperature     float32       `envconfig:"TEMPERATURE" default:"0.7"`
	Timeout         time.Duration `envconfig:"TIMEOUT" default:"10s"`
}

// findConfigFile looks for .env file in current directory and parent directories
func findConfigFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(".env file not found in current or parent directories")
}

// LoadConfig loads configuration from environment variables
// It will first try to load a .env file if present, then process environment variables
func LoadConfig() (*Config, error) {
	// Try to find and load .env file
	if configPath, err := findConfigFile(); err == nil {
		if err := godotenv.Load(configPath); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("error processing config: %w", err)
	}

	return &cfg, nil
}

// String returns a string representation of the config (with sensitive values redacted)
func (c *Config) String() string {
	return fmt.Sprintf("Config{AnthropicAPIKey: <redacted>, LogLevel: %s, MaxTokens: %d, DefaultModel: %s, Temperature: %.1f, Timeout: %s}",
		c.LogLevel, c.MaxTokens, c.DefaultModel, c.Temperature, c.Timeout)
}
