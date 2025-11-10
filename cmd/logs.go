package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var (
	followLogs bool
	emberLogs  bool
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View logs for the active environment",
	Long: `View Rails or Ember logs for the currently running environment.

By default shows Rails logs (last 100 lines). Use -f to follow.
Use --ember to show Ember logs instead.

Examples:
  denver logs              # Show last 100 lines of Rails logs
  denver logs -f           # Follow Rails logs
  denver logs --ember      # Show Ember logs
  denver logs --ember -f   # Follow Ember logs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showLogs()
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output")
	logsCmd.Flags().BoolVar(&emberLogs, "ember", false, "Show Ember logs instead of Rails")
}

func showLogs() error {
	// Check if an environment is running
	active, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("failed to check active environment: %w", err)
	}

	if active == nil {
		return fmt.Errorf("no environment is currently running")
	}

	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, active.Name)
	discourseDir := filepath.Join(envDir, "discourse")

	var logPath string
	if emberLogs {
		// Ember logs (written by denver start command)
		logPath = filepath.Join(discourseDir, "log", "ember.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			fmt.Println("Ember log file not found. Ember may not have started yet.")
			fmt.Printf("Expected: %s\n", logPath)
			return nil
		}
	} else {
		// Rails logs (written by denver start command)
		logPath = filepath.Join(discourseDir, "log", "rails.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			return fmt.Errorf("rails log file not found at: %s", logPath)
		}
	}

	// Build tail command
	args := []string{}
	if followLogs {
		args = append(args, "-f")
	} else {
		args = append(args, "-n", "100")
	}
	args = append(args, logPath)

	cmd := exec.Command("tail", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
