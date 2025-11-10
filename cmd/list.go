package cmd

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments",
	Long: `List all Denver environments with their status.

Examples:
  denver list`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listEnvironments()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

type envInfo struct {
	name     string
	modified time.Time
	isActive bool
}

func listEnvironments() error {
	// Get environments directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	// Check if environments directory exists
	if _, err := os.Stat(envsDir); os.IsNotExist(err) {
		fmt.Println("No environments found")
		fmt.Println("\nCreate one with: denver create <name> --profile <profile>")
		return nil
	}

	// Read environments
	entries, err := os.ReadDir(envsDir)
	if err != nil {
		return fmt.Errorf("failed to read environments directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No environments found")
		fmt.Println("\nCreate one with: denver create <name> --profile <profile>")
		return nil
	}

	// Get active environment
	active, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("failed to load active environment: %w", err)
	}

	// Collect environment info
	var envs []envInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		isActive := false
		if active != nil && active.Name == entry.Name() {
			isActive = true
		}

		envs = append(envs, envInfo{
			name:     entry.Name(),
			modified: info.ModTime(),
			isActive: isActive,
		})
	}

	// Sort by modification time (most recent first)
	sort.Slice(envs, func(i, j int) bool {
		return envs[i].modified.After(envs[j].modified)
	})

	// Print environments
	fmt.Printf("Environments (%d):\n\n", len(envs))

	for _, env := range envs {
		status := " "
		if env.isActive {
			status = "●"
		}

		// Show relative time
		relTime := formatRelativeTime(env.modified)
		fmt.Printf("  %s %s (modified %s)\n", status, env.name, relTime)
	}

	if active != nil {
		fmt.Printf("\n● = running\n")
	}

	return nil
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}

	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	return t.Format("Jan 2, 2006")
}
