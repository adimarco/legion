package main

import (
	"context"
	"fmt"
	"log"

	"gofast/internal/agent"
	"gofast/internal/llm"
)

func main() {
	// Create a new LLM instance (using passthrough for testing)
	llmInstance := llm.NewPassthroughLLM("test-llm")

	// Set up a test context
	ctx := context.Background()

	// Initialize the LLM with a fixed response for testing
	msg := llm.Message{
		Type:    llm.MessageTypeUser,
		Content: "***FIXED_RESPONSE The size is approximately 3,475 kilometers in diameter",
	}
	_, err := llmInstance.Generate(ctx, msg, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create an agent instance
	sizer := agent.NewAgent(agent.AgentConfig{
		Name:        "Agent Example",
		Instruction: "Given an object, respond only with an estimate of its size.",
	}, llmInstance)

	// Run the agent
	runningAgent, err := sizer.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Example of single message interaction
	response, err := runningAgent.Send("the moon")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Size estimate: %s\n", response)

	// Start interactive chat
	if err := runningAgent.Chat(); err != nil {
		log.Fatal(err)
	}
}
