package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy <name>",
	Short: "Destroy a Discourse environment",
	Long: `Destroy an existing Discourse development environment.

This will delete the entire environment directory including the cloned
discourse repository and all plugins. This action cannot be undone.

Examples:
  denver destroy test-env
  denver destroy yaks`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return destroyEnvironment(name)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}

func destroyEnvironment(name string) error {
	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, name)

	// Check if environment exists
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", name)
	}

	fmt.Printf("Destroying environment '%s'...\n", name)
	fmt.Printf("Removing: %s\n", envDir)

	// Remove the entire environment directory
	if err := os.RemoveAll(envDir); err != nil {
		return fmt.Errorf("failed to destroy environment: %w", err)
	}

	fmt.Printf("✓ Environment '%s' destroyed successfully\n", name)

	return nil
}
