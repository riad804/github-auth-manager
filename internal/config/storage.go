package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/riad804/github-auth-manager/internal/models"
)

type Storage interface {
	Load() (*Config, error)
	Save(*Config) error
}

type FileStorage struct {
	path string
}

func NewFileStorage() *FileStorage {
	configDir, _ := os.UserConfigDir()
	return &FileStorage{
		path: filepath.Join(configDir, "github-auth-manager", "config.json"),
	}
}

func (fs *FileStorage) Load() (*Config, error) {
	if _, err := os.Stat(fs.path); os.IsNotExist(err) {
		return &Config{
			Contexts: make(map[string]models.Context),
		}, nil
	}

	data, err := os.ReadFile(fs.path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Contexts == nil {
		config.Contexts = make(map[string]models.Context)
	}

	return &config, nil
}

func (fs *FileStorage) Save(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fs.path), 0755); err != nil {
		return err
	}

	return os.WriteFile(fs.path, data, 0600)
}
