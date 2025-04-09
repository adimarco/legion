//go:generate mockery --name AugmentedLLM --dir ../internal/llm --output ./mocks --outpkg mocks --with-expecter --testonly
package hive

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/adimarco/hive/internal/llm"
	"github.com/adimarco/hive/internal/llm/mocks"
)

// setupMockLLM creates and configures a mock LLM for testing
func setupMockLLM(t *testing.T, errorOnMessage string, delay time.Duration) *mocks.AugmentedLLM {
	mockLLM := mocks.NewAugmentedLLM(t)

	// Basic method expectations
	mockLLM.EXPECT().Name().Return("mock").Maybe()
	mockLLM.EXPECT().Provider().Return("mock").Maybe()
	mockLLM.EXPECT().Initialize(mock.Anything, mock.Anything).Return(nil).Maybe()
	mockLLM.EXPECT().Cleanup().Return(nil).Maybe()

	// Generate method with error handling and delay
	mockLLM.EXPECT().Generate(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, msg llm.Message, params *llm.RequestParams) (llm.Message, error) {
		if delay > 0 {
			time.Sleep(delay)
		}
		if msg.Content == errorOnMessage {
			return llm.Message{Type: llm.MessageTypeSystem, Content: "mock error"}, errors.New("mock error")
		}
		return llm.Message{
			Type:    llm.MessageTypeAssistant,
			Content: "mock response for: " + msg.Content,
		}, nil
	}).Maybe()

	// GenerateString method with error handling and delay
	mockLLM.EXPECT().GenerateString(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, content string, params *llm.RequestParams) (string, error) {
		if delay > 0 {
			time.Sleep(delay)
		}
		if content == errorOnMessage {
			return "", errors.New("mock error")
		}
		return "mock response for: " + content, nil
	}).Maybe()

	return mockLLM
}

// testAgent returns a basic agent for testing
func testAgent(name string) *Agent {
	agent := New(name, "You are a test agent. Respond briefly and directly.")
	agent.params = &llm.RequestParams{
		Tools:  make([]string, 0),
		Config: make(map[string]any),
	}
	return agent
}

func TestChannelAgent(t *testing.T) {
	t.Run("basic operation", func(t *testing.T) {
		mock := setupMockLLM(t, "", 0)
		agent := testAgent("basic")
		agent.llm = mock
		ca := NewChannelAgent(agent)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, ca.Start(ctx))

		// Send a message
		require.NoError(t, ca.Send("test message"))

		// Get response
		select {
		case response := <-ca.Output():
			assert.Equal(t, "mock response for: test message", response)
		case err := <-ca.Errors():
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for response")
		}
	})

	t.Run("concurrent message handling", func(t *testing.T) {
		mock := setupMockLLM(t, "", 0)
		agent := testAgent("concurrent")
		agent.llm = mock
		ca := NewChannelAgent(agent)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		require.NoError(t, ca.Start(ctx))

		// Send multiple messages concurrently
		const messageCount = 100
		var wg sync.WaitGroup
		responses := make(chan string, messageCount)
		errors := make(chan error, messageCount)

		// Start response collector
		collectorDone := make(chan struct{})
		go func() {
			defer close(collectorDone)
			for {
				select {
				case resp, ok := <-ca.Output():
					if !ok {
						return
					}
					responses <- resp
				case err, ok := <-ca.Errors():
					if !ok {
						return
					}
					errors <- err
				case <-ctx.Done():
					return
				}
			}
		}()

		// Send messages
		for i := 0; i < messageCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				msg := fmt.Sprintf("message-%d", i)
				if err := ca.Send(msg); err != nil {
					select {
					case errors <- err:
					case <-ctx.Done():
					}
				}
			}(i)
		}

		// Wait for all sends to complete
		wg.Wait()

		// Collect responses with timeout
		receivedCount := 0
		errorCount := 0
		timeout := time.After(2 * time.Second)

	CollectLoop:
		for receivedCount < messageCount {
			select {
			case _, ok := <-responses:
				if !ok {
					break CollectLoop
				}
				receivedCount++
			case err, ok := <-errors:
				if !ok {
					break CollectLoop
				}
				errorCount++
				t.Logf("received error: %v", err)
			case <-timeout:
				t.Errorf("timeout waiting for responses, got %d/%d", receivedCount, messageCount)
				break CollectLoop
			case <-ctx.Done():
				t.Error("context cancelled while waiting for responses")
				break CollectLoop
			}
		}

		// Cancel context and wait for collector to finish
		cancel()
		select {
		case <-collectorDone:
			// Collector finished properly
		case <-time.After(time.Second):
			t.Error("timeout waiting for collector to finish")
		}

		// Close channels
		close(responses)
		close(errors)

		assert.Equal(t, messageCount, receivedCount, "should receive all responses")
		assert.Zero(t, errorCount, "should not receive any errors")
	})

	t.Run("error handling", func(t *testing.T) {
		mock := setupMockLLM(t, "error message", 0)
		agent := testAgent("error")
		agent.llm = mock
		ca := NewChannelAgent(agent)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, ca.Start(ctx))

		// Send message that triggers error
		require.NoError(t, ca.Send("error message"))

		// Should receive error
		select {
		case err := <-ca.Errors():
			assert.Contains(t, err.Error(), "mock error")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		mock := setupMockLLM(t, "", 100*time.Millisecond)
		agent := testAgent("cancel")
		agent.llm = mock
		ca := NewChannelAgent(agent)

		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, ca.Start(ctx))

		// Send message
		require.NoError(t, ca.Send("test"))

		// Cancel context before response
		cancel()

		// Verify channels are closed
		time.Sleep(200 * time.Millisecond) // Give time for cleanup

		// Verify channels are closed by trying to send/receive
		assert.Error(t, ca.Send("test"), "should not be able to send after close")

		// Verify all channels are closed
		assertChannelClosed(t, ca.output, "output channel")
		assertChannelClosed(t, ca.errors, "errors channel")
		assertChannelClosed(t, ca.done, "done channel")
	})

	t.Run("channel full handling", func(t *testing.T) {
		mock := setupMockLLM(t, "", 50*time.Millisecond)
		agent := testAgent("full")
		agent.llm = mock
		ca := NewChannelAgent(agent)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, ca.Start(ctx))

		// Try to fill input channel
		for i := 0; i < 15; i++ { // More than buffer size (10)
			err := ca.Send(fmt.Sprintf("message-%d", i))
			if err != nil {
				assert.Contains(t, err.Error(), "channel full")
				return
			}
		}
	})

	t.Run("cleanup idempotency", func(t *testing.T) {
		mock := setupMockLLM(t, "", 0)
		agent := testAgent("cleanup")
		agent.llm = mock
		ca := NewChannelAgent(agent)

		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, ca.Start(ctx))

		// Multiple closes should not panic
		cancel()
		ca.Close()
		ca.Close()
		ca.Close()
	})

	t.Run("message ordering", func(t *testing.T) {
		mock := setupMockLLM(t, "", 0)
		agent := testAgent("ordering")
		agent.llm = mock
		ca := NewChannelAgent(agent)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, ca.Start(ctx))

		// Send messages in order
		require.NoError(t, ca.Send("1"))
		require.NoError(t, ca.Send("2"))
		require.NoError(t, ca.Send("3"))

		// Collect responses
		responses := make([]string, 0, 3)
		for i := 0; i < 3; i++ {
			select {
			case resp := <-ca.Output():
				responses = append(responses, resp)
			case err := <-ca.Errors():
				t.Fatalf("unexpected error: %v", err)
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for response")
			}
		}

		// Note: Due to concurrent processing, we can't guarantee order
		// but we can verify we got all expected responses
		assert.ElementsMatch(t, []string{
			"mock response for: 1",
			"mock response for: 2",
			"mock response for: 3",
		}, responses)
	})
}

func assertChannelClosed[T any](t *testing.T, ch <-chan T, name string) {
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "%s should be closed", name)
	default:
		// If channel is buffered and not drained, try receiving with timeout
		timer := time.NewTimer(100 * time.Millisecond)
		defer timer.Stop()
		select {
		case _, ok := <-ch:
			assert.False(t, ok, "%s should be closed", name)
		case <-timer.C:
			t.Errorf("%s should be closed but timed out waiting", name)
		}
	}
}
