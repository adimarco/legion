package main

import (
	"fmt"
	"log"

	"gofast/internal/fastagent"
)

func init() {
	fastagent.NewArchetype("Visionary").
		WithRole("You are a technology strategist focused on identifying emerging trends and their potential impact on industry.").
		WithHistory().
		Register()

	fastagent.NewArchetype("Engineer").
		WithRole("You are a systems engineer focused on technical feasibility, implementation challenges, and practical limitations.").
		WithHistory().
		Register()

	fastagent.NewArchetype("Analyst").
		WithRole("You are a market analyst focused on business opportunities, competitive analysis, and growth metrics.").
		WithHistory().
		Register()
}

func main() {
	// Load configuration
	cfg, err := fastagent.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create research team
	team := fastagent.Team("TrendHunters", cfg,
		fastagent.New("Visionary", "You are a technology strategist focused on identifying emerging trends and their impact."),
		fastagent.New("Engineer", "You are a systems engineer focused on technical feasibility and implementation challenges."),
		fastagent.New("Analyst", "You are a market analyst focused on business opportunities and competitive analysis."),
	)
	defer team.Close()

	task := fastagent.NewTask("What's the real potential of quantum computing in the next 2 years?").
		AssignTo("Engineer", "What are the key technical limitations of current quantum systems?").
		AssignTo("Analyst", "Which quantum startups have the strongest market position?").
		AssignTo("Visionary", "How will quantum tech disrupt traditional computing?")

	// Run task and collect responses
	responses, err := task.Run(team)
	if err != nil {
		log.Fatal(err)
	}

	// Send synthesis request
	summary := fastagent.NewSynthesisRequest().
		WithResponses(responses).
		WithPrompt("Give a concise vision for quantum's near-term impact.")

	if _, err := summary.SendTo(team, "Visionary"); err != nil {
		log.Printf("Failed to get synthesis: %v", err)
	}

	fmt.Println("\nAnalysis completed")
}
