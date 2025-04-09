package main

import (
	"context"
	"fmt"
	"log"

	"gofast/internal/agent"
	"gofast/internal/llm"
)

func main() {
	// Create a new LLM instance using Anthropic
	llmInstance := llm.NewAnthropicLLM("test-llm")

	// Set up a test context
	ctx := context.Background()

	// Initialize the LLM
	if err := llmInstance.Initialize(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// Create an agent instance
	sizer := agent.NewAgent(agent.AgentConfig{
		Name:        "Size Estimator",
		Instruction: "You are a helpful AI assistant that specializes in estimating the sizes of objects. Given an object, respond only with an estimate of its size in appropriate units.",
		Model:       "claude-3-sonnet-20240320",
	}, llmInstance)

	// Run the agent
	runningAgent, err := sizer.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Example of single message interaction
	response, err := runningAgent.Send("How big is the moon?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Size estimate: %s\n", response)

	// Start interactive chat
	if err := runningAgent.Chat(); err != nil {
		log.Fatal(err)
	}
}
