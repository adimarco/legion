package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/adimarco/hive/config"
	"github.com/adimarco/hive/logging"
	"github.com/adimarco/hive/tools"
)

// AnthropicLLM implements the AugmentedLLM interface using Anthropic's Claude API
type AnthropicLLM struct {
	client   *anthropic.Client
	name     string
	memory   Memory
	logger   logging.Logger
	defaults *RequestParams
	tools    *tools.SimpleToolRegistry
}

// NewAnthropicLLM creates a new AnthropicLLM instance
func NewAnthropicLLM(name string) *AnthropicLLM {
	return &AnthropicLLM{
		name:   name,
		memory: NewSimpleMemory(),
		logger: logging.GetLogger("llm.anthropic"),
		tools:  tools.NewSimpleToolRegistry(),
		defaults: &RequestParams{
			Model:         "claude-3-haiku-20240307",
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
		Model:     reqParams.Model,
		Messages:  messages,
		MaxTokens: int64(reqParams.MaxTokens),
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

	// Add tools if specified
	if len(reqParams.Tools) > 0 {
		// Track unique tools by their name
		uniqueToolParams := make(map[string]anthropic.ToolParam)

		for _, toolName := range reqParams.Tools {
			tool, err := l.tools.Get(toolName)
			if err != nil {
				l.logger.Error(ctx, "Tool not found", logging.WithData(map[string]interface{}{
					"tool":  toolName,
					"error": err.Error(),
				}))
				continue
			}

			// Skip if we already added this tool name
			if _, exists := uniqueToolParams[tool.Name]; exists {
				l.logger.Info(ctx, "Skipping duplicate tool", logging.WithData(map[string]interface{}{
					"tool": tool.Name,
				}))
				continue
			}

			// Convert tool schema to Anthropic's format
			var schemaMap map[string]interface{}
			if err := json.Unmarshal(tool.Schema, &schemaMap); err != nil {
				l.logger.Error(ctx, "Failed to parse tool schema", logging.WithData(map[string]interface{}{
					"tool":  toolName,
					"error": err.Error(),
				}))
				continue
			}

			// Create tool parameter using just the tool name (no version)
			uniqueToolParams[tool.Name] = anthropic.ToolParam{
				Name:        tool.Name, // Use just the name without version
				Description: anthropic.String(tool.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Type:       "object",
					Properties: schemaMap["properties"].(map[string]interface{}),
				},
			}

			l.logger.Info(ctx, "Added tool", logging.WithData(map[string]interface{}{
				"tool":   tool.Name,
				"schema": tool.Schema,
			}))
		}

		// Convert the unique tools to the format Anthropic expects
		toolParams := make([]anthropic.ToolParam, 0, len(uniqueToolParams))
		for _, param := range uniqueToolParams {
			toolParams = append(toolParams, param)
		}

		if len(toolParams) > 0 {
			// Convert to union type and add to request
			tools := make([]anthropic.ToolUnionParam, len(toolParams))
			for i, param := range toolParams {
				tools[i] = anthropic.ToolUnionParam{OfTool: &param}
			}
			req.Tools = tools

			l.logger.Info(ctx, "Request with tools", logging.WithData(map[string]interface{}{
				"tools_count": len(tools),
				"tool_names": func() []string {
					names := make([]string, len(toolParams))
					for i, p := range toolParams {
						names[i] = p.Name
					}
					return names
				}(),
			}))
		}
	}

	// Make initial API call
	resp, err := l.client.Messages.New(ctx, req)
	if err != nil {
		l.logger.Error(ctx, "Anthropic API error", logging.WithData(map[string]interface{}{
			"error": err.Error(),
		}))
		return Message{}, fmt.Errorf("anthropic API error: %w", err)
	}

	l.logger.Info(ctx, "API response", logging.WithData(map[string]interface{}{
		"content": resp.Content,
	}))

	// Process the response and handle tool calls
	var finalResponse string
	messages = append(messages, resp.ToParam())

	// Tool calling loop - continue until no more tool calls or max iterations reached
	iterCount := 0
	maxIterations := reqParams.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10 // Default max iterations if not specified
	}

	for iterCount < maxIterations {
		iterCount++
		l.logger.Info(ctx, "Tool iteration", logging.WithData(map[string]interface{}{
			"iteration": iterCount,
			"max":       maxIterations,
		}))

		var toolResults []anthropic.ContentBlockParamUnion
		var hasToolCalls bool

		// Process each block in the response
		for _, block := range resp.Content {
			switch variant := block.AsAny().(type) {
			case anthropic.TextBlock:
				finalResponse = variant.Text
			case anthropic.ToolUseBlock:
				hasToolCalls = true
				l.logger.Info(ctx, "Tool use request", logging.WithData(map[string]interface{}{
					"tool":  variant.Name,
					"input": variant.JSON.Input.Raw(),
				}))

				// Parse the input
				var args map[string]any
				if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &args); err != nil {
					errMsg := fmt.Sprintf("Failed to parse tool input: %s", err.Error())
					l.logger.Error(ctx, errMsg, logging.WithData(map[string]interface{}{
						"tool":  variant.Name,
						"error": err.Error(),
					}))

					// Add error result so LLM can handle it
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						variant.ID,
						errMsg,
						true,
					))
					continue
				}

				// Execute the tool using the name as is (no version info)
				result, err := l.tools.Call(ctx, variant.Name, args)
				if err != nil {
					errMsg := fmt.Sprintf("Tool execution failed: %s", err.Error())
					l.logger.Error(ctx, errMsg, logging.WithData(map[string]interface{}{
						"tool":  variant.Name,
						"error": err.Error(),
					}))

					// Add error result so LLM can handle it
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						variant.ID,
						errMsg,
						true,
					))
					continue
				}

				l.logger.Info(ctx, "Tool result", logging.WithData(map[string]interface{}{
					"tool":   variant.Name,
					"result": result,
				}))

				// Add successful tool result
				toolResults = append(toolResults, anthropic.NewToolResultBlock(
					variant.ID,
					result.Content,
					result.IsError,
				))
			}
		}

		// If no tool calls, we're done
		if !hasToolCalls || len(toolResults) == 0 {
			break
		}

		// Add tool results and make another API call
		l.logger.Info(ctx, "Sending tool results back to LLM", logging.WithData(map[string]interface{}{
			"results_count": len(toolResults),
		}))

		messages = append(messages, anthropic.NewUserMessage(toolResults...))
		req.Messages = messages

		// Make follow-up API call with tool results
		resp, err = l.client.Messages.New(ctx, req)
		if err != nil {
			l.logger.Error(ctx, "Anthropic API error after tool calls", logging.WithData(map[string]interface{}{
				"error": err.Error(),
			}))
			return Message{}, fmt.Errorf("anthropic API error after tool calls: %w", err)
		}

		l.logger.Info(ctx, "API response after tool calls", logging.WithData(map[string]interface{}{
			"content": resp.Content,
		}))

		messages = append(messages, resp.ToParam())

		// Update final response from the latest response
		for _, block := range resp.Content {
			if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
				finalResponse = textBlock.Text
			}
		}
	}

	// Check if we hit the max iterations limit
	if iterCount >= maxIterations {
		l.logger.Error(ctx, "Reached maximum tool call iterations", logging.WithData(map[string]interface{}{
			"max_iterations": maxIterations,
		}))
	}

	// Build the final response message
	response := Message{
		Type:    MessageTypeAssistant,
		Content: finalResponse,
		Name:    l.name,
	}

	// Store response in history if enabled
	if reqParams.UseHistory {
		if err := l.memory.Add(response, false); err != nil {
			return Message{}, fmt.Errorf("failed to add response to history: %w", err)
		}
	}

	l.logger.Info(ctx, "Generated final response", logging.WithData(map[string]interface{}{
		"content":    response.Content,
		"iterations": iterCount,
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

// Tools returns the tool registry for this LLM
func (l *AnthropicLLM) Tools() tools.ToolRegistry {
	return l.tools
}

// ExecuteTool executes a specific tool directly
func (l *AnthropicLLM) ExecuteTool(ctx context.Context, toolName string, args map[string]any) (tools.ToolResult, error) {
	l.logger.Info(ctx, "Executing tool directly", logging.WithData(map[string]interface{}{
		"tool": toolName,
		"args": args,
	}))

	// Delegate to the tool registry
	result, err := l.tools.Call(ctx, toolName, args)
	if err != nil {
		l.logger.Error(ctx, "Tool execution failed", logging.WithData(map[string]interface{}{
			"tool":  toolName,
			"error": err.Error(),
		}))
		return tools.ToolResult{}, fmt.Errorf("failed to execute tool %s: %w", toolName, err)
	}

	l.logger.Info(ctx, "Tool execution succeeded", logging.WithData(map[string]interface{}{
		"tool":   toolName,
		"result": result,
	}))

	return result, nil
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
