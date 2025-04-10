package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// Tool represents a callable tool with metadata and execution capabilities
type Tool struct {
	// Core metadata
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`

	// Schema and validation
	Schema json.RawMessage `json:"schema"` // JSON Schema for input validation

	// Execution
	Handler ToolHandler `json:"-"`    // Not serialized
	Cost    uint64      `json:"cost"` // Credits per use

	// Lifecycle hooks (not serialized)
	Initialize func(ctx context.Context) error `json:"-"`
	Cleanup    func(ctx context.Context) error `json:"-"`
}

// ToolHandler defines the function signature for tool execution
type ToolHandler func(ctx context.Context, args map[string]any) (ToolResult, error)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content   string         `json:"content"`
	IsError   bool           `json:"is_error"`
	Metadata  map[string]any `json:"metadata"`
	Cost      uint64         `json:"cost"`      // Actual cost incurred
	Resources []string       `json:"resources"` // Resources accessed
}

// ToolRegistry manages tool registration, discovery, and execution
type ToolRegistry interface {
	// Registration
	Register(tool Tool) error
	Unregister(name string) error

	// Discovery
	Get(name string) (Tool, error)
	List() []Tool
	Search(query map[string]any) []Tool

	// Execution
	Call(ctx context.Context, name string, args map[string]any) (ToolResult, error)
	Stream(ctx context.Context, name string, args map[string]any) (<-chan ToolResult, error)
}

// ToolProvider defines a system that can provide tools
type ToolProvider interface {
	// Discovery
	GetTools() []Tool

	// Lifecycle
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
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
	key := tool.Name

	// Validate the tool
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool handler is required")
	}

	// Initialize the tool if needed
	if tool.Initialize != nil {
		if err := tool.Initialize(context.Background()); err != nil {
			return fmt.Errorf("failed to initialize tool: %w", err)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate registration
	if _, exists := r.tools[key]; exists {
		return fmt.Errorf("tool %q already registered", tool.Name)
	}

	r.tools[key] = tool
	return nil
}

// Unregister implements ToolRegistry.Unregister
func (r *SimpleToolRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find all versions of the tool
	for key, tool := range r.tools {
		if tool.Name == name {
			// Cleanup if needed
			if tool.Cleanup != nil {
				if err := tool.Cleanup(context.Background()); err != nil {
					return fmt.Errorf("failed to cleanup tool: %w", err)
				}
			}
			delete(r.tools, key)
		}
	}
	return nil
}

// Get implements ToolRegistry.Get
func (r *SimpleToolRegistry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Parse name and version
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

// Search implements ToolRegistry.Search
func (r *SimpleToolRegistry) Search(query map[string]any) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []Tool
	for _, tool := range r.tools {
		if matchesQuery(tool, query) {
			results = append(results, tool)
		}
	}
	return results
}

// Call implements ToolRegistry.Call
func (r *SimpleToolRegistry) Call(ctx context.Context, name string, args map[string]any) (ToolResult, error) {
	tool, err := r.Get(name)
	if err != nil {
		return ToolResult{}, err
	}

	// Validate arguments
	if err := ValidateToolArgs(&tool, args); err != nil {
		return ToolResult{}, fmt.Errorf("invalid arguments: %w", err)
	}

	// Execute the tool
	result, err := tool.Handler(ctx, args)
	if err != nil {
		return NewErrorResult(err), nil
	}

	return result, nil
}

// Stream implements ToolRegistry.Stream
func (r *SimpleToolRegistry) Stream(ctx context.Context, name string, args map[string]any) (<-chan ToolResult, error) {
	resultChan := make(chan ToolResult)

	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	// Validate arguments
	if err := ValidateToolArgs(&tool, args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Start streaming in a goroutine
	go func() {
		defer close(resultChan)

		result, err := tool.Handler(ctx, args)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case resultChan <- NewErrorResult(err):
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		case resultChan <- result:
		}
	}()

	return resultChan, nil
}

// matchesQuery checks if a tool matches the search query
func matchesQuery(tool Tool, query map[string]any) bool {
	if len(query) == 0 {
		return true
	}

	for k, v := range query {
		switch k {
		case "name":
			if tool.Name != v.(string) {
				return false
			}
		case "category":
			if tool.Category != v.(string) {
				return false
			}
		case "tag":
			found := false
			for _, tag := range tool.Tags {
				if tag == v.(string) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

// GetArgsSchema implements ToolValidator.GetArgsSchema
func (t *Tool) GetArgsSchema() json.RawMessage {
	return t.Schema
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

// ToolBuilder provides a fluent interface for building tools
type ToolBuilder struct {
	tool Tool
}

// New creates a new ToolBuilder
func New(name string) *ToolBuilder {
	return &ToolBuilder{
		tool: Tool{
			Name: name,
		},
	}
}

// WithDescription sets the tool description
func (b *ToolBuilder) WithDescription(desc string) *ToolBuilder {
	b.tool.Description = desc
	return b
}

// WithCategory sets the tool category
func (b *ToolBuilder) WithCategory(category string) *ToolBuilder {
	b.tool.Category = category
	return b
}

// WithTags adds tags to the tool
func (b *ToolBuilder) WithTags(tags ...string) *ToolBuilder {
	b.tool.Tags = append(b.tool.Tags, tags...)
	return b
}

// WithSchema sets the JSON schema for input validation
func (b *ToolBuilder) WithSchema(schema json.RawMessage) *ToolBuilder {
	b.tool.Schema = schema
	return b
}

// WithHandler sets the tool handler function
func (b *ToolBuilder) WithHandler(handler func(context.Context, map[string]any) (string, error)) *ToolBuilder {
	b.tool.Handler = func(ctx context.Context, args map[string]any) (ToolResult, error) {
		content, err := handler(ctx, args)
		if err != nil {
			return NewErrorResult(err), nil
		}
		return NewToolResult(content), nil
	}
	return b
}

// Build creates the final Tool
func (b *ToolBuilder) Build() Tool {
	return b.tool
}

// Register registers the tool with the default registry
func Register(tool Tool) error {
	registry := NewSimpleToolRegistry()
	return registry.Register(tool)
}

// RegisterFunctionTool creates and registers a tool from a simple function
// It examines the function signature to generate an appropriate schema
func RegisterFunctionTool(registry ToolRegistry, name, description string, handler interface{}) error {
	// Convert the handler to a proper ToolHandler function
	toolHandler := func(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
		// Handle different function signatures
		switch h := handler.(type) {
		case func() string:
			// Simple no-arg function
			return NewToolResult(h()), nil

		case func(string) string:
			// Function that takes a single string argument
			input, _ := args["input"].(string)
			return NewToolResult(h(input)), nil

		case func(map[string]interface{}) string:
			// Function that takes a map of arguments
			return NewToolResult(h(args)), nil

		case func(context.Context, map[string]interface{}) (string, error):
			// Full handler with context and error
			result, err := h(ctx, args)
			if err != nil {
				return NewErrorResult(err), nil
			}
			return NewToolResult(result), nil

		case func() (string, error):
			// No-arg function that can error
			result, err := h()
			if err != nil {
				return NewErrorResult(err), nil
			}
			return NewToolResult(result), nil

		default:
			// Unsupported handler type
			return ToolResult{}, fmt.Errorf("unsupported handler type: %T", handler)
		}
	}

	// Generate schema based on function signature
	var schema json.RawMessage
	switch handler.(type) {
	case func() string, func() (string, error):
		// No parameters
		schema = json.RawMessage(`{"type":"object","properties":{},"required":[]}`)

	case func(string) string:
		// Single string input
		schema = json.RawMessage(`{
			"type": "object",
			"properties": {
				"input": {"type": "string", "description": "Input text"}
			},
			"required": ["input"]
		}`)

	default:
		// Default to accepting any object
		schema = json.RawMessage(`{"type":"object","properties":{},"additionalProperties":true}`)
	}

	// Create and register the tool
	tool := Tool{
		Name:        name,
		Description: description,
		Schema:      schema,
		Handler:     toolHandler,
	}

	return registry.Register(tool)
}
