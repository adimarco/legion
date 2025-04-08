package context

import (
	"fmt"
	"sync"

	"gofast/internal/config"
)

// Context represents the global application context that is shared across components.
type Context struct {
	// Configuration settings
	Config *config.Settings

	// Component registries and handlers will be added as we implement them
	// Executor        Executor
	// HumanInput     HumanInputHandler
	// SignalNotifier SignalNotifier
	// ServerRegistry ServerRegistry
	// TaskRegistry   TaskRegistry
	// Logger         Logger
	// Tracer         Tracer

	// Internal state
	initialized bool
	mu          sync.RWMutex
}

var (
	// Global context instance
	globalContext *Context
	globalMu      sync.RWMutex
)

// Initialize creates a new Context instance with the provided configuration.
// If config is nil, it will attempt to load the default configuration.
func Initialize(cfg *config.Settings) (*Context, error) {
	if cfg == nil {
		var err error
		cfg, err = config.LoadSettings("")
		if err != nil {
			return nil, fmt.Errorf("failed to load default config: %w", err)
		}
	}

	ctx := &Context{
		Config:      cfg,
		initialized: true,
	}

	return ctx, nil
}

// InitializeGlobal creates and sets the global context instance.
// This is thread-safe and will return an error if called multiple times.
func InitializeGlobal(cfg *config.Settings) (*Context, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalContext != nil {
		return nil, fmt.Errorf("global context already initialized")
	}

	ctx, err := Initialize(cfg)
	if err != nil {
		return nil, err
	}

	globalContext = ctx
	return ctx, nil
}

// GetGlobal returns the global context instance.
// Returns an error if the global context hasn't been initialized.
func GetGlobal() (*Context, error) {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if globalContext == nil {
		return nil, fmt.Errorf("global context not initialized")
	}

	return globalContext, nil
}

// Cleanup performs cleanup of context resources.
// This should be called when the context is no longer needed.
func (c *Context) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		return nil
	}

	// Cleanup will be expanded as we add more components
	c.initialized = false
	return nil
}

// CleanupGlobal cleans up the global context instance.
func CleanupGlobal() error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalContext == nil {
		return nil
	}

	if err := globalContext.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup global context: %w", err)
	}

	globalContext = nil
	return nil
}

// IsInitialized returns whether the context has been initialized.
func (c *Context) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}
