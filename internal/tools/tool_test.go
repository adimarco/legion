package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

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
			Version:     "1.0.0",
			Description: "Basic calculator",
			Category:    "math",
			Tags:        []string{"math", "arithmetic"},
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"operation": {"type": "string"},
					"a": {"type": "number"},
					"b": {"type": "number"}
				},
				"required": ["operation", "a", "b"]
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
		tool, err := registry.Get("calculator@1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "calculator", tool.Name)
		assert.Equal(t, "1.0.0", tool.Version)

		// Test latest version
		tool, err = registry.Get("calculator")
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", tool.Version)

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
		result, err = registry.Call(ctx, "calculator", map[string]any{
			"operation": "unknown",
			"a":         2.0,
			"b":         3.0,
		})
		require.NoError(t, err)
		assert.True(t, result.IsError)
		assert.Contains(t, result.Content, "unknown operation")
	})

	t.Run("tool dependencies", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		// Create a formatter tool
		formatterTool := Tool{
			Name:        "formatter",
			Version:     "1.0.0",
			Description: "Text formatter",
			Category:    "text",
			Tags:        []string{"text", "formatting"},
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"text": {"type": "string"}
				},
				"required": ["text"]
			}`),
			Handler: func(ctx context.Context, args map[string]any) (ToolResult, error) {
				return NewToolResult(fmt.Sprintf("formatted: %s", args["text"])), nil
			},
		}

		// Create a tool that depends on the formatter
		processorTool := Tool{
			Name:        "processor",
			Version:     "1.0.0",
			Description: "Text processor",
			Category:    "text",
			Tags:        []string{"text", "processing"},
			Requires:    []string{"formatter@1.0.0"},
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"text": {"type": "string"}
				},
				"required": ["text"]
			}`),
			Handler: func(ctx context.Context, args map[string]any) (ToolResult, error) {
				return NewToolResult(fmt.Sprintf("processed: %s", args["text"])), nil
			},
		}

		// Test dependency validation
		err := registry.Register(processorTool)
		assert.Error(t, err) // Should fail because formatter is not registered

		// Register formatter and try again
		err = registry.Register(formatterTool)
		require.NoError(t, err)

		err = registry.Register(processorTool)
		require.NoError(t, err)

		// Test execution
		result, err := registry.Call(ctx, "processor", map[string]any{
			"text": "hello",
		})
		require.NoError(t, err)
		assert.Equal(t, "processed: hello", result.Content)
	})

	t.Run("tool configuration", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		// Create a tool with configuration
		config := json.RawMessage(`{
			"prefix": "test:"
		}`)
		tool := Tool{
			Name:        "configured",
			Version:     "1.0.0",
			Description: "Configured tool",
			Category:    "test",
			Tags:        []string{"test", "config"},
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"input": {"type": "string"}
				},
				"required": ["input"],
				"config": {
					"type": "object",
					"properties": {
						"prefix": {"type": "string"}
					},
					"required": ["prefix"]
				}
			}`),
			Config: config,
			Handler: func(ctx context.Context, args map[string]any) (ToolResult, error) {
				var cfg struct {
					Prefix string `json:"prefix"`
				}
				if err := json.Unmarshal(config, &cfg); err != nil {
					return NewErrorResult(err), nil
				}
				return NewToolResult(fmt.Sprintf("%s %s", cfg.Prefix, args["input"])), nil
			},
		}

		// Test registration with config validation
		err := registry.Register(tool)
		require.NoError(t, err)

		// Test execution
		result, err := registry.Call(ctx, "configured", map[string]any{
			"input": "hello",
		})
		require.NoError(t, err)
		assert.Equal(t, "test: hello", result.Content)
	})

	t.Run("streaming execution", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		// Create a streaming tool
		streamingTool := Tool{
			Name:        "streamer",
			Version:     "1.0.0",
			Description: "Streaming tool",
			Category:    "test",
			Tags:        []string{"test", "streaming"},
			Schema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"count": {"type": "integer"}
				},
				"required": ["count"]
			}`),
			Handler: func(ctx context.Context, args map[string]any) (ToolResult, error) {
				// Handle both int and float64 (JSON numbers come as float64)
				var count int
				switch v := args["count"].(type) {
				case int:
					count = v
				case float64:
					count = int(v)
				default:
					return NewErrorResult(fmt.Errorf("invalid count type: %T", args["count"])), nil
				}
				return NewToolResult(fmt.Sprintf("count: %d", count)), nil
			},
		}

		// Register the tool
		err := registry.Register(streamingTool)
		require.NoError(t, err)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test streaming
		resultChan, err := registry.Stream(ctx, "streamer", map[string]any{
			"count": 3,
		})
		require.NoError(t, err)

		// Read results with timeout
		select {
		case result := <-resultChan:
			assert.Equal(t, "count: 3", result.Content)
			assert.False(t, result.IsError)
		case <-ctx.Done():
			t.Fatal("timeout waiting for result")
		}

		// Verify channel is closed
		select {
		case _, ok := <-resultChan:
			assert.False(t, ok, "channel should be closed")
		case <-ctx.Done():
			t.Fatal("timeout waiting for channel close")
		}
	})

	t.Run("search functionality", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		// Register some tools
		tools := []Tool{
			{
				Name:        "tool1",
				Version:     "1.0.0",
				Description: "Tool 1",
				Category:    "category1",
				Tags:        []string{"tag1", "tag2"},
				Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			},
			{
				Name:        "tool2",
				Version:     "1.0.0",
				Description: "Tool 2",
				Category:    "category2",
				Tags:        []string{"tag2", "tag3"},
				Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			},
		}

		for _, tool := range tools {
			err := registry.Register(tool)
			require.NoError(t, err)
		}

		// Test search by name
		results := registry.Search(map[string]any{"name": "tool1"})
		assert.Len(t, results, 1)
		assert.Equal(t, "tool1", results[0].Name)

		// Test search by category
		results = registry.Search(map[string]any{"category": "category2"})
		assert.Len(t, results, 1)
		assert.Equal(t, "tool2", results[0].Name)

		// Test search by tag
		results = registry.Search(map[string]any{"tag": "tag2"})
		assert.Len(t, results, 2)
	})

	t.Run("lifecycle hooks", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		initialized := false
		cleaned := false

		// Create a tool with lifecycle hooks
		lifecycleTool := Tool{
			Name:        "lifecycle",
			Version:     "1.0.0",
			Description: "Lifecycle tool",
			Category:    "test",
			Tags:        []string{"test"},
			Handler:     func(ctx context.Context, args map[string]any) (ToolResult, error) { return ToolResult{}, nil },
			Initialize: func(ctx context.Context) error {
				initialized = true
				return nil
			},
			Cleanup: func(ctx context.Context) error {
				cleaned = true
				return nil
			},
		}

		// Test initialization on registration
		err := registry.Register(lifecycleTool)
		require.NoError(t, err)
		assert.True(t, initialized)

		// Test cleanup on unregister
		err = registry.Unregister("lifecycle")
		require.NoError(t, err)
		assert.True(t, cleaned)
	})

	t.Run("concurrent access", func(t *testing.T) {
		registry := NewSimpleToolRegistry()

		// Create a tool
		tool := Tool{
			Name:        "concurrent",
			Version:     "1.0.0",
			Description: "Concurrent tool",
			Category:    "test",
			Tags:        []string{"test"},
			Handler: func(ctx context.Context, args map[string]any) (ToolResult, error) {
				time.Sleep(10 * time.Millisecond) // Simulate work
				return NewToolResult("ok"), nil
			},
		}

		// Register tool
		require.NoError(t, registry.Register(tool))

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Run concurrent calls
		var wg sync.WaitGroup
		errs := make(chan error, 10)
		results := make(chan string, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result, err := registry.Call(ctx, "concurrent", nil)
				if err != nil {
					errs <- err
					return
				}
				results <- result.Content
			}()
		}

		// Wait with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Check results
			close(results)
			close(errs)

			// Verify no errors
			for err := range errs {
				assert.NoError(t, err)
			}

			// Verify results
			count := 0
			for result := range results {
				assert.Equal(t, "ok", result)
				count++
			}
			assert.Equal(t, 10, count, "should receive all results")

		case <-ctx.Done():
			t.Fatal("timeout waiting for concurrent operations")
		}
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

	t.Run("result with resources", func(t *testing.T) {
		result := ToolResult{
			Content:   "test",
			Resources: []string{"res1", "res2"},
			Cost:      100,
		}
		assert.Equal(t, uint64(100), result.Cost)
		assert.Equal(t, []string{"res1", "res2"}, result.Resources)
	})
}
