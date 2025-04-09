/*
Package llm provides memory management for LLM conversations.

The memory system is designed around these key principles:
1. Thread Safety: All operations are protected by mutexes for concurrent access
2. Separation of Concerns: Prompt messages are stored separately from conversation history
3. Flexibility: The simple interface allows for different implementations
4. Performance: In-memory storage provides fast access for active conversations
*/
package llm

// SimpleMemory is a basic in-memory implementation of the Memory interface
type SimpleMemory struct {
	messages []Message
	prompts  []Message
}

// NewSimpleMemory creates a new SimpleMemory instance
func NewSimpleMemory() *SimpleMemory {
	return &SimpleMemory{
		messages: make([]Message, 0),
		prompts:  make([]Message, 0),
	}
}

// Add adds a message to history
func (m *SimpleMemory) Add(msg Message, isPrompt bool) error {
	if isPrompt {
		m.prompts = append(m.prompts, msg)
	} else {
		m.messages = append(m.messages, msg)
	}
	return nil
}

// Get retrieves messages from memory
func (m *SimpleMemory) Get(includeHistory bool) ([]Message, error) {
	result := make([]Message, len(m.prompts))
	copy(result, m.prompts)

	if includeHistory {
		result = append(result, m.messages...)
	}

	return result, nil
}

// Clear clears the specified message types
func (m *SimpleMemory) Clear(clearPrompts bool) error {
	m.messages = make([]Message, 0)
	if clearPrompts {
		m.prompts = make([]Message, 0)
	}
	return nil
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
