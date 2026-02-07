package authmanager

import (
	"fmt"

	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
)

// accountsCmd represents the accounts command
var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List all logged-in accounts",
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := storage.GetAllAccounts()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(accounts) == 0 {
			fmt.Println("No accounts logged in.")
			return
		}

		fmt.Println("Logged-in accounts:")
		for _, acc := range accounts {
			fmt.Printf("- ID: %s, Provider: %s, Email: %s\n", acc.ID, acc.Provider, acc.Email)
		}
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}
