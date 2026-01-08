package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
)

type MockAuthHandler struct {
	config *config.Config
}

func NewMockAuthHandler(cfg *config.Config) *MockAuthHandler {
	return &MockAuthHandler{
		config: cfg,
	}
}

// MockLogin generates a mock token for testing
func (h *MockAuthHandler) MockLogin(w http.ResponseWriter, r *http.Request) {
	// Get user info from query params (optional, defaults provided)
	userID := r.URL.Query().Get("user_id")
	orgID := r.URL.Query().Get("org_id")
	name := r.URL.Query().Get("name")
	
	var roles []string
	if rolesParam := r.URL.Query().Get("roles"); rolesParam != "" {
		// Simple comma-separated roles
		roles = []string{rolesParam}
	}

	// Generate mock token
	token, err := auth.GenerateMockToken(userID, orgID, name, roles)
	if err != nil {
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return token as JSON
	response := map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterMockAuthRoutes registers mock authentication routes
func RegisterMockAuthRoutes(r chi.Router, cfg *config.Config) {
	mockAuthHandler := NewMockAuthHandler(cfg)
	
	// Public mock auth route (no auth required)
	r.Post("/auth/mock/login", mockAuthHandler.MockLogin)
}

