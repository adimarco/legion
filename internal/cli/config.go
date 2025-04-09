package cli

import (
	"fmt"
	"os"

	"github.com/adimarco/hive/internal/config"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Hive configuration",
		Long:  `View and modify Hive configuration settings.`,
	}

	cmd.AddCommand(configShowCmd())
	return cmd
}

func configShowCmd() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long: `Display the current Hive configuration.
Example:
  hive config show
  hive config show --file=/path/to/config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadSettings(configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
				return err
			}

			// Pretty print the configuration
			fmt.Printf("Current Hive Configuration:\n\n")
			fmt.Printf("Default Model: %s\n", cfg.DefaultModel)
			fmt.Printf("Log Level: %s\n", cfg.Logger.Level)

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Path to configuration file")
	return cmd
}
