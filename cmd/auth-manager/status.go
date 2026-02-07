package authmanager

import (
	"fmt"

	"github.com/riad804/auth-manager/internals/git"
	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current repo's linked account",
	Run: func(cmd *cobra.Command, args []string) {
		repoDir, err := git.FindGitRepo()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		accountID, err := storage.GetRepoMapping(repoDir)
		if err != nil {
			fmt.Println("No account linked.")
			return
		}

		account, err := storage.GetAccount(accountID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Linked to: ID: %s, Provider: %s, Email: %s\n", account.ID, account.Provider, account.Email)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
