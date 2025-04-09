package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gofast/internal/agent"
	"gofast/internal/llm"
)

func main() {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create LLM instance
	llmInstance := llm.NewAnthropicLLM("test-llm")
	if err := llmInstance.Initialize(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// Create channel agent
	channelAgent := agent.NewChannelAgent(agent.AgentConfig{
		Name:        "Concurrent Helper",
		Instruction: "You are a helpful AI assistant. Keep responses brief and to the point.",
		Type:        agent.AgentTypeBasic,
		Model:       "claude-3-sonnet-20240320",
	}, llmInstance)

	// Start the agent
	if err := channelAgent.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Create wait group for demo goroutines
	var wg sync.WaitGroup

	// Start error handler
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case err := <-channelAgent.Errors():
				log.Printf("Error: %v\n", err)
			case <-channelAgent.Done():
				return
			}
		}
	}()

	// Start response handler
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case response := <-channelAgent.Output():
				fmt.Printf("\nResponse received: %s\n", response)
			case <-channelAgent.Done():
				return
			}
		}
	}()

	// Send multiple messages concurrently
	messages := []string{
		"What is the capital of France?",
		"What is the largest planet?",
		"Who wrote Romeo and Juliet?",
		"What is the speed of light?",
		"What is the chemical symbol for gold?",
	}

	fmt.Println("Sending messages concurrently...")
	for _, msg := range messages {
		wg.Add(1)
		go func(message string) {
			defer wg.Done()
			fmt.Printf("\nSending: %s\n", message)
			if err := channelAgent.Send(message); err != nil {
				log.Printf("Failed to send message: %v\n", err)
			}
		}(msg)
		// Small delay to make output more readable
		time.Sleep(100 * time.Millisecond)
	}

	// Wait a bit for responses
	time.Sleep(5 * time.Second)

	// Clean shutdown
	fmt.Println("\nShutting down...")
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("Done!")
}
