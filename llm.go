package hive

import (
	"context"
	"fmt"

	"github.com/adimarco/hive/internal/config"
	"github.com/adimarco/hive/internal/llm"
)

// NewAnthropicLLM creates and initializes a new Anthropic LLM with sensible defaults.
// It will use environment variables and default configuration unless overridden.
func NewAnthropicLLM(name string, opts ...LLMOption) (llm.AugmentedLLM, error) {
	// Create base LLM
	l := llm.NewAnthropicLLM(name)

	// Load default settings
	settings := &config.Settings{
		DefaultModel: "claude-3-haiku-20240307", // Fast, cheap model by default
		Logger: config.LoggerSettings{
			Level: "info",
			Type:  "console",
		},
	}

	// Apply any options
	for _, opt := range opts {
		opt(settings)
	}

	// Initialize the LLM
	if err := l.Initialize(context.Background(), settings); err != nil {
		return nil, fmt.Errorf("failed to initialize LLM: %w", err)
	}

	return l, nil
}

// LLMOption allows customizing the LLM configuration
type LLMOption func(*config.Settings)

// WithModel sets the model to use
func WithModel(model string) LLMOption {
	return func(s *config.Settings) {
		s.DefaultModel = model
	}
}

// WithLogLevel sets the logging level
func WithLogLevel(level string) LLMOption {
	return func(s *config.Settings) {
		s.Logger.Level = level
	}
}

// WithLogType sets the logger type (console, file)
func WithLogType(logType string) LLMOption {
	return func(s *config.Settings) {
		s.Logger.Type = logType
	}
}
