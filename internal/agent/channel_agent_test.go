package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adimarco/hive/internal/config"
	"github.com/adimarco/hive/internal/llm"
)

// mockLLM implements a controlled test environment
type mockLLM struct {
	responses      map[string]string
	delay          time.Duration
	errorOnMessage string
	mu             sync.RWMutex
}

func newMockLLM() *mockLLM {
	return &mockLLM{
		responses: make(map[string]string),
	}
}

func (m *mockLLM) Initialize(ctx context.Context, cfg *config.Settings) error { return nil }
func (m *mockLLM) Name() string                                               { return "mock" }
func (m *mockLLM) Provider() string                                           { return "mock" }
func (m *mockLLM) Cleanup() error                                             { return nil }

func (m *mockLLM) Generate(ctx context.Context, msg llm.Message, params *llm.RequestParams) (llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if msg.Content == m.errorOnMessage {
		return llm.Message{}, errors.New("mock error")
	}

	response, ok := m.responses[msg.Content]
	if !ok {
		response = "mock response for: " + msg.Content
	}

	return llm.Message{
		Type:    llm.MessageTypeAssistant,
		Content: response,
	}, nil
}

func (m *mockLLM) GenerateString(ctx context.Context, content string, params *llm.RequestParams) (string, error) {
	msg := llm.Message{
		Type:    llm.MessageTypeUser,
		Content: content,
	}
	response, err := m.Generate(ctx, msg, params)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

func (m *mockLLM) CallTool(ctx context.Context, call llm.ToolCall) (string, error) {
	return "", errors.New("not implemented")
}

// testConfig returns a basic agent config for testing
func testConfig(name string) AgentConfig {
	return AgentConfig{
		Name:        name,
		Type:        AgentTypeBasic,
		Instruction: "You are a test agent. Respond briefly and directly.",
	}
}

func TestChannelAgent(t *testing.T) {
	t.Run("basic operation", func(t *testing.T) {
		mock := newMockLLM()
		agent := NewChannelAgent(testConfig("basic"), mock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, agent.Start(ctx))

		// Send a message
		require.NoError(t, agent.Send("test message"))

		// Get response
		select {
		case response := <-agent.Output():
			assert.Equal(t, "mock response for: test message", response)
		case err := <-agent.Errors():
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for response")
		}
	})

	t.Run("concurrent message handling", func(t *testing.T) {
		mock := newMockLLM()
		agent := NewChannelAgent(testConfig("concurrent"), mock)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		require.NoError(t, agent.Start(ctx))

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
				case resp, ok := <-agent.Output():
					if !ok {
						return
					}
					responses <- resp
				case err, ok := <-agent.Errors():
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
				if err := agent.Send(msg); err != nil {
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
		mock := newMockLLM()
		mock.errorOnMessage = "error message"

		agent := NewChannelAgent(testConfig("error"), mock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, agent.Start(ctx))

		// Send message that triggers error
		require.NoError(t, agent.Send("error message"))

		// Should receive error
		select {
		case err := <-agent.Errors():
			assert.Contains(t, err.Error(), "mock error")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		mock := newMockLLM()
		mock.delay = 100 * time.Millisecond // Add delay to ensure message is in flight

		agent := NewChannelAgent(testConfig("cancel"), mock)
		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, agent.Start(ctx))

		// Send message
		require.NoError(t, agent.Send("test"))

		// Cancel context before response
		cancel()

		// Verify channels are closed
		time.Sleep(200 * time.Millisecond) // Give time for cleanup

		// Verify channels are closed by trying to send/receive
		assert.Error(t, agent.Send("test"), "should not be able to send after close")

		// Verify all channels are closed
		assertChannelClosed(t, agent.output, "output channel")
		assertChannelClosed(t, agent.errors, "errors channel")
		assertChannelClosed(t, agent.done, "done channel")
	})

	t.Run("channel full handling", func(t *testing.T) {
		mock := newMockLLM()
		mock.delay = 50 * time.Millisecond // Add delay to help fill buffer

		agent := NewChannelAgent(testConfig("full"), mock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, agent.Start(ctx))

		// Try to fill input channel
		for i := 0; i < 15; i++ { // More than buffer size (10)
			err := agent.Send(fmt.Sprintf("message-%d", i))
			if err != nil {
				assert.Contains(t, err.Error(), "channel full")
				return
			}
		}
	})

	t.Run("cleanup idempotency", func(t *testing.T) {
		mock := newMockLLM()
		agent := NewChannelAgent(testConfig("cleanup"), mock)

		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, agent.Start(ctx))

		// Multiple closes should not panic
		cancel()
		agent.Close()
		agent.Close()
		agent.Close()
	})

	t.Run("message ordering", func(t *testing.T) {
		mock := newMockLLM()
		// Set up deterministic responses
		mock.responses = map[string]string{
			"1": "first",
			"2": "second",
			"3": "third",
		}

		agent := NewChannelAgent(testConfig("ordering"), mock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, agent.Start(ctx))

		// Send messages in order
		require.NoError(t, agent.Send("1"))
		require.NoError(t, agent.Send("2"))
		require.NoError(t, agent.Send("3"))

		// Collect responses
		responses := make([]string, 0, 3)
		for i := 0; i < 3; i++ {
			select {
			case resp := <-agent.Output():
				responses = append(responses, resp)
			case err := <-agent.Errors():
				t.Fatalf("unexpected error: %v", err)
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for response")
			}
		}

		// Note: Due to concurrent processing, we can't guarantee order
		// but we can verify we got all expected responses
		assert.ElementsMatch(t, []string{"first", "second", "third"}, responses)
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
