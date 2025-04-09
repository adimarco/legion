package main

import (
	"fmt"
	"log"
	"os"

	"gofast/internal/config"
	"gofast/internal/fastagent"
	"gofast/internal/llm"
)

func main() {
	// Create a local registry
	registry, err := fastagent.NewRegistry("memory://local")
	if err != nil {
		log.Fatal(err)
	}

	// Register team members
	registry.PublishAgent("research/analyst",
		"You are an analyst focused on interpreting research data and identifying key patterns.",
		fastagent.AgentConfig{
			Version:    "1.0.0",
			UseHistory: true,
			Tools:      []string{"search", "summarize"},
			Model:      "claude-3-haiku-20240307", // Fast, cheap model
		},
	)

	registry.PublishAgent("research/critic",
		"You are a research critic focused on identifying limitations and potential issues.",
		fastagent.AgentConfig{
			Version:    "1.0.0",
			UseHistory: true,
			Tools:      []string{"validate", "factcheck"},
			Model:      "claude-3-haiku-20240307", // Fast, cheap model
		},
	)

	researcher := registry.PublishAgent("custom/researcher",
		"You are a research specialist focused on gathering and analyzing primary data.",
		fastagent.AgentConfig{
			Version:    "1.0.0",
			UseHistory: true,
			Tools:      []string{"search", "analyze"},
			Model:      "claude-3-haiku-20240307", // Fast, cheap model
			Metadata: map[string]any{
				"author": "research_team_demo",
				"tags":   []string{"research", "analysis"},
			},
		},
	)

	// Choose LLM based on environment
	var teamLLM llm.AugmentedLLM
	switch os.Getenv("FASTAGENT_LLM") {
	case "passthrough":
		teamLLM = llm.NewPassthroughLLM("test-research")
		teamLLM.Initialize(nil, nil)
	default:
		// Load config for real LLM
		cfg, err := fastagent.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		teamLLM = llm.NewAnthropicLLM("research")
		if err := teamLLM.Initialize(nil, &config.Settings{
			DefaultModel: "claude-3-haiku-20240307", // Override config to use fast model
			Logger: config.LoggerSettings{
				Level: cfg.LogLevel,
			},
		}); err != nil {
			log.Fatal(err)
		}
	}

	// Create team with chosen LLM
	team := fastagent.TeamWithLLM("Research Team", teamLLM,
		researcher.ToAgent(),
		registry.MustGetAgent("research/analyst").ToAgent(),
		registry.MustGetAgent("research/critic").ToAgent(),
	)
	defer team.Close()

	// Define research task
	task := fastagent.NewTask("What are the key quantum computing cybersecurity risks in 2024?").
		AssignTo("custom/researcher", "What quantum algorithms pose the biggest threat to current encryption?").
		AssignTo("research/analyst", "Which systems are most vulnerable to quantum attacks?").
		AssignTo("research/critic", "What are the limitations of quantum-based attacks?")

	// Run task and collect responses
	responses, err := task.Run(team)
	if err != nil {
		log.Fatal(err)
	}

	// Get synthesis from the analyst
	summary := fastagent.NewSynthesisRequest().
		WithResponses(responses).
		WithPrompt("Create a balanced assessment of quantum computing risks.")

	if _, err := summary.SendTo(team, "research/analyst"); err != nil {
		log.Printf("Failed to get synthesis: %v", err)
	}

	fmt.Println("\nResearch completed")
}
