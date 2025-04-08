package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"gofast/internal/config"
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
	// Load settings from file
	settings, err := config.LoadSettings(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// Display the configuration
	color.Blue("\nFastAgent Configuration")
	fmt.Printf("\nConfiguration file: %s\n\n", configFile)

	// Display default model
	color.Green("default_model: %s\n", settings.DefaultModel)

	// Display logger settings
	color.Green("\nlogger:")
	fmt.Printf("  type: %s\n", settings.Logger.Type)
	fmt.Printf("  level: %s\n", settings.Logger.Level)
	fmt.Printf("  progress_display: %v\n", settings.Logger.ProgressDisplay)
	fmt.Printf("  path: %s\n", settings.Logger.Path)
	fmt.Printf("  batch_size: %d\n", settings.Logger.BatchSize)

	// Display MCP server settings
	color.Green("\nmcp:")
	color.Green("  servers:")
	for name, server := range settings.MCP.Servers {
		fmt.Printf("    %s:\n", name)
		fmt.Printf("      name: %s\n", server.Name)
		if server.Description != "" {
			fmt.Printf("      description: %s\n", server.Description)
		}
		fmt.Printf("      transport: %s\n", server.Transport)
		if server.Command != "" {
			fmt.Printf("      command: %s\n", server.Command)
		}
		if len(server.Args) > 0 {
			fmt.Printf("      args: %v\n", server.Args)
		}
		if server.URL != "" {
			fmt.Printf("      url: %s\n", server.URL)
		}
		if len(server.Env) > 0 {
			fmt.Printf("      env:\n")
			for k, v := range server.Env {
				fmt.Printf("        %s: %s\n", k, v)
			}
		}
	}

	// Show some helpful tips
	if verbose {
		fmt.Println("\nTips:")
		fmt.Println("- Configuration can be overridden with environment variables")
		fmt.Println("- API keys can be set in fastagent.secrets.yaml or as environment variables")
		fmt.Println("- Use --verbose flag to see more details about configuration")
	}

	return nil
}
