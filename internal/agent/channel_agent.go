package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/adimarco/hive/internal/llm"
)

// ChannelAgent extends Agent with channel-based message handling
type ChannelAgent struct {
	*Agent                  // Embed base Agent
	input     chan string   // Channel for incoming messages
	output    chan string   // Channel for responses
	done      chan struct{} // Channel for shutdown signaling
	errors    chan error    // Channel for error reporting
	closeOnce sync.Once     // Ensure cleanup happens once
	closed    bool          // Track closed state
	mu        sync.RWMutex  // Protect closed state
}

// NewChannelAgent creates a new ChannelAgent with the given configuration
func NewChannelAgent(config AgentConfig, llm llm.AugmentedLLM) *ChannelAgent {
	return &ChannelAgent{
		Agent:  NewAgent(config, llm),
		input:  make(chan string, 100), // Increased buffer for high concurrency
		output: make(chan string, 100), // Increased buffer for high concurrency
		done:   make(chan struct{}),    // Unbuffered for clean shutdown
		errors: make(chan error, 100),  // Increased buffer for high concurrency
	}
}

// Start begins processing messages in a separate goroutine
func (ca *ChannelAgent) Start(ctx context.Context) error {
	// Create running agent
	ra, err := ca.Agent.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Start message processing loop
	go ca.processMessages(ctx, ra)

	return nil
}

// processMessages handles the main message processing loop
func (ca *ChannelAgent) processMessages(ctx context.Context, ra *RunningAgent) {
	defer ca.Close()

	// Use WaitGroup to track in-flight messages
	var wg sync.WaitGroup
	defer wg.Wait() // Wait for all messages to complete before closing

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, mark as closed and wait for in-flight messages
			ca.mu.Lock()
			ca.closed = true
			ca.mu.Unlock()
			return

		case msg, ok := <-ca.input:
			if !ok {
				// Input channel closed
				return
			}

			// Check if we're closed before processing
			ca.mu.RLock()
			closed := ca.closed
			ca.mu.RUnlock()
			if closed {
				return
			}

			// Process message in separate goroutine
			wg.Add(1)
			go func(message string) {
				defer wg.Done()

				// Check context before processing
				select {
				case <-ctx.Done():
					return
				default:
				}

				response, err := ra.Send(message)
				if err != nil {
					// Try to send error, don't block if channel is full
					select {
					case ca.errors <- fmt.Errorf("failed to process message: %w", err):
					case <-ctx.Done():
					default:
						// Error channel full, log or handle appropriately
					}
					return
				}

				// Try to send response, respect context and channel state
				select {
				case <-ctx.Done():
					return
				default:
					ca.mu.RLock()
					closed := ca.closed
					ca.mu.RUnlock()

					if !closed {
						select {
						case ca.output <- response:
							// Response sent successfully
						case <-ctx.Done():
							return
						default:
							// Output channel full, send error
							select {
							case ca.errors <- fmt.Errorf("output channel full"):
							case <-ctx.Done():
							default:
								// Error channel full, log or handle appropriately
							}
						}
					}
				}
			}(msg)
		}
	}
}

// Send queues a message for processing
// Returns immediately, responses come through the Output() channel
func (ca *ChannelAgent) Send(msg string) error {
	// Check if agent is closed
	ca.mu.RLock()
	closed := ca.closed
	ca.mu.RUnlock()

	if closed {
		return fmt.Errorf("agent is closed")
	}

	// Try to send message, don't block if channel is full
	select {
	case ca.input <- msg:
		return nil
	default:
		return fmt.Errorf("input channel full")
	}
}

// Input returns the channel for sending messages
func (ca *ChannelAgent) Input() chan<- string {
	return ca.input
}

// Output returns the channel for receiving responses
func (ca *ChannelAgent) Output() <-chan string {
	return ca.output
}

// Errors returns the channel for receiving errors
func (ca *ChannelAgent) Errors() <-chan error {
	return ca.errors
}

// Done returns a channel that's closed when the agent stops
func (ca *ChannelAgent) Done() <-chan struct{} {
	return ca.done
}

// Close shuts down the agent and closes all channels
func (ca *ChannelAgent) Close() {
	ca.closeOnce.Do(func() {
		// Mark as closed first to prevent new messages
		ca.mu.Lock()
		ca.closed = true
		ca.mu.Unlock()

		// Close channels in order:
		// 1. input - stops new messages
		// 2. done - signals shutdown
		close(ca.input)
		close(ca.done)

		// Wait a short time for in-flight messages to complete
		time.Sleep(100 * time.Millisecond)

		// Close remaining channels
		close(ca.errors)
		close(ca.output)
	})
}
