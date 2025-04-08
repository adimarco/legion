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

// SimpleMemory provides a basic thread-safe in-memory implementation of Memory.
// This implementation is suitable for:
// - Development and testing scenarios
// - Short-lived conversations
// - Single-instance deployments
//
// For production deployments with specific requirements (persistence,
// distributed storage, etc.), you should implement a custom Memory
// implementation.
type SimpleMemory struct {
	mu             sync.RWMutex
	history        []Message // Regular conversation history
	promptMessages []Message // System prompts and other context
}

// NewSimpleMemory creates a new SimpleMemory instance.
// The memory starts empty and can be populated through the Add method.
// Both history slices are initialized with zero capacity since we
// can't predict the conversation length.
func NewSimpleMemory() *SimpleMemory {
	return &SimpleMemory{
		history:        make([]Message, 0),
		promptMessages: make([]Message, 0),
	}
}

// Add adds a message to history.
// Messages are categorized into two types:
// 1. Prompt messages (isPrompt=true): System prompts, permanent context
// 2. History messages (isPrompt=false): Regular conversation messages
//
// This separation allows for:
// - Clearing conversation history while preserving prompts
// - Selectively including/excluding history in requests
// - Different retention policies for different message types
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

// Get retrieves messages from memory.
// The includeHistory parameter controls whether conversation history
// is included in the result:
// - true: Returns prompt messages + history messages
// - false: Returns only prompt messages
//
// The returned slice is a deep copy of the stored messages to prevent
// external modifications to the internal state.
func (m *SimpleMemory) Get(includeHistory bool) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Always include prompt messages
	result := make([]Message, len(m.promptMessages))
	for i, msg := range m.promptMessages {
		result[i] = deepCopyMessage(msg)
	}

	if includeHistory {
		// Append copies of history messages
		for _, msg := range m.history {
			result = append(result, deepCopyMessage(msg))
		}
	}

	return result, nil
}

// deepCopyMessage creates a deep copy of a Message
func deepCopyMessage(msg Message) Message {
	copy := Message{
		Type:    msg.Type,
		Content: msg.Content,
		Name:    msg.Name,
	}

	// Deep copy metadata
	if msg.Metadata != nil {
		copy.Metadata = make(map[string]any, len(msg.Metadata))
		for k, v := range msg.Metadata {
			copy.Metadata[k] = v
		}
	}

	// Deep copy tool calls
	if len(msg.ToolCalls) > 0 {
		copy.ToolCalls = make([]ToolCall, len(msg.ToolCalls))
		for i, call := range msg.ToolCalls {
			copy.ToolCalls[i] = ToolCall{
				ID:       call.ID,
				Name:     call.Name,
				Response: call.Response,
			}
			if call.Args != nil {
				copy.ToolCalls[i].Args = make(map[string]any, len(call.Args))
				for k, v := range call.Args {
					copy.ToolCalls[i].Args[k] = v
				}
			}
		}
	}

	// Deep copy message parts
	if len(msg.Parts) > 0 {
		copy.Parts = make([]MessagePart, len(msg.Parts))
		for i, part := range msg.Parts {
			copy.Parts[i] = MessagePart{
				Type:    part.Type,
				Content: part.Content,
			}
			if part.Data != nil {
				copy.Parts[i].Data = make(map[string]any, len(part.Data))
				for k, v := range part.Data {
					copy.Parts[i].Data[k] = v
				}
			}
		}
	}

	return copy
}

// Clear clears the specified message types.
// The clearPrompts parameter controls what gets cleared:
// - true: Clears both history and prompt messages
// - false: Clears only history messages
//
// This flexibility allows for:
// - Starting new conversations while keeping context
// - Complete reset of all state
// - Clearing history without affecting system prompts
func (m *SimpleMemory) Clear(clearPrompts bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = make([]Message, 0)
	if clearPrompts {
		m.promptMessages = make([]Message, 0)
	}
	return nil
}
