package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/gitutils"
	"github.com/spf13/cobra"
)

var repoAssignCmd = &cobra.Command{
	Use:   "assign <context-name> [path-to-repo-root]",
	Short: "Assign a GitHub context to a local repository",
	Long: `Assigns a previously defined GitHub context to a local Git repository.
The assignment is made to the root of the Git repository.
If [path-to-repo-root] is not provided, the current directory is assumed to be within the repository.`,
	Args: cobra.RangeArgs(1, 2), // context-name is required, path is optional
	RunE: func(cmd *cobra.Command, args []string) error {
		contextName := args[0]
		repoPathArg := "." // Default to current directory
		if len(args) == 2 {
			repoPathArg = args[1]
		}

		absPath, err := filepath.Abs(repoPathArg)
		if err != nil {
			return fmt.Errorf("invalid repository path argument '%s': %w", repoPathArg, err)
		}

		repoRoot, err := gitutils.FindRepoRoot(absPath)
		if err != nil {
			return fmt.Errorf("failed to find Git repository root at or above '%s': %w. Please provide the path to the root of a Git repository", absPath, err)
		}

		if _, found := config.FindContext(contextName); !found {
			return fmt.Errorf("context '%s' does not exist. Use 'gham context list' to see available contexts", contextName)
		}

		err = config.AssignRepoContext(repoRoot, contextName)
		if err != nil {
			return fmt.Errorf("failed to assign context '%s' to repository '%s': %w", contextName, repoRoot, err)
		}

		fmt.Printf("Context '%s' assigned to repository at '%s'.\n", contextName, repoRoot)
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoAssignCmd)
}
