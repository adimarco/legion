package fastagent

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

// Build creates the FastAgent team
func (b *TeamBuilder) Build() *FastAgent {
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return nil
	}

	agents := make([]*Agent, 0, len(b.specialists)+1)
	if b.coordinator != nil {
		agents = append(agents, b.coordinator)
	}
	agents = append(agents, b.specialists...)
	return Team(b.name, cfg, agents...)
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
