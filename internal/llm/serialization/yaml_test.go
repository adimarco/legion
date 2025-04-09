package serialization

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gofast/internal/llm"
)

func TestYAMLSerialization(t *testing.T) {
	t.Run("conversation round trip", func(t *testing.T) {
		// Create a test conversation
		conv := &SerializedConversation{
			Name:        "Test Conversation",
			Description: "A test conversation",
			Messages: []SerializedMessage{
				{
					Role: "user",
					Content: []SerializedContent{
						NewTextContent("hello"),
					},
				},
				{
					Role: "assistant",
					Content: []SerializedContent{
						NewTextContent("hi there"),
						NewImageContent("test.jpg", map[string]interface{}{
							"width": 100,
						}),
					},
				},
			},
			Metadata: map[string]interface{}{
				"test": true,
			},
		}

		// Create a temporary directory for test files
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "conversation.yaml")

		// Save the conversation
		err := SaveConversation(conv, path)
		require.NoError(t, err)

		// Load the conversation
		loaded, err := LoadConversation(path)
		require.NoError(t, err)

		// Verify the loaded conversation matches the original
		assert.Equal(t, conv.Name, loaded.Name)
		assert.Equal(t, conv.Description, loaded.Description)
		assert.Equal(t, len(conv.Messages), len(loaded.Messages))
		assert.Equal(t, conv.Metadata["test"], loaded.Metadata["test"])

		// Verify message content
		assert.Equal(t, "hello", loaded.Messages[0].Content[0].Text)
		assert.Equal(t, "hi there", loaded.Messages[1].Content[0].Text)
		assert.Equal(t, "test.jpg", loaded.Messages[1].Content[1].Path)
	})

	t.Run("messages round trip", func(t *testing.T) {
		// Create test messages
		messages := []SerializedMessage{
			{
				Role: "user",
				Content: []SerializedContent{
					NewTextContent("test message"),
				},
			},
			{
				Role: "assistant",
				Content: []SerializedContent{
					NewTextContent("response"),
				},
				Metadata: map[string]interface{}{
					"confidence": 0.9,
				},
			},
		}

		// Create a temporary directory for test files
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "messages.yaml")

		// Save the messages
		err := SaveMessages(messages, path)
		require.NoError(t, err)

		// Load the messages
		loaded, err := LoadMessages(path)
		require.NoError(t, err)

		// Verify the loaded messages match the original
		assert.Equal(t, len(messages), len(loaded))
		assert.Equal(t, messages[0].Role, loaded[0].Role)
		assert.Equal(t, messages[0].Content[0].Text, loaded[0].Content[0].Text)
		assert.Equal(t, messages[1].Metadata["confidence"], loaded[1].Metadata["confidence"])
	})

	t.Run("write to buffer", func(t *testing.T) {
		// Create test data
		messages := []SerializedMessage{
			{
				Role: "user",
				Content: []SerializedContent{
					NewTextContent("hello"),
				},
			},
		}

		conv := &SerializedConversation{
			Name:     "Test",
			Messages: messages,
		}

		// Test WriteConversation
		var buf bytes.Buffer
		err := WriteConversation(&buf, conv)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "name: Test")
		assert.Contains(t, buf.String(), "hello")

		// Test WriteMessages
		buf.Reset()
		err = WriteMessages(&buf, messages)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "role: user")
		assert.Contains(t, buf.String(), "hello")
	})

	t.Run("error handling", func(t *testing.T) {
		tmpDir := t.TempDir()

		t.Run("invalid yaml file", func(t *testing.T) {
			path := filepath.Join(tmpDir, "invalid.yaml")
			err := os.WriteFile(path, []byte("invalid: ][yaml"), 0644)
			require.NoError(t, err)

			_, err = LoadConversation(path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to unmarshal YAML")

			_, err = LoadMessages(path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to unmarshal YAML")
		})

		t.Run("nonexistent file", func(t *testing.T) {
			path := filepath.Join(tmpDir, "nonexistent.yaml")

			_, err := LoadConversation(path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to read file")

			_, err = LoadMessages(path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to read file")
		})

		t.Run("invalid directory", func(t *testing.T) {
			path := filepath.Join(os.DevNull, "test.yaml")

			err := SaveConversation(&SerializedConversation{}, path)
			assert.Error(t, err)

			err = SaveMessages(nil, path)
			assert.Error(t, err)
		})
	})
}

func TestConversionHelpers(t *testing.T) {
	t.Run("convert core messages", func(t *testing.T) {
		messages := []llm.Message{
			{
				Type:    llm.MessageTypeUser,
				Content: "hello",
				Name:    "user1",
			},
			{
				Type:    llm.MessageTypeAssistant,
				Content: "hi",
				Parts: []llm.MessagePart{
					{Type: "text", Content: "extra"},
				},
			},
		}

		// Convert to serialized format
		conv := NewConversation("Test", "Description", messages)
		assert.Equal(t, "Test", conv.Name)
		assert.Equal(t, "Description", conv.Description)
		assert.Len(t, conv.Messages, 2)

		// Convert back to core messages
		converted, err := conv.ToMessages()
		require.NoError(t, err)
		assert.Len(t, converted, 2)
		assert.Equal(t, messages[0].Content, converted[0].Content)
		assert.Equal(t, messages[1].Parts[0].Content, converted[1].Parts[0].Content)
	})
}
