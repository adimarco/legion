package hive

import (
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

// WithAgent creates and registers a new agent with the app
func (a *App) WithAgent(instruction string) *Agent {
	agent := NewDefaultAgent(instruction).WithLLM(a.llm)
	a.agents[agent.name] = agent
	return agent
}

// WithTool registers a tool with the app's LLM
func (a *App) WithTool(tool tools.Tool) error {
	return a.llm.Tools().Register(tool)
}

// Close cleans up app resources
func (a *App) Close() error {
	if a.llm != nil {
		return a.llm.Cleanup()
	}
	return nil
}
