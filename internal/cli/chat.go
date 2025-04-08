package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gofast/internal/llm"
)

func chatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start a chat session with an LLM",
		Long: `Start an interactive chat session with an LLM.
Currently using a passthrough LLM for testing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startChat(cmd.Context())
		},
	}

	return cmd
}

func startChat(ctx context.Context) error {
	// Create and initialize the LLM
	llmInstance := llm.NewPassthroughLLM("test-llm")
	if err := llmInstance.Initialize(ctx, nil); err != nil {
		return fmt.Errorf("failed to initialize LLM: %w", err)
	}
	defer llmInstance.Cleanup()

	fmt.Println("\nStarting chat session (type 'exit' to quit)...")
	fmt.Println("Using PassthroughLLM for testing - it will echo your messages back.")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Get user input
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "exit" {
			break
		}

		if input == "" {
			continue
		}

		// Generate response
		response, err := llmInstance.GenerateString(ctx, input, nil)
		if err != nil {
			return fmt.Errorf("failed to generate response: %w", err)
		}

		fmt.Printf("\nAssistant: %s\n", response)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	fmt.Println("\nChat session ended.")
	return nil
}
