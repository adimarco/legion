package serialization

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"gofast/internal/llm"
)

// LoadConversation loads a conversation from a YAML file.
// The file should contain a SerializedConversation.
func LoadConversation(path string) (*SerializedConversation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var conv SerializedConversation
	if err := yaml.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &conv, nil
}

// SaveConversation saves a conversation to a YAML file.
func SaveConversation(conv *SerializedConversation, path string) error {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(conv)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadMessages loads a sequence of messages from a YAML file.
// The file should contain a list of SerializedMessage.
func LoadMessages(path string) ([]SerializedMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var messages []SerializedMessage
	if err := yaml.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return messages, nil
}

// SaveMessages saves a sequence of messages to a YAML file.
func SaveMessages(messages []SerializedMessage, path string) error {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// WriteConversation writes a conversation in YAML format to a writer.
// This is useful for displaying conversations or writing to non-file destinations.
func WriteConversation(w io.Writer, conv *SerializedConversation) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	return encoder.Encode(conv)
}

// WriteMessages writes messages in YAML format to a writer.
func WriteMessages(w io.Writer, messages []SerializedMessage) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	return encoder.Encode(messages)
}

// ConvertMessages converts a slice of core Messages to SerializedMessages.
func ConvertMessages(messages []llm.Message) []SerializedMessage {
	result := make([]SerializedMessage, len(messages))
	for i, msg := range messages {
		result[i] = FromMessage(msg)
	}
	return result
}

// ConvertToMessages converts a slice of SerializedMessages to core Messages.
func ConvertToMessages(messages []SerializedMessage) ([]llm.Message, error) {
	result := make([]llm.Message, len(messages))
	for i, msg := range messages {
		var err error
		result[i], err = msg.ToMessage()
		if err != nil {
			return nil, fmt.Errorf("failed to convert message %d: %w", i, err)
		}
	}
	return result, nil
}

// NewConversation creates a new SerializedConversation with the given messages.
func NewConversation(name string, description string, messages []llm.Message) *SerializedConversation {
	return &SerializedConversation{
		Name:        name,
		Description: description,
		Messages:    ConvertMessages(messages),
	}
}

// ToMessages converts a SerializedConversation to a slice of core Messages.
func (sc *SerializedConversation) ToMessages() ([]llm.Message, error) {
	return ConvertToMessages(sc.Messages)
}
