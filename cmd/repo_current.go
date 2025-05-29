package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/gitutils"
	"github.com/spf13/cobra"
)

var repoCurrentCmd = &cobra.Command{
	Use:   "current [path-to-repo]",
	Short: "Show the assigned GHAM context for the current (or specified) repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lookupPath := "."
		if len(args) == 1 {
			lookupPath = args[0]
		}

		absPath, err := filepath.Abs(lookupPath)
		if err != nil {
			return fmt.Errorf("invalid path '%s': %w", lookupPath, err)
		}

		repoRoot, err := gitutils.FindRepoRoot(absPath)
		if err != nil {
			// This is not an error for the command itself, just info
			fmt.Printf("Directory '%s' is not within a Git repository tracked by GHAM or is not a Git repository.\n", absPath)
			return nil
		}

		contextName, found := config.GetRepoContextName(repoRoot)
		if !found {
			fmt.Printf("No GHAM context is explicitly assigned to the repository at: %s\n", repoRoot)
			fmt.Println("Git operations will use your global or system Git configuration.")
			return nil
		}

		fmt.Printf("Repository: %s\n", repoRoot)
		fmt.Printf("Assigned GHAM Context: %s\n", contextName)

		// Optionally, display more info about the context
		if ctx, ctxFound := config.FindContext(contextName); ctxFound {
			fmt.Printf("  Username: %s\n", ctx.Username)
			fmt.Printf("  Email: %s\n", ctx.Email)
		} else {
			fmt.Printf("  Warning: Context '%s' is assigned but its definition was not found in GHAM configuration.\n", contextName)
		}
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoCurrentCmd)
}
