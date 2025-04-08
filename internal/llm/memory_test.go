package llm

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleMemory(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		mem := NewSimpleMemory()

		// Add some messages
		promptMsg := Message{
			Type:    MessageTypeSystem,
			Content: "system prompt",
		}
		require.NoError(t, mem.Add(promptMsg, true))

		historyMsg := Message{
			Type:    MessageTypeUser,
			Content: "user message",
		}
		require.NoError(t, mem.Add(historyMsg, false))

		// Get with history
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		assert.Len(t, msgs, 2)
		assert.Equal(t, promptMsg, msgs[0])
		assert.Equal(t, historyMsg, msgs[1])

		// Get without history
		msgs, err = mem.Get(false)
		require.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Equal(t, promptMsg, msgs[0])
	})

	t.Run("message ordering", func(t *testing.T) {
		mem := NewSimpleMemory()

		// Add messages in specific order
		messages := []struct {
			msg      Message
			isPrompt bool
		}{
			{Message{Type: MessageTypeSystem, Content: "prompt1"}, true},
			{Message{Type: MessageTypeUser, Content: "msg1"}, false},
			{Message{Type: MessageTypeSystem, Content: "prompt2"}, true},
			{Message{Type: MessageTypeUser, Content: "msg2"}, false},
		}

		for _, m := range messages {
			require.NoError(t, mem.Add(m.msg, m.isPrompt))
		}

		// Verify order is maintained
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		assert.Len(t, msgs, 4)

		// Prompts should come first, in order
		assert.Equal(t, "prompt1", msgs[0].Content)
		assert.Equal(t, "prompt2", msgs[1].Content)
		// Then history messages, in order
		assert.Equal(t, "msg1", msgs[2].Content)
		assert.Equal(t, "msg2", msgs[3].Content)
	})

	t.Run("clear operations", func(t *testing.T) {
		mem := NewSimpleMemory()

		// Add messages
		require.NoError(t, mem.Add(Message{Type: MessageTypeSystem, Content: "prompt"}, true))
		require.NoError(t, mem.Add(Message{Type: MessageTypeUser, Content: "history"}, false))

		// Clear history only
		require.NoError(t, mem.Clear(false))
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Equal(t, MessageTypeSystem, msgs[0].Type)

		// Clear everything
		require.NoError(t, mem.Clear(true))
		msgs, err = mem.Get(true)
		require.NoError(t, err)
		assert.Empty(t, msgs)
	})

	t.Run("concurrent operations", func(t *testing.T) {
		mem := NewSimpleMemory()
		var wg sync.WaitGroup
		numGoroutines := 10

		// Test concurrent writes
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(i int) {
				defer wg.Done()
				msg := Message{
					Type:    MessageTypeUser,
					Content: fmt.Sprintf("message %d", i),
				}
				_ = mem.Add(msg, false)
			}(i)
		}
		wg.Wait()

		// Test concurrent reads and writes
		wg.Add(numGoroutines * 2)
		for i := 0; i < numGoroutines; i++ {
			// Reader
			go func() {
				defer wg.Done()
				_, _ = mem.Get(true)
			}()

			// Writer
			go func(i int) {
				defer wg.Done()
				msg := Message{
					Type:    MessageTypeSystem,
					Content: fmt.Sprintf("prompt %d", i),
				}
				_ = mem.Add(msg, true)
			}(i)
		}
		wg.Wait()

		// Verify final state
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		assert.Len(t, msgs, numGoroutines*2) // numGoroutines messages + numGoroutines prompts
	})

	t.Run("message immutability", func(t *testing.T) {
		mem := NewSimpleMemory()

		// Add a message
		original := Message{
			Type:    MessageTypeUser,
			Content: "original",
			Metadata: map[string]any{
				"key": "value",
			},
		}
		require.NoError(t, mem.Add(original, false))

		// Get the message and modify it
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		require.Len(t, msgs, 1)

		msgs[0].Content = "modified"
		msgs[0].Metadata["key"] = "new value"

		// Get again and verify original is unchanged
		msgs2, err := mem.Get(true)
		require.NoError(t, err)
		require.Len(t, msgs2, 1)
		assert.Equal(t, "original", msgs2[0].Content)
		assert.Equal(t, "value", msgs2[0].Metadata["key"])
	})

	t.Run("empty state handling", func(t *testing.T) {
		mem := NewSimpleMemory()

		// Get from empty memory
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		assert.Empty(t, msgs)

		// Clear empty memory
		require.NoError(t, mem.Clear(true))

		// Add and clear repeatedly
		for i := 0; i < 5; i++ {
			require.NoError(t, mem.Add(Message{Content: "test"}, false))
			require.NoError(t, mem.Clear(true))
			msgs, err = mem.Get(true)
			require.NoError(t, err)
			assert.Empty(t, msgs)
		}
	})

	t.Run("message type handling", func(t *testing.T) {
		mem := NewSimpleMemory()

		// Test all message types
		types := []MessageType{
			MessageTypeSystem,
			MessageTypeUser,
			MessageTypeAssistant,
			MessageTypeTool,
		}

		for _, typ := range types {
			msg := Message{
				Type:    typ,
				Content: fmt.Sprintf("test %s", typ),
			}
			require.NoError(t, mem.Add(msg, false))
		}

		// Verify all messages were stored
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		assert.Len(t, msgs, len(types))

		// Verify each type is present
		for i, typ := range types {
			assert.Equal(t, typ, msgs[i].Type)
		}
	})
}
