package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
	noColor bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hive",
	Short: "Hive - Intelligent Agent Framework",
	Long: `Hive is a framework for building and deploying intelligent agents.
It provides tools for agent composition, team coordination, and deployment.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Apply global flags
		if noColor {
			color.NoColor = true
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return showWelcome()
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Disable all output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")

	// Initialize subcommands
	rootCmd.AddCommand(
		setupCmd(),
		bootstrapCmd(),
		configCmd(),
	)

	// Disable the completion command for now since we haven't implemented it
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func showWelcome() error {
	if !noColor {
		color.New(color.Bold).Printf("\nhive v0.1.0\n")
	} else {
		fmt.Printf("\nhive v0.1.0\n")
	}
	fmt.Println("Build effective agents using Model Context Protocol (MCP)")

	fmt.Println("\nAvailable Commands:")
	fmt.Println("  setup      Set up a new agent project with configuration files")
	fmt.Println("  bootstrap  Create example applications (workflow, researcher, etc.)")
	fmt.Println("  config     Manage FastAgent configuration")

	fmt.Println("\nGetting Started:")
	fmt.Println("1. Set up a new project:")
	fmt.Println("   hive setup")
	fmt.Println("\n2. Create Building Effective Agents workflow examples:")
	fmt.Println("   hive bootstrap workflow")
	fmt.Println("\n3. Explore other examples:")
	fmt.Println("   hive bootstrap")

	fmt.Println("\nUse --help with any command for more information")
	fmt.Println("Example: hive bootstrap --help")

	return nil
}

func showVersion() {
	if color.NoColor {
		fmt.Printf("\nhive v0.1.0\n")
	} else {
		color.New(color.Bold).Printf("\nhive v0.1.0\n")
	}
}

func showHelp(cmd *cobra.Command, args []string) {
	fmt.Println("\nCommon Commands:")
	fmt.Println("   hive setup")
	fmt.Println("   hive bootstrap workflow")
	fmt.Println("   hive bootstrap")
	fmt.Println("")
	fmt.Println("Example: hive bootstrap --help")
}
