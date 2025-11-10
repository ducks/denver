package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <name>",
	Short: "Setup a Discourse environment",
	Long: `Install dependencies and setup database for a Discourse environment.

This command runs:
- bundle install (Ruby dependencies)
- bundle exec rake db:create db:migrate (database setup)
- yarn install (JavaScript dependencies)

Run this once after creating an environment, before starting it.

Examples:
  denver setup yaks
  denver setup test-pr`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return setupEnvironment(name)
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func setupEnvironment(name string) error {
	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, name)
	discourseDir := filepath.Join(envDir, "discourse")

	// Check if environment exists
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", name)
	}

	fmt.Printf("Setting up environment '%s'...\n\n", name)

	// Bundle install
	fmt.Println("Installing Ruby dependencies (bundle install)...")
	if err := runCommand(discourseDir, "bundle", "install"); err != nil {
		return fmt.Errorf("bundle install failed: %w", err)
	}
	fmt.Println("✓ Ruby dependencies installed")

	// Pnpm install (must run before db:migrate because db:migrate precompiles assets)
	fmt.Println("\nInstalling JavaScript dependencies (pnpm install)...")
	if err := runCommand(discourseDir, "pnpm", "install"); err != nil {
		return fmt.Errorf("pnpm install failed: %w", err)
	}
	fmt.Println("✓ JavaScript dependencies installed")

	// Database setup (runs after pnpm because db:migrate precompiles assets)
	fmt.Println("\nSetting up database...")
	if err := runCommand(discourseDir, "bundle", "exec", "rake", "db:create", "db:migrate"); err != nil {
		return fmt.Errorf("database setup failed: %w", err)
	}
	fmt.Println("✓ Database created and migrated")

	fmt.Printf("\n✓ Environment '%s' is ready!\n", name)
	fmt.Printf("Use 'denver start %s' to start the servers\n", name)

	return nil
}

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
