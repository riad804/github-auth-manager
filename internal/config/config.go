package config

import "github.com/riad804/github-auth-manager/internal/models"

type Config struct {
	CurrentContext string                    `json:"currentContext"`
	Contexts       map[string]models.Context `json:"contexts"`
}
