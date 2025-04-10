package llm

// ToolCall represents a request to call a tool.
// The design supports both synchronous and asynchronous tool execution,
// with the Response field allowing for result storage.
type ToolCall struct {
	// ID uniquely identifies this tool call
	ID string `json:"id"`
	// Name identifies which tool to call
	Name string `json:"name"`
	// Args holds the parameters for the tool call
	Args map[string]any `json:"args"`
	// Response stores the result of the tool call
	Response string `json:"response,omitempty"`
}
