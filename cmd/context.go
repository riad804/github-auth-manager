package cmd

import (
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage GitHub authentication contexts",
	Long: `Provides subcommands to add, list, remove, and manage GitHub contexts.
Each context typically represents a different GitHub identity (e.g., personal, work).`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help() // Show help if 'gham context' is called without subcommand
		}
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
