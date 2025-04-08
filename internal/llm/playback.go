package llm

import (
	"context"
	"fmt"

	"gofast/internal/config"
)

// PlaybackLLM extends PassthroughLLM to provide message playback capabilities.
// Instead of echoing messages, it plays back pre-recorded responses in sequence.
type PlaybackLLM struct {
	*PassthroughLLM
	messages     []Message
	currentIndex int
	overage      int // tracks attempts to read past the end
}

// NewPlaybackLLM creates a new PlaybackLLM instance
func NewPlaybackLLM(name string) *PlaybackLLM {
	return &PlaybackLLM{
		PassthroughLLM: NewPassthroughLLM(name),
		messages:       make([]Message, 0),
		currentIndex:   -1, // -1 indicates not initialized
		overage:        -1,
	}
}

// LoadMessages initializes the playback sequence with a set of messages
func (l *PlaybackLLM) LoadMessages(msgs []Message) {
	l.messages = msgs
	l.currentIndex = -1 // Reset to uninitialized state
	l.overage = -1
}

// getNextAssistantMessage returns the next assistant message in the sequence
func (l *PlaybackLLM) getNextAssistantMessage() Message {
	// If we've exhausted messages, return an overage message
	if l.currentIndex >= len(l.messages) {
		l.overage++
		return Message{
			Type:    MessageTypeAssistant,
			Content: fmt.Sprintf("MESSAGES EXHAUSTED (list size %d) (%d overage)", len(l.messages), l.overage),
			Name:    l.Name(),
		}
	}

	// Find next assistant message
	for l.currentIndex < len(l.messages) {
		if l.messages[l.currentIndex].Type == MessageTypeAssistant {
			msg := l.messages[l.currentIndex]
			msg.Name = l.Name() // Ensure name is set
			l.currentIndex++
			return msg
		}
		l.currentIndex++
	}

	// If we get here, no more assistant messages
	l.overage++
	return Message{
		Type:    MessageTypeAssistant,
		Content: fmt.Sprintf("MESSAGES EXHAUSTED (list size %d) (%d overage)", len(l.messages), l.overage),
		Name:    l.Name(),
	}
}

// Generate overrides PassthroughLLM to provide playback functionality
func (l *PlaybackLLM) Generate(ctx context.Context, msg Message, params *RequestParams) (Message, error) {
	// Store the user message in history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(msg, false); err != nil {
			return Message{}, fmt.Errorf("failed to add message to history: %w", err)
		}
	}

	var response Message

	// First call: initialize with messages
	if l.currentIndex == -1 {
		// Store all messages as prompts
		for _, m := range l.messages {
			if err := l.memory.Add(m, true); err != nil {
				return Message{}, fmt.Errorf("failed to add message to history: %w", err)
			}
		}

		l.currentIndex = 0
		response = Message{
			Type:    MessageTypeAssistant,
			Content: fmt.Sprintf("HISTORY LOADED (%d messages)", len(l.messages)),
			Name:    l.Name(),
		}
	} else {
		// Get next message in sequence
		response = l.getNextAssistantMessage()
	}

	// Store the response in history if using history
	if params != nil && params.UseHistory {
		if err := l.memory.Add(response, false); err != nil {
			return Message{}, fmt.Errorf("failed to add response to history: %w", err)
		}
	}

	return response, nil
}

// GenerateString overrides PassthroughLLM to use playback functionality
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

// Initialize sets up the LLM with configuration
func (l *PlaybackLLM) Initialize(ctx context.Context, cfg *config.Settings) error {
	l.currentIndex = -1 // Reset index on initialization
	l.overage = -1
	return l.PassthroughLLM.Initialize(ctx, cfg)
}

// Cleanup performs necessary cleanup
func (l *PlaybackLLM) Cleanup() error {
	l.messages = nil
	l.currentIndex = -1
	l.overage = -1
	return l.PassthroughLLM.Cleanup()
}
