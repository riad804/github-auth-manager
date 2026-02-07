package authmanager

import (
	"fmt"

	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
)

var logoutID string

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from an account",
	Long:  `Removes the stored token and account metadata for the specified ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Remove from keyring
		err := storage.DeleteToken(logoutID)
		if err != nil {
			fmt.Printf("Error deleting token: %v\n", err)
			return
		}

		// Remove from config
		err = storage.DeleteAccount(logoutID)
		if err != nil {
			fmt.Printf("Error deleting account: %v\n", err)
			return
		}

		fmt.Printf("Successfully logged out from account %s\n", logoutID)
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
	logoutCmd.Flags().StringVar(&logoutID, "id", "", "Account ID to logout")
	logoutCmd.MarkFlagRequired("id")
}
