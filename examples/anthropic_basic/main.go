package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"

	"gofast/internal/config"
	"gofast/internal/fastagent"
	"gofast/internal/llm"
	"gofast/internal/tools"
)

// getCurrentTime is a simple tool that returns the current time
func getCurrentTime() string {
	return time.Now().Format(time.RFC1123)
}

func main() {
	// Load environment variables from .env file in project root
	projectRoot := filepath.Join("..", "..")
	if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Create a tool registry
	registry := tools.NewSimpleToolRegistry()

	// Create and register our time tool
	timeTool := tools.Tool{
		Name:        "getCurrentTime",
		Version:     "1.0.0",
		Description: "Returns the current time in RFC1123 format",
		Category:    "utility",
		Tags:        []string{"time", "utility"},
		Schema:      json.RawMessage(`{}`), // No arguments needed
		Handler: func(ctx context.Context, args map[string]any) (tools.ToolResult, error) {
			return tools.NewToolResult(time.Now().Format(time.RFC1123)), nil
		},
	}

	if err := registry.Register(timeTool); err != nil {
		fmt.Printf("Failed to register time tool: %v\n", err)
		os.Exit(1)
	}

	// Load config for LLM
	cfg, err := fastagent.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create and initialize the LLM
	teamLLM := llm.NewAnthropicLLM("example-llm")
	if err := teamLLM.Initialize(context.Background(), &config.Settings{
		DefaultModel: "claude-3-haiku", // Fast, cheap model
		Logger: config.LoggerSettings{
			Level: cfg.LogLevel,
		},
	}); err != nil {
		fmt.Printf("Failed to initialize LLM: %v\n", err)
		os.Exit(1)
	}

	// Create an agent that can use our tool
	agent := fastagent.New("time-agent",
		"You are an assistant that helps with time-related tasks. "+
			"Use the getCurrentTime tool to get the current time and suggest appropriate activities.",
	).WithTools("getCurrentTime@1.0.0")

	// Create a team with our agent and LLM
	team := fastagent.TeamWithLLM("Time Tutorial", teamLLM, agent)
	defer team.Close()

	// Ask the agent to use the tool
	response, err := team.Send("time-agent",
		"What time is it right now? Please use the getCurrentTime tool to check "+
			"and suggest some appropriate activities for this time of day.")
	if err != nil {
		fmt.Printf("Error getting response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAgent Response: %s\n", response)
}
