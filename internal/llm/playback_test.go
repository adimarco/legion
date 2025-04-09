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

		// First call should echo back since playback hasn't started
		response, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "start"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "start", response.Content)

		// Start playback
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "!start"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "Starting playback")
		assert.Contains(t, response.Content, "4 messages")

		// Get first assistant message
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "hi there", response.Content)

		// Get second assistant message
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

		// Start playback
		response, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "!start"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "Starting playback")

		// Get the one assistant message
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "hi there", response.Content)

		// Should get exhausted message
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "MESSAGES EXHAUSTED")
		assert.Contains(t, response.Content, "0 overage")

		// Should increment overage counter
		response, err = llm.Generate(ctx, Message{Type: MessageTypeUser, Content: "next"}, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "MESSAGES EXHAUSTED")
		assert.Contains(t, response.Content, "1 overage")
	})

	t.Run("record and trigger responses", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Record some responses
		recordMsg := Message{
			Type:    MessageTypeUser,
			Content: "!record hello Hi there, how can I help?",
		}
		response, err := llm.Generate(ctx, recordMsg, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "Recorded response for trigger")

		// Test invalid record command
		invalidRecord := Message{
			Type:    MessageTypeUser,
			Content: "!record",
		}
		_, err = llm.Generate(ctx, invalidRecord, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid record response command")

		// Test trigger matching
		triggers := []struct {
			input    string
			expected string
		}{
			{"hello world", "Hi there, how can I help?"},
			{"saying hello to you", "Hi there, how can I help?"},
			{"different message", "different message"}, // Should echo back
		}

		for _, tt := range triggers {
			msg := Message{Type: MessageTypeUser, Content: tt.input}
			response, err := llm.Generate(ctx, msg, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, response.Content)
		}
	})

	t.Run("history management", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Generate with history enabled
		params := &RequestParams{UseHistory: true}

		// Record and trigger some responses
		messages := []struct {
			input    string
			expected string
		}{
			{"!record hello Hi there!", "Recorded response for trigger: hello"},
			{"hello world", "Hi there!"},
			{"another message", "another message"},
		}

		for _, msg := range messages {
			response, err := llm.Generate(ctx, Message{Type: MessageTypeUser, Content: msg.input}, params)
			require.NoError(t, err)
			assert.Equal(t, msg.expected, response.Content)
		}

		// Verify history contains all messages
		history, err := llm.memory.Get(true)
		require.NoError(t, err)
		assert.Len(t, history, len(messages)*2) // Each exchange has user message + response
	})

	t.Run("cleanup and reinitialization", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Record some responses and load messages
		recordMsg := Message{
			Type:    MessageTypeUser,
			Content: "!record hello Hi there!",
		}
		_, err := llm.Generate(ctx, recordMsg, nil)
		require.NoError(t, err)

		llm.LoadMessages([]Message{
			{Type: MessageTypeAssistant, Content: "test message"},
		})

		// Cleanup should reset everything
		require.NoError(t, llm.Cleanup())
		assert.Empty(t, llm.messages)
		assert.Equal(t, 0, llm.currentIndex)
		assert.Equal(t, 0, llm.overage)
		assert.Empty(t, llm.responses)
		assert.False(t, llm.playbackStarted)

		// Should be able to reinitialize
		require.NoError(t, llm.Initialize(ctx, nil))

		// Verify clean state
		msg := Message{Type: MessageTypeUser, Content: "hello"}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", response.Content) // Should echo back, no recorded response
	})

	t.Run("concurrent operations", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Record a response
		recordMsg := Message{
			Type:    MessageTypeUser,
			Content: "!record hello Hi there!",
		}
		_, err := llm.Generate(ctx, recordMsg, nil)
		require.NoError(t, err)

		// Run concurrent operations
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				msg := Message{Type: MessageTypeUser, Content: "hello"}
				response, err := llm.Generate(ctx, msg, nil)
				require.NoError(t, err)
				assert.Equal(t, "Hi there!", response.Content)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("tool call integration", func(t *testing.T) {
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		// Record a response that includes a tool call
		recordMsg := Message{
			Type:    MessageTypeUser,
			Content: "!record search " + ToolCallPrefix + "search {\"query\": \"test\"}",
		}
		_, err := llm.Generate(ctx, recordMsg, nil)
		require.NoError(t, err)

		// Trigger the tool call
		msg := Message{Type: MessageTypeUser, Content: "please search for something"}
		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, "Tool call: search")
		assert.Contains(t, response.Content, `"query": "test"`)
	})

	t.Run("tool calls", func(t *testing.T) {
		msg := Message{
			Type:    MessageTypeUser,
			Content: ToolCallPrefix + ` search {"query": "test"}`,
		}
		llm := NewPlaybackLLM("test")
		ctx := context.Background()
		require.NoError(t, llm.Initialize(ctx, nil))

		response, err := llm.Generate(ctx, msg, nil)
		require.NoError(t, err)
		assert.Contains(t, response.Content, ToolCallPrefix)
		assert.Contains(t, response.Content, "search")
		assert.Contains(t, response.Content, `"query": "test"`)
	})
}
