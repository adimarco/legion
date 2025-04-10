/*
Package llm provides memory management for LLM conversations.

The memory system is designed around these key principles:
1. Thread Safety: All operations are protected by mutexes for concurrent access
2. Separation of Concerns: Prompt messages are stored separately from conversation history
3. Flexibility: The simple interface allows for different implementations
4. Performance: In-memory storage provides fast access for active conversations
*/
package llm

import (
	"sync"
)

// Memory manages conversation history and prompt storage
type Memory interface {
	// Add adds a message to history
	Add(msg Message, isPrompt bool) error

	// Get retrieves messages from memory
	Get(includeHistory bool) ([]Message, error)

	// Clear clears the specified message types
	Clear(clearPrompts bool) error
}

// SimpleMemory provides a basic thread-safe in-memory implementation
type SimpleMemory struct {
	mu      sync.RWMutex
	history []Message
	prompts []Message
}

// NewSimpleMemory creates a new SimpleMemory instance
func NewSimpleMemory() *SimpleMemory {
	return &SimpleMemory{
		history: make([]Message, 0),
		prompts: make([]Message, 0),
	}
}

// Add adds a message to history
func (m *SimpleMemory) Add(msg Message, isPrompt bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a deep copy of the message
	msgCopy := deepCopyMessage(msg)

	if isPrompt {
		m.prompts = append(m.prompts, msgCopy)
	} else {
		m.history = append(m.history, msgCopy)
	}
	return nil
}

// Get retrieves messages from memory
func (m *SimpleMemory) Get(includeHistory bool) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Start with prompt messages
	result := make([]Message, len(m.prompts))
	for i, msg := range m.prompts {
		result[i] = deepCopyMessage(msg)
	}

	// Add history if requested
	if includeHistory {
		for _, msg := range m.history {
			result = append(result, deepCopyMessage(msg))
		}
	}

	return result, nil
}

// Clear clears the specified message types
func (m *SimpleMemory) Clear(clearPrompts bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = make([]Message, 0)
	if clearPrompts {
		m.prompts = make([]Message, 0)
	}
	return nil
}

// deepCopyMessage creates a deep copy of a Message
func deepCopyMessage(msg Message) Message {
	return Message{
		Type:      msg.Type,
		Content:   msg.Content,
		Name:      msg.Name,
		Metadata:  deepCopyMap(msg.Metadata),
		ToolCalls: deepCopyToolCalls(msg.ToolCalls),
		Parts:     deepCopyMessageParts(msg.Parts),
	}
}

// deepCopyMap creates a deep copy of a map[string]any
func deepCopyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	copy := make(map[string]any, len(m))
	for k, v := range m {
		copy[k] = v // Note: This assumes values don't need deep copying
	}
	return copy
}

// deepCopyToolCalls creates a deep copy of a slice of ToolCalls
func deepCopyToolCalls(calls []ToolCall) []ToolCall {
	if len(calls) == 0 {
		return nil
	}
	copy := make([]ToolCall, len(calls))
	for i, call := range calls {
		copy[i] = ToolCall{
			ID:       call.ID,
			Name:     call.Name,
			Response: call.Response,
			Args:     deepCopyMap(call.Args),
		}
	}
	return copy
}

// deepCopyMessageParts creates a deep copy of a slice of MessageParts
func deepCopyMessageParts(parts []MessagePart) []MessagePart {
	if len(parts) == 0 {
		return nil
	}
	copy := make([]MessagePart, len(parts))
	for i, part := range parts {
		copy[i] = MessagePart{
			Type:    part.Type,
			Content: part.Content,
			Data:    deepCopyMap(part.Data),
		}
	}
	return copy
}
