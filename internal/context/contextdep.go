package context

import (
	"fmt"
	"sync"
)

// ContextDependent is an interface for components that need context access.
// It provides methods for getting and setting context, with fallback to global context.
type ContextDependent interface {
	// Context returns the current context, falling back to global context if needed.
	Context() (*Context, error)

	// WithContext returns a new instance with the given context.
	WithContext(ctx *Context) ContextDependent

	// Cleanup performs any necessary cleanup when the component is done.
	Cleanup() error
}

// BaseContextDependent provides a base implementation of ContextDependent.
// Embed this struct in components that need context access.
type BaseContextDependent struct {
	ctx *Context
	mu  sync.RWMutex
}

// NewBaseContextDependent creates a new BaseContextDependent instance.
func NewBaseContextDependent(ctx *Context) *BaseContextDependent {
	return &BaseContextDependent{ctx: ctx}
}

// Context returns the current context, falling back to global context if needed.
func (b *BaseContextDependent) Context() (*Context, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// First try instance context
	if b.ctx != nil {
		return b.ctx, nil
	}

	// Fall back to global context
	ctx, err := GetGlobal()
	if err != nil {
		return nil, fmt.Errorf("no context available: %w", err)
	}

	return ctx, nil
}

// WithContext returns a new instance with the given context.
// This is meant to be used by embedding structs to implement their own WithContext.
func (b *BaseContextDependent) WithContext(ctx *Context) ContextDependent {
	return NewBaseContextDependent(ctx)
}

// Cleanup performs any necessary cleanup.
func (b *BaseContextDependent) Cleanup() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.ctx = nil
	return nil
}

// MustContext is a helper that returns the context or panics if not available.
// This is useful in situations where context access should never fail.
func (b *BaseContextDependent) MustContext() *Context {
	ctx, err := b.Context()
	if err != nil {
		panic(fmt.Sprintf("context not available: %v", err))
	}
	return ctx
}
