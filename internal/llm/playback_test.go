package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaybackLLM(t *testing.T) {
	t.Run("basic playback", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Create a sequence of messages
		messages := []Message{
			{Type: MessageTypeUser, Content: "hello"},
			{Type: MessageTypeAssistant, Content: "hi there"},
			{Type: MessageTypeUser, Content: "how are you?"},
			{Type: MessageTypeAssistant, Content: "I'm good"},
		}
		llm.LoadMessages(messages)

		// First call should return HISTORY LOADED
		response, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "start"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "HISTORY LOADED")
		assert.Contains(t, response.Content, "4 messages")

		// Subsequent calls should return assistant messages in sequence
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "hi there", response.Content)

		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "I'm good", response.Content)
	})

	t.Run("message exhaustion", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Load just one assistant message
		messages := []Message{
			{Type: MessageTypeUser, Content: "hello"},
			{Type: MessageTypeAssistant, Content: "hi there"},
		}
		llm.LoadMessages(messages)

		// First call: HISTORY LOADED
		response, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "start"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "HISTORY LOADED")

		// Second call: get the one assistant message
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "hi there", response.Content)

		// Third call: should get exhausted message
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "MESSAGES EXHAUSTED")
		assert.Contains(t, response.Content, "0 overage")

		// Fourth call: should increment overage counter
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "MESSAGES EXHAUSTED")
		assert.Contains(t, response.Content, "1 overage")
	})

	t.Run("history management", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		messages := []Message{
			{Type: MessageTypeUser, Content: "hello"},
			{Type: MessageTypeAssistant, Content: "hi there"},
		}
		llm.LoadMessages(messages)

		// Generate with history enabled
		params := &RequestParams{UseHistory: true}

		// First call should store messages as prompts and return HISTORY LOADED
		response, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "start"}, params)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "HISTORY LOADED")

		// Verify history contains loaded messages plus the user message and response
		history, err := llm.memory.Get(true)
		require.NoError(t, err)
		assert.Len(t, history, 4) // 2 loaded messages + user message + HISTORY LOADED response
		assert.Equal(t, "hello", history[0].Content)
		assert.Equal(t, "hi there", history[1].Content)
		assert.Equal(t, "start", history[2].Content)
		assert.Contains(t, history[3].Content, "HISTORY LOADED")
	})

	t.Run("cleanup and reinitialization", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Load and use some messages
		messages := []Message{
			{Type: MessageTypeAssistant, Content: "hi there"},
		}
		llm.LoadMessages(messages)

		_, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "start"}, nil)
		require.NoError(t, err)

		// Cleanup should reset everything
		require.NoError(t, llm.Cleanup())
		assert.Empty(t, llm.messages)
		assert.Equal(t, -1, llm.currentIndex)
		assert.Equal(t, -1, llm.overage)

		// Should be able to reinitialize
		require.NoError(t, llm.Initialize(ctx, nil))
		llm.LoadMessages(messages)
		response, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "start"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "HISTORY LOADED")
	})
}
