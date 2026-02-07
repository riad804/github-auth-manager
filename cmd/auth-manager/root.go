package authmanager

import (
	"os"
	"path/filepath"

	"github.com/riad804/auth-manager/internals/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gham",
	Short: "Manage multiple Git accounts per repository",
	Long:  `A CLI tool to handle repository-scoped Git authentication for multiple accounts across providers.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in ~/.config/git-auth-manager
	configDir := filepath.Join(home, ".config", "git-auth-manager")
	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Create config dir if not exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0700) // Secure permissions
		cobra.CheckErr(err)
	}

	// Read config, create if not exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Create empty config
			err = viper.SafeWriteConfig()
			cobra.CheckErr(err)
		} else {
			cobra.CheckErr(err)
		}
	}

	// Initialize accounts map if not present
	if !viper.IsSet("accounts") {
		viper.Set("accounts", map[string]interface{}{})
		err = viper.WriteConfig()
		cobra.CheckErr(err)
	}

	// Initialize keyring service
	storage.InitKeyring()
}
