package cmd

import (
	"fmt"
	"strings"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/keyring"
	"github.com/riad804/github-auth-manager/internal/utils"
	"github.com/spf13/cobra"
)

var (
	flagContextAddToken    string
	flagContextAddEmail    string
	flagContextAddUsername string
)

var contextAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new GitHub context",
	Long: `Adds a new GitHub context with a unique name.
It will prompt for the Personal Access Token (PAT) if not provided via --token.
Email and username for Git commits can also be provided.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		contextName := strings.TrimSpace(args[0])
		if contextName == "" {
			return fmt.Errorf("context name cannot be empty")
		}

		// Handle Token
		token := strings.TrimSpace(flagContextAddToken)
		var err error
		if token == "" {
			fmt.Printf("Adding context '%s'.\n", contextName)
			token, err = utils.PromptForInput("Enter Personal Access Token (PAT) (will not be echoed): ", true)
			if err != nil {
				return err
			}
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}
		}

		// Handle Email (optional, can prompt or leave empty)
		email := strings.TrimSpace(flagContextAddEmail)
		if email == "" {
			promptEmail, err := utils.PromptForInput(fmt.Sprintf("Enter Git commit email for context '%s' (optional, press Enter to skip): ", contextName), false)
			if err != nil {
				// Non-fatal for optional fields, or make it fatal if you prefer
				fmt.Printf("Warning: could not read email: %v\n", err)
			} else {
				email = promptEmail
			}
		}

		// Handle Username (optional, can prompt or leave empty for default)
		username := strings.TrimSpace(flagContextAddUsername)
		if username == "" {
			promptUsername, err := utils.PromptForInput(fmt.Sprintf("Enter Git commit username for context '%s' (optional, press Enter for default '%s'): ", contextName, config.DefaultUserName), false)
			if err != nil {
				fmt.Printf("Warning: could not read username: %v\n", err)
			} else {
				username = promptUsername
			}
		}

		newCtx := config.Context{
			Name:     contextName,
			Username: username, // Will use default if empty, handled in config.AddContext
			Email:    email,
		}

		if err := config.AddContext(newCtx); err != nil {
			return fmt.Errorf("failed to add context to configuration: %w", err)
		}

		if err := keyring.StoreToken(contextName, token); err != nil {
			// Attempt to roll back adding context from config if token storage fails
			// Best effort, ignore error from RemoveContext here as we're already in an error state.
			_, _ = config.RemoveContext(contextName)
			return fmt.Errorf("failed to store token securely: %w. Context '%s' has not been fully added", err, contextName)
		}

		fmt.Printf("Context '%s' added successfully.\n", contextName)
		if newCtx.Email == "" {
			fmt.Println("Warning: No email specified for this context. Git commits might use global config email.")
		}
		if newCtx.Username == "" || newCtx.Username == config.DefaultUserName {
			fmt.Printf("Using default username '%s' for this context. Git commits might use global config username if this default is not what you intend.\n", config.DefaultUserName)
		}
		return nil
	},
}

func init() {
	contextCmd.AddCommand(contextAddCmd)

	contextAddCmd.Flags().StringVarP(&flagContextAddToken, "token", "t", "", "Personal Access Token (PAT) for the context")
	contextAddCmd.Flags().StringVarP(&flagContextAddEmail, "email", "e", "", "Email for Git commits for this context")
	contextAddCmd.Flags().StringVarP(&flagContextAddUsername, "username", "u", "", fmt.Sprintf("Username for Git commits (defaults to '%s' if not set)", config.DefaultUserName))
}
