package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	oauthClientID     = "claude-cli"
	oauthTokenURL     = "https://auth.anthropic.com/oauth/token"
	oauthAuthorizeURL = "https://auth.anthropic.com/oauth/authorize"
	oauthCallbackURL  = "http://localhost:9876/callback"
)

// OAuthToken holds OAuth token data
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
}

// IsExpired returns true if the token has expired
func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.Expiry)
}

// OAuthManager handles OAuth authentication flow
type OAuthManager struct {
	tokenPath string
}

// NewOAuthManager creates a new OAuth manager
func NewOAuthManager() *OAuthManager {
	return &OAuthManager{
		tokenPath: filepath.Join(homeDir(), ".claude", "oauth_token.json"),
	}
}

// StartLoginFlow initiates the OAuth login flow
func (m *OAuthManager) StartLoginFlow(ctx context.Context) (string, error) {
	state := generateState()
	authURL := fmt.Sprintf(
		"%s?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=read",
		oauthAuthorizeURL,
		url.QueryEscape(oauthClientID),
		url.QueryEscape(oauthCallbackURL),
		state,
	)

	if err := openBrowser(authURL); err != nil {
		return authURL, fmt.Errorf("failed to open browser: %w", err)
	}

	return authURL, nil
}

// ExchangeCode exchanges an authorization code for tokens
func (m *OAuthManager) ExchangeCode(ctx context.Context, code string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", oauthCallbackURL)
	data.Set("client_id", oauthClientID)

	req, err := http.NewRequestWithContext(ctx, "POST", oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	if token.Expiry.IsZero() {
		token.Expiry = time.Now().Add(1 * time.Hour)
	}

	if err := m.saveToken(&token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return &token, nil
}

// RefreshToken refreshes an expired OAuth token
func (m *OAuthManager) RefreshToken(ctx context.Context) (*OAuthToken, error) {
	token, err := m.loadToken()
	if err != nil {
		return nil, err
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token.RefreshToken)
	data.Set("client_id", oauthClientID)

	req, err := http.NewRequestWithContext(ctx, "POST", oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var newToken OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&newToken); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	if newToken.Expiry.IsZero() {
		newToken.Expiry = time.Now().Add(1 * time.Hour)
	}

	if err := m.saveToken(&newToken); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return &newToken, nil
}

// GetValidToken returns a valid token, refreshing if necessary
func (m *OAuthManager) GetValidToken(ctx context.Context) (*OAuthToken, error) {
	token, err := m.loadToken()
	if err != nil {
		return nil, err
	}

	if !token.IsExpired() {
		return token, nil
	}

	return m.RefreshToken(ctx)
}

// ClearToken removes the stored OAuth token
func (m *OAuthManager) ClearToken() error {
	return os.Remove(m.tokenPath)
}

func (m *OAuthManager) saveToken(token *OAuthToken) error {
	tokenDir := filepath.Dir(m.tokenPath)
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	return os.WriteFile(m.tokenPath, data, 0o600)
}

func (m *OAuthManager) loadToken() (*OAuthToken, error) {
	data, err := os.ReadFile(m.tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no OAuth token found")
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token OAuthToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func generateState() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "fallback-state"
	}
	return base64.URLEncoding.EncodeToString(b)
}
