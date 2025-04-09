package serialization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adimarco/hive/internal/llm"
)

func TestSerializedMessage_Conversion(t *testing.T) {
	t.Run("basic text message", func(t *testing.T) {
		// Create a serialized message
		sm := SerializedMessage{
			Role: "user",
			Content: []SerializedContent{
				NewTextContent("hello world"),
			},
			Name: "test-user",
			Metadata: map[string]interface{}{
				"key": "value",
			},
		}

		// Convert to core Message
		msg, err := sm.ToMessage()
		require.NoError(t, err)

		// Verify conversion
		assert.Equal(t, llm.MessageTypeUser, msg.Type)
		assert.Equal(t, "hello world", msg.Content)
		assert.Equal(t, "test-user", msg.Name)
		assert.Equal(t, "value", msg.Metadata["key"])

		// Convert back to SerializedMessage
		sm2 := FromMessage(msg)

		// Verify round-trip
		assert.Equal(t, "user", sm2.Role)
		assert.Equal(t, "hello world", sm2.Content[0].Text)
		assert.Equal(t, "test-user", sm2.Name)
		assert.Equal(t, "value", sm2.Metadata["key"])
	})

	t.Run("multipart message", func(t *testing.T) {
		// Create a serialized message with multiple parts
		sm := SerializedMessage{
			Role: "assistant",
			Content: []SerializedContent{
				NewTextContent("main content"),
				NewTextContent("additional text"),
				NewImageContent("image.jpg", map[string]interface{}{
					"width":  float64(800),
					"height": float64(600),
				}),
				NewResourceContent("data.json", map[string]interface{}{
					"type": "application/json",
				}),
			},
		}

		// Convert to core Message
		msg, err := sm.ToMessage()
		require.NoError(t, err)

		// Verify conversion
		assert.Equal(t, llm.MessageTypeAssistant, msg.Type)
		assert.Equal(t, "main content", msg.Content)
		require.Len(t, msg.Parts, 3)
		assert.Equal(t, "text", msg.Parts[0].Type)
		assert.Equal(t, "additional text", msg.Parts[0].Content)
		assert.Equal(t, "image", msg.Parts[1].Type)
		assert.Equal(t, "image.jpg", msg.Parts[1].Content)
		assert.Equal(t, "resource", msg.Parts[2].Type)
		assert.Equal(t, "data.json", msg.Parts[2].Content)

		// Convert back to SerializedMessage
		sm2 := FromMessage(msg)

		// Verify round-trip
		assert.Equal(t, "assistant", sm2.Role)
		require.Len(t, sm2.Content, 4)
		assert.Equal(t, "main content", sm2.Content[0].Text)
		assert.Equal(t, "additional text", sm2.Content[1].Text)
		assert.Equal(t, "image.jpg", sm2.Content[2].Path)
		assert.Equal(t, float64(800), sm2.Content[2].Data["width"])
		assert.Equal(t, "data.json", sm2.Content[3].Path)
	})

	t.Run("invalid role", func(t *testing.T) {
		sm := SerializedMessage{
			Role: "invalid",
			Content: []SerializedContent{
				NewTextContent("test"),
			},
		}

		_, err := sm.ToMessage()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid message role")
	})
}

func TestContentHelpers(t *testing.T) {
	t.Run("text content", func(t *testing.T) {
		content := NewTextContent("hello")
		assert.Equal(t, ContentTypeText, content.Type)
		assert.Equal(t, "hello", content.Text)
	})

	t.Run("image content", func(t *testing.T) {
		data := map[string]interface{}{"width": float64(100)}
		content := NewImageContent("test.jpg", data)
		assert.Equal(t, ContentTypeImage, content.Type)
		assert.Equal(t, "test.jpg", content.Path)
		assert.Equal(t, data, content.Data)
	})

	t.Run("resource content", func(t *testing.T) {
		data := map[string]interface{}{"type": "text/plain"}
		content := NewResourceContent("test.txt", data)
		assert.Equal(t, ContentTypeResource, content.Type)
		assert.Equal(t, "test.txt", content.Path)
		assert.Equal(t, data, content.Data)
	})
}

func TestIsImagePath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.jpg", true},
		{"test.jpeg", true},
		{"test.png", true},
		{"test.gif", true},
		{"test.webp", true},
		{"test.txt", false},
		{"test", false},
		{"test.jpg.txt", false},
		{".jpg", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsImagePath(tt.path))
		})
	}
}

func TestConvertMessages(t *testing.T) {
	messages := []llm.Message{
		{
			Type:    llm.MessageTypeUser,
			Content: "hello",
		},
		{
			Type:    llm.MessageTypeAssistant,
			Content: "hi there",
			Parts: []llm.MessagePart{
				{Type: "text", Content: "additional"},
			},
		},
	}

	// Convert to serialized format
	serialized := ConvertMessages(messages)
	require.Len(t, serialized, 2)

	// Convert back to core messages
	converted, err := ConvertToMessages(serialized)
	require.NoError(t, err)
	require.Len(t, converted, 2)

	// Verify round-trip
	assert.Equal(t, messages[0].Type, converted[0].Type)
	assert.Equal(t, messages[0].Content, converted[0].Content)
	assert.Equal(t, messages[1].Type, converted[1].Type)
	assert.Equal(t, messages[1].Content, converted[1].Content)
	assert.Equal(t, messages[1].Parts[0].Content, converted[1].Parts[0].Content)
}
