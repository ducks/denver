package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of the running environment",
	Long: `Show detailed information about the currently running Discourse environment
including branch, plugins, ports, and process information.

Examples:
  denver status`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func showStatus() error {
	// Load active environment
	active, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("failed to load active environment: %w", err)
	}

	if active == nil {
		fmt.Println("No environment is currently running")
		fmt.Println("\nUse 'denver start <name>' to start an environment")
		return nil
	}

	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, active.Name)
	discourseDir := filepath.Join(envDir, "discourse")
	pluginsDir := filepath.Join(envDir, "plugins")

	fmt.Printf("Environment: %s (running)\n", active.Name)
	fmt.Printf("Location: %s\n\n", envDir)

	// Get git branch info
	branch, err := getGitBranch(discourseDir)
	if err != nil {
		fmt.Printf("Branch: (unknown)\n")
	} else {
		fmt.Printf("Branch: %s\n", branch)
	}

	// Get commit info
	commit, err := getGitCommit(discourseDir)
	if err != nil {
		fmt.Printf("Commit: (unknown)\n")
	} else {
		fmt.Printf("Commit: %s\n", commit)
	}

	// Show ports and PIDs
	fmt.Printf("\nServers:\n")
	fmt.Printf("  Rails:  http://localhost:%d (PID: %d)\n", active.RailsPort, active.RailsPID)
	fmt.Printf("  Ember:  http://localhost:%d (PID: %d)\n", active.EmberPort, active.EmberPID)

	// List plugins
	plugins, err := listPlugins(pluginsDir)
	if err != nil || len(plugins) == 0 {
		fmt.Printf("\nPlugins: (none)\n")
	} else {
		fmt.Printf("\nPlugins (%d):\n", len(plugins))
		for _, plugin := range plugins {
			fmt.Printf("  - %s\n", plugin)
		}
	}

	return nil
}

func getGitBranch(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getGitCommit(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "log", "-1", "--pretty=format:%h - %s")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func listPlugins(pluginsDir string) ([]string, error) {
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			plugins = append(plugins, entry.Name())
		}
	}

	return plugins, nil
}
