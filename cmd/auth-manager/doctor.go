package authmanager

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/riad804/auth-manager/internals/git"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check and setup the tool",
	Long:  `Verifies Git installation, credential helper setup, and offers fixes.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check Git installed
		_, err := exec.LookPath("git")
		if err != nil {
			fmt.Println("Git not installed.")
			return
		}
		fmt.Println("Git installed: OK")

		// Check credential helper
		helpers, _ := git.GetGlobalConfig("credential.helper")
		// ourHelper := "!" + exec.Command("sh", "-c", "which git-auth-manager").Output() // Approximate, actual is "!git-auth-manager credential"
		ourHelperStr := "!git-auth-manager credential"
		if !strings.Contains(helpers, ourHelperStr) {
			fmt.Println("Credential helper not set. Setting now...")
			err = git.SetGlobalConfigAdd("credential.helper", ourHelperStr)
			if err != nil {
				fmt.Printf("Error setting helper: %v\n", err)
				return
			}
			fmt.Println("Set credential.helper: OK")
		} else {
			fmt.Println("Credential helper set: OK")
		}

		// OS detection
		fmt.Printf("OS: %s\n", git.DetectOS())

		fmt.Println("Setup complete.")
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
