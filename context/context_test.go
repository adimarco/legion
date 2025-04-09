package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adimarco/hive/internal/config"
)

func TestContext_Initialize(t *testing.T) {
	// Create test config
	cfg := &config.Settings{
		DefaultModel: "test-model",
		Logger: config.LoggerSettings{
			Type:  "console",
			Level: "info",
		},
	}

	// Test initialization with config
	ctx, err := Initialize(cfg)
	require.NoError(t, err)
	assert.NotNil(t, ctx)
	assert.Equal(t, cfg, ctx.Config)
	assert.True(t, ctx.IsInitialized())

	// Test initialization with nil config (should load default)
	ctx, err = Initialize(nil)
	require.NoError(t, err)
	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.Config)
	assert.True(t, ctx.IsInitialized())
}

func TestContext_GlobalContext(t *testing.T) {
	// Ensure we start with a clean state
	_ = CleanupGlobal()

	// Test getting uninitialized global context
	ctx, err := GetGlobal()
	assert.Error(t, err)
	assert.Nil(t, ctx)

	// Initialize global context
	cfg := &config.Settings{DefaultModel: "test-model"}
	ctx, err = InitializeGlobal(cfg)
	require.NoError(t, err)
	assert.NotNil(t, ctx)

	// Test getting initialized global context
	ctx, err = GetGlobal()
	require.NoError(t, err)
	assert.NotNil(t, ctx)
	assert.Equal(t, cfg, ctx.Config)

	// Test initializing global context again (should fail)
	ctx, err = InitializeGlobal(cfg)
	assert.Error(t, err)
	assert.Nil(t, ctx)

	// Test cleanup
	err = CleanupGlobal()
	require.NoError(t, err)

	// Verify cleanup worked
	ctx, err = GetGlobal()
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

func TestContext_Cleanup(t *testing.T) {
	cfg := &config.Settings{DefaultModel: "test-model"}
	ctx, err := Initialize(cfg)
	require.NoError(t, err)

	// Test cleanup
	err = ctx.Cleanup()
	require.NoError(t, err)
	assert.False(t, ctx.IsInitialized())

	// Test cleanup of already cleaned up context
	err = ctx.Cleanup()
	assert.NoError(t, err)
}

func TestContext_Concurrency(t *testing.T) {
	// This test ensures our mutex protection works
	// We'll initialize and cleanup the context concurrently

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			cfg := &config.Settings{DefaultModel: "test-model"}
			ctx, err := Initialize(cfg)
			if err == nil && ctx != nil {
				_ = ctx.Cleanup()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}
}
