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

	t.Run("tool calls", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Test tool call
		call := ToolCall{
			ID:   "test-call",
			Name: "test-tool",
			Args: map[string]any{
				"arg1": "value1",
				"arg2": 42,
			},
		}

		result, err := llm.CallTool(ctx, call)
		require.NoError(t, err)
		assert.Contains(t, result, "Tool call: test-tool")
		assert.Contains(t, result, "value1")
		assert.Contains(t, result, "42")
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
}
