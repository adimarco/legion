package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleToolRegistry(t *testing.T) {
	ctx := context.Background()

	t.Run("basic registration and execution", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		// Create a simple calculator tool
		calcTool := Tool{
			Name:        "calculator",
			Description: "Basic calculator",
			InputSchema: json.RawMessage(`{
                "type": "object",
                "properties": {
                    "operation": {"type": "string"},
                    "a": {"type": "number"},
                    "b": {"type": "number"}
                }
            }`),
			Handler: func(ctx context.Context, args map[string]any) (ToolResult, error) {
				a := args["a"].(float64)
				b := args["b"].(float64)
				switch args["operation"].(string) {
				case "add":
					return NewToolResult(fmt.Sprintf("%f", a+b)), nil
				default:
					return NewErrorResult(fmt.Errorf("unknown operation")), nil
				}
			},
		}

		// Test registration
		err := registry.Register(calcTool)
		require.NoError(t, err)

		// Test duplicate registration
		err = registry.Register(calcTool)
		assert.Error(t, err)

		// Test retrieval
		tool, err := registry.Get("calculator")
		require.NoError(t, err)
		assert.Equal(t, "calculator", tool.Name)

		// Test listing
		tools := registry.List()
		assert.Len(t, tools, 1)
		assert.Equal(t, "calculator", tools[0].Name)

		// Test execution
		result, err := registry.Call(ctx, "calculator", map[string]any{
			"operation": "add",
			"a":         2.0,
			"b":         3.0,
		})
		require.NoError(t, err)
		assert.Equal(t, "5.000000", result.Content)
		assert.False(t, result.IsError)

		// Test error handling
		result2, err := registry.Call(ctx, "calculator", map[string]any{
			"operation": "unknown",
			"a":         2.0,
			"b":         3.0,
		})
		require.NoError(t, err)
		assert.True(t, result2.IsError)
		assert.Contains(t, result2.Content, "unknown operation")
	})

	t.Run("validation", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		// Test empty name
		err := registry.Register(Tool{
			Description: "test",
			Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
		})
		assert.Error(t, err)

		// Test nil handler
		err = registry.Register(Tool{
			Name:        "test",
			Description: "test",
		})
		assert.Error(t, err)
	})

	t.Run("concurrent access", func(t *testing.T) {
		registry := NewSimpleToolRegistry()
		tool := Tool{
			Name:        "test",
			Description: "test",
			Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return NewToolResult("ok"), nil },
		}

		// Register tool
		require.NoError(t, registry.Register(tool))

		// Run concurrent calls
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result, err := registry.Call(ctx, "test", nil)
				assert.NoError(t, err)
				assert.Equal(t, "ok", result.Content)
			}()
		}
		wg.Wait()
	})
}

func TestToolResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := NewToolResult("test")
		assert.Equal(t, "test", result.Content)
		assert.False(t, result.IsError)
	})

	t.Run("error result", func(t *testing.T) {
		err := fmt.Errorf("test error")
		result := NewErrorResult(err)
		assert.Equal(t, "test error", result.Content)
		assert.True(t, result.IsError)
	})
}
