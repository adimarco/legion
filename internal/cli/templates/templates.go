package templates

const (
	ConfigTemplate = `# FastAgent Configuration File

# Default Model Configuration:
# Takes format:
#   <provider>.<model_string>.<reasoning_effort?> (e.g. anthropic.claude-3-5-sonnet-20241022 or openai.o3-mini.low)
# Accepts aliases for Anthropic Models: haiku, haiku3, sonnet, sonnet35, opus, opus3
# and OpenAI Models: gpt-4o-mini, gpt-4o, o1, o1-mini, o3-mini

default_model: haiku

# Logging and Console Configuration:
logger:
    progress_display: true
    show_chat: true
    show_tools: true
    truncate_tools: true

# MCP Servers
mcp:
    servers:
        fetch:
            command: "uvx"
            args: ["mcp-server-fetch"]
        filesystem:
            command: "npx"
            args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
`

	SecretsTemplate = `# FastAgent Secrets Configuration
# WARNING: Keep this file secure and never commit to version control

# Alternatively set OPENAI_API_KEY and ANTHROPIC_API_KEY environment variables. Config file takes precedence.

openai:
    api_key: <your-api-key-here>
anthropic:
    api_key: <your-api-key-here>

# Example of setting an MCP Server environment variable
mcp:
    servers:
        brave:
            env:
                BRAVE_API_KEY: <your_api_key_here>
`

	GitignoreTemplate = `# FastAgent secrets file
fastagent.secrets.yaml

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo

# Environment
.env
`

	MainTemplate = `package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropic-ai/claude-go"
)

func main() {
	// Create a new FastAgent instance
	client, err := claude.NewClient(claude.WithAPIKey("<your-api-key>"))
	if err != nil {
		log.Fatal(err)
	}

	// Example message
	msg := claude.Message{
		Role: "user",
		Content: []claude.Content{
			{Type: "text", Text: "Hello! What can you help me with today?"},
		},
	}

	// Send message to Claude
	resp, err := client.CreateMessage(context.Background(), &claude.CreateMessageRequest{
		Model:    claude.ModelHaiku,
		Messages: []claude.Message{msg},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", resp.Content[0].Text)
}
`

	ModTemplate = `module %s

go 1.21

require (
	github.com/anthropic-ai/claude-go v0.0.0-20240308222815-20c05b6b4ad5
)
`
)
