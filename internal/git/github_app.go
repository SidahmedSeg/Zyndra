package git

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// GitHubAppConfig holds GitHub App configuration
type GitHubAppConfig struct {
	AppID            int64
	ClientID         string
	ClientSecret     string
	PrivateKeyBase64 string
	AppName          string
	CallbackURL      string
}

// GitHubAppInstallation represents a GitHub App installation
type GitHubAppInstallation struct {
	InstallationID int64
	AccountLogin   string
	AccountID      int64
	AccountType    string // "User" or "Organization"
	Repositories   []string
}

// GetGitHubAppInstallURL returns the URL for users to install the GitHub App
func GetGitHubAppInstallURL(appName string) string {
	return fmt.Sprintf("https://github.com/apps/%s/installations/new", appName)
}

// GetGitHubAppInstallURLForAccount returns the URL for users to modify installation on specific account
func GetGitHubAppInstallURLForAccount(appName string, installationID int64) string {
	return fmt.Sprintf("https://github.com/apps/%s/installations/%d", appName, installationID)
}

// ParsePrivateKey parses the base64-encoded private key
func ParsePrivateKey(privateKeyBase64 string) (*rsa.PrivateKey, error) {
	// Decode base64
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Parse PEM block
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	// Parse RSA private key
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
	}

	return key, nil
}

// GenerateJWT generates a JWT for GitHub App authentication
func GenerateJWT(appID int64, privateKey *rsa.PrivateKey) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": appID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GetInstallationToken gets an access token for a specific installation
func GetInstallationToken(ctx context.Context, cfg *GitHubAppConfig, installationID int64) (string, time.Time, error) {
	// Parse private key
	privateKey, err := ParsePrivateKey(cfg.PrivateKeyBase64)
	if err != nil {
		return "", time.Time{}, err
	}

	// Generate JWT
	jwtToken, err := GenerateJWT(cfg.AppID, privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Create authenticated client with JWT
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwtToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get installation token
	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create installation token: %w", err)
	}

	return token.GetToken(), token.GetExpiresAt().Time, nil
}

// ListInstallations lists all installations for the GitHub App
func ListInstallations(ctx context.Context, cfg *GitHubAppConfig) ([]*github.Installation, error) {
	// Parse private key
	privateKey, err := ParsePrivateKey(cfg.PrivateKeyBase64)
	if err != nil {
		return nil, err
	}

	// Generate JWT
	jwtToken, err := GenerateJWT(cfg.AppID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Create authenticated client with JWT
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwtToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// List installations
	installations, _, err := client.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}

	return installations, nil
}

// GetInstallationRepositories lists repositories for an installation
func GetInstallationRepositories(ctx context.Context, cfg *GitHubAppConfig, installationID int64) ([]*github.Repository, error) {
	// Get installation token
	token, _, err := GetInstallationToken(ctx, cfg, installationID)
	if err != nil {
		return nil, err
	}

	// Create client with installation token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// List repositories
	repos, _, err := client.Apps.ListRepos(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repos.Repositories, nil
}

// NewGitHubAppClient creates a new GitHub client authenticated with an installation token
func NewGitHubAppClient(ctx context.Context, cfg *GitHubAppConfig, installationID int64) (*github.Client, error) {
	// Get installation token
	token, _, err := GetInstallationToken(ctx, cfg, installationID)
	if err != nil {
		return nil, err
	}

	// Create client with installation token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc), nil
}

// GetGitHubAppOAuthURL generates the OAuth URL for GitHub App user authorization
// This is used when you need both app installation AND user authorization
func GetGitHubAppOAuthURL(cfg *GitHubAppConfig, state string) string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s",
		cfg.ClientID,
		cfg.CallbackURL,
		state,
	)
}

// ExchangeGitHubAppCode exchanges the OAuth code for an access token
func ExchangeGitHubAppCode(ctx context.Context, cfg *GitHubAppConfig, code string) (*oauth2.Token, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// GetUserInstallations gets all GitHub App installations for a specific user
func GetUserInstallations(ctx context.Context, userAccessToken string) ([]*github.Installation, error) {
	// Create client with user access token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: userAccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// List user's installations
	installations, _, err := client.Apps.ListUserInstallations(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list user installations: %w", err)
	}

	return installations, nil
}

// ParseInstallationCallback parses the callback from GitHub App installation
func ParseInstallationCallback(r *http.Request) (installationID int64, setupAction string, err error) {
	installationIDStr := r.URL.Query().Get("installation_id")
	if installationIDStr == "" {
		return 0, "", fmt.Errorf("missing installation_id")
	}

	var id int64
	if _, err := fmt.Sscanf(installationIDStr, "%d", &id); err != nil {
		return 0, "", fmt.Errorf("invalid installation_id: %w", err)
	}

	setupAction = r.URL.Query().Get("setup_action")
	return id, setupAction, nil
}

