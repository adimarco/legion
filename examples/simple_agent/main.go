package main

import (
	"fmt"
	"log"
	"time"

	"github.com/adimarco/hive"
)

// getCurrentTime returns the current time in RFC1123 format
func getCurrentTime() string {
	return time.Now().Format(time.RFC1123)
}

func main() {
	// Create a new app
	app := hive.NewApp("time-assistant")
	defer app.Close()

	// Register time tool with simple function
	if err := app.Tool("getCurrentTime", getCurrentTime); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}

	// Create agent with simplified interface
	agent := app.Agent("Help with time-related tasks")

	// Send a message and get response
	response, err := agent.Send("What time is it right now? Please suggest some activities.")
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}

	fmt.Printf("\nAgent Response: %s\n", response)
}
