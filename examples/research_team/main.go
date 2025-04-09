package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"gofast/internal/fastagent"
)

func init() {
	// Register team archetypes
	fastagent.NewArchetype("Researcher").
		WithRole("You are a research specialist focused on gathering and analyzing primary data and evidence.").
		WithHistory().
		Register()

	fastagent.NewArchetype("Analyst").
		WithRole("You are an analyst focused on interpreting data and identifying patterns and implications.").
		WithHistory().
		Register()

	fastagent.NewArchetype("Critic").
		WithRole("You are a critical thinker focused on identifying limitations, assumptions, and potential issues.").
		WithHistory().
		Register()
}

func main() {
	// Load configuration
	cfg, err := fastagent.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.AnthropicAPIKey == "" {
		fmt.Println("ANTHROPIC_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Create research team
	team := fastagent.NewTeam("Research Team").
		WithArchetype("Researcher").
		WithArchetype("Analyst").
		WithArchetype("Critic").
		Build()
	defer team.Close()

	// Define research tasks
	tasks := []struct {
		question    string
		assignments map[string]string
	}{
		{
			question: "What are the key quantum computing cybersecurity risks in 2024?",
			assignments: map[string]string{
				"Researcher": "What quantum algorithms pose the biggest threat to current encryption?",
				"Analyst":    "Which systems are most vulnerable to quantum attacks?",
				"Critic":     "What are the limitations of quantum-based attacks?",
			},
		},
		{
			question: "How will climate change impact agriculture in 2025?",
			assignments: map[string]string{
				"Researcher": "What crop yields will be most affected by temperature changes?",
				"Analyst":    "Which regions face the highest agricultural risk?",
				"Critic":     "What adaptation strategies show the most promise?",
			},
		},
	}

	// Process each task
	for _, task := range tasks {
		fmt.Printf("\nResearching: %s\n", task.question)

		// Create task context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		var wg sync.WaitGroup
		results := make(map[string]string)
		resultsChan := make(chan struct {
			agent string
			resp  string
		}, len(task.assignments))

		// Launch research assignments
		for agent, question := range task.assignments {
			wg.Add(1)
			go func(agent, question string) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				default:
					if response, err := team.Send(agent, question); err == nil {
						select {
						case resultsChan <- struct {
							agent string
							resp  string
						}{agent, response}:
						case <-ctx.Done():
							return
						}
					}
				}
			}(agent, question)
		}

		// Wait for results or timeout
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// Collect results
		for result := range resultsChan {
			results[result.agent] = result.resp
		}

		// Synthesize findings
		if len(results) == len(task.assignments) {
			fmt.Println("\nSynthesizing findings:")
			for agent, response := range results {
				fmt.Printf("- %s: %s\n", agent, response)
			}
		}

		cancel()
	}

	fmt.Println("\nResearch completed")
}
