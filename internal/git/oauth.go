package git

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	gitlaboauth "golang.org/x/oauth2/gitlab"
)

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	GitLabClientID     string
	GitLabClientSecret string
	GitLabRedirectURL  string
	GitLabBaseURL      string // Optional, for self-hosted GitLab
}

// OAuthState stores OAuth state for CSRF protection
type OAuthState struct {
	Provider   string
	OrgID      string
	UserID     string
	ExpiresAt  time.Time
	StateToken string
}

// GenerateOAuthState generates a secure OAuth state token
func GenerateOAuthState(provider, orgID, userID string) (*OAuthState, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	stateToken := base64.URLEncoding.EncodeToString(bytes)

	return &OAuthState{
		Provider:   provider,
		OrgID:      orgID,
		UserID:     userID,
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		StateToken: stateToken,
	}, nil
}

// GetGitHubOAuthURL generates the GitHub OAuth authorization URL
func GetGitHubOAuthURL(cfg *OAuthConfig, state string) string {
	// Ensure redirect URL doesn't have trailing slash (GitHub is strict about exact match)
	redirectURL := cfg.GitHubRedirectURL
	if len(redirectURL) > 0 && redirectURL[len(redirectURL)-1] == '/' {
		redirectURL = redirectURL[:len(redirectURL)-1]
	}
	
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  redirectURL,
		// Use public_repo scope for single repository access (more limited than full repo access)
		// Note: GitHub OAuth Apps don't support true single-repo scopes, but public_repo is more limited
		Scopes:       []string{"public_repo", "read:user"},
		Endpoint:     github.Endpoint,
	}

	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetGitLabOAuthURL generates the GitLab OAuth authorization URL
func GetGitLabOAuthURL(cfg *OAuthConfig, state string) string {
	endpoint := gitlaboauth.Endpoint
	if cfg.GitLabBaseURL != "" {
		// Custom GitLab instance
		endpoint = oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", cfg.GitLabBaseURL),
			TokenURL: fmt.Sprintf("%s/oauth/token", cfg.GitLabBaseURL),
		}
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GitLabClientID,
		ClientSecret: cfg.GitLabClientSecret,
		RedirectURL:  cfg.GitLabRedirectURL,
		Scopes:       []string{"api", "read_user", "read_repository"},
		Endpoint:     endpoint,
	}

	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeGitHubCode exchanges authorization code for access token
func ExchangeGitHubCode(ctx context.Context, cfg *OAuthConfig, code string) (*oauth2.Token, error) {
	// Ensure redirect URL doesn't have trailing slash
	redirectURL := cfg.GitHubRedirectURL
	if len(redirectURL) > 0 && redirectURL[len(redirectURL)-1] == '/' {
		redirectURL = redirectURL[:len(redirectURL)-1]
	}
	
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  redirectURL,
		// Note: GitHub OAuth Apps don't support single-repository scopes
		// To restrict to a single repo, you need to use GitHub Apps (not OAuth Apps)
		// For now, using minimal scopes - user will need to select repo after auth
		// public_repo: access to public repositories only (more limited)
		Scopes:       []string{"public_repo", "read:user"},
		Endpoint:     github.Endpoint,
	}

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// ExchangeGitLabCode exchanges authorization code for access token
func ExchangeGitLabCode(ctx context.Context, cfg *OAuthConfig, code string) (*oauth2.Token, error) {
	endpoint := gitlaboauth.Endpoint
	if cfg.GitLabBaseURL != "" {
		endpoint = oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", cfg.GitLabBaseURL),
			TokenURL: fmt.Sprintf("%s/oauth/token", cfg.GitLabBaseURL),
		}
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GitLabClientID,
		ClientSecret: cfg.GitLabClientSecret,
		RedirectURL:  cfg.GitLabRedirectURL,
		Scopes:       []string{"api", "read_user", "read_repository"},
		Endpoint:     endpoint,
	}

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// GetGitHubUser gets the authenticated GitHub user
func GetGitHubUser(ctx context.Context, token string) (*GitUser, error) {
	client := NewGitHubClient(token)
	user, _, err := client.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub user: %w", err)
	}

	return &GitUser{
		ID:       int64(user.GetID()),
		Login:    user.GetLogin(),
		Name:     user.GetName(),
		Email:    user.GetEmail(),
		AvatarURL: user.GetAvatarURL(),
	}, nil
}

// GetGitLabUser gets the authenticated GitLab user
func GetGitLabUser(ctx context.Context, token, baseURL string) (*GitUser, error) {
	client := NewGitLabClient(token, baseURL)
	user, _, err := client.client.Users.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitLab user: %w", err)
	}

	return &GitUser{
		ID:        int64(user.ID),
		Login:     user.Username,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}, nil
}

// GitUser represents a Git user (GitHub or GitLab)
type GitUser struct {
	ID        int64
	Login     string
	Name      string
	Email     string
	AvatarURL string
}

// ParseOAuthCallback parses OAuth callback parameters
func ParseOAuthCallback(r *http.Request) (code, state string, err error) {
	code = r.URL.Query().Get("code")
	state = r.URL.Query().Get("state")

	if code == "" {
		return "", "", fmt.Errorf("missing code parameter")
	}
	if state == "" {
		return "", "", fmt.Errorf("missing state parameter")
	}

	return code, state, nil
}

// BuildRedirectURL builds the OAuth redirect URL
func BuildRedirectURL(baseURL, path string) string {
	u, _ := url.Parse(baseURL)
	u.Path = path
	return u.String()
}

