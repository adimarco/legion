package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassthroughLLM(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()

		// Test initialization
		require.NoError(t, llm.Initialize(ctx, nil))
		assert.Equal(t, "test", llm.Name())
		assert.Equal(t, "passthrough", llm.Provider())

		// Test message generation
		msg := Message{
			Type:    MessageTypeUser,
			Content: "hello",
		}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Equal(t, MessageTypeAssistant, response.Type)
		assert.Equal(t, msg.Content, response.Content)
		assert.Equal(t, llm.Name(), response.Name)

		// Test string generation
		content := "test message"
		str, err := llm.GenerateString(ctx, content, nil)
		require.NoError(t, err)
		assert.Equal(t, content, str)
	})

	t.Run("fixed response", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()

		// Set fixed response
		msg := Message{
			Type:    MessageTypeUser,
			Content: FIXED_RESPONSE_INDICATOR + " fixed output",
		}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Equal(t, "fixed output", response.Content)

		// Subsequent messages should return fixed response
		msg = Message{
			Type:    MessageTypeUser,
			Content: "other message",
		}
		response, err = llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Equal(t, "fixed output", response.Content)
	})

	t.Run("fixed response empty", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()

		// Empty fixed response should be ignored
		msg := Message{
			Type:    MessageTypeUser,
			Content: FIXED_RESPONSE_INDICATOR,
		}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Equal(t, FIXED_RESPONSE_INDICATOR, response.Content)

		// Next message should echo as normal
		msg = Message{
			Type:    MessageTypeUser,
			Content: "normal message",
		}
		response, err = llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Equal(t, "normal message", response.Content)
	})

	t.Run("tool calls", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()

		t.Run("no args", func(t *testing.T) {
			name, args, err := llm.parseToolCall(CALL_TOOL_INDICATOR + " test_tool")
			require.NoError(t, err)
			assert.Equal(t, "test_tool", name)
			assert.Nil(t, args)
		})

		t.Run("with args", func(t *testing.T) {
			name, args, err := llm.parseToolCall(CALL_TOOL_INDICATOR + ` test_tool {"arg": "value", "num": 42}`)
			require.NoError(t, err)
			assert.Equal(t, "test_tool", name)
			require.NotNil(t, args)
			assert.Equal(t, "value", args["arg"])
			assert.Equal(t, float64(42), args["num"]) // JSON numbers are float64
		})

		t.Run("invalid json", func(t *testing.T) {
			_, _, err := llm.parseToolCall(CALL_TOOL_INDICATOR + ` test_tool {bad json}`)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid JSON")
		})

		t.Run("tool execution", func(t *testing.T) {
			msg := Message{
				Type:    MessageTypeUser,
				Content: CALL_TOOL_INDICATOR + ` test_tool {"arg": "value"}`,
			}
			response, err := llm.Generate(ctx, msg, nil)
			require.NoError(t, err)
			assert.Contains(t, response.Content, "Tool call: test_tool")
			assert.Contains(t, response.Content, "value")
		})
	})

	t.Run("history management", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Generate with history enabled
		params := &RequestParams{UseHistory: true}
		msg1 := Message{Type: MessageTypeUser, Content: "first"}
		msg2 := Message{Type: MessageTypeUser, Content: "second"}

		_, err := llm.Generate(ctx, msg1, params)
		require.NoError(t, err)
		_, err = llm.Generate(ctx, msg2, params)
		require.NoError(t, err)

		// Verify history
		history, err := llm.memory.Get(true)
		require.NoError(t, err)
		assert.Len(t, history, 4) // 2 user messages + 2 responses
		assert.Equal(t, msg1.Content, history[0].Content)
		assert.Equal(t, msg1.Content, history[1].Content)
		assert.Equal(t, msg2.Content, history[2].Content)
		assert.Equal(t, msg2.Content, history[3].Content)
	})

	t.Run("cleanup", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Add some messages
		params := &RequestParams{UseHistory: true}
		_, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "test"}, params)
		require.NoError(t, err)

		// Cleanup should clear memory
		require.NoError(t, llm.Cleanup())
		history, err := llm.memory.Get(true)
		require.NoError(t, err)
		assert.Empty(t, history)
	})

	t.Run("multipart messages", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()

		// Test message with multiple parts
		msg := Message{
			Type:    MessageTypeUser,
			Content: "main content",
			Parts: []MessagePart{
				{Type: "text", Content: "part 1"},
				{Type: "text", Content: "part 2"},
			},
		}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "main content")
		assert.Contains(t, response.Content, "part 1")
		assert.Contains(t, response.Content, "part 2")
	})

	t.Run("tool call formatting", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()

		// Test tool call with no arguments
		msg := Message{
			Type:    MessageTypeUser,
			Content: CALL_TOOL_INDICATOR + " test_tool",
		}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "Tool call: test_tool")
		assert.Contains(t, response.Content, "no arguments")

		// Test tool call with arguments
		msg = Message{
			Type:    MessageTypeUser,
			Content: CALL_TOOL_INDICATOR + ` test_tool {"key": "value", "num": 42}`,
		}
		response, err = llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "Tool call: test_tool")
		assert.Contains(t, response.Content, `"key": "value"`)
		assert.Contains(t, response.Content, `"num": 42`)

		// Test tool call with response
		result, err := llm.CallTool(ctx, ToolCall{
			Name:     "test_tool",
			Response: "success!",
		})
		require.NoError(t, err)
		assert.Contains(t, result, "Tool 'test_tool' result: success!")
	})
}
