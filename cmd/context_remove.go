package cmd

import (
	"fmt"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/keyring"
	"github.com/spf13/cobra"
)

var contextRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Remove a GitHub context and its stored token",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		contextName := args[0]

		// Check if context exists before attempting removal
		if _, found := config.FindContext(contextName); !found {
			return fmt.Errorf("context '%s' not found", contextName)
		}

		// First, remove from keyring
		err := keyring.DeleteToken(contextName)
		if err != nil {
			// Log warning but proceed, as user might want to remove config even if keyring fails
			fmt.Printf("Warning: could not remove token for '%s' from keyring: %v\n", contextName, err)
			fmt.Println("Proceeding to remove context from configuration.")
		}

		// Then, remove from config
		removed, err := config.RemoveContext(contextName)
		if err != nil {
			return fmt.Errorf("failed to remove context '%s' from configuration: %w", contextName, err)
		}

		if !removed { // Should not happen if FindContext found it, but as a safeguard
			fmt.Printf("Context '%s' was not found in the configuration (this is unexpected).\n", contextName)
			return nil
		}

		fmt.Printf("Context '%s' and its associated token (if present in keyring) removed successfully.\n", contextName)
		fmt.Println("Any repositories previously assigned to this context have been unassigned.")
		return nil
	},
}

func init() {
	contextCmd.AddCommand(contextRemoveCmd)
}
