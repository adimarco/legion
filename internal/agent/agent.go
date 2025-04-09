package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/adimarco/hive/internal/llm"
)

// AgentType represents different types of agents
type AgentType string

const (
	// AgentTypeBasic represents a simple agent that processes messages
	AgentTypeBasic AgentType = "agent"
	// AgentTypeOrchestrator coordinates multiple agents
	AgentTypeOrchestrator AgentType = "orchestrator"
	// AgentTypeRouter routes messages to appropriate agents
	AgentTypeRouter AgentType = "router"
	// AgentTypeChain chains multiple agents in sequence
	AgentTypeChain AgentType = "chain"
	// AgentTypeParallel runs multiple agents in parallel
	AgentTypeParallel AgentType = "parallel"
)

// AgentConfig holds configuration for an agent instance
type AgentConfig struct {
	// Name uniquely identifies this agent
	Name string
	// Instruction provides the agent's base behavior
	Instruction string
	// Type determines how this agent operates
	Type AgentType
	// Model specifies which LLM model to use
	Model string
	// UseHistory determines if conversation history is maintained
	UseHistory bool
	// RequestParams provides additional LLM configuration
	RequestParams *llm.RequestParams

	// Additional configuration for specific agent types
	ChildAgents   []string       // For orchestrator
	RouterAgents  []string       // For router
	ChainSequence []string       // For chain
	FanOutAgents  []string       // For parallel
	FanInAgent    string         // For parallel
	Metadata      map[string]any // For custom data
}

// Agent represents a configured agent instance
type Agent struct {
	config AgentConfig
	llm    llm.AugmentedLLM
	output io.Writer // For configurable output
}

// NewAgent creates a new Agent with the given configuration
func NewAgent(config AgentConfig, llm llm.AugmentedLLM) *Agent {
	return &Agent{
		config: config,
		llm:    llm,
		output: os.Stdout, // Default to stdout
	}
}

// SetOutput configures where the agent writes output
func (a *Agent) SetOutput(w io.Writer) {
	a.output = w
}

// Run starts an agent session and returns a RunningAgent
func (a *Agent) Run(ctx context.Context) (*RunningAgent, error) {
	// Validate configuration
	if a.config.Name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	if a.config.Instruction == "" {
		return nil, fmt.Errorf("agent instruction is required")
	}

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

// Send sends a single message to the agent and returns the response
func (ra *RunningAgent) Send(msg string) (string, error) {
	// Check context cancellation
	select {
	case <-ra.ctx.Done():
		return "", ra.ctx.Err()
	default:
	}

	message := llm.Message{
		Type:    llm.MessageTypeUser,
		Content: msg,
	}

	params := &llm.RequestParams{
		SystemPrompt: ra.agent.config.Instruction,
		UseHistory:   ra.agent.config.UseHistory,
	}
	if ra.agent.config.RequestParams != nil {
		params = ra.agent.config.RequestParams
	}

	response, err := ra.agent.llm.Generate(ra.ctx, message, params)
	if err != nil {
		return "", fmt.Errorf("failed to get LLM completion: %w", err)
	}

	return response.Content, nil
}

// Chat starts an interactive chat session with the agent
func (ra *RunningAgent) Chat() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprintln(ra.agent.output, "Starting chat session. Type 'exit' to end.")
	fmt.Fprintln(ra.agent.output, "Instruction:", ra.agent.config.Instruction)

	for {
		select {
		case <-ra.ctx.Done():
			return ra.ctx.Err()
		default:
		}

		fmt.Fprint(ra.agent.output, "\nUser: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			return nil
		}

		response, err := ra.Send(input)
		if err != nil {
			return fmt.Errorf("failed to get response: %w", err)
		}

		fmt.Fprintf(ra.agent.output, "Assistant: %s\n", response)
	}
}
