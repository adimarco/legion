package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassthroughLLM(t *testing.T) {
	ctx := context.Background()

	t.Run("basic", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		require.NoError(t, llm.Initialize(ctx, nil))

		msg := Message{
			Type:    MessageTypeUser,
			Content: "test message",
		}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Equal(t, "test message", response.Content)
	})

	t.Run("fixed response", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		require.NoError(t, llm.Initialize(ctx, nil))

		msg := Message{
			Type:    MessageTypeUser,
			Content: FixedResponsePrefix + " fixed output",
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

	t.Run("tool calls", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
		require.NoError(t, llm.Initialize(ctx, nil))

		t.Run("no args", func(t *testing.T) {
			name, args, err := llm.parseToolCall(ToolCallPrefix + " test_tool")
			require.NoError(t, err)
			assert.Equal(t, "test_tool", name)
			assert.Nil(t, args)
		})

		t.Run("with args", func(t *testing.T) {
			name, args, err := llm.parseToolCall(ToolCallPrefix + ` test_tool {"arg": "value", "num": 42}`)
			require.NoError(t, err)
			assert.Equal(t, "test_tool", name)
			require.NotNil(t, args)
			assert.Equal(t, "value", args["arg"])
			assert.Equal(t, float64(42), args["num"]) // JSON numbers are float64
		})

		t.Run("invalid json", func(t *testing.T) {
			_, _, err := llm.parseToolCall(ToolCallPrefix + ` test_tool {bad json}`)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid JSON")
		})

		t.Run("tool execution", func(t *testing.T) {
			msg := Message{
				Type:    MessageTypeUser,
				Content: ToolCallPrefix + ` test_tool {"arg": "value"}`,
			}
			response, err := llm.Generate(ctx, msg, nil)
			require.NoError(t, err)
			assert.Contains(t, response.Content, "Tool call: test_tool")
			assert.Contains(t, response.Content, "value")
		})
	})

	t.Run("history", func(t *testing.T) {
		llm := NewPassthroughLLM("test")
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
}
