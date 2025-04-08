package llm

import (
	"sync"
)

// SimpleMemory provides a basic thread-safe in-memory implementation of Memory
type SimpleMemory struct {
	mu             sync.RWMutex
	history        []Message
	promptMessages []Message
}

// NewSimpleMemory creates a new SimpleMemory instance
func NewSimpleMemory() *SimpleMemory {
	return &SimpleMemory{
		history:        make([]Message, 0),
		promptMessages: make([]Message, 0),
	}
}

// Add adds a message to history
func (m *SimpleMemory) Add(msg Message, isPrompt bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if isPrompt {
		m.promptMessages = append(m.promptMessages, msg)
	} else {
		m.history = append(m.history, msg)
	}
	return nil
}

// Get retrieves messages from memory
func (m *SimpleMemory) Get(includeHistory bool) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Always include prompt messages
	result := make([]Message, len(m.promptMessages))
	copy(result, m.promptMessages)

	if includeHistory {
		result = append(result, m.history...)
	}

	return result, nil
}

// Clear clears the specified message types
func (m *SimpleMemory) Clear(clearPrompts bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = make([]Message, 0)
	if clearPrompts {
		m.promptMessages = make([]Message, 0)
	}
	return nil
}
