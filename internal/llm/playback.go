/*
Package llm provides advanced testing utilities for LLM interactions.

The PlaybackLLM extends PassthroughLLM to provide sophisticated message
playback capabilities. It supports two main modes of operation:

 1. Response Recording: Using the !record command, you can associate specific
    trigger phrases with predetermined responses. This allows for testing
    how agents handle different responses in various contexts.

 2. Message Sequence Playback: You can pre-load a sequence of messages that
    will be played back in order, simulating a real conversation flow.

Key Features:
- Trigger-based responses for testing specific scenarios
- Ordered message playback for testing conversation flow
- Exhaustion handling to detect when tests exceed expected messages
- Full history management and logging
- Tool call support inherited from PassthroughLLM

This implementation is particularly useful for:
- Integration testing of complex agent workflows
- Regression testing with recorded conversations
- Testing error handling and edge cases
- Simulating specific conversation patterns
*/
package llm

import (
	"context"
	"fmt"
	"strings"

	"gofast/internal/config"
	"gofast/internal/logging"
)

const (
	// RECORD_RESPONSE_INDICATOR records a trigger-response pair
	// Format: !record <trigger> <response>
	RECORD_RESPONSE_INDICATOR = "!record"

	// START_PLAYBACK_INDICATOR starts message sequence playback
	START_PLAYBACK_INDICATOR = "!start"
)

// PlaybackLLM extends PassthroughLLM to provide message playback capabilities.
// Instead of echoing messages, it plays back pre-recorded responses in sequence
// or responds based on trigger phrases.
//
// The implementation supports two modes:
// 1. Trigger-based: Responses are played based on message content matching
// 2. Sequential: Messages are played back in a predefined order
type PlaybackLLM struct {
	*PassthroughLLM                   // Inherit PassthroughLLM functionality
	messages        []Message         // Pre-loaded message sequence
	currentIndex    int               // Current position in message sequence
	overage         int               // Tracks attempts to read past the end
	name            string            // Instance identifier
	memory          Memory            // Conversation history
	cfg             *config.Settings  // Configuration
	defaults        *RequestParams    // Default parameters
	logger          logging.Logger    // Structured logging
	responses       map[string]string // Trigger-based responses
	playbackStarted bool              // Whether playback has been started
}

// NewPlaybackLLM creates a new PlaybackLLM instance.
// The name parameter is used to identify this instance in logs
// and responses, making it easier to track in multi-LLM scenarios.
func NewPlaybackLLM(name string) *PlaybackLLM {
	llm := &PlaybackLLM{
		PassthroughLLM: NewPassthroughLLM(name),
		name:           name,
		memory:         NewSimpleMemory(),
		responses:      make(map[string]string),
	}
	llm.logger = logging.GetLogger("llm.playback")
	return llm
}

// LoadMessages initializes the playback sequence with a set of messages.
// This method resets the playback state and prepares for sequential
// message delivery. The messages will be played back in order after
// receiving the !start command.
func (l *PlaybackLLM) LoadMessages(msgs []Message) {
	l.messages = msgs
	l.currentIndex = 0 // Initialize to start of sequence
	l.overage = 0
	l.playbackStarted = false
}

// getNextAssistantMessage returns the next assistant message in the sequence.
// This internal method handles:
// - Finding the next assistant message in the sequence
// - Tracking sequence position
// - Generating exhaustion messages when we run out
// - Maintaining the overage count
func (l *PlaybackLLM) getNextAssistantMessage() Message {
	// If playback hasn't started, return empty message
	if !l.playbackStarted {
		return Message{
			Type:    MessageTypeAssistant,
			Content: "Playback not started. Use !start to begin playback.",
			Name:    l.name,
		}
	}

	// Find next assistant message
	for l.currentIndex < len(l.messages) {
		msg := l.messages[l.currentIndex]
		l.currentIndex++
		if msg.Type == MessageTypeAssistant {
			l.logger.Debug(context.Background(), "Playing back message", logging.WithData(map[string]interface{}{
				"content": msg.Content,
				"index":   l.currentIndex - 1,
			}))
			return msg
		}
	}

	// If no more messages, return exhausted message
	l.overage++
	l.logger.Debug(context.Background(), "Message sequence exhausted", logging.WithData(map[string]interface{}{
		"overage": l.overage,
	}))
	return Message{
		Type:    MessageTypeAssistant,
		Content: fmt.Sprintf("MESSAGES EXHAUSTED (list size %d) (%d overage)", len(l.messages), l.overage-1),
		Name:    l.name,
	}
}

// Generate processes a message and returns a pre-recorded response.
// The response selection follows this priority:
// 1. Record Response Command: Stores a new trigger-response pair
// 2. Start Playback Command: Begins message sequence playback
// 3. Trigger Matching: Returns response for matching trigger phrase
// 4. Message Playback: Returns next message in sequence if started
// 5. Default Behavior: Falls back to PassthroughLLM echo
func (l *PlaybackLLM) Generate(ctx context.Context, msg Message, params *RequestParams) (Message, error) {
	// Log user message
	l.logger.Info(ctx, "Received user message", logging.WithData(map[string]interface{}{
		"content": msg.Content,
		"type":    msg.Type,
	}))

	// Store the user message in history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(msg, false); err != nil {
			l.logger.Error(ctx, "Failed to add message to history", logging.WithData(map[string]interface{}{
				"error": err.Error(),
			}))
			return Message{}, fmt.Errorf("failed to add message to history: %w", err)
		}
	}

	var response Message

	// Check for special commands
	if strings.HasPrefix(msg.Content, RECORD_RESPONSE_INDICATOR) {
		content := strings.TrimSpace(strings.TrimPrefix(msg.Content, RECORD_RESPONSE_INDICATOR))
		if content == "" {
			l.logger.Error(ctx, "Invalid record response command", logging.WithData(map[string]interface{}{
				"content": msg.Content,
			}))
			return Message{}, fmt.Errorf("invalid record response command: %s", msg.Content)
		}

		// Split remaining content into trigger and response
		parts := strings.SplitN(content, " ", 2)
		if len(parts) != 2 {
			l.logger.Error(ctx, "Invalid record response command", logging.WithData(map[string]interface{}{
				"content": msg.Content,
			}))
			return Message{}, fmt.Errorf("invalid record response command: %s", msg.Content)
		}

		trigger := strings.TrimSpace(parts[0])
		responseText := strings.TrimSpace(parts[1])
		l.responses[trigger] = responseText

		l.logger.Debug(ctx, "Recorded response", logging.WithData(map[string]interface{}{
			"trigger":  trigger,
			"response": responseText,
		}))

		response = Message{
			Type:    MessageTypeAssistant,
			Content: fmt.Sprintf("Recorded response for trigger: %s", trigger),
			Name:    l.name,
		}
	} else if msg.Content == START_PLAYBACK_INDICATOR {
		l.playbackStarted = true
		response = Message{
			Type:    MessageTypeAssistant,
			Content: fmt.Sprintf("Starting playback of %d messages", len(l.messages)),
			Name:    l.name,
		}
	} else {
		// Check for trigger matches
		matched := false
		for trigger, responseText := range l.responses {
			if strings.Contains(strings.ToLower(msg.Content), strings.ToLower(trigger)) {
				l.logger.Debug(ctx, "Found matching trigger", logging.WithData(map[string]interface{}{
					"trigger":  trigger,
					"message":  msg.Content,
					"response": responseText,
				}))

				// If response is a tool call, handle it through PassthroughLLM
				if strings.HasPrefix(responseText, "***CALL_TOOL") {
					toolMsg := Message{
						Type:    MessageTypeUser,
						Content: responseText,
						Name:    l.name,
					}
					return l.PassthroughLLM.Generate(ctx, toolMsg, params)
				}

				response = Message{
					Type:    MessageTypeAssistant,
					Content: responseText,
					Name:    l.name,
				}
				matched = true
				break
			}
		}

		// If no trigger match and playback started, return next message
		if !matched {
			if l.playbackStarted {
				response = l.getNextAssistantMessage()
			} else {
				response = Message{
					Type:    MessageTypeAssistant,
					Content: msg.Content,
					Name:    l.name,
				}
			}
		}
	}

	// Add response to history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(response, false); err != nil {
			l.logger.Error(ctx, "Failed to add response to history", logging.WithData(map[string]interface{}{
				"error": err.Error(),
			}))
			return Message{}, fmt.Errorf("failed to add response to history: %w", err)
		}
	}

	return response, nil
}

// GenerateString overrides PassthroughLLM to use playback functionality.
// This is a convenience method for simple testing scenarios where
// you don't need the full Message structure.
func (l *PlaybackLLM) GenerateString(ctx context.Context, content string, params *RequestParams) (string, error) {
	msg := Message{
		Type:    MessageTypeUser,
		Content: content,
	}
	response, err := l.Generate(ctx, msg, params)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// Provider returns the provider name
func (l *PlaybackLLM) Provider() string {
	return "playback"
}

// Initialize sets up the LLM with configuration.
// This resets the playback state to ensure a clean starting point.
func (l *PlaybackLLM) Initialize(ctx context.Context, cfg *config.Settings) error {
	l.currentIndex = 0
	l.overage = 0
	l.playbackStarted = false
	l.responses = make(map[string]string)
	return l.PassthroughLLM.Initialize(ctx, cfg)
}

// Cleanup performs necessary cleanup.
// This resets all state, including:
// - Message sequence
// - Playback position
// - Overage counter
// - Recorded responses
// - Inherited PassthroughLLM state
func (l *PlaybackLLM) Cleanup() error {
	l.messages = nil
	l.currentIndex = 0
	l.overage = 0
	l.playbackStarted = false
	l.responses = make(map[string]string)
	return l.PassthroughLLM.Cleanup()
}
