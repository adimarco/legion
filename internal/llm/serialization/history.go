package serialization

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adimarco/hive/internal/llm"
)

// HistoryMetadata contains metadata about a saved conversation history
type HistoryMetadata struct {
	// Timestamp when the history was saved
	SavedAt time.Time `yaml:"saved_at" json:"saved_at"`
	// Name of the LLM that generated the responses
	LLMName string `yaml:"llm_name,omitempty" json:"llm_name,omitempty"`
	// Provider of the LLM (e.g., "anthropic", "openai")
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
	// Model used for generation (e.g., "gpt-4", "claude-3")
	Model string `yaml:"model,omitempty" json:"model,omitempty"`
	// Additional metadata specific to the conversation
	Custom map[string]interface{} `yaml:"custom,omitempty" json:"custom,omitempty"`
}

// SaveHistory saves a conversation history to a file.
// The history is saved in a structured format that includes:
// - Metadata about when and how it was generated
// - The sequence of messages in the conversation
// - Any additional context or custom metadata
func SaveHistory(memory llm.Memory, path string, metadata *HistoryMetadata) error {
	// Get all messages from memory
	messages, err := memory.Get(true)
	if err != nil {
		return fmt.Errorf("failed to get messages from memory: %w", err)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create conversation with metadata
	conv := &SerializedConversation{
		Name:        filepath.Base(path),
		Description: "Saved conversation history",
		Messages:    ConvertMessages(messages),
		Metadata: map[string]interface{}{
			"saved_at": time.Now(),
		},
	}

	// Add metadata if provided
	if metadata != nil {
		conv.Metadata["llm_name"] = metadata.LLMName
		conv.Metadata["provider"] = metadata.Provider
		conv.Metadata["model"] = metadata.Model
		for k, v := range metadata.Custom {
			conv.Metadata[k] = v
		}
	}

	// Save to file
	return SaveConversation(conv, path)
}

// LoadHistory loads a conversation history from a file and adds it to memory.
// If clearExisting is true, existing messages in memory are cleared first.
// Returns the metadata from the loaded history.
func LoadHistory(memory llm.Memory, path string, clearExisting bool) (*HistoryMetadata, error) {
	// Load the conversation
	conv, err := LoadConversation(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}

	// Clear existing history if requested
	if clearExisting {
		if err := memory.Clear(false); err != nil {
			return nil, fmt.Errorf("failed to clear memory: %w", err)
		}
	}

	// Convert messages
	messages, err := ConvertToMessages(conv.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Add messages to memory
	for _, msg := range messages {
		if err := memory.Add(msg, false); err != nil {
			return nil, fmt.Errorf("failed to add message to memory: %w", err)
		}
	}

	// Extract metadata
	metadata := &HistoryMetadata{
		Custom: make(map[string]interface{}),
	}

	if savedAt, ok := conv.Metadata["saved_at"].(time.Time); ok {
		metadata.SavedAt = savedAt
	}
	if llmName, ok := conv.Metadata["llm_name"].(string); ok {
		metadata.LLMName = llmName
	}
	if provider, ok := conv.Metadata["provider"].(string); ok {
		metadata.Provider = provider
	}
	if model, ok := conv.Metadata["model"].(string); ok {
		metadata.Model = model
	}

	// Copy any other metadata to Custom
	for k, v := range conv.Metadata {
		switch k {
		case "saved_at", "llm_name", "provider", "model":
			continue
		default:
			metadata.Custom[k] = v
		}
	}

	return metadata, nil
}

// SaveHistoryToDir saves a conversation history to a timestamped file in a directory.
// The filename will be in the format: YYYYMMDD_HHMMSS_name.yaml
func SaveHistoryToDir(memory llm.Memory, dir string, name string, metadata *HistoryMetadata) (string, error) {
	// Create timestamp for filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.yaml", timestamp, name)
	path := filepath.Join(dir, filename)

	// Save the history
	if err := SaveHistory(memory, path, metadata); err != nil {
		return "", err
	}

	return path, nil
}

// ListHistoryFiles returns a list of history files in a directory.
// The files are sorted by name (which puts them in chronological order if using timestamped names).
func ListHistoryFiles(dir string) ([]string, error) {
	// List all .yaml files
	pattern := filepath.Join(dir, "*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list history files: %w", err)
	}

	return matches, nil
}

// GetHistoryMetadata loads just the metadata from a history file without loading the messages.
func GetHistoryMetadata(path string) (*HistoryMetadata, error) {
	// Load the conversation
	conv, err := LoadConversation(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}

	// Extract metadata
	metadata := &HistoryMetadata{
		Custom: make(map[string]interface{}),
	}

	if savedAt, ok := conv.Metadata["saved_at"].(time.Time); ok {
		metadata.SavedAt = savedAt
	}
	if llmName, ok := conv.Metadata["llm_name"].(string); ok {
		metadata.LLMName = llmName
	}
	if provider, ok := conv.Metadata["provider"].(string); ok {
		metadata.Provider = provider
	}
	if model, ok := conv.Metadata["model"].(string); ok {
		metadata.Model = model
	}

	// Copy any other metadata to Custom
	for k, v := range conv.Metadata {
		switch k {
		case "saved_at", "llm_name", "provider", "model":
			continue
		default:
			metadata.Custom[k] = v
		}
	}

	return metadata, nil
}
