package auth

import "fmt"

// Provider defines OAuth endpoints and configs
type Provider struct {
	Name       string
	AuthURL    string
	TokenURL   string
	UserURL    string
	Scopes     []string
	ClientID   string // Placeholder: replace with your registered app's client_id
	AuthHeader string // For user info: "token %s" or "Bearer %s"
}

var providers = map[string]Provider{
	"github": {
		Name:       "github",
		AuthURL:    "https://github.com/login/oauth/authorize",
		TokenURL:   "https://github.com/login/oauth/access_token",
		UserURL:    "https://api.github.com/user",
		Scopes:     []string{"repo", "user:email"},
		ClientID:   "Ov23liPCCpJ195rICwIC",
		AuthHeader: "token %s",
	},
	"gitlab": {
		Name:       "gitlab",
		AuthURL:    "https://gitlab.com/oauth/authorize",
		TokenURL:   "https://gitlab.com/oauth/token",
		UserURL:    "https://gitlab.com/api/v4/user",
		Scopes:     []string{"api", "read_user"},
		ClientID:   "replace_with_your_gitlab_client_id",
		AuthHeader: "Bearer %s",
	},
	"bitbucket": {
		Name:       "bitbucket",
		AuthURL:    "https://bitbucket.org/site/oauth2/authorize",
		TokenURL:   "https://bitbucket.org/site/oauth2/access_token",
		UserURL:    "https://api.bitbucket.org/2.0/user",
		Scopes:     []string{"account", "repository"},
		ClientID:   "replace_with_your_bitbucket_client_id",
		AuthHeader: "Bearer %s",
	},
}

// GetProvider returns the provider config
func GetProvider(name string) (Provider, error) {
	p, ok := providers[name]
	if !ok {
		return Provider{}, fmt.Errorf("unsupported provider: %s", name)
	}
	return p, nil
}
