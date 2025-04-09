package hive

import (
	"context"
	"fmt"

	"github.com/adimarco/hive/internal/llm"
	"github.com/adimarco/hive/internal/tools"
)

// App provides a high-level interface for creating and managing agents and tools
type App struct {
	llm    llm.AugmentedLLM
	agents map[string]*Agent
}

// NewApp creates a new App instance with default configuration
func NewApp(name string) *App {
	llm, err := NewAnthropicLLM(name)
	if err != nil {
		// For now, we'll panic on initialization errors
		// In the future, we can return error and let caller handle it
		panic(fmt.Sprintf("failed to create LLM: %v", err))
	}
	return &App{
		llm:    llm,
		agents: make(map[string]*Agent),
	}
}

// Agent creates a new agent with the given instruction
func (a *App) Agent(instruction string) *Agent {
	agent := NewDefaultAgent(instruction).WithLLM(a.llm)

	// Add all available tools automatically
	tools := a.llm.Tools().List()
	for _, tool := range tools {
		agent.WithTools(fmt.Sprintf("%s@%s", tool.Name, tool.Version))
	}

	a.agents[agent.name] = agent
	return agent
}

// Tool creates and registers a new tool with a simple function handler
func (a *App) Tool(name string, handler interface{}) error {
	// Convert the handler to a proper tool handler function
	toolHandler := func(ctx context.Context, args map[string]any) (string, error) {
		// For now, we'll just handle simple functions that return a string
		// Later we can add support for more function signatures
		if h, ok := handler.(func() string); ok {
			result := tools.NewToolResult(h())
			return result.Content, nil
		}
		return "", fmt.Errorf("unsupported handler type")
	}

	// Create and register the tool
	tool := Tool(name, toolHandler).
		WithDescription(fmt.Sprintf("Tool '%s' that returns a result", name)).
		Build()

	return a.llm.Tools().Register(tool)
}

// Close cleans up app resources
func (a *App) Close() error {
	if a.llm != nil {
		return a.llm.Cleanup()
	}
	return nil
}
