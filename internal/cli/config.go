package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage FastAgent configuration",
		Long: `View and modify FastAgent configuration settings.

The config command allows you to view and manage your FastAgent configuration:

  - View current configuration:
    gofast config show

  - View configuration from a specific file:
    gofast config show --file=/path/to/config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(configShowCmd())
	return cmd
}

func configShowCmd() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long: `Display the current FastAgent configuration settings.

By default, looks for fastagent.config.yaml in the current directory.
Use --file to specify a different configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showConfig(configFile)
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "fastagent.config.yaml", "Configuration file to display")
	return cmd
}

func showConfig(configFile string) error {
	// Resolve the config file path
	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return fmt.Errorf("failed to resolve config file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", absPath)
	}

	// Read and parse the config file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into a generic map
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Display the configuration
	color.Blue("\nFastAgent Configuration")
	fmt.Printf("\nConfiguration file: %s\n\n", absPath)

	// Display the configuration in a formatted way
	displayConfig(config, 0)

	// Show some helpful tips
	if verbose {
		fmt.Println("\nTips:")
		fmt.Println("- Configuration can be overridden with environment variables")
		fmt.Println("- API keys can be set in fastagent.secrets.yaml or as environment variables")
		fmt.Println("- Use --verbose flag to see more details about configuration")
	}

	return nil
}

func displayConfig(config map[string]interface{}, indent int) {
	for key, value := range config {
		indentStr := fmt.Sprintf("%*s", indent*2, "")

		switch v := value.(type) {
		case map[string]interface{}:
			color.Green("%s%s:", indentStr, key)
			displayConfig(v, indent+1)
		case []interface{}:
			color.Green("%s%s:", indentStr, key)
			for _, item := range v {
				fmt.Printf("%s  - %v\n", indentStr, item)
			}
		default:
			if verbose {
				// In verbose mode, show more details about configuration values
				fmt.Printf("%s%s: %v", indentStr, key, value)
				if key == "default_model" {
					fmt.Printf(" (current)")
				}
				fmt.Println()
			} else {
				fmt.Printf("%s%s: %v\n", indentStr, key, value)
			}
		}
	}
}
