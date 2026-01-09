package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/domain"
	"github.com/intelifox/click-deploy/internal/git"
	"github.com/intelifox/click-deploy/internal/store"
)

type GitHandler struct {
	store  *store.DB
	config *config.Config
}

func NewGitHandler(store *store.DB, cfg *config.Config) *GitHandler {
	return &GitHandler{
		store:  store,
		config: cfg,
	}
}

// RegisterGitRoutes registers all git-related routes
func RegisterGitRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewGitHandler(db, cfg)

	// OAuth initiation (returns JSON with URL for frontend redirect)
	r.Get("/git/connect/github/url", h.GetGitHubOAuthURL)
	r.Get("/git/connect/gitlab/url", h.GetGitLabOAuthURL)
	
	// OAuth initiation (direct redirect - kept for backward compatibility)
	r.Get("/git/connect/github", h.ConnectGitHub)
	r.Get("/git/connect/gitlab", h.ConnectGitLab)

	// OAuth callbacks are registered separately in main.go at root level

	// Git connections management
	r.Get("/git/connections", h.ListConnections)
	r.Delete("/git/connections/{id}", h.DeleteConnection)

	// Repository operations
	r.Get("/git/repos", h.ListRepositories)
	r.Get("/git/repos/{owner}/{repo}/branches", h.ListBranches)
	r.Get("/git/repos/{owner}/{repo}/tree", h.GetRepositoryTree)
}

// GetGitHubOAuthURL returns the GitHub OAuth URL as JSON (for frontend to redirect)
func (h *GitHandler) GetGitHubOAuthURL(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	userID := auth.GetUserID(r.Context())

	if orgID == "" || userID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID or User ID not found in token"))
		return
	}

	// Check if GitHub OAuth is configured
	if h.config.GitHubClientID == "" || h.config.GitHubClientSecret == "" {
		WriteError(w, domain.NewInvalidInputError("GitHub OAuth is not configured. Please set GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET environment variables."))
		return
	}

	state, err := git.GenerateOAuthState("github", orgID, userID)
	if err != nil {
		WriteError(w, domain.ErrInternal.WithError(err))
		return
	}

	// TODO: Store state in cache/DB for validation

	oauthConfig := &git.OAuthConfig{
		GitHubClientID:     h.config.GitHubClientID,
		GitHubClientSecret: h.config.GitHubClientSecret,
		GitHubRedirectURL:  h.config.GitHubRedirectURL,
	}

	authURL := git.GetGitHubOAuthURL(oauthConfig, state.StateToken)
	
	WriteJSON(w, http.StatusOK, map[string]string{
		"auth_url": authURL,
	})
}

// ConnectGitHub initiates GitHub OAuth flow (redirects immediately)
func (h *GitHandler) ConnectGitHub(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	userID := auth.GetUserID(r.Context())

	if orgID == "" || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	state, err := git.GenerateOAuthState("github", orgID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Store state in cache/DB for validation

	oauthConfig := &git.OAuthConfig{
		GitHubClientID:     h.config.GitHubClientID,
		GitHubClientSecret: h.config.GitHubClientSecret,
		GitHubRedirectURL:  h.config.GitHubRedirectURL,
	}

	authURL := git.GetGitHubOAuthURL(oauthConfig, state.StateToken)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// GetGitLabOAuthURL returns the GitLab OAuth URL as JSON (for frontend to redirect)
func (h *GitHandler) GetGitLabOAuthURL(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	userID := auth.GetUserID(r.Context())

	if orgID == "" || userID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID or User ID not found in token"))
		return
	}

	// Check if GitLab OAuth is configured
	if h.config.GitLabClientID == "" || h.config.GitLabClientSecret == "" {
		WriteError(w, domain.NewInvalidInputError("GitLab OAuth is not configured. Please set GITLAB_CLIENT_ID and GITLAB_CLIENT_SECRET environment variables."))
		return
	}

	state, err := git.GenerateOAuthState("gitlab", orgID, userID)
	if err != nil {
		WriteError(w, domain.ErrInternal.WithError(err))
		return
	}

	// TODO: Store state in cache/DB for validation

	oauthConfig := &git.OAuthConfig{
		GitLabClientID:     h.config.GitLabClientID,
		GitLabClientSecret: h.config.GitLabClientSecret,
		GitLabRedirectURL:  h.config.GitLabRedirectURL,
		GitLabBaseURL:      h.config.GitLabBaseURL,
	}

	authURL := git.GetGitLabOAuthURL(oauthConfig, state.StateToken)
	
	WriteJSON(w, http.StatusOK, map[string]string{
		"auth_url": authURL,
	})
}

// ConnectGitLab initiates GitLab OAuth flow (redirects immediately)
func (h *GitHandler) ConnectGitLab(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	userID := auth.GetUserID(r.Context())

	if orgID == "" || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	state, err := git.GenerateOAuthState("gitlab", orgID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Store state in cache/DB for validation

	oauthConfig := &git.OAuthConfig{
		GitLabClientID:     h.config.GitLabClientID,
		GitLabClientSecret: h.config.GitLabClientSecret,
		GitLabRedirectURL:  h.config.GitLabRedirectURL,
		GitLabBaseURL:      h.config.GitLabBaseURL,
	}

	authURL := git.GetGitLabOAuthURL(oauthConfig, state.StateToken)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// CallbackGitHub handles GitHub OAuth callback
func (h *GitHandler) CallbackGitHub(w http.ResponseWriter, r *http.Request) {
	code, state, err := git.ParseOAuthCallback(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse state token to get orgID and userID (callback is public, no auth context)
	orgID, userID, err := git.ParseOAuthState(state)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid or expired state token: %v", err), http.StatusBadRequest)
		return
	}

	if orgID == "" || userID == "" {
		http.Error(w, "Missing orgID or userID in state token", http.StatusBadRequest)
		return
	}

	oauthConfig := &git.OAuthConfig{
		GitHubClientID:     h.config.GitHubClientID,
		GitHubClientSecret: h.config.GitHubClientSecret,
		GitHubRedirectURL:  h.config.GitHubRedirectURL,
	}

	token, err := git.ExchangeGitHubCode(r.Context(), oauthConfig, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info
	gitUser, err := git.GetGitHubUser(r.Context(), token.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if connection already exists
	existing, err := h.store.GetGitConnectionByOrgAndProvider(r.Context(), orgID, "github")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var refreshToken sql.NullString
	if token.RefreshToken != "" {
		refreshToken = sql.NullString{String: token.RefreshToken, Valid: true}
	}

	var tokenExpiresAt sql.NullTime
	if !token.Expiry.IsZero() {
		tokenExpiresAt = sql.NullTime{Time: token.Expiry, Valid: true}
	}

	connection := &store.GitConnection{
		CasdoorOrgID:   orgID,
		Provider:       "github",
		AccessToken:    token.AccessToken, // TODO: Encrypt
		RefreshToken:   refreshToken,
		TokenExpiresAt: tokenExpiresAt,
		AccountName:    sql.NullString{String: gitUser.Login, Valid: true},
		AccountID:      sql.NullString{String: fmt.Sprintf("%d", gitUser.ID), Valid: true},
		ConnectedBy:    sql.NullString{String: userID, Valid: true},
	}

	if existing != nil {
		// Update existing connection
		err = h.store.UpdateGitConnection(r.Context(), existing.ID, connection)
	} else {
		// Create new connection
		err = h.store.CreateGitConnection(r.Context(), connection)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to frontend callback page
	// Determine frontend URL from config or default
	frontendURL := h.config.BaseURL
	if frontendURL == "" || frontendURL == "http://localhost:8080" {
		frontendURL = "https://zyndra.armonika.cloud"
	}
	// Convert backend URL to frontend URL if needed
	if frontendURL == "https://api.zyndra.armonika.cloud" {
		frontendURL = "https://zyndra.armonika.cloud"
	}
	
	// Redirect to frontend callback page
	redirectURL := fmt.Sprintf("%s/git/callback?success=true&provider=github", frontendURL)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// CallbackGitLab handles GitLab OAuth callback
func (h *GitHandler) CallbackGitLab(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	userID := auth.GetUserID(r.Context())

	code, _, err := git.ParseOAuthCallback(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Validate state token (currently ignored)

	oauthConfig := &git.OAuthConfig{
		GitLabClientID:     h.config.GitLabClientID,
		GitLabClientSecret: h.config.GitLabClientSecret,
		GitLabRedirectURL:  h.config.GitLabRedirectURL,
		GitLabBaseURL:      h.config.GitLabBaseURL,
	}

	token, err := git.ExchangeGitLabCode(r.Context(), oauthConfig, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info
	gitUser, err := git.GetGitLabUser(r.Context(), token.AccessToken, h.config.GitLabBaseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if connection already exists
	existing, err := h.store.GetGitConnectionByOrgAndProvider(r.Context(), orgID, "gitlab")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var refreshToken sql.NullString
	if token.RefreshToken != "" {
		refreshToken = sql.NullString{String: token.RefreshToken, Valid: true}
	}

	var tokenExpiresAt sql.NullTime
	if !token.Expiry.IsZero() {
		tokenExpiresAt = sql.NullTime{Time: token.Expiry, Valid: true}
	}

	connection := &store.GitConnection{
		CasdoorOrgID:   orgID,
		Provider:       "gitlab",
		AccessToken:    token.AccessToken, // TODO: Encrypt
		RefreshToken:   refreshToken,
		TokenExpiresAt: tokenExpiresAt,
		AccountName:    sql.NullString{String: gitUser.Login, Valid: true},
		AccountID:      sql.NullString{String: fmt.Sprintf("%d", gitUser.ID), Valid: true},
		ConnectedBy:    sql.NullString{String: userID, Valid: true},
	}

	if existing != nil {
		// Update existing connection
		err = h.store.UpdateGitConnection(r.Context(), existing.ID, connection)
	} else {
		// Create new connection
		err = h.store.CreateGitConnection(r.Context(), connection)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to success page or return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"connection": connection,
	})
}

// ListConnections lists all git connections for the organization
func (h *GitHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	connections, err := h.store.ListGitConnectionsByOrg(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Don't expose access tokens
	for _, conn := range connections {
		conn.AccessToken = ""
		if conn.RefreshToken.Valid {
			conn.RefreshToken = sql.NullString{}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

// DeleteConnection deletes a git connection
func (h *GitHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}

	err = h.store.DeleteGitConnection(r.Context(), id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Connection not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListRepositories lists repositories for the authenticated user
func (h *GitHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "github" // Default
	}

	// Get connection for this provider
	connection, err := h.store.GetGitConnectionByOrgAndProvider(r.Context(), orgID, provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if connection == nil {
		WriteError(w, domain.NewNotFoundError(fmt.Sprintf("No %s connection found. Please connect your %s account first.", provider, provider)))
		return
	}

	var repos []*git.Repository
	switch provider {
	case "github":
		client := git.NewGitHubClient(connection.AccessToken)
		repos, err = client.GetUserRepositories(r.Context())
	case "gitlab":
		client := git.NewGitLabClient(connection.AccessToken, h.config.GitLabBaseURL)
		repos, err = client.GetUserRepositories(r.Context())
	default:
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

// ListBranches lists branches for a repository
func (h *GitHandler) ListBranches(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "github"
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	// Get connection
	connection, err := h.store.GetGitConnectionByOrgAndProvider(r.Context(), orgID, provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if connection == nil {
		http.Error(w, "No connection found", http.StatusNotFound)
		return
	}

	var branches []*git.Branch
	switch provider {
	case "github":
		client := git.NewGitHubClient(connection.AccessToken)
		branches, err = client.GetBranches(r.Context(), owner, repo)
	case "gitlab":
		client := git.NewGitLabClient(connection.AccessToken, h.config.GitLabBaseURL)
		branches, err = client.GetBranches(r.Context(), owner, repo)
	default:
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branches)
}

// GetRepositoryTree gets the directory tree for a repository
func (h *GitHandler) GetRepositoryTree(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "github"
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	branch := r.URL.Query().Get("branch")
	path := r.URL.Query().Get("path")

	// Get connection
	connection, err := h.store.GetGitConnectionByOrgAndProvider(r.Context(), orgID, provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if connection == nil {
		http.Error(w, "No connection found", http.StatusNotFound)
		return
	}

	var tree []*git.TreeEntry
	switch provider {
	case "github":
		client := git.NewGitHubClient(connection.AccessToken)
		tree, err = client.GetRepositoryTree(r.Context(), owner, repo, branch, path)
	case "gitlab":
		client := git.NewGitLabClient(connection.AccessToken, h.config.GitLabBaseURL)
		tree, err = client.GetRepositoryTree(r.Context(), owner, repo, branch, path)
	default:
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

