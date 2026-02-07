package authmanager

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/riad804/auth-manager/internals/git"
	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
)

var linkID string

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link current repo to an account",
	Long:  `Binds the current Git repository to a specific account for authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Detect current repo
		repoDir, err := git.FindGitRepo()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Detect provider from remote
		remoteURL, err := git.GetRemoteURL("origin")
		if err != nil {
			fmt.Printf("Error getting remote: %v\n", err)
			return
		}
		provider := git.DetectProvider(remoteURL)

		// Get accounts for provider
		accounts, err := storage.GetAccountsByProvider(provider)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if len(accounts) == 0 {
			fmt.Printf("No accounts for %s. Please login first.\n", provider)
			return
		}

		// If ID provided, use it
		if linkID != "" {
			// Verify exists
			_, err := storage.GetAccount(linkID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
		} else {
			// Show list
			fmt.Printf("Available accounts for %s:\n", provider)
			for i, acc := range accounts {
				fmt.Printf("%d: ID: %s, Email: %s\n", i+1, acc.ID, acc.Email)
			}

			// Prompt selection
			fmt.Print("Select account number: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			num, err := strconv.Atoi(input[:len(input)-1])
			if err != nil || num < 1 || num > len(accounts) {
				fmt.Println("Invalid selection.")
				return
			}
			linkID = accounts[num-1].ID
		}

		// Store mapping
		err = storage.StoreRepoMapping(repoDir, linkID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Repo linked to account %s\n", linkID)
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.Flags().StringVar(&linkID, "id", "", "Account ID to link (optional, prompts if not set)")
}
