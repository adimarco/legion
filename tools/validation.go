package tools

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// ToolValidator represents a tool that can be validated
type ToolValidator interface {
	// GetArgsSchema returns the JSON schema for validating tool arguments
	GetArgsSchema() json.RawMessage
}

// ValidateArgs validates the provided arguments against a JSON schema.
// If the schema is empty, validation is skipped.
func ValidateArgs(schema json.RawMessage, args map[string]interface{}) error {
	if len(schema) == 0 {
		return nil
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schema))
	documentLoader := gojsonschema.NewGoLoader(args)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, desc := range result.Errors() {
			errMsgs = append(errMsgs, desc.String())
		}
		return fmt.Errorf("invalid arguments: %v", errMsgs)
	}

	return nil
}

// ValidateToolArgs validates the arguments for a specific tool using its schema
func ValidateToolArgs(tool ToolValidator, args map[string]interface{}) error {
	return ValidateArgs(tool.GetArgsSchema(), args)
}

// ValidateToolConfig validates a tool's configuration against its schema
func ValidateToolConfig(tool *Tool) error {
	if len(tool.Config) == 0 {
		return nil // No config to validate
	}

	// Parse the config schema from the tool's schema
	var schema struct {
		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(tool.Schema, &schema); err != nil {
		return fmt.Errorf("failed to parse config schema: %w", err)
	}

	if len(schema.Config) == 0 {
		return fmt.Errorf("tool has config but no config schema defined")
	}

	// Parse the actual config
	var config map[string]interface{}
	if err := json.Unmarshal(tool.Config, &config); err != nil {
		return fmt.Errorf("failed to parse tool config: %w", err)
	}

	return ValidateArgs(schema.Config, config)
}

// ValidateToolDependencies validates that all required tools exist and are compatible
func ValidateToolDependencies(tool *Tool, registry ToolRegistry) error {
	for _, dep := range tool.Requires {
		depTool, err := registry.Get(dep)
		if err != nil {
			return fmt.Errorf("required tool %q not found: %w", dep, err)
		}

		// Validate the dependency's configuration
		if err := ValidateToolConfig(&depTool); err != nil {
			return fmt.Errorf("invalid configuration for dependency %q: %w", dep, err)
		}
	}
	return nil
}

// ValidateToolMetadata validates a tool's metadata is complete and valid
func ValidateToolMetadata(tool *Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Version == "" {
		return fmt.Errorf("tool version is required")
	}
	if tool.Description == "" {
		return fmt.Errorf("tool description is required")
	}
	if tool.Category == "" {
		return fmt.Errorf("tool category is required")
	}
	if len(tool.Tags) == 0 {
		return fmt.Errorf("tool must have at least one tag")
	}
	return nil
}

// ValidateTool performs comprehensive validation of a tool
func ValidateTool(tool *Tool, registry ToolRegistry) error {
	// Validate metadata
	if err := ValidateToolMetadata(tool); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	// Validate configuration
	if err := ValidateToolConfig(tool); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Validate dependencies
	if err := ValidateToolDependencies(tool, registry); err != nil {
		return fmt.Errorf("invalid dependencies: %w", err)
	}

	// Validate handler
	if tool.Handler == nil {
		return fmt.Errorf("tool handler is required")
	}

	return nil
}
