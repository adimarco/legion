package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adimarco/hive"
	"github.com/adimarco/hive/internal/tools"
)

func main() {
	// Create a new app
	app := hive.NewApp("time-assistant")
	defer app.Close()

	// Register time tool with simplified interface
	timeTool := hive.Tool("getCurrentTime", func(ctx context.Context, args map[string]any) (string, error) {
		result := tools.NewToolResult(time.Now().Format(time.RFC1123))
		return result.Content, nil
	}).
		WithDescription("Gets the current time in RFC1123 format (e.g. 'Wed, 09 Apr 2025 12:15:53 EDT')").
		Build()

	if err := app.WithTool(timeTool); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}

	// Create agent with simplified interface
	agent := app.WithAgent(
		"You help with time-related tasks. Suggest activities appropriate for the current time.").
		WithTools("getCurrentTime@1.0.0")

	// Send a message and get response
	response, err := agent.Send("What time is it right now? Please suggest some activities.")
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}

	fmt.Printf("\nAgent Response: %s\n", response)
}
