package agent

import (
	"context"

	"gofast/internal/llm"
)

// FastAgent is the main entry point for creating and running agents
type FastAgent struct {
	name string
	llm  llm.AugmentedLLM
}

// NewFastAgent creates a new FastAgent instance
func NewFastAgent(name string, llm llm.AugmentedLLM) *FastAgent {
	return &FastAgent{
		name: name,
		llm:  llm,
	}
}

// AgentConfig holds configuration for a single agent instance
type AgentConfig struct {
	Instruction string
}

// Agent represents a single configured agent instance
type Agent struct {
	config AgentConfig
	fa     *FastAgent
}

// NewAgent creates a new Agent with the given configuration
func (fa *FastAgent) NewAgent(config AgentConfig) *Agent {
	return &Agent{
		config: config,
		fa:     fa,
	}
}

// Run starts an agent session and returns a RunningAgent that can be used for interaction
func (a *Agent) Run(ctx context.Context) (*RunningAgent, error) {
	return &RunningAgent{
		agent: a,
		ctx:   ctx,
	}, nil
}

// RunningAgent represents an active agent session
type RunningAgent struct {
	agent *Agent
	ctx   context.Context
}
