package main

import (
	"fmt"
	"log"
	"os"

	"github.com/riad804/github-auth-manager/internal/auth"
	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/models"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "github-auth",
		Usage: "Manage multiple GitHub authentication contexts",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a new GitHub authentication context",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Context name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "token",
						Usage:    "GitHub personal access token",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "username",
						Usage: "GitHub username (optional)",
					},
					&cli.StringFlag{
						Name:  "email",
						Usage: "GitHub email (optional)",
					},
				},
				Action: addContext,
			},
			{
				Name:  "use",
				Usage: "Set active context",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Context name to activate",
						Required: true,
					},
				},
				Action: useContext,
			},
			{
				Name:   "list",
				Usage:  "List all contexts",
				Action: listContexts,
			},
			{
				Name:   "current",
				Usage:  "Show current context",
				Action: currentContext,
			},
			{
				Name:  "remove",
				Usage: "Remove a context",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Context name to remove",
						Required: true,
					},
				},
				Action: removeContext,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func addContext(c *cli.Context) error {
	manager := auth.NewManager(config.NewFileStorage())
	return manager.AddContext(models.Context{
		Name:     c.String("name"),
		Token:    c.String("token"),
		Username: c.String("username"),
		Email:    c.String("email"),
	})
}

func useContext(c *cli.Context) error {
	manager := auth.NewManager(config.NewFileStorage())
	return manager.UseContext(c.String("name"))
}

func listContexts(c *cli.Context) error {
	manager := auth.NewManager(config.NewFileStorage())
	contexts, err := manager.ListContexts()
	if err != nil {
		return err
	}

	current, _ := manager.CurrentContext()

	for _, ctx := range contexts {
		prefix := " "
		if ctx.Name == current {
			prefix = "*"
		}
		fmt.Printf("%s %s (user: %s)\n", prefix, ctx.Name, ctx.Username)
	}
	return nil
}

func currentContext(c *cli.Context) error {
	manager := auth.NewManager(config.NewFileStorage())
	current, err := manager.CurrentContext()
	if err != nil {
		return err
	}

	ctx, err := manager.GetContext(current)
	if err != nil {
		return err
	}

	fmt.Printf("Current context: %s\n", current)
	fmt.Printf("Username: %s\n", ctx.Username)
	fmt.Printf("Email: %s\n", ctx.Email)
	return nil
}

func removeContext(c *cli.Context) error {
	manager := auth.NewManager(config.NewFileStorage())
	return manager.RemoveContext(c.String("name"))
}
