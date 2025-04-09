package fastagent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"

	"gofast/internal/agent"
	"gofast/internal/config"
	"gofast/internal/llm"
)

// Agent represents a configured agent that can be customized with options
type Agent struct {
	name        string
	instruction string
	model       string
	useHistory  bool
	humanInput  bool
	params      *llm.RequestParams
}

// FastAgent manages a collection of agents
type FastAgent struct {
	name   string
	llm    llm.AugmentedLLM
	agents map[string]*agent.ChannelAgent
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Agent with basic configuration
func New(name, instruction string) *Agent {
	return &Agent{
		name:        name,
		instruction: instruction,
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

// Team creates a new FastAgent instance with the given agents
func Team(name string, cfg *Config, agents ...*Agent) *FastAgent {
	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
	}()

	f := &FastAgent{
		name:   name,
		agents: make(map[string]*agent.ChannelAgent),
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize LLM
	f.llm = llm.NewAnthropicLLM(name)
	if err := f.llm.Initialize(ctx, &config.Settings{
		DefaultModel: cfg.DefaultModel,
		Logger: config.LoggerSettings{
			Level: cfg.LogLevel,
		},
	}); err != nil {
		fmt.Printf("Error initializing LLM: %v\n", err)
		cancel()
		return f
	}

	// Initialize all agents
	for _, a := range agents {
		config := agent.AgentConfig{
			Name:        a.name,
			Instruction: a.instruction,
			Type:        agent.AgentTypeBasic,
			Model:       a.model,
			UseHistory:  a.useHistory,
		}

		channelAgent := agent.NewChannelAgent(config, f.llm)
		if err := channelAgent.Start(ctx); err != nil {
			fmt.Printf("Error starting agent %s: %v\n", a.name, err)
			continue
		}

		f.agents[a.name] = channelAgent
	}

	return f
}

// TeamWithLLM creates a new FastAgent instance with a specific LLM
func TeamWithLLM(name string, llm llm.AugmentedLLM, agents ...*Agent) *FastAgent {
	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
	}()

	f := &FastAgent{
		name:   name,
		agents: make(map[string]*agent.ChannelAgent),
		ctx:    ctx,
		cancel: cancel,
		llm:    llm,
	}

	// Initialize all agents
	for _, a := range agents {
		config := agent.AgentConfig{
			Name:        a.name,
			Instruction: a.instruction,
			Type:        agent.AgentTypeBasic,
			Model:       a.model,
			UseHistory:  a.useHistory,
		}

		channelAgent := agent.NewChannelAgent(config, f.llm)
		if err := channelAgent.Start(ctx); err != nil {
			fmt.Printf("Error starting agent %s: %v\n", a.name, err)
			continue
		}

		f.agents[a.name] = channelAgent
	}

	return f
}

// formatResponse handles message formatting for display
func formatResponse(response string) string {
	// Handle empty responses
	if response == "" {
		return ""
	}

	// Format bold terms using ANSI escape codes
	// Note: This assumes terminal output. For other outputs,
	// formatting should be handled by the display layer
	response = strings.ReplaceAll(response, "*", "\033[1m")
	response = strings.ReplaceAll(response, "*", "\033[0m")

	return strings.TrimSpace(response)
}

// Send sends a message to the specified agent and returns a formatted response
func (f *FastAgent) Send(agentName, message string) (string, error) {
	agent, ok := f.agents[agentName]
	if !ok {
		return "", fmt.Errorf("agent %q not found", agentName)
	}

	// Print the assignment with color
	color.New(color.FgHiBlack).Printf("\n→ %s: %s\n", agentName, message)

	if err := agent.Send(message); err != nil {
		return "", fmt.Errorf("failed to send message to %s: %w", agentName, err)
	}

	// Wait for response with proper error handling
	select {
	case response := <-agent.Output():
		formatted := formatResponse(response)
		fmt.Printf("\n%s: %s\n", color.New(color.Bold).Sprint(agentName), formatted)
		return response, nil
	case err := <-agent.Errors():
		color.Red("Error from %s: %v", agentName, err)
		return "", fmt.Errorf("error from %s: %w", agentName, err)
	case <-f.ctx.Done():
		return "", fmt.Errorf("context cancelled while waiting for response from %s: %w", agentName, f.ctx.Err())
	}
}

// Chat starts an interactive chat session with the specified agent
func (f *FastAgent) Chat(agentName string) error {
	channelAgent, ok := f.agents[agentName]
	if !ok {
		return fmt.Errorf("agent %q not found", agentName)
	}

	// Use the existing RunningAgent's Chat method
	runningAgent, err := channelAgent.Agent.Run(f.ctx)
	if err != nil {
		return fmt.Errorf("failed to start chat: %w", err)
	}

	fmt.Printf("\nStarting chat with %s. Press Ctrl+C to exit.\n", agentName)
	return runningAgent.Chat()
}

// Close cleans up all resources
func (f *FastAgent) Close() error {
	f.cancel()
	for _, agent := range f.agents {
		agent.Close()
	}
	return nil
}
