package llm

import (
	"context"

	"gofast/internal/config"
)

// MessageType represents the role of a message in a conversation
type MessageType string

const (
	MessageTypeSystem    MessageType = "system"
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeTool      MessageType = "tool"
)

// Message represents a generic message in a conversation
type Message struct {
	Type      MessageType    `json:"type"`
	Content   string         `json:"content"`
	Name      string         `json:"name,omitempty"`
	ToolCalls []ToolCall     `json:"tool_calls,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ToolCall represents a request to call a tool
type ToolCall struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Args     map[string]any `json:"args"`
	Response string         `json:"response,omitempty"`
}

// RequestParams configures how the LLM should process the request
type RequestParams struct {
	Model         string         `json:"model"`
	SystemPrompt  string         `json:"system_prompt,omitempty"`
	MaxTokens     int            `json:"max_tokens,omitempty"`
	Temperature   float32        `json:"temperature,omitempty"`
	StopSequences []string       `json:"stop_sequences,omitempty"`
	UseHistory    bool           `json:"use_history"`
	ParallelTools bool           `json:"parallel_tools"`
	MaxIterations int            `json:"max_iterations"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// AugmentedLLM represents an LLM enhanced with tools, memory, and context management
type AugmentedLLM interface {
	// Initialize sets up the LLM with configuration
	Initialize(ctx context.Context, cfg *config.Settings) error

	// Generate processes a message and returns a response
	Generate(ctx context.Context, msg Message, params *RequestParams) (Message, error)

	// GenerateString is a convenience method that returns just the content string
	GenerateString(ctx context.Context, content string, params *RequestParams) (string, error)

	// CallTool executes a tool call and returns the result
	CallTool(ctx context.Context, call ToolCall) (string, error)

	// Name returns the identifier for this LLM instance
	Name() string

	// Provider returns the LLM provider (e.g., "anthropic", "openai")
	Provider() string

	// Cleanup performs any necessary cleanup
	Cleanup() error
}

// Memory manages conversation history and prompt storage
type Memory interface {
	// Add adds a message to history
	Add(msg Message, isPrompt bool) error

	// Get retrieves messages from memory
	Get(includeHistory bool) ([]Message, error)

	// Clear clears the specified message types
	Clear(clearPrompts bool) error
}

// Provider represents an LLM provider (e.g., Anthropic, OpenAI)
type Provider interface {
	// Initialize sets up the provider with configuration
	Initialize(ctx context.Context, cfg *config.Settings) error

	// CreateLLM creates a new LLM instance
	CreateLLM(name string, params *RequestParams) (AugmentedLLM, error)

	// Name returns the provider name
	Name() string
}
