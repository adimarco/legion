package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/adimarco/hive/internal/config"
	"github.com/adimarco/hive/internal/logging"
)

// AnthropicLLM implements the AugmentedLLM interface using Anthropic's Claude API
type AnthropicLLM struct {
	client   *anthropic.Client
	name     string
	memory   Memory
	logger   logging.Logger
	defaults *RequestParams
}

// NewAnthropicLLM creates a new AnthropicLLM instance
func NewAnthropicLLM(name string) *AnthropicLLM {
	return &AnthropicLLM{
		name:   name,
		memory: NewSimpleMemory(),
		logger: logging.GetLogger("llm.anthropic"),
		defaults: &RequestParams{
			Model:         "claude-3-7-sonnet-latest",
			UseHistory:    true,
			ParallelTools: true,
			MaxIterations: 10,
			MaxTokens:     1024,
		},
	}
}

// Initialize sets up the LLM with configuration
func (l *AnthropicLLM) Initialize(ctx context.Context, cfg *config.Settings) error {
	// For now, we'll use an environment variable. Later we can add proper config
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	l.client = &client
	return nil
}

// Generate processes a message and returns a response
func (l *AnthropicLLM) Generate(ctx context.Context, msg Message, params *RequestParams) (Message, error) {
	l.logger.Info(ctx, "Generating response", logging.WithData(map[string]interface{}{
		"content": msg.Content,
		"type":    msg.Type,
	}))

	// Store user message in history if enabled
	if params != nil && params.UseHistory {
		if err := l.memory.Add(msg, false); err != nil {
			return Message{}, fmt.Errorf("failed to add message to history: %w", err)
		}
	}

	// Build message list including history if needed
	var messages []anthropic.MessageParam
	if params != nil && params.UseHistory {
		history, err := l.memory.Get(true)
		if err != nil {
			return Message{}, err
		}
		messages = convertToAnthropicMessages(history)
	}

	// Add current message
	messages = append(messages, anthropic.MessageParam{
		Role: anthropic.MessageParamRoleUser,
		Content: []anthropic.ContentBlockParamUnion{
			{
				OfRequestTextBlock: &anthropic.TextBlockParam{
					Text: msg.Content,
				},
			},
		},
	})

	// Prepare request parameters
	reqParams := l.defaults
	if params != nil {
		reqParams = params
	}

	// Create message request
	req := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		Messages:  messages,
		MaxTokens: 1024,
	}

	if reqParams.SystemPrompt != "" {
		req.System = []anthropic.TextBlockParam{
			{
				Text: reqParams.SystemPrompt,
			},
		}
	}
	if reqParams.Temperature > 0 {
		req.Temperature = anthropic.Float(float64(reqParams.Temperature))
	}

	// Make API call
	resp, err := l.client.Messages.New(ctx, req)
	if err != nil {
		l.logger.Error(ctx, "Anthropic API error", logging.WithData(map[string]interface{}{
			"error": err.Error(),
		}))
		return Message{}, fmt.Errorf("Anthropic API error: %w", err)
	}

	// Convert response
	response := Message{
		Type:    MessageTypeAssistant,
		Content: resp.Content[0].Text,
		Name:    l.name,
	}

	// Store response in history if enabled
	if params != nil && params.UseHistory {
		if err := l.memory.Add(response, false); err != nil {
			return Message{}, fmt.Errorf("failed to add response to history: %w", err)
		}
	}

	l.logger.Info(ctx, "Generated response", logging.WithData(map[string]interface{}{
		"content": response.Content,
	}))

	return response, nil
}

// GenerateString is a convenience method for simple text interactions
func (l *AnthropicLLM) GenerateString(ctx context.Context, content string, params *RequestParams) (string, error) {
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

// CallTool executes a tool call and returns the result
func (l *AnthropicLLM) CallTool(ctx context.Context, call ToolCall) (string, error) {
	// For now, we'll just format the call details
	// In a future implementation, we can properly integrate with Anthropic's tool calling
	args := "no arguments"
	if len(call.Args) > 0 {
		argsJSON, err := json.Marshal(call.Args)
		if err != nil {
			args = fmt.Sprintf("error formatting arguments: %v", err)
		} else {
			args = string(argsJSON)
		}
	}
	return fmt.Sprintf("Tool call: %s\nArguments:\n%s", call.Name, args), nil
}

// Name returns the identifier for this LLM instance
func (l *AnthropicLLM) Name() string {
	return l.name
}

// Provider returns the LLM provider
func (l *AnthropicLLM) Provider() string {
	return "anthropic"
}

// Cleanup performs any necessary cleanup
func (l *AnthropicLLM) Cleanup() error {
	return l.memory.Clear(true)
}

// Helper functions

func convertToAnthropicMessages(msgs []Message) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, len(msgs))
	for i, msg := range msgs {
		role := anthropic.MessageParamRoleUser
		switch msg.Type {
		case MessageTypeAssistant:
			role = anthropic.MessageParamRoleAssistant
		case MessageTypeSystem:
			// System messages are handled differently in Anthropic's API
			continue
		}
		result[i] = anthropic.MessageParam{
			Role: role,
			Content: []anthropic.ContentBlockParamUnion{
				{
					OfRequestTextBlock: &anthropic.TextBlockParam{
						Text: msg.Content,
					},
				},
			},
		}
	}
	return result
}
