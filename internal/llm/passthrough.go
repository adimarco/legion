package llm

import (
	"context"
	"fmt"

	"gofast/internal/config"
)

// PassthroughLLM is a simple implementation that just echoes messages back
// Useful for testing and development without making actual API calls
type PassthroughLLM struct {
	name     string
	memory   Memory
	cfg      *config.Settings
	defaults *RequestParams
}

// NewPassthroughLLM creates a new PassthroughLLM instance
func NewPassthroughLLM(name string) *PassthroughLLM {
	return &PassthroughLLM{
		name:   name,
		memory: NewSimpleMemory(),
		defaults: &RequestParams{
			Model:         "passthrough",
			UseHistory:    true,
			ParallelTools: true,
			MaxIterations: 10,
		},
	}
}

// Initialize sets up the LLM with configuration
func (l *PassthroughLLM) Initialize(ctx context.Context, cfg *config.Settings) error {
	l.cfg = cfg
	return nil
}

// Generate processes a message and returns a response
func (l *PassthroughLLM) Generate(ctx context.Context, msg Message, params *RequestParams) (Message, error) {
	// Store the user message in history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(msg, false); err != nil {
			return Message{}, fmt.Errorf("failed to add message to history: %w", err)
		}
	}

	// Create response message
	response := Message{
		Type:    MessageTypeAssistant,
		Content: msg.Content, // Simply echo back the content
		Name:    l.name,
	}

	// Store the response in history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(response, false); err != nil {
			return Message{}, fmt.Errorf("failed to add response to history: %w", err)
		}
	}

	return response, nil
}

// GenerateString is a convenience method that returns just the content string
func (l *PassthroughLLM) GenerateString(ctx context.Context, content string, params *RequestParams) (string, error) {
	msg := Message{
		Type:    MessageTypeUser,
		Content: content,
	}
	response, err := l.Generate(ctx, msg, params)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// CallTool executes a tool call and returns the result
func (l *PassthroughLLM) CallTool(ctx context.Context, call ToolCall) (string, error) {
	// For passthrough, just return a message about the tool call
	return fmt.Sprintf("Tool call: %s(%v)", call.Name, call.Args), nil
}

// Name returns the identifier for this LLM instance
func (l *PassthroughLLM) Name() string {
	return l.name
}

// Provider returns the LLM provider (e.g., "anthropic", "openai")
func (l *PassthroughLLM) Provider() string {
	return "passthrough"
}

// Cleanup performs any necessary cleanup
func (l *PassthroughLLM) Cleanup() error {
	return l.memory.Clear(true)
}
