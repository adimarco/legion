package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/adimarco/hive"
	"github.com/adimarco/hive/internal/tools"
)

func main() {
	// Create and initialize the LLM with defaults
	agentLLM, err := hive.NewAnthropicLLM("time-assistant")
	if err != nil {
		log.Fatal(err)
	}

	// Create a simple time tool
	timeTool := tools.Tool{
		Name:        "getCurrentTime",
		Version:     "1.0.0",
		Description: "Returns the current time in RFC1123 format",
		Category:    "utility",
		Tags:        []string{"time", "utility"},
		Schema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"format": {
					"type": "string",
					"description": "The time format to use (defaults to RFC1123)",
					"enum": ["RFC1123"]
				}
			}
		}`),
		Handler: func(ctx context.Context, args map[string]any) (tools.ToolResult, error) {
			return tools.NewToolResult(time.Now().Format(time.RFC1123)), nil
		},
	}

	// Register the tool with the LLM's registry
	if err := agentLLM.Tools().Register(timeTool); err != nil {
		log.Fatal(err)
	}

	// Create a simple time agent using fluent interface
	agent := hive.New("time-agent",
		"You are an assistant that helps with time-related tasks. "+
			"You have access to a tool called getCurrentTime that returns the current time in RFC1123 format. "+
			"When asked about the current time, follow these steps:\n"+
			"1. Call the getCurrentTime tool to get the current time\n"+
			"2. Parse the time from the tool's response\n"+
			"3. Suggest appropriate activities based on the time of day\n"+
			"4. Consider the timezone information when making suggestions\n"+
							"Always be specific in your suggestions and explain your reasoning.").
		WithModel("claude-3-haiku-20240307"). // Fast, cheap model
		WithTools("getCurrentTime@1.0.0").    // Add our tool with version
		WithHistory().                        // Enable chat history
		WithLLM(agentLLM)                     // Set the LLM

	// Create team with single agent
	team := hive.TeamWithLLM("Time Tutorial", agentLLM, agent)
	defer team.Close()

	// Send a message and get response
	response, err := team.Send("time-agent",
		"What time is it right now? Please use the getCurrentTime tool to check "+
			"and suggest some appropriate activities for this time of day.")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAgent Response: %s\n", response)
}
