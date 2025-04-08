package config

import (
	"fmt"
	"strings"
)

// Valid logger types
var validLoggerTypes = map[string]bool{
	"none":    true,
	"console": true,
	"file":    true,
}

// Valid log levels
var validLogLevels = map[string]bool{
	"debug":   true,
	"info":    true,
	"warning": true,
	"error":   true,
}

// Valid transport types
var validTransportTypes = map[string]bool{
	"stdio": true,
	"sse":   true,
}

// Validate checks if the settings are valid
func (s *Settings) Validate() error {
	// Validate logger settings
	if err := s.Logger.Validate(); err != nil {
		return fmt.Errorf("invalid logger settings: %w", err)
	}

	// Validate MCP settings
	if err := s.MCP.Validate(); err != nil {
		return fmt.Errorf("invalid MCP settings: %w", err)
	}

	return nil
}

// Validate checks if the logger settings are valid
func (s *LoggerSettings) Validate() error {
	// Check logger type
	if !validLoggerTypes[s.Type] {
		return fmt.Errorf("invalid logger type %q, must be one of: %s",
			s.Type, strings.Join(mapKeys(validLoggerTypes), ", "))
	}

	// Check log level
	if !validLogLevels[s.Level] {
		return fmt.Errorf("invalid log level %q, must be one of: %s",
			s.Level, strings.Join(mapKeys(validLogLevels), ", "))
	}

	// Check file path if type is file
	if s.Type == "file" && s.Path == "" {
		return fmt.Errorf("path is required for file logger")
	}

	// Check batch size
	if s.BatchSize < 1 {
		return fmt.Errorf("batch size must be greater than 0")
	}

	return nil
}

// Validate checks if the MCP settings are valid
func (s *MCPSettings) Validate() error {
	for name, server := range s.Servers {
		if err := server.Validate(); err != nil {
			return fmt.Errorf("invalid server %q: %w", name, err)
		}
	}
	return nil
}

// Validate checks if the server settings are valid
func (s *MCPServerSettings) Validate() error {
	// Check transport type
	if !validTransportTypes[s.Transport] {
		return fmt.Errorf("invalid transport %q, must be one of: %s",
			s.Transport, strings.Join(mapKeys(validTransportTypes), ", "))
	}

	// Validate stdio transport requirements
	if s.Transport == "stdio" {
		if s.Command == "" {
			return fmt.Errorf("command is required for stdio transport")
		}
	}

	// Validate SSE transport requirements
	if s.Transport == "sse" {
		if s.URL == "" {
			return fmt.Errorf("url is required for sse transport")
		}
	}

	return nil
}

// mapKeys returns a sorted slice of map keys
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
