package agent

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gofast/internal/llm"
)

// Send sends a single message to the agent and returns the response
func (ra *RunningAgent) Send(msg string) (string, error) {
	message := llm.Message{
		Type:    llm.MessageTypeUser,
		Content: msg,
	}

	params := &llm.RequestParams{
		SystemPrompt: ra.agent.config.Instruction,
		UseHistory:   true,
	}

	response, err := ra.agent.fa.llm.Generate(ra.ctx, message, params)
	if err != nil {
		return "", fmt.Errorf("failed to get LLM completion: %w", err)
	}

	return response.Content, nil
}

// Chat starts an interactive chat session with the agent
func (ra *RunningAgent) Chat() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Starting chat session. Type 'exit' to end.")
	fmt.Println("Instruction:", ra.agent.config.Instruction)

	for {
		fmt.Print("\nUser: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			return nil
		}

		response, err := ra.Send(input)
		if err != nil {
			return fmt.Errorf("failed to get response: %w", err)
		}

		fmt.Printf("Assistant: %s\n", response)
	}
}
