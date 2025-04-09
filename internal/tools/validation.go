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
