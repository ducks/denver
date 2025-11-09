package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a Discourse environment",
	Long: `Start Rails and Ember servers for a Discourse environment.

Only one environment can be running at a time. Servers will run on:
- Rails: http://localhost:3000
- Ember: http://localhost:4200

Examples:
  denver start yaks
  denver start test-pr`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return startEnvironment(name)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func startEnvironment(name string) error {
	// Check if another environment is already running
	active, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("failed to check active environment: %w", err)
	}

	if active != nil {
		return fmt.Errorf("environment '%s' is already running. Stop it first with: denver stop", active.Name)
	}

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

	fmt.Printf("Starting environment '%s'...\n", name)

	// Start Rails server
	fmt.Println("Starting Rails server on port 3000...")
	railsCmd := exec.Command("bundle", "exec", "rails", "server", "-p", "3000")
	railsCmd.Dir = discourseDir
	railsCmd.Stdout = nil // Run in background
	railsCmd.Stderr = nil

	if err := railsCmd.Start(); err != nil {
		return fmt.Errorf("failed to start Rails server: %w", err)
	}

	railsPID := railsCmd.Process.Pid
	fmt.Printf("✓ Rails server started (PID: %d)\n", railsPID)

	// Start Ember server
	fmt.Println("Starting Ember server on port 4200...")
	emberCmd := exec.Command("ember", "serve", "--port", "4200")
	emberCmd.Dir = discourseDir
	emberCmd.Stdout = nil // Run in background
	emberCmd.Stderr = nil

	if err := emberCmd.Start(); err != nil {
		// If ember fails, kill rails
		railsCmd.Process.Kill()
		return fmt.Errorf("failed to start Ember server: %w", err)
	}

	emberPID := emberCmd.Process.Pid
	fmt.Printf("✓ Ember server started (PID: %d)\n", emberPID)

	// Save active environment
	activeEnv := &config.ActiveEnvironment{
		Name:      name,
		RailsPID:  railsPID,
		EmberPID:  emberPID,
		RailsPort: 3000,
		EmberPort: 4200,
	}

	if err := config.SaveActive(activeEnv); err != nil {
		// Cleanup on failure
		railsCmd.Process.Kill()
		emberCmd.Process.Kill()
		return fmt.Errorf("failed to save active environment: %w", err)
	}

	fmt.Printf("\n✓ Environment '%s' is now running\n", name)
	fmt.Println("  Rails:  http://localhost:3000")
	fmt.Println("  Ember:  http://localhost:4200")
	fmt.Println("\nUse 'denver stop' to stop the servers")

	return nil
}
