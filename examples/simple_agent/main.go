package main

import (
	"fmt"
	"log"

	"gofast/internal/fastagent"
)

func main() {
	// Create a size estimator agent
	sizer := fastagent.New("sizer", "Given an object, respond only with an estimate of its size in appropriate units.")
	team := fastagent.Team("Simple Agent Demo", sizer)
	defer team.Close()

	// Send a message and get response
	response, err := team.Send("sizer", "How big is the moon?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Size estimate: %s\n", response)

	// Start interactive chat (press Ctrl+C to exit)
	if err := team.Chat("sizer"); err != nil {
		log.Fatal(err)
	}
}
