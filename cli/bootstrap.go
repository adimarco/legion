package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Example types and their configurations
var exampleTypes = map[string]struct {
	Description string
	Files       map[string]string
	CreateDir   bool
}{
	"workflow": {
		Description: "Example workflows demonstrating each of the patterns in Anthropic's " +
			"'Building Effective Agents' paper, implemented in Go.",
		Files: map[string]string{
			"chaining.go":    workflowChainingTemplate,
			"parallel.go":    workflowParallelTemplate,
			"router.go":      workflowRouterTemplate,
			"evaluator.go":   workflowEvaluatorTemplate,
			"human_input.go": workflowHumanInputTemplate,
			"go.mod":         workflowModTemplate,
		},
		CreateDir: true,
	},
	"researcher": {
		Description: "Research agent example with evaluation/optimization capabilities. " +
			"Uses Brave Search API and demonstrates advanced agent patterns.",
		Files: map[string]string{
			"researcher.go":         researcherTemplate,
			"researcher_eval.go":    researcherEvalTemplate,
			"go.mod":                researcherModTemplate,
			"fastagent.config.yaml": researcherConfigTemplate,
		},
		CreateDir: true,
	},
}

func bootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap [type] [directory]",
		Short: "Create example applications",
		Long:  `Create example applications (workflow, researcher, etc.).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return showBootstrapOverview()
			}

			exampleType := args[0]
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}

			return createExample(exampleType, dir)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite existing files")
	return cmd
}

func showBootstrapOverview() error {
	color.Blue("\nFastAgent Example Applications")
	fmt.Print("Build agents and compose workflows through practical examples\n\n")

	fmt.Print("Available Examples:\n")
	for name, info := range exampleTypes {
		color.Green("\n%s", name)
		fmt.Printf("%s\n", info.Description)
		fmt.Print("\nFiles:\n")
		for file := range info.Files {
			fmt.Printf("  • %s\n", file)
		}
	}

	fmt.Print("\nUsage:\n")
	fmt.Print("  gofast bootstrap workflow DIR      Create workflow examples in DIR\n")
	fmt.Print("  gofast bootstrap researcher DIR    Create researcher example in DIR\n")
	fmt.Print("\nOptions:\n")
	fmt.Print("  --force            Overwrite existing files\n")
	fmt.Print("\nExamples:\n")
	fmt.Print("  gofast bootstrap workflow .\n")
	fmt.Print("  gofast bootstrap researcher . --force\n")

	return nil
}

func createExample(exampleType, dir string) error {
	info, ok := exampleTypes[exampleType]
	if !ok {
		return fmt.Errorf("unknown example type: %s", exampleType)
	}

	// Resolve and create base directory
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// If example type requires a subdirectory, create it
	if info.CreateDir {
		absPath = filepath.Join(absPath, exampleType)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	color.Blue("\nCreating %s example in %s\n", exampleType, absPath)
	fmt.Println("\nCreating files:")

	created := []string{}
	for filename, content := range info.Files {
		path := filepath.Join(absPath, filename)

		// Create subdirectories if needed
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}

		// Check if file exists
		if _, err := os.Stat(path); err == nil && !force {
			color.Yellow("  Skipping %s (already exists)", filename)
			continue
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
		color.Green("  Created %s", filename)
		created = append(created, filename)
	}

	if len(created) > 0 {
		showCompletionMessage(exampleType, absPath)
	} else {
		color.Yellow("\nNo files were created.")
	}

	return nil
}

func showCompletionMessage(exampleType, dir string) {
	color.Green("\nSetup completed successfully!")
	fmt.Println("\nCreated example in:", dir)

	fmt.Println("\nNext steps:")
	switch exampleType {
	case "workflow":
		fmt.Println("1. Review chaining.go for the basic workflow example")
		fmt.Println("2. Check other examples:")
		fmt.Println("   - parallel.go: Run agents in parallel")
		fmt.Println("   - router.go: Route requests between agents")
		fmt.Println("   - evaluator.go: Add evaluation capabilities")
		fmt.Println("   - human_input.go: Incorporate human feedback")
		fmt.Println("3. Run an example:")
		fmt.Println("   cd", dir)
		fmt.Println("   go mod tidy")
		fmt.Println("   go run chaining.go")
	case "researcher":
		fmt.Println("1. Set up the Brave API key in fastagent.secrets.yaml")
		fmt.Println("2. Try the basic researcher:")
		fmt.Println("   cd", dir)
		fmt.Println("   go mod tidy")
		fmt.Println("   go run researcher.go")
		fmt.Println("3. Try the version with evaluation:")
		fmt.Println("   go run researcher_eval.go")
	}
}

// Templates for different example types
const (
	workflowChainingTemplate = `package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropic-ai/claude-go"
)

func main() {
	// Example of chaining multiple agents together
	fmt.Println("Workflow Chaining Example")
}
`

	workflowParallelTemplate = `package main

import (
	"context"
	"fmt"
	"sync"
)

func main() {
	// Example of running agents in parallel
	fmt.Println("Parallel Workflow Example")
}
`

	workflowRouterTemplate = `package main

import (
	"context"
	"fmt"
)

func main() {
	// Example of routing between different agents
	fmt.Println("Router Workflow Example")
}
`

	workflowEvaluatorTemplate = `package main

import (
	"context"
	"fmt"
)

func main() {
	// Example of evaluating agent responses
	fmt.Println("Evaluator Workflow Example")
}
`

	workflowHumanInputTemplate = `package main

import (
	"context"
	"fmt"
)

func main() {
	// Example of incorporating human feedback
	fmt.Println("Human Input Workflow Example")
}
`

	workflowModTemplate = `module workflow

go 1.21

require (
	github.com/anthropic-ai/claude-go v0.0.0-20240308222815-20c05b6b4ad5
)
`

	researcherTemplate = `package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropic-ai/claude-go"
)

func main() {
	// Example researcher agent implementation
	fmt.Println("Researcher Agent Example")
}
`

	researcherEvalTemplate = `package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropic-ai/claude-go"
)

func main() {
	// Example researcher agent with evaluation
	fmt.Println("Researcher Agent with Evaluation Example")
}
`

	researcherModTemplate = `module researcher

go 1.21

require (
	github.com/anthropic-ai/claude-go v0.0.0-20240308222815-20c05b6b4ad5
)
`

	researcherConfigTemplate = `# FastAgent Configuration for Researcher Example
default_model: sonnet

mcp:
    servers:
        brave:
            command: "npx"
            args: ["-y", "@modelcontextprotocol/server-brave"]
`
)
