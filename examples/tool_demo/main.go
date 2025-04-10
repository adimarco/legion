package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adimarco/hive"
)

func main() {
	// Create a new app with tool support
	app := hive.NewApp("tool-demo")
	defer app.Close()

	// Register tool categories
	if err := registerFileTools(app); err != nil {
		fmt.Printf("Error registering file tools: %v\n", err)
		os.Exit(1)
	}

	if err := registerSystemTools(app); err != nil {
		fmt.Printf("Error registering system tools: %v\n", err)
		os.Exit(1)
	}

	if err := registerDateTimeTools(app); err != nil {
		fmt.Printf("Error registering date/time tools: %v\n", err)
		os.Exit(1)
	}

	// Create an agent with access to tools
	agent := app.Agent("You are a helpful agent with access to system tools. Use them to help answer user questions.")

	// Start an interactive session
	runningAgent, err := agent.Run(context.Background())
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		os.Exit(1)
	}

	// Run an interactive chat session
	fmt.Println("Tool Demo Agent - Type 'exit' to quit")
	fmt.Println("Try asking about files in the current directory, environment variables, or the current date and time.")
	fmt.Println("---------------------------------------")

	if err := runningAgent.Chat(); err != nil {
		fmt.Printf("Error in chat session: %v\n", err)
		os.Exit(1)
	}
}

// registerFileTools registers tools for file operations
func registerFileTools(app *hive.App) error {
	// Read file contents
	if err := app.Tool("readFile", func(ctx context.Context, args map[string]interface{}) (string, error) {
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("path must be a string")
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		return string(data), nil
	}); err != nil {
		return err
	}

	// List directory contents
	if err := app.Tool("listDir", func(ctx context.Context, args map[string]interface{}) (string, error) {
		path, ok := args["path"].(string)
		if !ok {
			path = "."
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("failed to read directory: %w", err)
		}

		var result strings.Builder
		for _, entry := range entries {
			entryType := "file"
			if entry.IsDir() {
				entryType = "dir"
			}
			fmt.Fprintf(&result, "%s [%s]\n", entry.Name(), entryType)
		}

		return result.String(), nil
	}); err != nil {
		return err
	}

	return nil
}

// registerSystemTools registers tools for system operations
func registerSystemTools(app *hive.App) error {
	// Get working directory
	if err := app.Tool("getWorkingDir", func() string {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return dir
	}); err != nil {
		return err
	}

	// Get environment variable
	if err := app.Tool("getEnv", func(ctx context.Context, args map[string]interface{}) (string, error) {
		name, ok := args["name"].(string)
		if !ok {
			return "", fmt.Errorf("name must be a string")
		}

		value := os.Getenv(name)
		if value == "" {
			return fmt.Sprintf("Environment variable %s is not set", name), nil
		}

		return value, nil
	}); err != nil {
		return err
	}

	return nil
}

// registerDateTimeTools registers tools for date and time operations
func registerDateTimeTools(app *hive.App) error {
	// Get current date and time
	if err := app.Tool("getDateTime", func() string {
		now := time.Now()
		return now.Format(time.RFC3339)
	}); err != nil {
		return err
	}

	// Format a date
	if err := app.Tool("formatDate", func(ctx context.Context, args map[string]interface{}) (string, error) {
		layout, ok := args["layout"].(string)
		if !ok {
			layout = "2006-01-02 15:04:05"
		}

		timeStr, ok := args["time"].(string)
		if !ok {
			return time.Now().Format(layout), nil
		}

		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return "", fmt.Errorf("invalid time format, expected RFC3339: %w", err)
		}

		return t.Format(layout), nil
	}); err != nil {
		return err
	}

	return nil
}
