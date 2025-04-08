package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gofast/internal/config"
)

const (
	// CALL_TOOL_INDICATOR prefixes a message that should trigger a tool call
	CALL_TOOL_INDICATOR = "***CALL_TOOL"
	// FIXED_RESPONSE_INDICATOR sets a fixed response for all subsequent calls
	FIXED_RESPONSE_INDICATOR = "***FIXED_RESPONSE"
)

// PassthroughLLM is a simple implementation that just echoes messages back
// Useful for testing and development without making actual API calls
type PassthroughLLM struct {
	name          string
	memory        Memory
	cfg           *config.Settings
	defaults      *RequestParams
	fixedResponse string // if set, always return this response
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

	// Check for special commands
	if strings.HasPrefix(msg.Content, FIXED_RESPONSE_INDICATOR) {
		parts := strings.SplitN(msg.Content, FIXED_RESPONSE_INDICATOR, 2)
		if len(parts) > 1 {
			l.fixedResponse = strings.TrimSpace(parts[1])
		}
	}

	if strings.HasPrefix(msg.Content, CALL_TOOL_INDICATOR) {
		toolName, args, err := l.parseToolCall(msg.Content)
		if err != nil {
			return Message{}, fmt.Errorf("failed to parse tool call: %w", err)
		}
		result, err := l.CallTool(ctx, ToolCall{
			Name: toolName,
			Args: args,
		})
		if err != nil {
			return Message{}, fmt.Errorf("failed to call tool: %w", err)
		}
		return Message{
			Type:    MessageTypeAssistant,
			Content: result,
			Name:    l.name,
		}, nil
	}

	// Create response message
	var content string
	if l.fixedResponse != "" {
		content = l.fixedResponse
	} else {
		content = msg.Content
	}

	response := Message{
		Type:    MessageTypeAssistant,
		Content: content,
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

// parseToolCall parses a tool call message into name and arguments
func (l *PassthroughLLM) parseToolCall(content string) (name string, args map[string]any, err error) {
	// Remove the indicator
	content = strings.TrimPrefix(content, CALL_TOOL_INDICATOR)
	content = strings.TrimSpace(content)

	// Split into name and args
	parts := strings.SplitN(content, " ", 2)
	name = parts[0]

	// Parse args if present
	if len(parts) > 1 {
		if err := json.Unmarshal([]byte(parts[1]), &args); err != nil {
			return "", nil, fmt.Errorf("invalid JSON arguments: %w", err)
		}
	}

	return name, args, nil
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
