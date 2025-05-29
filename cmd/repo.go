package cmd

import (
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repository-specific context assignments",
	Long:  `Assign contexts to specific repositories or view current assignments.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
	// repo_assign.go and repo_current.go will add their commands to repoCmd
}
