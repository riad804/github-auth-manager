package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	AppName         = "gham"
	ConfigFileName  = "config.yaml"
	KeyringService  = "GHAM_PAT_Storage_v1" // Consider versioning if format changes
	DefaultUserName = "GHAM User"
)

type Context struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username,omitempty"` // omitempty to not write if default
	Email    string `yaml:"email,omitempty"`
}

type RepoConfig struct {
	Path        string `yaml:"path"`
	ContextName string `yaml:"contextName"`
}

type AppConfig struct {
	Contexts     []Context    `yaml:"contexts"`
	Repositories []RepoConfig `yaml:"repositories"`
}

var GlobalConfig AppConfig
var configFilePath string // Store the actual path for error reporting

func GetConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(userConfigDir, AppName), nil
}

// GetConfigFilePathForError returns the path for error messages if InitConfig fails early.
func GetConfigFilePathForError() string {
	if configFilePath != "" {
		return configFilePath
	}
	// Attempt to construct it if not yet set (e.g., MkdirAll failed)
	dir, err := GetConfigDir()
	if err != nil {
		return "unknown (could not determine user config directory)"
	}
	return filepath.Join(dir, ConfigFileName)
}

func InitConfig() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0750); err != nil { // 0750: rwx for user, rx for group
		return fmt.Errorf("failed to create config directory '%s': %w", configDir, err)
	}

	configFilePath = filepath.Join(configDir, ConfigFileName)

	viper.SetConfigName(strings.TrimSuffix(ConfigFileName, filepath.Ext(ConfigFileName)))
	viper.SetConfigType(filepath.Ext(ConfigFileName)[1:])
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create it with empty structure
			fmt.Printf("Config file not found at %s. Creating a new one.\n", configFilePath)
			GlobalConfig = AppConfig{
				Contexts:     []Context{},
				Repositories: []RepoConfig{},
			}
			return SaveConfig() // Ensure new file is written
		}
		return fmt.Errorf("failed to read config file '%s': %w", configFilePath, err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config from '%s': %w", configFilePath, err)
	}
	return nil
}

func SaveConfig() error {
	if configFilePath == "" {
		// This case should ideally be prevented by InitConfig always setting it
		return fmt.Errorf("config file path not initialized. Call InitConfig first")
	}

	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to ensure config directory '%s' exists for saving: %w", configDir, err)
	}

	yamlData, err := yaml.Marshal(&GlobalConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write with 0600 permissions (rw for user only) as it might contain sensitive paths
	if err = os.WriteFile(configFilePath, yamlData, 0600); err != nil {
		return fmt.Errorf("failed to write config file '%s': %w", configFilePath, err)
	}
	// fmt.Printf("DEBUG: Configuration saved to %s\n", configFilePath)
	return nil
}

func FindContext(name string) (*Context, bool) {
	for i, ctx := range GlobalConfig.Contexts {
		if ctx.Name == name {
			return &GlobalConfig.Contexts[i], true
		}
	}
	return nil, false
}

func AddContext(newCtx Context) error {
	if _, found := FindContext(newCtx.Name); found {
		return fmt.Errorf("context with name '%s' already exists", newCtx.Name)
	}
	if newCtx.Username == "" { // Apply default if username is empty
		newCtx.Username = DefaultUserName
	}
	GlobalConfig.Contexts = append(GlobalConfig.Contexts, newCtx)
	return SaveConfig()
}

func RemoveContext(name string) (bool, error) {
	found := false
	var updatedContexts []Context
	for _, ctx := range GlobalConfig.Contexts {
		if ctx.Name == name {
			found = true
		} else {
			updatedContexts = append(updatedContexts, ctx)
		}
	}

	if !found {
		return false, nil // Not found, no error, but signal not found
	}
	GlobalConfig.Contexts = updatedContexts

	// Also remove repository assignments using this context
	var updatedRepos []RepoConfig
	repoAssignmentRemoved := false
	for _, repoCfg := range GlobalConfig.Repositories {
		if repoCfg.ContextName != name {
			updatedRepos = append(updatedRepos, repoCfg)
		} else {
			repoAssignmentRemoved = true
		}
	}
	if repoAssignmentRemoved {
		GlobalConfig.Repositories = updatedRepos
	}

	return true, SaveConfig()
}

func AssignRepoContext(repoPath, contextName string) error {
	// Path should already be absolute and validated as repo root by caller
	// Remove existing assignment for this path, if any
	var updatedRepos []RepoConfig
	for _, rc := range GlobalConfig.Repositories {
		if rc.Path == repoPath {
			if rc.ContextName == contextName { // Already assigned to the same context
				return nil // No change needed
			}
			// If different context, it will be overwritten by adding the new one below
			// and not adding this old one to updatedRepos.
		} else {
			updatedRepos = append(updatedRepos, rc)
		}
	}
	// If it was found and different, it's excluded. If not found, this adds it.
	// If found and same, we returned nil above.
	// This logic simplifies to just re-adding/replacing:
	var newRepoList []RepoConfig
	for _, rc := range GlobalConfig.Repositories {
		if rc.Path != repoPath {
			newRepoList = append(newRepoList, rc)
		}
	}
	newRepoList = append(newRepoList, RepoConfig{Path: repoPath, ContextName: contextName})
	GlobalConfig.Repositories = newRepoList

	return SaveConfig()
}

func GetRepoContextName(repoPath string) (string, bool) {
	// Path should already be absolute and repo root
	for _, rc := range GlobalConfig.Repositories {
		if rc.Path == repoPath {
			return rc.ContextName, true
		}
	}
	return "", false
}
