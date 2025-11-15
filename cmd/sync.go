package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ducks/denver/internal/config"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <env> [target]",
	Short: "Update discourse core and/or plugins in an environment",
	Long: `Update discourse core and/or plugins by pulling latest changes from git.

Target options:
  (none)          - Update discourse core and all remote plugins
  discourse       - Update only discourse core
  <plugin-name>   - Update only the specified plugin

Local plugins (created with --local) are skipped since they're symlinks
to your development directories.

Examples:
  denver sync frndr                    # Update discourse + all remote plugins
  denver sync frndr discourse          # Update only discourse core
  denver sync frndr discourse-chat     # Update only discourse-chat plugin`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		envName := args[0]
		target := ""
		if len(args) > 1 {
			target = args[1]
		}
		return syncEnvironment(envName, target)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func syncEnvironment(envName string, target string) error {
	// Get environment directory
	envsDir, err := config.GetEnvironmentsDir()
	if err != nil {
		return err
	}

	envDir := filepath.Join(envsDir, envName)
	discourseDir := filepath.Join(envDir, "discourse")
	pluginsDir := filepath.Join(envDir, "plugins")

	// Check if environment exists
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", envName)
	}

	// Load environment config to know which plugins are local vs remote
	envConfig, err := config.LoadEnvironmentConfig(envDir)
	if err != nil {
		return fmt.Errorf("failed to load environment config: %w", err)
	}

	// If target is specified, sync only that target
	if target != "" {
		if target == "discourse" {
			return syncDiscourse(discourseDir)
		}

		// Find the plugin in config
		var plugin *config.PluginConfig
		for _, p := range envConfig.Plugins {
			if p.Name == target {
				plugin = &p
				break
			}
		}

		if plugin == nil {
			return fmt.Errorf("plugin '%s' not found in environment '%s'", target, envName)
		}

		if plugin.Local != "" {
			fmt.Printf("Plugin '%s' is local (symlinked to %s), skipping sync\n", plugin.Name, plugin.Local)
			return nil
		}

		return syncPlugin(pluginsDir, plugin.Name)
	}

	// No target specified, sync everything
	fmt.Printf("Syncing environment '%s'...\n\n", envName)

	// Sync discourse core
	if err := syncDiscourse(discourseDir); err != nil {
		return err
	}

	// Sync all remote plugins
	remotePlugins := 0
	localPlugins := 0

	for _, plugin := range envConfig.Plugins {
		if plugin.Local != "" {
			localPlugins++
			fmt.Printf("  ⊙ %s (local, skipping)\n", plugin.Name)
			continue
		}

		if err := syncPlugin(pluginsDir, plugin.Name); err != nil {
			fmt.Printf("  ✗ %s failed: %v\n", plugin.Name, err)
			continue
		}
		remotePlugins++
	}

	fmt.Printf("\n✓ Environment '%s' synced\n", envName)
	fmt.Printf("  Discourse: updated\n")
	fmt.Printf("  Remote plugins: %d updated\n", remotePlugins)
	if localPlugins > 0 {
		fmt.Printf("  Local plugins: %d skipped\n", localPlugins)
	}

	return nil
}

func syncDiscourse(discourseDir string) error {
	fmt.Println("Syncing discourse core...")

	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = discourseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to sync discourse: %w", err)
	}

	fmt.Println("✓ Discourse core synced")
	return nil
}

func syncPlugin(pluginsDir string, pluginName string) error {
	pluginPath := filepath.Join(pluginsDir, pluginName)

	// Check if plugin directory exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin directory not found: %s", pluginPath)
	}

	fmt.Printf("  Syncing %s...\n", pluginName)

	cmd := exec.Command("git", "pull", "--rebase")
	cmd.Dir = pluginPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to sync plugin: %w", err)
	}

	fmt.Printf("  ✓ %s synced\n", pluginName)
	return nil
}
