package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/riad804/auth-manager/internals/models"
	"github.com/riad804/auth-manager/internals/storage"
	_ "golang.org/x/crypto/pbkdf2"
)

// mutex for token access
var tokenMutex = &sync.Mutex{}

// PerformOAuth handles the full OAuth flow
func PerformOAuth(provider Provider) (*models.Token, error) {
	// Generate state and verifier
	state := generateRandomString(16)
	verifier := generateRandomString(43) // 128 bits entropy
	challenge := base64URLEncode(sha256.Sum256([]byte(verifier)))

	// Use fixed port for callback
	const port = "1236"
	redirectURI := "http://127.0.0.1:" + port + "/callback"

	// Start local server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server := &http.Server{
		Addr: "127.0.0.1:" + port,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errChan <- fmt.Errorf("state mismatch")
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in response")
			http.Error(w, "no code", http.StatusBadRequest)
			return
		}
		codeChan <- code
		fmt.Fprint(w, "Login successful. You can close this tab.")
		// Shutdown server after sending response
		go func() {
			time.Sleep(500 * time.Millisecond)
			server.Shutdown(context.Background())
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Give server time to start listening
	time.Sleep(100 * time.Millisecond)

	// Build auth URL
	authParams := url.Values{
		"client_id":             {provider.ClientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {strings.Join(provider.Scopes, " ")},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	authURL := provider.AuthURL + "?" + authParams.Encode()

	// Open browser
	err := openBrowser(authURL)
	if err != nil {
		return nil, err
	}

	// Wait for code or error
	select {
	case code := <-codeChan:
		// Close server
		server.Shutdown(context.Background())

		// Exchange code for token
		tokenParams := url.Values{
			"client_id":     {provider.ClientID},
			"client_secret": {provider.ClientSecret}, // Required by GitHub
			"code":          {code},
			"grant_type":    {"authorization_code"},
			"redirect_uri":  {redirectURI},
			"code_verifier": {verifier},
		}
		resp, err := http.PostForm(provider.TokenURL, tokenParams)
		if err != nil {
			return nil, fmt.Errorf("token exchange failed: %w", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read token response failed: %w", err)
		}

		var token models.Token
		err = json.Unmarshal(body, &token)
		if err != nil {
			// For GitHub, it's form encoded
			values, _ := url.ParseQuery(string(body))
			token.AccessToken = values.Get("access_token")
			token.RefreshToken = values.Get("refresh_token")
			token.TokenType = values.Get("token_type")
			expiresIn, _ := strconv.Atoi(values.Get("expires_in"))
			if expiresIn > 0 {
				token.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
			}
		}

		if token.AccessToken == "" {
			return nil, fmt.Errorf("no access token in response: %s", string(body))
		}
		return &token, nil
	case err := <-errChan:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("OAuth timeout")
	}
}

// FetchUserEmail gets user email from API
func FetchUserEmail(provider Provider, accessToken string) (string, error) {
	req, err := http.NewRequest("GET", provider.UserURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf(provider.AuthHeader, accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var user map[string]interface{}
	json.Unmarshal(body, &user)

	// Get email
	email, ok := user["email"].(string)
	if !ok {
		// For GitHub, primary email from array? But for simplicity, use login + "@example.com" if no
		login, _ := user["login"].(string)
		email = login + "@github.com" // Placeholder
	}
	return email, nil
}

// MaybeRefreshToken handles refresh
func MaybeRefreshToken(provider Provider, accountID string, token *models.Token) *models.Token {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	now := time.Now()
	if token.Expiry.IsZero() {
		return token // No expiry
	}

	margin := 5 * time.Minute
	if token.Expiry.After(now.Add(margin)) {
		return token // Good
	}

	if token.RefreshToken == "" {
		return token // Can't refresh
	}

	if token.Expiry.After(now) {
		// Near expiry, async refresh
		go refreshToken(provider, accountID, token.RefreshToken)
		return token
	}

	// Expired, sync refresh
	err := refreshToken(provider, accountID, token.RefreshToken)
	if err != nil {
		// Log? But return old, Git will fail
		return token
	}

	// Reload
	newToken, _ := storage.GetToken(accountID)
	return newToken
}

// refreshToken performs the refresh
func refreshToken(provider Provider, accountID string, refreshToken string) error {
	params := url.Values{
		"client_id":     {provider.ClientID},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	resp, err := http.PostForm(provider.TokenURL, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var newToken models.Token
	err = json.Unmarshal(body, &newToken)
	if err != nil {
		values, _ := url.ParseQuery(string(body))
		newToken.AccessToken = values.Get("access_token")
		newToken.RefreshToken = values.Get("refresh_token")
		newToken.TokenType = values.Get("token_type")
		expiresIn, _ := strconv.Atoi(values.Get("expires_in"))
		if expiresIn > 0 {
			newToken.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
		}
	}

	return storage.StoreToken(accountID, &newToken)
}

// Helper functions
func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)[:length]
}

func base64URLEncode(data [32]byte) string {
	return base64.RawURLEncoding.EncodeToString(data[:])
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return fmt.Errorf("unsupported OS")
	}

	return exec.Command(cmd, args...).Start()
}
