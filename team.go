package hive

import (
	"context"
	"fmt"

	"github.com/adimarco/hive/internal/llm"
)

// Archetype defines a specialist role with its behavior
type Archetype struct {
	Name        string
	Instruction string
	UseHistory  bool
}

var archetypeRegistry = make(map[string]Archetype)

// RegisterArchetype adds an archetype to the global registry
func RegisterArchetype(name string, archetype Archetype) {
	archetype.Name = name // Ensure name matches key
	archetypeRegistry[name] = archetype
}

// GetArchetype retrieves an archetype from the registry
func GetArchetype(name string) (Archetype, bool) {
	a, ok := archetypeRegistry[name]
	return a, ok
}

// Team manages a collection of Agents
type Team struct {
	name   string
	llm    llm.AugmentedLLM
	agents map[string]*Agent
	ctx    context.Context
	cancel context.CancelFunc
}

// TeamWithLLM creates a new Team with the given LLM and agents
func TeamWithLLM(name string, llm llm.AugmentedLLM, agents ...*Agent) *Team {
	ctx, cancel := context.WithCancel(context.Background())
	team := &Team{
		name:   name,
		llm:    llm,
		agents: make(map[string]*Agent),
		ctx:    ctx,
		cancel: cancel,
	}

	for _, agent := range agents {
		team.agents[agent.name] = agent
	}

	return team
}

// Send sends a message to a specific agent and returns its response
func (t *Team) Send(agentName, message string) (string, error) {
	agent, ok := t.agents[agentName]
	if !ok {
		return "", fmt.Errorf("agent %q not found", agentName)
	}

	params := &llm.RequestParams{
		Model:      agent.model,
		UseHistory: agent.useHistory,
		Tools:      agent.params.Tools,
		Config:     agent.params.Config,
	}

	return t.llm.GenerateString(t.ctx, message, params)
}

// Close cleans up team resources
func (t *Team) Close() {
	if t.cancel != nil {
		t.cancel()
	}
}

// TeamBuilder provides a fluent interface for building teams
type TeamBuilder struct {
	name        string
	coordinator *Agent
	specialists []*Agent
}

// NewTeam creates a new TeamBuilder
func NewTeam(name string) *TeamBuilder {
	return &TeamBuilder{
		name:        name,
		specialists: make([]*Agent, 0),
	}
}

// WithCoordinator adds a coordinator to the team
func (b *TeamBuilder) WithCoordinator(instruction string) *TeamBuilder {
	b.coordinator = New("Coordinator", instruction).WithHistory()
	return b
}

// WithArchetype adds a specialist to the team using a registered archetype
func (b *TeamBuilder) WithArchetype(name string) *TeamBuilder {
	if archetype, ok := GetArchetype(name); ok {
		agent := New(name, archetype.Instruction)
		if archetype.UseHistory {
			agent.WithHistory()
		}
		b.specialists = append(b.specialists, agent)
	}
	return b
}

// WithSpecialist adds a specialist to the team
func (b *TeamBuilder) WithSpecialist(name, instruction string) *TeamBuilder {
	b.specialists = append(b.specialists, New(name, instruction))
	return b
}

// Build creates the Team
func (b *TeamBuilder) Build(llm llm.AugmentedLLM) *Team {
	agents := make([]*Agent, 0, len(b.specialists)+1)
	if b.coordinator != nil {
		agents = append(agents, b.coordinator)
	}
	agents = append(agents, b.specialists...)
	return TeamWithLLM(b.name, llm, agents...)
}

// ArchetypeBuilder provides a fluent interface for building archetypes
type ArchetypeBuilder struct {
	archetype Archetype
}

// NewArchetype starts building a new archetype
func NewArchetype(name string) *ArchetypeBuilder {
	return &ArchetypeBuilder{
		archetype: Archetype{Name: name},
	}
}

// WithRole sets the role description
func (b *ArchetypeBuilder) WithRole(role string) *ArchetypeBuilder {
	b.archetype.Instruction = role
	return b
}

// WithHistory enables conversation history
func (b *ArchetypeBuilder) WithHistory() *ArchetypeBuilder {
	b.archetype.UseHistory = true
	return b
}

// Register adds the archetype to the registry
func (b *ArchetypeBuilder) Register() {
	RegisterArchetype(b.archetype.Name, b.archetype)
}

// Chat starts an interactive chat session with the specified agent
func (t *Team) Chat(agentName string) error {
	agent, ok := t.agents[agentName]
	if !ok {
		return fmt.Errorf("agent %q not found", agentName)
	}

	// Create running agent
	ra, err := agent.Run(t.ctx)
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Start chat session
	return ra.Chat()
}
