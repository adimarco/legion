package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adimarco/hive/internal/config"
)

// TestComponent implements ContextDependent for testing
type TestComponent struct {
	*BaseContextDependent
}

func NewTestComponent(ctx *Context) *TestComponent {
	return &TestComponent{
		BaseContextDependent: NewBaseContextDependent(ctx),
	}
}

// WithContext returns a new TestComponent with the given context
func (t *TestComponent) WithContext(ctx *Context) ContextDependent {
	return NewTestComponent(ctx)
}

func TestBaseContextDependent_Context(t *testing.T) {
	// Ensure we start with a clean state
	_ = CleanupGlobal()

	// Create a test config
	cfg := &config.Settings{DefaultModel: "test-model"}

	// Test with instance context
	ctx, err := Initialize(cfg)
	require.NoError(t, err)

	comp := NewTestComponent(ctx)
	compCtx, err := comp.Context()
	require.NoError(t, err)
	assert.Equal(t, ctx, compCtx)

	// Test fallback to global context
	comp = NewTestComponent(nil)
	_, err = comp.Context()
	assert.Error(t, err) // Should fail because no global context

	// Initialize global context
	globalCtx, err := InitializeGlobal(cfg)
	require.NoError(t, err)

	// Now should succeed with global context
	compCtx, err = comp.Context()
	require.NoError(t, err)
	assert.Equal(t, globalCtx, compCtx)

	// Cleanup
	_ = CleanupGlobal()
}

func TestBaseContextDependent_WithContext(t *testing.T) {
	// Create two different contexts
	cfg1 := &config.Settings{DefaultModel: "test-model-1"}
	cfg2 := &config.Settings{DefaultModel: "test-model-2"}

	ctx1, err := Initialize(cfg1)
	require.NoError(t, err)
	ctx2, err := Initialize(cfg2)
	require.NoError(t, err)

	// Create component with first context
	comp := NewTestComponent(ctx1)

	// Switch to second context
	comp2 := comp.WithContext(ctx2).(*TestComponent)

	// Verify contexts
	ctx1Got, err := comp.Context()
	require.NoError(t, err)
	assert.Equal(t, ctx1, ctx1Got)

	ctx2Got, err := comp2.Context()
	require.NoError(t, err)
	assert.Equal(t, ctx2, ctx2Got)
}

func TestBaseContextDependent_Cleanup(t *testing.T) {
	cfg := &config.Settings{DefaultModel: "test-model"}
	ctx, err := Initialize(cfg)
	require.NoError(t, err)

	comp := NewTestComponent(ctx)

	// Test cleanup
	err = comp.Cleanup()
	require.NoError(t, err)

	// Context should now be nil
	_, err = comp.Context()
	assert.Error(t, err)
}

func TestBaseContextDependent_MustContext(t *testing.T) {
	// Test with valid context
	cfg := &config.Settings{DefaultModel: "test-model"}
	ctx, err := Initialize(cfg)
	require.NoError(t, err)

	comp := NewTestComponent(ctx)
	assert.NotPanics(t, func() {
		compCtx := comp.MustContext()
		assert.Equal(t, ctx, compCtx)
	})

	// Test with no context (should panic)
	comp = NewTestComponent(nil)
	assert.Panics(t, func() {
		_ = comp.MustContext()
	})
}

func TestBaseContextDependent_Concurrency(t *testing.T) {
	cfg := &config.Settings{DefaultModel: "test-model"}
	ctx, err := Initialize(cfg)
	require.NoError(t, err)

	comp := NewTestComponent(ctx)

	// Test concurrent access to context
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = comp.Context()
			_ = comp.WithContext(ctx)
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}
}
