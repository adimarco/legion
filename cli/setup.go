package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/adimarco/hive/cli/templates"
)

var (
	force bool
)

func setupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup [directory]",
		Short: "Set up a new agent project",
		Long:  `Set up a new agent project with configuration files and example agent.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			return setupProject(dir)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite existing files")
	return cmd
}

func setupProject(dir string) error {
	// Resolve and create directory if needed
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get the module name from the directory name
	moduleName := filepath.Base(absPath)
	if moduleName == "." {
		// If in current directory, use the directory name
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		moduleName = filepath.Base(cwd)
	}

	// Create project structure
	files := map[string]string{
		"fastagent.config.yaml":  templates.ConfigTemplate,
		"fastagent.secrets.yaml": templates.SecretsTemplate,
		".gitignore":             templates.GitignoreTemplate,
		"main.go":                templates.MainTemplate,
		"go.mod":                 fmt.Sprintf(templates.ModTemplate, moduleName),
	}

	color.Blue("\nSetting up new FastAgent project in %s\n", absPath)
	fmt.Println("\nCreating files:")

	for filename, content := range files {
		path := filepath.Join(absPath, filename)

		// Check if file exists
		if _, err := os.Stat(path); err == nil && !force {
			color.Yellow("  Skipping %s (already exists)", filename)
			continue
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
		color.Green("  Created %s", filename)
	}

	color.Green("\nSetup completed successfully!")
	fmt.Println("\nImportant steps:")
	color.Yellow("1. Add your API keys to fastagent.secrets.yaml or set as environment variables:")
	fmt.Println("   - OPENAI_API_KEY")
	fmt.Println("   - ANTHROPIC_API_KEY")
	color.Yellow("2. Keep fastagent.secrets.yaml secure and never commit it to version control")
	color.Yellow("3. Update fastagent.config.yaml to configure your agent")

	fmt.Println("\nTo get started:")
	fmt.Println("1. cd", dir)
	fmt.Println("2. go mod tidy")
	fmt.Println("3. go run main.go")

	return nil
}
