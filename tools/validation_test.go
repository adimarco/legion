package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTool implements ToolValidator for testing
type TestTool struct {
	argsSchema json.RawMessage
}

func (t *TestTool) GetArgsSchema() json.RawMessage {
	return t.argsSchema
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name    string
		schema  json.RawMessage
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name:   "valid simple schema",
			schema: json.RawMessage(`{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "integer"}}, "required": ["name"]}`),
			args: map[string]interface{}{
				"name": "test",
				"age":  30,
			},
			wantErr: false,
		},
		{
			name:   "invalid type",
			schema: json.RawMessage(`{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "integer"}}, "required": ["name"]}`),
			args: map[string]interface{}{
				"name": "test",
				"age":  "30", // string instead of integer
			},
			wantErr: true,
		},
		{
			name:   "missing required field",
			schema: json.RawMessage(`{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "integer"}}, "required": ["name"]}`),
			args: map[string]interface{}{
				"age": 30,
			},
			wantErr: true,
		},
		{
			name:    "empty schema",
			schema:  json.RawMessage(""),
			args:    map[string]interface{}{"name": "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArgs(tt.schema, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateToolMetadata(t *testing.T) {
	tests := []struct {
		name    string
		tool    *Tool
		wantErr bool
	}{
		{
			name: "valid metadata",
			tool: &Tool{
				Name:        "test",
				Description: "Test tool",
				Category:    "test",
				Tags:        []string{"test"},
				Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tool: &Tool{
				Description: "Test tool",
				Category:    "test",
				Tags:        []string{"test"},
				Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			},
			wantErr: true,
		},
		{
			name: "missing description",
			tool: &Tool{
				Name:     "test",
				Category: "test",
				Tags:     []string{"test"},
				Handler:  func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			},
			wantErr: true,
		},
		{
			name: "missing category",
			tool: &Tool{
				Name:        "test",
				Description: "Test tool",
				Tags:        []string{"test"},
				Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			},
			wantErr: true,
		},
		{
			name: "missing tags",
			tool: &Tool{
				Name:        "test",
				Description: "Test tool",
				Category:    "test",
				Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolMetadata(tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTool(t *testing.T) {
	registry := NewSimpleToolRegistry()

	// Create a valid tool
	validTool := Tool{
		Name:        "test",
		Description: "Test tool",
		Category:    "test",
		Tags:        []string{"test"},
		Schema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"input": {"type": "string"}
			},
			"required": ["input"]
		}`),
		Handler: func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
	}

	// Test valid tool
	err := ValidateTool(&validTool, registry)
	assert.NoError(t, err)

	// Test invalid metadata
	invalidMetadata := validTool
	invalidMetadata.Name = ""
	err = ValidateTool(&invalidMetadata, registry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid metadata")

	// Test missing handler
	invalidHandler := validTool
	invalidHandler.Handler = nil
	err = ValidateTool(&invalidHandler, registry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool handler is required")
}
