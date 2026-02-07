package authmanager

import (
	"fmt"

	"github.com/riad804/auth-manager/internals/git"
	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
)

// unlinkCmd represents the unlink command
var unlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlink current repo from account",
	Run: func(cmd *cobra.Command, args []string) {
		repoDir, err := git.FindGitRepo()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		err = storage.DeleteRepoMapping(repoDir)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Repo unlinked.")
	},
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
}
