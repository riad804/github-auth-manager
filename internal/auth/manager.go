package auth

import (
	"errors"
	"fmt"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/models"
)

type Manager struct {
	storage config.Storage
}

func NewManager(storage config.Storage) *Manager {
	return &Manager{storage: storage}
}

func (m *Manager) AddContext(ctx models.Context) error {
	config, err := m.storage.Load()
	if err != nil {
		return err
	}

	if _, exists := config.Contexts[ctx.Name]; exists {
		return fmt.Errorf("context %s already exists", ctx.Name)
	}

	config.Contexts[ctx.Name] = ctx
	return m.storage.Save(config)
}

func (m *Manager) UseContext(name string) error {
	config, err := m.storage.Load()
	if err != nil {
		return err
	}

	if _, exists := config.Contexts[name]; !exists {
		return fmt.Errorf("context %s not found", name)
	}

	config.CurrentContext = name
	return m.storage.Save(config)
}

func (m *Manager) ListContexts() ([]models.Context, error) {
	config, err := m.storage.Load()
	if err != nil {
		return nil, err
	}

	var contexts []models.Context
	for _, ctx := range config.Contexts {
		contexts = append(contexts, ctx)
	}
	return contexts, nil
}

func (m *Manager) CurrentContext() (string, error) {
	config, err := m.storage.Load()
	if err != nil {
		return "", err
	}

	if config.CurrentContext == "" {
		return "", errors.New("no current context set")
	}

	return config.CurrentContext, nil
}

func (m *Manager) GetContext(name string) (models.Context, error) {
	config, err := m.storage.Load()
	if err != nil {
		return models.Context{}, err
	}

	ctx, exists := config.Contexts[name]
	if !exists {
		return models.Context{}, fmt.Errorf("context %s not found", name)
	}

	return ctx, nil
}

func (m *Manager) RemoveContext(name string) error {
	config, err := m.storage.Load()
	if err != nil {
		return err
	}

	if _, exists := config.Contexts[name]; !exists {
		return fmt.Errorf("context %s not found", name)
	}

	delete(config.Contexts, name)

	// If we're removing the current context, unset it
	if config.CurrentContext == name {
		config.CurrentContext = ""
	}

	return m.storage.Save(config)
}
