package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running Discourse environment",
	Long: `Stop the currently running Discourse environment's Rails and Ember servers.

Examples:
  denver stop`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopEnvironment()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func stopEnvironment() error {
	// Load active environment
	active, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("failed to load active environment: %w", err)
	}

	if active == nil {
		return fmt.Errorf("no environment is currently running")
	}

	fmt.Printf("Stopping environment '%s'...\n", active.Name)

	// Kill Rails server
	if err := killProcess(active.RailsPID, "Rails"); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// Kill Ember server
	if err := killProcess(active.EmberPID, "Ember"); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// Clear active file
	if err := config.ClearActive(); err != nil {
		return fmt.Errorf("failed to clear active environment: %w", err)
	}

	fmt.Printf("\n✓ Environment '%s' stopped\n", active.Name)

	return nil
}

func killProcess(pid int, name string) error {
	// Kill the entire process group (negative PID kills the group)
	// This ensures child processes (like ember's thread-loader workers) are also killed
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		// If process group kill fails, try killing just the process
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("failed to find %s process (PID: %d): %w", name, pid, err)
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to stop %s server (PID: %d): %w", name, pid, err)
		}
	}

	fmt.Printf("✓ %s server stopped (PID: %d)\n", name, pid)

	return nil
}
