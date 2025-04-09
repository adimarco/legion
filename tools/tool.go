package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Tool represents a callable tool with metadata and execution capabilities
type Tool struct {
	// Core metadata
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`

	// Schema and validation
	Schema json.RawMessage `json:"schema"` // JSON Schema for input validation

	// Dependencies
	Requires []string        `json:"requires"` // Required tools
	Config   json.RawMessage `json:"config"`   // Tool configuration

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
	key := fmt.Sprintf("%s@%s", tool.Name, tool.Version)

	// Validate the tool
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Version == "" {
		return fmt.Errorf("tool version is required")
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool handler is required")
	}

	// Check dependencies first, outside the lock
	for _, dep := range tool.Requires {
		if _, err := r.Get(dep); err != nil {
			return fmt.Errorf("missing dependency: %s", dep)
		}
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
		return fmt.Errorf("tool %q version %q already registered", tool.Name, tool.Version)
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
	if !strings.Contains(name, "@") {
		// Find latest version
		var latest Tool
		var latestVer string
		for key, tool := range r.tools {
			if strings.HasPrefix(key, name+"@") {
				if latestVer == "" || tool.Version > latestVer {
					latest = tool
					latestVer = tool.Version
				}
			}
		}
		if latestVer != "" {
			return latest, nil
		}
		return Tool{}, fmt.Errorf("tool %q not found", name)
	}

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

	// Check dependencies
	if err := r.checkDependencies(ctx, tool); err != nil {
		return ToolResult{}, fmt.Errorf("dependency check failed: %w", err)
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

	// Check dependencies
	if err := r.checkDependencies(ctx, tool); err != nil {
		return nil, fmt.Errorf("dependency check failed: %w", err)
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

// checkDependencies verifies that all required tools are available
func (r *SimpleToolRegistry) checkDependencies(ctx context.Context, tool Tool) error {
	for _, dep := range tool.Requires {
		if _, err := r.Get(dep); err != nil {
			return fmt.Errorf("required tool %q not found", dep)
		}
	}
	return nil
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
			Name:    name,
			Version: "1.0.0", // Default version
		},
	}
}

// WithVersion sets the tool version
func (b *ToolBuilder) WithVersion(version string) *ToolBuilder {
	b.tool.Version = version
	return b
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
