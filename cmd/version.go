package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GHAM",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GHAM version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
