package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"gofast/internal/llm"
)

func demoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo [scenario]",
		Short: "Run interactive demos of FastAgent capabilities",
		Long: `Run interactive demos that showcase FastAgent's capabilities.

Available scenarios:
  simple    - Basic chat interaction
  tools     - Tool calling demonstration
  fixed     - Fixed response handling
  playback  - Conversation replay`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scenario := "simple"
			if len(args) > 0 {
				scenario = args[0]
			}
			return runDemo(cmd.Context(), scenario)
		},
	}

	return cmd
}

func runDemo(ctx context.Context, scenario string) error {
	switch scenario {
	case "simple":
		return demoSimpleChat(ctx)
	case "tools":
		return demoToolCalls(ctx)
	case "fixed":
		return demoFixedResponses(ctx)
	case "playback":
		return demoPlayback(ctx)
	default:
		return fmt.Errorf("unknown scenario: %s", scenario)
	}
}

func demoSimpleChat(ctx context.Context) error {
	color.Blue("\nDemo: Simple Chat")
	fmt.Println("Demonstrating basic message handling with PassthroughLLM")

	assistant := llm.NewPassthroughLLM("Echo")
	messages := []string{
		"Hello, how are you?",
		"What's the weather like?",
		"Goodbye!",
	}

	for _, msg := range messages {
		time.Sleep(1 * time.Second) // Dramatic pause
		color.Green("\nUser: %s", msg)

		response, err := assistant.GenerateString(ctx, msg, nil)
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
		color.Yellow("Assistant: %s", response)
	}

	return nil
}

func demoToolCalls(ctx context.Context) error {
	color.Blue("\nDemo: Tool Calls")
	fmt.Println("Demonstrating tool call parsing and execution")

	assistant := llm.NewPassthroughLLM("ToolUser")
	messages := []string{
		"Let me check the weather.",
		llm.ToolCallPrefix + ` weather {"location": "Seattle", "units": "F"}`,
		"Now let me check the time.",
		llm.ToolCallPrefix + ` time {"timezone": "PST"}`,
	}

	for _, msg := range messages {
		time.Sleep(1 * time.Second)
		color.Green("\nUser: %s", msg)

		response, err := assistant.GenerateString(ctx, msg, nil)
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
		color.Yellow("Assistant: %s", response)
	}

	return nil
}

func demoFixedResponses(ctx context.Context) error {
	color.Blue("\nDemo: Fixed Responses")
	fmt.Println("Demonstrating fixed response handling")

	assistant := llm.NewPassthroughLLM("FixedBot")
	messages := []string{
		"What's your favorite color?",
		llm.FixedResponsePrefix + " Blue! Always blue!",
		"Are you sure?",
		"What about red?",
		"Maybe green?",
	}

	for _, msg := range messages {
		time.Sleep(1 * time.Second)
		color.Green("\nUser: %s", msg)

		response, err := assistant.GenerateString(ctx, msg, nil)
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
		color.Yellow("Assistant: %s", response)
	}

	return nil
}

func demoPlayback(ctx context.Context) error {
	color.Blue("\nDemo: Conversation Replay")
	fmt.Println("Demonstrating conversation playback capabilities")

	// Create a conversation
	conversation := []llm.Message{
		{Type: llm.MessageTypeUser, Content: "Tell me a story"},
		{Type: llm.MessageTypeAssistant, Content: "Once upon a time, in a magical forest..."},
		{Type: llm.MessageTypeUser, Content: "What happened next?"},
		{Type: llm.MessageTypeAssistant, Content: "A brave knight encountered a friendly dragon!"},
		{Type: llm.MessageTypeUser, Content: "And then what?"},
		{Type: llm.MessageTypeAssistant, Content: "They became best friends and opened a cozy coffee shop together."},
	}

	// Create playback LLM
	assistant := llm.NewPlaybackLLM("Storyteller")
	if err := assistant.Initialize(ctx, nil); err != nil {
		return fmt.Errorf("failed to initialize LLM: %w", err)
	}
	assistant.LoadMessages(conversation)

	// Play through the conversation
	messages := []string{
		"Ready to hear a story?",
		"What happens in the story?",
		"Then what happened?",
		"How does it end?",
		"Tell me another story!", // This should trigger exhaustion
	}

	for _, msg := range messages {
		time.Sleep(1 * time.Second)
		color.Green("\nUser: %s", msg)

		response, err := assistant.GenerateString(ctx, msg, nil)
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
		color.Yellow("Assistant: %s", response)
	}

	return nil
}
