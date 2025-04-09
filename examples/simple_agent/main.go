package main

import (
	"fmt"
	"log"

	"github.com/adimarco/hive"
)

func main() {
	// Create and initialize the LLM with defaults
	agentLLM, err := hive.NewAnthropicLLM("sizer")
	if err != nil {
		log.Fatal(err)
	}

	// Create a simple size estimator agent
	agent := hive.New("sizer", "Given an object, respond only with an estimate of its size in appropriate units.")
	agent.WithModel("claude-3-haiku-20240307") // Fast, cheap model
	agent.WithTools("measure", "convert")
	agent.WithHistory()

	// Create team with single agent
	team := hive.TeamWithLLM("Simple Agent Demo", agentLLM, agent)
	defer team.Close()

	// Send a message and get response
	response, err := team.Send("sizer", "How big is the sun?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Size estimate: %s\n", response)
}
