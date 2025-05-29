package cmd

import (
	"fmt"
	"os"

	"github.com/riad804/github-auth-manager/internal/gitutils"
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git [git command and arguments...]",
	Short: "Wrap git commands to inject context-specific credentials",
	Long: `Wraps any git command (e.g., clone, push, pull) and automatically injects
the appropriate GitHub credentials based on the repository's assigned GHAM context.
For example: 'gham git clone <url>' or 'gham git push'.
If no context is assigned to the current repository, it falls back to your system's Git configuration.`,
	DisableFlagParsing: true, // Pass all flags directly to the underlying git command
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Mimic git's behavior or show GHAM-specific help for 'gham git'
			fmt.Println("gham: 'git' requires a Git command.")
			fmt.Println("Example: gham git clone https://github.com/user/repo.git")
			fmt.Println("Example: gham git status")
			// Optionally, execute 'git --help'
			// return gitutils.ExecuteGitCommandWithContext([]string{"--help"})
			return nil // Return nil to avoid Cobra printing its own usage for this specific case
		}

		// For debugging purposes, you might want to see what GHAM is doing.
		// This should be behind a verbose flag in a real application.
		// fmt.Printf("[GHAM DEBUG] Wrapping: git %s\n", strings.Join(args, " "))

		err := gitutils.ExecuteGitCommandWithContext(args, os.Stdout, os.Stderr)
		if err != nil {
			// The error from ExecuteGitCommandWithContext should be descriptive enough.
			// Cobra will print it if not silenced, and main.go will os.Exit(1).
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
