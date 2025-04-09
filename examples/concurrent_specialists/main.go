package main

import (
	"fmt"
	"log"

	"gofast/internal/fastagent"
)

func init() {
	fastagent.NewArchetype("SysAdmin").
		WithRole("You are a systems administrator focused on diagnosing and resolving system-level performance issues, log analysis, and infrastructure maintenance.").
		WithHistory().
		Register()

	fastagent.NewArchetype("StackOverflow").
		WithRole("You are a software expert focused on code optimization, performance bottlenecks, and architectural improvements.").
		WithHistory().
		Register()

	fastagent.NewArchetype("CloudGuru").
		WithRole("You are a cloud infrastructure expert focused on scalability, containerization, and cloud-native solutions.").
		WithHistory().
		Register()
}

func main() {
	// Load configuration
	cfg, err := fastagent.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create and start the team
	team := fastagent.Team("Tech Support", cfg,
		fastagent.New("SysAdmin", "You are a systems administrator focused on diagnosing and resolving system-level performance issues."),
		fastagent.New("StackOverflow", "You are a software expert focused on code optimization and performance bottlenecks."),
		fastagent.New("CloudGuru", "You are a cloud infrastructure expert focused on scalability and cloud-native solutions."),
	)
	defer team.Close()

	// Define questions for each specialist
	task := fastagent.NewTask("Production system is running slow").
		AssignTo("SysAdmin", "What system-level issues could be causing the slowdown?").
		AssignTo("StackOverflow", "What are common performance bottlenecks and their solutions?").
		AssignTo("CloudGuru", "How can we improve our cloud infrastructure for better performance?")

	// Run task and collect responses
	responses, err := task.Run(team)
	if err != nil {
		log.Fatal(err)
	}

	// Get final recommendation
	summary := fastagent.NewSynthesisRequest().
		WithResponses(responses).
		WithPrompt("Based on these findings, what's our action plan?")

	if _, err := summary.SendTo(team, "CloudGuru"); err != nil {
		log.Printf("Failed to get action plan: %v", err)
	}

	fmt.Println("\nAnalysis completed")
}
