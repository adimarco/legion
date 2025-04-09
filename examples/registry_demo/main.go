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
	// Connect to registry (could be local, remote, etc.)
	registry, err := fastagent.NewRegistry("thoth://registry.agents.dev")
	if err != nil {
		log.Fatal(err)
	}

	// Publish a custom agent
	cpo := registry.PublishAgent("skiddie420/cpo",
		"A Chief Product Officer focused on product strategy and roadmap planning.",
		fastagent.AgentConfig{
			Version: "2.1.7",
			Model:   "claude-3-haiku-20240307", // Fast, cheap model
			Tools: []string{
				"jira", "figma", "github",
				"miro/mindmap", "slack/threads",
			},
			UseHistory: true,
			Metadata: map[string]any{
				"author":     "skiddie420",
				"tags":       []string{"product", "strategy", "leadership"},
				"repository": "https://github.com/skiddie420/cpo-agent",
				"license":    "MIT",
			},
		},
	)

	// Get existing agents
	devops, err := registry.GetAgent("musscope/devops")
	if err != nil {
		log.Fatal(err)
	}
	devops.UseMCPTools("aws-sdk", "terraform", "datadog")

	// Search for agents
	fmt.Println("\nFinding product-focused agents:")
	for _, a := range registry.SearchAgents(map[string]any{
		"tags": []string{"product"},
	}) {
		fmt.Printf("- %s (%s)\n  %s\n", a.Name, a.Version, a.Role)
	}

	// Choose LLM based on environment
	var teamLLM llm.AugmentedLLM
	switch os.Getenv("FASTAGENT_LLM") {
	case "passthrough":
		teamLLM = llm.NewPassthroughLLM("test-product")
		teamLLM.Initialize(nil, nil)
	default:
		// Load config for real LLM
		cfg, err := fastagent.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		teamLLM = llm.NewAnthropicLLM("product")
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
	team := fastagent.TeamWithLLM("Product Development", teamLLM,
		cpo.ToAgent(),
		devops.ToAgent(),
	)
	defer team.Close()

	// Run a task
	task := fastagent.NewTask("How should we improve our deployment process?").
		AssignTo("skiddie420/cpo", "What product requirements should we consider?").
		AssignTo("musscope/devops", "What technical improvements would have the biggest impact?")

	responses, err := task.Run(team)
	if err != nil {
		log.Fatal(err)
	}

	// Get synthesis
	summary := fastagent.NewSynthesisRequest().
		WithResponses(responses).
		WithPrompt("Create an action plan that balances product needs with technical improvements.")

	if _, err := summary.SendTo(team, "skiddie420/cpo"); err != nil {
		log.Printf("Failed to get synthesis: %v", err)
	}
}
