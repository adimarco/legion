package hive

import (
	"fmt"

	"github.com/adimarco/hive/llm"
	"github.com/adimarco/hive/tools"
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

	app := &App{
		llm:    llm,
		agents: make(map[string]*Agent),
	}

	return app
}

// Agent creates a new agent with the given instruction
func (a *App) Agent(instruction string) *Agent {
	agent := NewDefaultAgent(instruction).WithLLM(a.llm)

	// Add all available tools automatically
	tools := a.llm.Tools().List()

	// Track which tools we've already added to avoid duplicates
	addedTools := make(map[string]bool)

	for _, tool := range tools {
		// Skip tools we've already added
		if addedTools[tool.Name] {
			continue
		}

		agent.WithTools(tool.Name)
		addedTools[tool.Name] = true
	}

	a.agents[agent.name] = agent
	return agent
}

// Tool creates and registers a new tool with a simple function handler
func (a *App) Tool(name string, handler interface{}) error {
	// Use the new RegisterFunctionTool helper
	description := fmt.Sprintf("Tool '%s' provided by the application", name)
	return tools.RegisterFunctionTool(a.llm.Tools(), name, description, handler)
}

// Close cleans up app resources
func (a *App) Close() error {
	if a.llm != nil {
		return a.llm.Cleanup()
	}
	return nil
}
