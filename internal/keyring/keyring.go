package keyring

import (
	"errors"
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/riad804/github-auth-manager/internal/config"
)

var kr keyring.Keyring
var keyringErr error // Store init error

func init() {
	// Define allowed backends for more predictable behavior
	// Order matters: first successful one is used.
	allowedBackends := []keyring.BackendType{
		keyring.KeychainBackend,      // macOS
		keyring.SecretServiceBackend, // Linux
		keyring.WinCredBackend,       // Windows
		keyring.KWalletBackend,       // Linux (KDE)
		keyring.PassBackend,          // Linux/macOS (pass utility)
		// keyring.FileBackend,          // Fallback (less secure, consider if needed)
	}

	kr, keyringErr = keyring.Open(keyring.Config{
		ServiceName:              config.KeyringService,
		AllowedBackends:          allowedBackends,
		KeychainTrustApplication: true, // macOS: trust the application path
		// For Linux Secret Service:
		LibSecretCollectionName: "gham",
		// For pass:
		PassCmd: "pass",
		// PassDir: "~/.password-store", // if non-standard
	})

	if keyringErr != nil {
		// This error will be checked by functions using the keyring
		fmt.Fprintf(os.Stderr, "Warning: GHAM could not initialize system keyring: %v\n", keyringErr)
		fmt.Fprintln(os.Stderr, "PATs will not be stored securely. Commands requiring tokens may fail or be insecure.")
		fmt.Fprintln(os.Stderr, "Ensure you have a compatible keyring service (libsecret, GNOME Keyring, KWallet, macOS Keychain, Windows Credential Manager).")
		// For MVP, we allow the app to run, but operations needing tokens will fail more gracefully.
	}
}

func checkKeyring() error {
	if keyringErr != nil {
		return fmt.Errorf("keyring is not available: %w", keyringErr)
	}
	if kr == nil { // Should be covered by keyringErr, but as a safeguard
		return errors.New("keyring not initialized (keyring instance is nil)")
	}
	return nil
}

func StoreToken(contextName, token string) error {
	if err := checkKeyring(); err != nil {
		return err
	}
	// Key for keyring item should be unique per token. Context name is good.
	itemKey := contextName
	err := kr.Set(keyring.Item{
		Key:         itemKey,
		Data:        []byte(token),
		Label:       fmt.Sprintf("GHAM PAT for context '%s'", contextName),
		Description: "GitHub Personal Access Token managed by GHAM CLI.",
	})
	if err != nil {
		return fmt.Errorf("failed to store token for context '%s' in keyring: %w", contextName, err)
	}
	return nil
}

func GetToken(contextName string) (string, error) {
	if err := checkKeyring(); err != nil {
		return "", err
	}
	itemKey := contextName
	item, err := kr.Get(itemKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", fmt.Errorf("no token found for context '%s' in keyring. Please ensure it was added correctly", contextName)
		}
		return "", fmt.Errorf("failed to get token for context '%s' from keyring: %w", contextName, err)
	}
	return string(item.Data), nil
}

func DeleteToken(contextName string) error {
	if err := checkKeyring(); err != nil {
		// If keyring isn't available, we can't delete from it, but the config removal might still be desired.
		// Return the error so caller can decide how to handle.
		return err
	}
	itemKey := contextName
	err := kr.Remove(itemKey)
	// Do not error if key is not found, as it might have been deleted manually or never existed.
	if err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return fmt.Errorf("failed to delete token for context '%s' from keyring: %w", contextName, err)
	}
	return nil
}
