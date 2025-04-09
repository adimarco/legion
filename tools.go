package hive

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adimarco/hive/internal/tools"
)

// Tool creates a new tool with sensible defaults
func Tool(name string, handler func(context.Context, map[string]any) (string, error)) *tools.ToolBuilder {
	builder := tools.New(name).
		WithHandler(handler).
		WithVersion("1.0.0")

	// Create a basic schema for the tool
	schema := map[string]interface{}{
		"type":        "object",
		"description": fmt.Sprintf("Tool '%s' - use this tool to get results", name),
		"properties":  map[string]interface{}{},
		"required":    []string{},
	}

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		// In practice this should never happen with our simple schema
		panic(fmt.Sprintf("failed to marshal schema: %v", err))
	}

	return builder.WithSchema(schemaJSON)
}
