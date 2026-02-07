package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/riad804/auth-manager/internals/models"
	"github.com/spf13/viper"
)

// StoreAccount saves account to Viper config
func StoreAccount(account models.Account) error {
	accounts := viper.GetStringMap("accounts")
	accounts[account.ID] = account
	viper.Set("accounts", accounts)
	return viper.WriteConfig()
}

// GetAccount retrieves account by ID
func GetAccount(id string) (models.Account, error) {
	accounts := viper.GetStringMap("accounts")
	accInt, ok := accounts[id]
	if !ok {
		return models.Account{}, fmt.Errorf("account not found: %s", id)
	}
	accMap := accInt.(map[string]interface{})
	return models.Account{
		ID:       accMap["id"].(string),
		Provider: accMap["provider"].(string),
		Email:    accMap["email"].(string),
	}, nil
}

// GetAllAccounts lists all
func GetAllAccounts() ([]models.Account, error) {
	accounts := viper.GetStringMap("accounts")
	result := []models.Account{}
	for _, accInt := range accounts {
		accMap := accInt.(map[string]interface{})
		result = append(result, models.Account{
			ID:       accMap["id"].(string),
			Provider: accMap["provider"].(string),
			Email:    accMap["email"].(string),
		})
	}
	return result, nil
}

// GetAccountsByProvider filters by provider
func GetAccountsByProvider(provider string) ([]models.Account, error) {
	all, err := GetAllAccounts()
	if err != nil {
		return nil, err
	}
	result := []models.Account{}
	for _, acc := range all {
		if acc.Provider == provider {
			result = append(result, acc)
		}
	}
	return result, nil
}

// DeleteAccount removes account
func DeleteAccount(id string) error {
	accounts := viper.GetStringMap("accounts")
	delete(accounts, id)
	viper.Set("accounts", accounts)
	return viper.WriteConfig()
}

// StoreRepoMapping saves to .git/git-auth-manager.json
func StoreRepoMapping(repoDir, accountID string) error {
	mapping := map[string]string{"account_id": accountID}
	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	filePath := filepath.Join(repoDir, ".git", "git-auth-manager.json")
	return os.WriteFile(filePath, data, 0600)
}

// GetRepoMapping reads
func GetRepoMapping(repoDir string) (string, error) {
	filePath := filepath.Join(repoDir, ".git", "git-auth-manager.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	var mapping map[string]string
	json.Unmarshal(data, &mapping)
	return mapping["account_id"], nil
}

// DeleteRepoMapping removes
func DeleteRepoMapping(repoDir string) error {
	filePath := filepath.Join(repoDir, ".git", "git-auth-manager.json")
	return os.Remove(filePath)
}
