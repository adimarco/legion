package serialization

import (
	"fmt"
	"path/filepath"

	"gofast/internal/llm"
)

// ContentType represents the type of content in a message part
type ContentType string

const (
	// ContentTypeText represents plain text content
	ContentTypeText ContentType = "text"
	// ContentTypeImage represents image data
	ContentTypeImage ContentType = "image"
	// ContentTypeResource represents an embedded resource (file, etc.)
	ContentTypeResource ContentType = "resource"
)

// SerializedContent represents a single piece of content in a message.
// This maps to the MCPContentType in Python but with Go idioms.
type SerializedContent struct {
	// Type indicates what kind of content this is
	Type ContentType `yaml:"type" json:"type"`
	// Text holds the text content if Type is ContentTypeText
	Text string `yaml:"text,omitempty" json:"text,omitempty"`
	// Data holds binary data or structured data for other types
	Data map[string]interface{} `yaml:"data,omitempty" json:"data,omitempty"`
	// Path holds a file path for resource types
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// SerializedMessage represents a message that can be serialized to/from YAML/JSON.
// This provides a more flexible format than the core Message type, supporting
// multiple content parts and additional metadata.
type SerializedMessage struct {
	// Role is the message role (user, assistant, system, tool)
	Role string `yaml:"role" json:"role"`
	// Content holds one or more content parts
	Content []SerializedContent `yaml:"content" json:"content"`
	// Name optionally identifies the sender
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Metadata allows for additional structured data
	Metadata map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// SerializedConversation represents a sequence of messages that can be
// saved/loaded as a unit. This is useful for test scenarios and history.
type SerializedConversation struct {
	// Name identifies this conversation
	Name string `yaml:"name" json:"name"`
	// Description explains the purpose/content
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Messages contains the conversation messages in order
	Messages []SerializedMessage `yaml:"messages" json:"messages"`
	// Metadata allows for additional structured data
	Metadata map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// ToMessage converts a SerializedMessage to a core Message.
// This handles converting the flexible serialized format to the
// more structured core type.
func (sm *SerializedMessage) ToMessage() (llm.Message, error) {
	msg := llm.Message{
		Name:     sm.Name,
		Metadata: sm.Metadata,
	}

	// Convert role to MessageType
	switch sm.Role {
	case "user":
		msg.Type = llm.MessageTypeUser
	case "assistant":
		msg.Type = llm.MessageTypeAssistant
	case "system":
		msg.Type = llm.MessageTypeSystem
	case "tool":
		msg.Type = llm.MessageTypeTool
	default:
		return msg, fmt.Errorf("invalid message role: %s", sm.Role)
	}

	// Convert content parts
	for _, content := range sm.Content {
		switch content.Type {
		case ContentTypeText:
			if msg.Content == "" {
				msg.Content = content.Text
			} else {
				// Additional text parts become MessageParts
				msg.Parts = append(msg.Parts, llm.MessagePart{
					Type:    string(content.Type),
					Content: content.Text,
				})
			}
		case ContentTypeImage, ContentTypeResource:
			msg.Parts = append(msg.Parts, llm.MessagePart{
				Type:    string(content.Type),
				Content: content.Path,
				Data:    content.Data,
			})
		}
	}

	return msg, nil
}

// FromMessage converts a core Message to a SerializedMessage.
// This expands the core type into the more flexible serialized format.
func FromMessage(msg llm.Message) SerializedMessage {
	sm := SerializedMessage{
		Name:     msg.Name,
		Metadata: msg.Metadata,
	}

	// Convert MessageType to role
	switch msg.Type {
	case llm.MessageTypeUser:
		sm.Role = "user"
	case llm.MessageTypeAssistant:
		sm.Role = "assistant"
	case llm.MessageTypeSystem:
		sm.Role = "system"
	case llm.MessageTypeTool:
		sm.Role = "tool"
	}

	// Convert main content to first text part
	if msg.Content != "" {
		sm.Content = append(sm.Content, SerializedContent{
			Type: ContentTypeText,
			Text: msg.Content,
		})
	}

	// Convert additional parts
	for _, part := range msg.Parts {
		content := SerializedContent{
			Type: ContentType(part.Type),
			Data: part.Data,
		}
		switch ContentType(part.Type) {
		case ContentTypeText:
			content.Text = part.Content
		case ContentTypeImage, ContentTypeResource:
			content.Path = part.Content
		}
		sm.Content = append(sm.Content, content)
	}

	return sm
}

// NewTextContent creates a SerializedContent for text.
func NewTextContent(text string) SerializedContent {
	return SerializedContent{
		Type: ContentTypeText,
		Text: text,
	}
}

// NewImageContent creates a SerializedContent for an image.
func NewImageContent(path string, data map[string]interface{}) SerializedContent {
	return SerializedContent{
		Type: ContentTypeImage,
		Path: path,
		Data: data,
	}
}

// NewResourceContent creates a SerializedContent for a resource.
func NewResourceContent(path string, data map[string]interface{}) SerializedContent {
	return SerializedContent{
		Type: ContentTypeResource,
		Path: path,
		Data: data,
	}
}

// IsImagePath returns true if the path appears to be an image file.
func IsImagePath(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}
