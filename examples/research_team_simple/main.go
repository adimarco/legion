package main

import (
	"fmt"
	"log"

	"gofast/internal/fastagent"
)

func main() {
	// Create a local registry
	registry, err := fastagent.NewRegistry("memory://local")
	if err != nil {
		log.Fatal(err)
	}

	// Create our team members
	registry.PublishAgent("tech/visionary",
		"You are a technology strategist focused on identifying emerging trends and their potential impact.",
		fastagent.AgentConfig{
			Version:    "1.0.0",
			UseHistory: true,
			Tools:      []string{"research", "analyze"},
			Metadata: map[string]any{
				"tags": []string{"strategy", "trends", "innovation"},
			},
		},
	)

	registry.PublishAgent("tech/engineer",
		"You are a systems engineer focused on technical feasibility and implementation challenges.",
		fastagent.AgentConfig{
			Version:    "1.0.0",
			UseHistory: true,
			Tools:      []string{"benchmark", "prototype"},
			Metadata: map[string]any{
				"tags": []string{"engineering", "systems", "implementation"},
			},
		},
	)

	analyst := registry.PublishAgent("tech/analyst",
		"You are a market analyst focused on business opportunities and competitive analysis.",
		fastagent.AgentConfig{
			Version:    "1.0.0",
			UseHistory: true,
			Tools:      []string{"market_research", "competitor_analysis"},
			Metadata: map[string]any{
				"tags": []string{"market", "business", "analysis"},
			},
		},
	)

	// Load config
	cfg, err := fastagent.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create team using agents from registry
	team := fastagent.Team("TrendHunters", cfg,
		registry.MustGetAgent("tech/visionary").ToAgent(),
		registry.MustGetAgent("tech/engineer").ToAgent(),
		analyst.ToAgent(),
	)
	defer team.Close()

	// Define and run task
	task := fastagent.NewTask("What's the real potential of quantum computing in the next 2 years?").
		AssignTo("tech/engineer", "What are the key technical limitations of current quantum systems?").
		AssignTo("tech/analyst", "Which quantum startups have the strongest market position?").
		AssignTo("tech/visionary", "How will quantum tech disrupt traditional computing?")

	responses, err := task.Run(team)
	if err != nil {
		log.Fatal(err)
	}

	// Get synthesis from the visionary
	summary := fastagent.NewSynthesisRequest().
		WithResponses(responses).
		WithPrompt("Give a concise vision for quantum's near-term impact.")

	if _, err := summary.SendTo(team, "tech/visionary"); err != nil {
		log.Printf("Failed to get synthesis: %v", err)
	}

	fmt.Println("\nAnalysis completed")
}
