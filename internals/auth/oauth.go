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
)

// mutex for token access
var tokenMutex = &sync.Mutex{}

// ==========================
// PKCE helpers (RFC 7636)
// ==========================

func generateCodeVerifier() (string, error) {
	const length = 64 // MUST be 43–128

	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789-._~"

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ==========================
// OAuth flow
// ==========================

func PerformOAuth(provider Provider) (*models.Token, error) {
	state, err := generateState()
	if err != nil {
		return nil, err
	}

	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}

	challenge := generateCodeChallenge(verifier)

	// Get random free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errChan <- fmt.Errorf("state mismatch")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errChan <- fmt.Errorf("no code in response")
			return
		}

		fmt.Fprint(w, "Login successful. You can close this tab.")
		codeChan <- code

		go func() {
			time.Sleep(300 * time.Millisecond)
			server.Shutdown(context.Background())
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

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

	if err := openBrowser(authURL); err != nil {
		return nil, err
	}

	select {
	case code := <-codeChan:
		tokenParams := url.Values{
			"client_id":     {provider.ClientID},
			"code":          {code},
			"grant_type":    {"authorization_code"},
			"redirect_uri":  {redirectURI},
			"code_verifier": {verifier},
		}

		// GitHub still allows client_secret for classic OAuth apps
		if provider.ClientSecret != "" {
			tokenParams.Set("client_secret", provider.ClientSecret)
		}

		req, _ := http.NewRequest("POST", provider.TokenURL, strings.NewReader(tokenParams.Encode()))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var token models.Token
		if err := json.Unmarshal(body, &token); err != nil {
			values, _ := url.ParseQuery(string(body))
			token.AccessToken = values.Get("access_token")
			token.RefreshToken = values.Get("refresh_token")
			token.TokenType = values.Get("token_type")
			if exp, _ := strconv.Atoi(values.Get("expires_in")); exp > 0 {
				token.Expiry = time.Now().Add(time.Duration(exp) * time.Second)
			}
		}

		if token.AccessToken == "" {
			return nil, fmt.Errorf("no access token in response: %s", string(body))
		}

		return &token, nil

	case err := <-errChan:
		return nil, err

	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("OAuth timeout")
	}
}

// ==========================
// Token refresh (unchanged)
// ==========================

func MaybeRefreshToken(provider Provider, accountID string, token *models.Token) *models.Token {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if token.Expiry.IsZero() || token.Expiry.After(time.Now().Add(5*time.Minute)) {
		return token
	}

	if token.RefreshToken == "" {
		return token
	}

	go refreshToken(provider, accountID, token.RefreshToken)
	return token
}

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

	body, _ := io.ReadAll(resp.Body)

	var newToken models.Token
	if err := json.Unmarshal(body, &newToken); err != nil {
		values, _ := url.ParseQuery(string(body))
		newToken.AccessToken = values.Get("access_token")
		newToken.RefreshToken = values.Get("refresh_token")
		newToken.TokenType = values.Get("token_type")
		if exp, _ := strconv.Atoi(values.Get("expires_in")); exp > 0 {
			newToken.Expiry = time.Now().Add(time.Duration(exp) * time.Second)
		}
	}

	return storage.StoreToken(accountID, &newToken)
}

// ==========================
// Browser helper
// ==========================

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
