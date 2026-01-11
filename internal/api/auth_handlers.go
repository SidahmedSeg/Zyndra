package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
)

// CustomAuthHandler handles custom JWT authentication endpoints
type CustomAuthHandler struct {
	db         *store.DB
	jwtService *auth.JWTService
	config     *config.Config
}

// NewCustomAuthHandler creates a new custom auth handler
func NewCustomAuthHandler(db *store.DB, cfg *config.Config) *CustomAuthHandler {
	jwtConfig := auth.DefaultJWTConfig(cfg.JWTSecret)
	if cfg.JWTAccessExpiry > 0 {
		jwtConfig.AccessExpiry = cfg.JWTAccessExpiry
	}
	if cfg.JWTRefreshExpiry > 0 {
		jwtConfig.RefreshExpiry = cfg.JWTRefreshExpiry
	}
	
	return &CustomAuthHandler{
		db:         db,
		jwtService: auth.NewJWTService(jwtConfig),
		config:     cfg,
	}
}

// RegisterRequest is the request body for registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest is the request body for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest is the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse is the response for auth endpoints
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    time.Time   `json:"expires_at"`
	TokenType    string      `json:"token_type"`
	User         *UserResponse `json:"user"`
}

// UserResponse is the user info in auth responses
type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	EmailVerified bool   `json:"email_verified"`
	Org           *OrgResponse `json:"organization,omitempty"`
}

// OrgResponse is the organization info in auth responses
type OrgResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}

// Register handles user registration
func (h *CustomAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "Email, password, and name are required", http.StatusBadRequest)
		return
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	exists, err := h.db.UserExistsByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Validate password
	if err := auth.ValidatePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.db.CreateUser(r.Context(), req.Email, passwordHash, req.Name)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create default organization for the user
	org, err := h.db.CreateOrganization(r.Context(), req.Name+"'s Workspace", user.ID)
	if err != nil {
		http.Error(w, "Failed to create organization", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.Name, org.ID, org.Slug, "owner")
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Store refresh token
	expiresAt := time.Now().Add(h.jwtService.RefreshExpiry())
	_, err = h.db.CreateRefreshToken(r.Context(), user.ID, tokenPair.RefreshToken, expiresAt)
	if err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	// Return response
	resp := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User: &UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			AvatarURL:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
			Org: &OrgResponse{
				ID:   org.ID,
				Name: org.Name,
				Slug: org.Slug,
				Role: "owner",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login handles user login
func (h *CustomAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Get user's organizations
	orgs, err := h.db.ListUserOrganizations(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to get organizations", http.StatusInternalServerError)
		return
	}

	// Use the first organization (primary)
	var org *store.Organization
	var role string
	if len(orgs) > 0 {
		org = orgs[0]
		role, _ = h.db.GetUserRoleInOrg(r.Context(), org.ID, user.ID)
	}

	// Generate tokens
	var orgID, orgSlug string
	if org != nil {
		orgID = org.ID
		orgSlug = org.Slug
	}
	
	tokenPair, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.Name, orgID, orgSlug, role)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Store refresh token
	expiresAt := time.Now().Add(h.jwtService.RefreshExpiry())
	_, err = h.db.CreateRefreshToken(r.Context(), user.ID, tokenPair.RefreshToken, expiresAt)
	if err != nil {
		http.Error(w, "Failed to store refresh token", http.StatusInternalServerError)
		return
	}

	// Build response
	resp := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User: &UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			AvatarURL:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
		},
	}

	if org != nil {
		resp.User.Org = &OrgResponse{
			ID:   org.ID,
			Name: org.Name,
			Slug: org.Slug,
			Role: role,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Refresh handles token refresh
func (h *CustomAuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Validate refresh token
	rt, err := h.db.ValidateRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := h.db.GetUserByID(r.Context(), rt.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get user's organizations
	orgs, err := h.db.ListUserOrganizations(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to get organizations", http.StatusInternalServerError)
		return
	}

	// Use the first organization (primary)
	var org *store.Organization
	var role string
	if len(orgs) > 0 {
		org = orgs[0]
		role, _ = h.db.GetUserRoleInOrg(r.Context(), org.ID, user.ID)
	}

	// Generate new tokens
	var orgID, orgSlug string
	if org != nil {
		orgID = org.ID
		orgSlug = org.Slug
	}
	
	tokenPair, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.Name, orgID, orgSlug, role)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Rotate refresh token (revoke old, create new)
	expiresAt := time.Now().Add(h.jwtService.RefreshExpiry())
	_, err = h.db.RotateRefreshToken(r.Context(), req.RefreshToken, tokenPair.RefreshToken, user.ID, expiresAt)
	if err != nil {
		http.Error(w, "Failed to rotate refresh token", http.StatusInternalServerError)
		return
	}

	// Build response
	resp := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
		User: &UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			AvatarURL:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
		},
	}

	if org != nil {
		resp.User.Org = &OrgResponse{
			ID:   org.ID,
			Name: org.Name,
			Slug: org.Slug,
			Role: role,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout
func (h *CustomAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Revoke refresh token
	err := h.db.RevokeRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		// Don't expose if token doesn't exist
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// Me returns the current user's information
func (h *CustomAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user's organizations
	orgs, err := h.db.ListUserOrganizations(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to get organizations", http.StatusInternalServerError)
		return
	}

	resp := UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		AvatarURL:     user.AvatarURL,
		EmailVerified: user.EmailVerified,
	}

	// Include current org from context
	orgID := auth.GetOrgID(r.Context())
	for _, org := range orgs {
		if org.ID == orgID {
			role, _ := h.db.GetUserRoleInOrg(r.Context(), org.ID, user.ID)
			resp.Org = &OrgResponse{
				ID:   org.ID,
				Name: org.Name,
				Slug: org.Slug,
				Role: role,
			}
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetJWTService returns the JWT service for middleware use
func (h *CustomAuthHandler) GetJWTService() *auth.JWTService {
	return h.jwtService
}

// RegisterCustomAuthRoutes registers the custom auth routes
func RegisterCustomAuthRoutes(r chi.Router, db *store.DB, cfg *config.Config, authValidator auth.ValidatorInterface) *CustomAuthHandler {
	handler := NewCustomAuthHandler(db, cfg)

	r.Route("/auth", func(r chi.Router) {
		// Public routes
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/refresh", handler.Refresh)
		r.Post("/logout", handler.Logout)
		
		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(authValidator))
			r.Get("/me", handler.Me)
		})
	})

	return handler
}

