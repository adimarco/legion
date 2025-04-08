package llm

import (
	"fmt"
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
		done := make(chan bool)

		// Run concurrent operations
		for i := 0; i < 10; i++ {
			go func(i int) {
				msg := Message{
					Type:    MessageTypeUser,
					Content: fmt.Sprintf("message %d", i),
				}
				_ = mem.Add(msg, false)
				_, _ = mem.Get(true)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify we can still operate on the memory
		msgs, err := mem.Get(true)
		require.NoError(t, err)
		assert.Len(t, msgs, 10)
	})
}
