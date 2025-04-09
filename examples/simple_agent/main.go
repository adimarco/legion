package main

import (
	"context"
	"fmt"
	"log"

	"github.com/adimarco/hive/internal/config"
	"github.com/adimarco/hive/internal/fastagent"
	"github.com/adimarco/hive/internal/llm"
)

func main() {
	// Create a local registry
	registry, err := fastagent.NewRegistry("memory://local")
	if err != nil {
		log.Fatal(err)
	}

	// Create a size estimator agent
	sizer := registry.PublishAgent("tools/sizer",
		"Given an object, respond only with an estimate of its size in appropriate units.",
		fastagent.AgentConfig{
			Version:    "1.0.0",
			UseHistory: true,
			Tools:      []string{"measure", "convert"},
			Model:      "claude-3-haiku-20240307", // Fast, cheap model
			Metadata: map[string]any{
				"tags": []string{"measurement", "estimation"},
			},
		},
	)

	// Load config for LLM
	cfg, err := fastagent.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create and initialize the LLM
	teamLLM := llm.NewAnthropicLLM("sizer")
	if err := teamLLM.Initialize(context.TODO(), &config.Settings{
		DefaultModel: cfg.DefaultModel,
		Logger: config.LoggerSettings{
			Level: cfg.LogLevel,
		},
	}); err != nil {
		log.Fatal(err)
	}

	// Create team with chosen LLM
	team := fastagent.TeamWithLLM("Simple Agent Demo", teamLLM, sizer.ToAgent())
	defer team.Close()

	// Send a message and get response
	response, err := team.Send("tools/sizer", "How big is the moon?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Size estimate: %s\n", response)

	// Start interactive chat (press Ctrl+C to exit)
	if err := team.Chat("tools/sizer"); err != nil {
		log.Fatal(err)
	}
}
