/*
Package llm provides the core abstractions and implementations for Large Language Model (LLM) interactions.

Design Philosophy:
The package is built around several key principles:

 1. Provider Agnosticism: The core interfaces are designed to be provider-agnostic,
    allowing seamless integration of different LLM providers (e.g., Anthropic, OpenAI)
    without changing the consuming code.

 2. Testing First: The package includes robust testing tools (PassthroughLLM, PlaybackLLM)
    that enable thorough testing of agent logic without making actual API calls.

 3. Flexible Memory Management: The Memory interface allows for different memory
    implementations while maintaining a consistent interface for history management.

 4. Tool Integration: The design supports tool calls and responses as first-class
    citizens, enabling complex agent behaviors and workflows.

Usage:
Most users will interact with this package through the AugmentedLLM interface,
which provides high-level operations for generating responses and managing context.
For testing, the PassthroughLLM and PlaybackLLM implementations provide
deterministic behaviors without external dependencies.
*/
package llm

import (
	"context"
	"strings"

	"gofast/internal/config"
)

// MessageType represents the role of a message in a conversation.
// This follows the common pattern used by major LLM providers where
// messages have distinct roles in the conversation.
type MessageType string

const (
	// MessageTypeSystem represents system-level instructions or context
	MessageTypeSystem MessageType = "system"
	// MessageTypeUser represents messages from the user/human
	MessageTypeUser MessageType = "user"
	// MessageTypeAssistant represents messages from the AI assistant
	MessageTypeAssistant MessageType = "assistant"
	// MessageTypeTool represents messages related to tool operations
	MessageTypeTool MessageType = "tool"
)

// Message represents a generic message in a conversation.
// The design is intentionally provider-agnostic, allowing for conversion
// to provider-specific formats as needed. It supports both simple text
// messages and complex interactions like tool calls.
type Message struct {
	// Type indicates the role of this message in the conversation
	Type MessageType `json:"type"`
	// Content holds the primary text content of the message
	Content string `json:"content"`
	// Name optionally identifies the sender (useful for tool calls)
	Name string `json:"name,omitempty"`
	// ToolCalls holds any tool operations requested by this message
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// Metadata allows for provider-specific or custom data
	Metadata map[string]any `json:"metadata,omitempty"`
	// Parts supports multipart messages (e.g., text + images)
	Parts []MessagePart `json:"parts,omitempty"`
}

// MessagePart represents a part of a multipart message.
// This abstraction allows for future support of multi-modal
// interactions (text, images, audio, etc.) while maintaining
// backward compatibility.
type MessagePart struct {
	// Type indicates the content type of this part
	Type string `json:"type"`
	// Content holds the actual content
	Content string `json:"content"`
	// Data holds type-specific metadata
	Data map[string]any `json:"data,omitempty"`
}

// GetAllText returns all text content from a message, including parts.
// This is particularly useful when you need to process all text content
// regardless of its location in the message structure.
func (m *Message) GetAllText() string {
	texts := []string{m.Content}
	for _, part := range m.Parts {
		if part.Content != "" {
			texts = append(texts, part.Content)
		}
	}
	return strings.Join(texts, "\n")
}

// ToolCall represents a request to call a tool.
// The design supports both synchronous and asynchronous tool execution,
// with the Response field allowing for result storage.
type ToolCall struct {
	// ID uniquely identifies this tool call
	ID string `json:"id"`
	// Name identifies which tool to call
	Name string `json:"name"`
	// Args holds the parameters for the tool call
	Args map[string]any `json:"args"`
	// Response stores the result of the tool call
	Response string `json:"response,omitempty"`
}

// RequestParams holds parameters for an LLM request
type RequestParams struct {
	SystemPrompt  string         // System prompt to use
	Model         string         // Model to use
	Temperature   float32        // Temperature for sampling
	MaxTokens     int            // Maximum tokens to generate
	UseHistory    bool           // Whether to include conversation history
	ParallelTools bool           // Whether to run tools in parallel
	MaxIterations int            // Maximum number of tool call iterations
	Tools         []string       // Required MCP tools
	Config        map[string]any // Additional configuration
}

// AugmentedLLM represents an LLM enhanced with tools, memory, and context management.
// This is the primary interface for interacting with LLMs in the system.
// The interface is designed to be:
// 1. Provider-agnostic: Works with any LLM provider
// 2. Context-aware: All operations receive a context.Context
// 3. Tool-capable: Supports tool calls and responses
// 4. Memory-enabled: Can maintain conversation history
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

// Provider represents an LLM provider (e.g., Anthropic, OpenAI).
// This abstraction allows for:
// 1. Lazy initialization of provider resources
// 2. Provider-specific configuration handling
// 3. Factory pattern for creating LLM instances
type Provider interface {
	// Initialize sets up the provider with configuration
	Initialize(ctx context.Context, cfg *config.Settings) error

	// CreateLLM creates a new LLM instance
	CreateLLM(name string, params *RequestParams) (AugmentedLLM, error)

	// Name returns the provider name
	Name() string
}
