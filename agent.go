// Package hive provides a framework for building and deploying intelligent agents.
package hive

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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

// Agent represents a configured agent instance
type Agent struct {
	name        string
	instruction string
	agentType   AgentType
	model       string
	useHistory  bool
	humanInput  bool
	params      *llm.RequestParams
	llm         llm.AugmentedLLM
	output      io.Writer // For configurable output
}

// New creates a new Agent with basic configuration
func New(name, instruction string) *Agent {
	return &Agent{
		name:        name,
		instruction: instruction,
		agentType:   AgentTypeBasic,
		output:      os.Stdout,
	}
}

// WithModel sets the model for the agent
func (a *Agent) WithModel(model string) *Agent {
	a.model = model
	return a
}

// WithHistory enables conversation history
func (a *Agent) WithHistory() *Agent {
	a.useHistory = true
	return a
}

// WithHumanInput enables human input requests
func (a *Agent) WithHumanInput() *Agent {
	a.humanInput = true
	return a
}

// WithParams sets additional request parameters
func (a *Agent) WithParams(params *llm.RequestParams) *Agent {
	a.params = params
	return a
}

// WithTools adds MCP tools to the agent
func (a *Agent) WithTools(tools ...string) *Agent {
	if a.params == nil {
		a.params = &llm.RequestParams{}
	}
	if a.params.Tools == nil {
		a.params.Tools = make([]string, 0)
	}
	a.params.Tools = append(a.params.Tools, tools...)
	return a
}

// WithConfig adds additional configuration to the agent
func (a *Agent) WithConfig(cfg map[string]any) *Agent {
	if a.params == nil {
		a.params = &llm.RequestParams{}
	}
	if a.params.Config == nil {
		a.params.Config = make(map[string]any)
	}
	for k, v := range cfg {
		a.params.Config[k] = v
	}
	return a
}

// WithLLM sets the LLM for the agent
func (a *Agent) WithLLM(llm llm.AugmentedLLM) *Agent {
	a.llm = llm
	return a
}

// WithType sets the agent type
func (a *Agent) WithType(agentType AgentType) *Agent {
	a.agentType = agentType
	return a
}

// SetOutput configures where the agent writes output
func (a *Agent) SetOutput(w io.Writer) {
	a.output = w
}

// Run starts an agent session and returns a RunningAgent
func (a *Agent) Run(ctx context.Context) (*RunningAgent, error) {
	// Validate configuration
	if a.name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	if a.instruction == "" {
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
		SystemPrompt: ra.agent.instruction,
		Model:        ra.agent.model,
		UseHistory:   ra.agent.useHistory,
		Tools:        ra.agent.params.Tools,
		Config:       ra.agent.params.Config,
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
	fmt.Fprintln(ra.agent.output, "Instruction:", ra.agent.instruction)

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

// NewDefaultAgent creates a new agent with sensible defaults
func NewDefaultAgent(instruction string) *Agent {
	// Generate a unique name if not provided
	name := fmt.Sprintf("agent-%d", time.Now().UnixNano())

	// Create agent with sensible defaults
	return New(name, instruction).
		WithModel("claude-3-haiku-20240307").
		WithHistory()
}

// Send sends a message to the agent and returns the response
func (a *Agent) Send(msg string) (string, error) {
	// Create running agent with background context
	ra, err := a.Run(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to start agent: %w", err)
	}

	// Send message and get response
	return ra.Send(msg)
}
