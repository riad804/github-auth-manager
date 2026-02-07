package authmanager

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/riad804/auth-manager/internals/auth"
	"github.com/riad804/auth-manager/internals/git"
	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
)

// credentialCmd is the hidden command for Git credential helper mode
var credentialCmd = &cobra.Command{
	Use:    "credential",
	Short:  "Internal Git credential helper",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			os.Exit(1)
		}
		action := args[0]

		if action != "get" {
			// Ignore store/erase
			os.Exit(0)
		}

		// Parse stdin
		input := make(map[string]string)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				input[parts[0]] = parts[1]
			}
		}

		// Get repo dir
		repoDir, err := git.FindGitRepo()
		if err != nil {
			os.Exit(0) // Silent fail, Git falls back
		}

		// Get mapping
		accountID, err := storage.GetRepoMapping(repoDir)
		if err != nil {
			os.Exit(0)
		}

		// Get account
		account, err := storage.GetAccount(accountID)
		if err != nil {
			os.Exit(0)
		}

		// Get provider
		provider, err := auth.GetProvider(account.Provider)
		if err != nil {
			os.Exit(0)
		}

		// Get token (with refresh if needed)
		token, err := storage.GetToken(accountID)
		if err != nil {
			os.Exit(0)
		}

		// Refresh if necessary
		token = auth.MaybeRefreshToken(provider, accountID, token)

		// Output per provider
		var username, password string
		switch account.Provider {
		case "github":
			username = token.AccessToken
			password = ""
		case "gitlab":
			username = "oauth2"
			password = token.AccessToken
		case "bitbucket":
			username = "x-token-auth"
			password = token.AccessToken
		}

		fmt.Printf("username=%s\n", username)
		fmt.Printf("password=%s\n", password)
	},
}

func init() {
	rootCmd.AddCommand(credentialCmd)
}
