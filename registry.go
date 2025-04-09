package hive

import (
	"fmt"
	"net/url"
	"sync"
)

// Registry represents an agent registry (local or remote)
type Registry struct {
	url   *url.URL
	store *AgentStore
	mu    sync.RWMutex
}

// AgentStore is the in-memory store for agent versions
type AgentStore struct {
	agents map[string]map[string]*AgentPackage // name -> version -> package
}

// AgentPackage represents a versioned agent configuration
type AgentPackage struct {
	Name       string         // e.g. "skiddie420/cpo"
	Version    string         // e.g. "latest", "1.0.0"
	Role       string         // Core instruction/role
	Tools      []string       // Required MCP tools
	Config     map[string]any // Additional configuration
	UseHistory bool
	Metadata   map[string]any // Author, tags, etc.
}

// NewRegistry creates a new registry from a URL
// Supported schemes: thoth://, memory://, http://, https://
func NewRegistry(uri string) (*Registry, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid registry URL: %w", err)
	}

	r := &Registry{
		url: u,
		store: &AgentStore{
			agents: make(map[string]map[string]*AgentPackage),
		},
	}

	return r, nil
}

// AgentConfig holds configuration for publishing an agent
type AgentConfig struct {
	Version    string         // Semantic version
	Model      string         // LLM model to use
	Tools      []string       // Required MCP tools
	Config     map[string]any // Additional configuration
	UseHistory bool
	Metadata   map[string]any // Author, tags, etc.
}

// PublishAgent adds a new agent to the registry
func (r *Registry) PublishAgent(name, role string, cfg AgentConfig) *AgentPackage {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions, exists := r.store.agents[name]
	if !exists {
		versions = make(map[string]*AgentPackage)
		r.store.agents[name] = versions
	}

	agent := &AgentPackage{
		Name:       name,
		Version:    cfg.Version,
		Role:       role,
		Tools:      cfg.Tools,
		Config:     cfg.Config,
		UseHistory: cfg.UseHistory,
		Metadata:   cfg.Metadata,
	}

	versions[cfg.Version] = agent
	versions["latest"] = agent // Update latest pointer

	return agent
}

// GetAgent retrieves an agent from the registry
func (r *Registry) GetAgent(name string) (*AgentPackage, error) {
	return r.GetAgentVersion(name, "latest")
}

// GetAgentVersion retrieves a specific version of an agent
func (r *Registry) GetAgentVersion(name, version string) (*AgentPackage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.store.agents[name]
	if !exists {
		return nil, fmt.Errorf("agent %q not found", name)
	}

	agent, exists := versions[version]
	if !exists {
		return nil, fmt.Errorf("version %q of agent %q not found", version, name)
	}

	return agent, nil
}

// SearchAgents searches for agents by metadata
func (r *Registry) SearchAgents(query map[string]any) []*AgentPackage {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*AgentPackage, 0)
	for _, versions := range r.store.agents {
		agent := versions["latest"]
		if agent.matchesQuery(query) {
			results = append(results, agent)
		}
	}

	return results
}

// matchesQuery checks if an agent matches the search query
func (a *AgentPackage) matchesQuery(query map[string]any) bool {
	if len(query) == 0 {
		return true
	}

	for k, v := range query {
		if mv, ok := a.Metadata[k]; !ok || mv != v {
			return false
		}
	}

	return true
}

// UseMCPTools adds MCP tools to an agent
func (a *AgentPackage) UseMCPTools(tools ...string) {
	a.Tools = append(a.Tools, tools...)
}

// WithConfig adds configuration to an agent
func (a *AgentPackage) WithConfig(cfg map[string]any) *AgentPackage {
	a.Config = cfg
	return a
}

// WithHistory enables conversation history
func (a *AgentPackage) WithHistory() *AgentPackage {
	a.UseHistory = true
	return a
}

// ToAgent converts an agent package to an Agent instance
func (a *AgentPackage) ToAgent() *Agent {
	agent := New(a.Name, a.Role)
	if a.UseHistory {
		agent.WithHistory()
	}
	if len(a.Tools) > 0 {
		agent.WithTools(a.Tools...)
	}
	if a.Config != nil {
		agent.WithConfig(a.Config)
	}
	return agent
}

// MustGetAgent retrieves an agent from the registry, panicking if not found
func (r *Registry) MustGetAgent(name string) *AgentPackage {
	agent, err := r.GetAgent(name)
	if err != nil {
		panic(fmt.Sprintf("agent %q not found: %v", name, err))
	}
	return agent
}
