package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/keyring"
	"github.com/spf13/cobra"
)

var contextListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all configured GitHub contexts",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(config.GlobalConfig.Contexts) == 0 {
			fmt.Println("No contexts configured yet. Use 'gham context add <name>' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0) // minwidth, tabwidth, padding, padchar, flags
		fmt.Fprintln(w, "NAME\tUSERNAME\tEMAIL\tTOKEN STORED?")
		fmt.Fprintln(w, "----\t--------\t-----\t-------------")

		for _, ctx := range config.GlobalConfig.Contexts {
			_, err := keyring.GetToken(ctx.Name)
			tokenStored := "Yes"
			if err != nil {
				tokenStored = "No / Error" // More informative if keyring access fails
			}
			username := ctx.Username
			if username == "" {
				username = "(default)"
			}
			email := ctx.Email
			if email == "" {
				email = "(not set)"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ctx.Name, username, email, tokenStored)
		}
		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}
		return nil
	},
}

func init() {
	contextCmd.AddCommand(contextListCmd)
}
