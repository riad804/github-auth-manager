package storage

import (
	"encoding/json"

	"github.com/riad804/auth-manager/internals/models"
	"github.com/zalando/go-keyring"
)

const serviceName = "git-auth-manager"

// InitKeyring dummy, go-keyring init on use
func InitKeyring() {}

// StoreToken saves token to keyring
func StoreToken(id string, token *models.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return keyring.Set(serviceName, id, string(data))
}

// GetToken retrieves token
func GetToken(id string) (*models.Token, error) {
	data, err := keyring.Get(serviceName, id)
	if err != nil {
		return nil, err
	}
	var token models.Token
	err = json.Unmarshal([]byte(data), &token)
	return &token, err
}

// DeleteToken removes token
func DeleteToken(id string) error {
	return keyring.Delete(serviceName, id)
}
