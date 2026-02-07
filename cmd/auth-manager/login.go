package authmanager

import (
	"fmt"

	"github.com/riad804/auth-manager/internals/auth"
	"github.com/riad804/auth-manager/internals/models"
	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
)

var loginProvider string
var loginID string

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to a Git provider using OAuth",
	Long:  `Initiates browser-based OAuth login and securely stores the token.`,
	Run: func(cmd *cobra.Command, args []string) {
		provider, err := auth.GetProvider(loginProvider)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Perform OAuth flow
		token, err := auth.PerformOAuth(provider)
		if err != nil {
			fmt.Printf("Error during OAuth: %v\n", err)
			return
		}

		// Fetch user info for email
		email, err := auth.FetchUserEmail(provider, token.AccessToken)
		if err != nil {
			fmt.Printf("Error fetching user info: %v\n", err)
			return
		}

		// Create account
		account := models.Account{
			ID:       loginID,
			Provider: loginProvider,
			Email:    email,
		}

		// Store token in keyring
		err = storage.StoreToken(loginID, token)
		if err != nil {
			fmt.Printf("Error storing token: %v\n", err)
			return
		}

		// Store account metadata in config
		err = storage.StoreAccount(account)
		if err != nil {
			fmt.Printf("Error storing account: %v\n", err)
			return
		}

		fmt.Printf("Successfully logged in as %s for %s (%s)\n", email, loginProvider, loginID)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginProvider, "provider", "", "Git provider (github, gitlab, bitbucket)")
	loginCmd.MarkFlagRequired("provider")
	loginCmd.Flags().StringVar(&loginID, "id", "", "Unique account ID (e.g., work, personal)")
	loginCmd.MarkFlagRequired("id")
}
