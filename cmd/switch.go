package cmd

import (
	"fmt"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch to a different Discourse environment",
	Long: `Stop the currently running environment (if any) and start a different one.

This is a convenience command that combines 'denver stop' and 'denver start'.

Examples:
  denver switch yaks
  denver switch test-pr`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return switchEnvironment(name)
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func switchEnvironment(name string) error {
	// Check if an environment is currently running
	active, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("failed to check active environment: %w", err)
	}

	// If same environment is already running, nothing to do
	if active != nil && active.Name == name {
		fmt.Printf("Environment '%s' is already running\n", name)
		return nil
	}

	// Stop current environment if one is running
	if active != nil {
		fmt.Printf("Stopping current environment '%s'...\n", active.Name)
		if err := stopEnvironment(); err != nil {
			return fmt.Errorf("failed to stop current environment: %w", err)
		}
		fmt.Println()
	}

	// Start the new environment
	return startEnvironment(name)
}
