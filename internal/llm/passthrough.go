/*
Package llm provides testing utilities for LLM interactions.

The PassthroughLLM is a testing implementation that provides predictable,
deterministic behavior without external dependencies. It supports:

1. Message Echo: By default, simply returns the input message
2. Tool Calls: Parses and formats tool calls for testing
3. Fixed Responses: Can be configured to always return specific responses
4. History Management: Tracks conversation history like a real LLM

This implementation is particularly useful for:
- Unit testing agent logic without API calls
- Development and debugging of tool integration
- Testing conversation flow and history management
- Simulating error conditions and edge cases
*/
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gofast/internal/config"
	"gofast/internal/logging"
)

const (
	// ToolCallPrefix prefixes a message that should trigger a tool call.
	// Format: ***CALL_TOOL <tool_name> <json_args>
	// Example: ***CALL_TOOL search {"query": "golang"}
	ToolCallPrefix = "***CALL_TOOL"

	// FixedResponsePrefix sets a fixed response for all subsequent calls.
	// Format: ***FIXED_RESPONSE <response_text>
	// Example: ***FIXED_RESPONSE I will always say this
	FixedResponsePrefix = "***FIXED_RESPONSE"
)

// PassthroughLLM is a simple implementation that just echoes messages back.
// It's designed for testing and development, providing predictable behavior
// without making actual API calls. This implementation helps test:
// - Basic message handling
// - Tool call parsing and execution
// - History management
// - Error conditions
type PassthroughLLM struct {
	name          string           // Identifier for this instance
	memory        Memory           // Conversation history storage
	cfg           *config.Settings // Configuration
	defaults      *RequestParams   // Default parameters
	fixedResponse string           // Optional fixed response
	logger        logging.Logger   // Structured logging
}

// NewPassthroughLLM creates a new PassthroughLLM instance.
// The name parameter is used to identify this instance in logs
// and responses, making it easier to track in multi-LLM scenarios.
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
		logger: logging.GetLogger("llm.passthrough"),
	}
}

// Initialize sets up the LLM with configuration.
// For PassthroughLLM, this is minimal since we don't need
// API keys or complex setup. We just store the config for
// potential future use.
func (l *PassthroughLLM) Initialize(ctx context.Context, cfg *config.Settings) error {
	l.cfg = cfg
	return nil
}

// Generate processes a message and returns a response.
// The behavior depends on the message content:
// 1. Normal messages: Echo back the content
// 2. Tool calls (***CALL_TOOL): Parse and format tool response
// 3. Fixed response commands: Set or use fixed response
//
// This method also handles:
// - History management (if UseHistory is true)
// - Logging of all operations
// - Special command processing
func (l *PassthroughLLM) Generate(ctx context.Context, msg Message, params *RequestParams) (Message, error) {
	// Log user message
	l.logger.Info(ctx, "Received user message", logging.WithData(map[string]interface{}{
		"content": msg.Content,
		"type":    msg.Type,
	}))

	// Store the user message in history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(msg, false); err != nil {
			return Message{}, fmt.Errorf("failed to add message to history: %w", err)
		}
	}

	// Check for special commands
	if strings.HasPrefix(msg.Content, FixedResponsePrefix) {
		parts := strings.SplitN(msg.Content, FixedResponsePrefix, 2)
		if len(parts) > 1 {
			l.fixedResponse = strings.TrimSpace(parts[1])
			l.logger.Debug(ctx, "Set fixed response", logging.WithData(map[string]interface{}{
				"response": l.fixedResponse,
			}))
		}
	}

	if strings.HasPrefix(msg.Content, ToolCallPrefix) {
		toolName, args, err := l.parseToolCall(msg.Content)
		if err != nil {
			l.logger.Error(ctx, "Failed to parse tool call", logging.WithData(map[string]interface{}{
				"error": err.Error(),
			}))
			return Message{}, fmt.Errorf("failed to parse tool call: %w", err)
		}

		l.logger.Debug(ctx, "Calling tool", logging.WithData(map[string]interface{}{
			"tool":      toolName,
			"arguments": args,
		}))

		result, err := l.CallTool(ctx, ToolCall{
			Name: toolName,
			Args: args,
		})
		if err != nil {
			l.logger.Error(ctx, "Failed to call tool", logging.WithData(map[string]interface{}{
				"error": err.Error(),
			}))
			return Message{}, fmt.Errorf("failed to call tool: %w", err)
		}

		response := Message{
			Type:    MessageTypeAssistant,
			Content: result,
			Name:    l.name,
		}

		l.logger.Info(ctx, "Tool call completed", logging.WithData(map[string]interface{}{
			"result": result,
		}))

		return response, nil
	}

	// Create response message
	var content string
	if l.fixedResponse != "" {
		content = l.fixedResponse
	} else {
		// Handle multipart messages by concatenating all parts
		content = msg.GetAllText()
	}

	response := Message{
		Type:    MessageTypeAssistant,
		Content: content,
		Name:    l.name,
	}

	// Log assistant response
	l.logger.Info(ctx, "Generated response", logging.WithData(map[string]interface{}{
		"content": response.Content,
		"type":    response.Type,
	}))

	// Store the response in history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(response, false); err != nil {
			return Message{}, fmt.Errorf("failed to add response to history: %w", err)
		}
	}

	return response, nil
}

// GenerateString is a convenience method that returns just the content string.
// This is useful for simple testing scenarios where you don't need
// the full Message structure.
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

// parseToolCall parses a tool call message into name and arguments.
// Format: ***CALL_TOOL <tool_name> <json_args>
// The JSON arguments are optional. If present, they must be valid JSON.
func (l *PassthroughLLM) parseToolCall(content string) (name string, args map[string]any, err error) {
	// Remove the indicator
	content = strings.TrimPrefix(content, ToolCallPrefix)
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

// CallTool executes a tool call and returns the result.
// For PassthroughLLM, this formats the call details into a readable
// string, which is useful for testing tool integration without
// actually executing the tools.
func (l *PassthroughLLM) CallTool(ctx context.Context, call ToolCall) (string, error) {
	// For passthrough, format the tool call result in a more detailed way
	if call.Response != "" {
		return fmt.Sprintf("Tool '%s' result: %s", call.Name, call.Response), nil
	}

	args := "no arguments"
	if len(call.Args) > 0 {
		argsJSON, err := json.MarshalIndent(call.Args, "", "  ")
		if err != nil {
			args = fmt.Sprintf("error formatting arguments: %v", err)
		} else {
			args = string(argsJSON)
		}
	}

	return fmt.Sprintf("Tool call: %s\nArguments:\n%s", call.Name, args), nil
}

// Name returns the identifier for this LLM instance
func (l *PassthroughLLM) Name() string {
	return l.name
}

// Provider returns the LLM provider (e.g., "anthropic", "openai")
func (l *PassthroughLLM) Provider() string {
	return "passthrough"
}

// Cleanup performs any necessary cleanup.
// For PassthroughLLM, this just clears the memory.
func (l *PassthroughLLM) Cleanup() error {
	return l.memory.Clear(true)
}
