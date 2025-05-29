package cmd

import (
	"fmt"
	"os"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/spf13/cobra"
)

// Version will be set by main.go from ldflags or default
var Version string

var rootCmd = &cobra.Command{
	Use:   "gham",
	Short: "GitHub Authentication Manager (GHAM) CLI",
	Long: `GHAM is a lightweight, secure CLI utility that enables seamless management
of GitHub authentication contexts. It helps developers working with multiple
GitHub accounts (personal, professional, client-based).`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config initialization for 'completion' and 'version' commands
		if cmd.Name() == "completion" || cmd.Name() == "version" || (cmd.Parent() != nil && cmd.Parent().Name() == "completion") {
			return nil
		}
		if err := config.InitConfig(); err != nil {
			// Provide a more user-friendly message if config init fails
			fmt.Fprintf(os.Stderr, "Error initializing configuration: %v\n", err)
			fmt.Fprintln(os.Stderr, "Please ensure your user config directory is writable.")
			fmt.Fprintf(os.Stderr, "Attempted config path: %s\n", config.GetConfigFilePathForError()) // Helper for error
			return err                                                                                // Return error to stop execution
		}
		return nil
	},
	// SilenceUsage is useful for CLI tools to not show usage on every error.
	// Errors are handled and printed, then os.Exit(1) is called.
	SilenceUsage: true,
	// SilenceErrors: true, // If you want to handle all error printing yourself
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra prints the error by default. If SilenceErrors is true, you'd print here.
		// fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	// rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}
