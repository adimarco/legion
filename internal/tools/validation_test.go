package tools

import (
	"encoding/json"
	"testing"
)

// TestTool implements ToolValidator for testing
type TestTool struct {
	name       string
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

func TestValidateToolArgs(t *testing.T) {
	tests := []struct {
		name    string
		tool    ToolValidator
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid calculator args",
			tool: &TestTool{
				name:       "calculator",
				argsSchema: json.RawMessage(`{"type": "object", "properties": {"x": {"type": "number"}, "y": {"type": "number"}}, "required": ["x", "y"]}`),
			},
			args: map[string]interface{}{
				"x": 10.5,
				"y": 20.0,
			},
			wantErr: false,
		},
		{
			name: "invalid calculator args",
			tool: &TestTool{
				name:       "calculator",
				argsSchema: json.RawMessage(`{"type": "object", "properties": {"x": {"type": "number"}, "y": {"type": "number"}}, "required": ["x", "y"]}`),
			},
			args: map[string]interface{}{
				"x": "not a number",
				"y": 20.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolArgs(tt.tool, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
