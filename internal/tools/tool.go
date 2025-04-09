package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Tool represents a callable tool with metadata and execution capabilities
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"` // JSON Schema for input validation
	ArgsSchema  json.RawMessage `json:"parameters"`
	Handler     ToolHandler     `json:"-"`    // Not serialized
	Cost        uint64          `json:"cost"` // Credits per use (future monetization)
}

// ToolHandler defines the function signature for tool execution
type ToolHandler func(ctx context.Context, args map[string]any) (ToolResult, error)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content  string         `json:"content"`  // Text content of the result
	IsError  bool           `json:"is_error"` // Whether this result represents an error
	Metadata map[string]any `json:"metadata"` // Additional result data
}

// ToolRegistry manages tool registration and execution
type ToolRegistry interface {
	// Register adds a tool to the registry
	Register(tool Tool) error

	// Get retrieves a tool by name
	Get(name string) (Tool, error)

	// List returns all registered tools
	List() []Tool

	// Call executes a tool by name
	Call(ctx context.Context, name string, args map[string]any) (ToolResult, error)
}

// ToolProvider defines an interface for systems that can provide tools
type ToolProvider interface {
	// GetTools returns the tools provided by this provider
	GetTools() []Tool
}

// SimpleToolRegistry provides a basic thread-safe implementation of ToolRegistry
type SimpleToolRegistry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewSimpleToolRegistry creates a new SimpleToolRegistry
func NewSimpleToolRegistry() *SimpleToolRegistry {
	return &SimpleToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register implements ToolRegistry.Register
func (r *SimpleToolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool %q already registered", tool.Name)
	}

	// Validate the tool
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool handler is required")
	}

	r.tools[tool.Name] = tool
	return nil
}

// Get implements ToolRegistry.Get
func (r *SimpleToolRegistry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return Tool{}, fmt.Errorf("tool %q not found", name)
	}
	return tool, nil
}

// List implements ToolRegistry.List
func (r *SimpleToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Call implements ToolRegistry.Call
func (r *SimpleToolRegistry) Call(ctx context.Context, name string, args map[string]any) (ToolResult, error) {
	tool, err := r.Get(name)
	if err != nil {
		return ToolResult{}, err
	}

	// TODO: Add input validation against schema
	// TODO: Add cost tracking

	return tool.Handler(ctx, args)
}

// NewToolResult creates a successful tool result
func NewToolResult(content string) ToolResult {
	return ToolResult{
		Content: content,
		IsError: false,
	}
}

// NewErrorResult creates an error tool result
func NewErrorResult(err error) ToolResult {
	return ToolResult{
		Content: err.Error(),
		IsError: true,
	}
}

// GetArgsSchema returns the JSON schema for validating tool arguments
func (t *Tool) GetArgsSchema() json.RawMessage {
	return t.InputSchema
}
