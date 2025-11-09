package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "denver",
	Short: "Denver - Discourse ENVironments managER",
	Long: `Denver makes it easy to create, manage, and switch between multiple
isolated Discourse development environments.

Create environments from YAML profiles that define plugin sets, themes,
site settings, and more. Each environment gets its own isolated dev
container with unique ports and volumes.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags would go here
}
