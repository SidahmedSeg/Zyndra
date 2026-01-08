package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/intelifox/click-deploy/internal/config"
)

type AuthHandler struct {
	config *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		config: cfg,
	}
}

// InitiateCasdoorLogin redirects to Casdoor OAuth login
func (h *AuthHandler) InitiateCasdoorLogin(w http.ResponseWriter, r *http.Request) {
	// Get redirect URL from query or use default
	redirectURL := r.URL.Query().Get("redirect_uri")
	if redirectURL == "" {
		// Default to frontend URL - construct from BASE_URL
		// If BASE_URL is api.zyndra.armonika.cloud, frontend is zyndra.armonika.cloud
		baseURL := h.config.BaseURL
		if strings.Contains(baseURL, "api.") {
			redirectURL = strings.Replace(baseURL, "api.", "", 1) + "/auth/callback"
		} else {
			// Fallback: assume frontend is on same domain
			redirectURL = baseURL + "/auth/callback"
		}
	}

	// Construct Casdoor OAuth URL
	// Casdoor OAuth flow: /login/oauth/authorize
	authURL := fmt.Sprintf("%s/login/oauth/authorize", h.config.CasdoorEndpoint)
	
	params := url.Values{}
	params.Set("client_id", h.config.CasdoorClientID)
	params.Set("redirect_uri", redirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email")
	params.Set("state", "random_state_string") // In production, use a secure random state

	authURLWithParams := authURL + "?" + params.Encode()

	http.Redirect(w, r, authURLWithParams, http.StatusFound)
}

// CallbackCasdoor handles the OAuth callback from Casdoor
func (h *AuthHandler) CallbackCasdoor(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	// Casdoor token endpoint: /api/login/oauth/access_token
	tokenURL := fmt.Sprintf("%s/api/login/oauth/access_token", h.config.CasdoorEndpoint)
	
	// Prepare token exchange request
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("client_id", h.config.CasdoorClientID)
	formData.Set("client_secret", h.config.CasdoorClientSecret)
	
	// Get redirect_uri from query or use default
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		baseURL := h.config.BaseURL
		if strings.Contains(baseURL, "api.") {
			redirectURI = strings.Replace(baseURL, "api.", "", 1) + "/auth/callback"
		} else {
			redirectURI = baseURL + "/auth/callback"
		}
	}
	formData.Set("redirect_uri", redirectURI)

	// Make token exchange request
	resp, err := http.PostForm(tokenURL, formData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response for debugging
		bodyBytes := make([]byte, 1024)
		resp.Body.Read(bodyBytes)
		http.Error(w, fmt.Sprintf("Token exchange failed: %d - %s", resp.StatusCode, string(bodyBytes)), http.StatusInternalServerError)
		return
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode token response: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to frontend with token
	// In production, use a more secure method (e.g., httpOnly cookie)
	baseURL := h.config.BaseURL
	var frontendURL string
	if strings.Contains(baseURL, "api.") {
		frontendURL = strings.Replace(baseURL, "api.", "", 1)
	} else {
		frontendURL = baseURL
	}
	redirectToFrontend := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, tokenResponse.AccessToken)
	
	http.Redirect(w, r, redirectToFrontend, http.StatusFound)
}

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r chi.Router, cfg *config.Config) {
	authHandler := NewAuthHandler(cfg)
	
	// Public auth routes (no auth required)
	r.Get("/auth/casdoor/login", authHandler.InitiateCasdoorLogin)
	r.Get("/auth/casdoor/callback", authHandler.CallbackCasdoor)
}

