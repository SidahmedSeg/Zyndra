package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/email"
	"github.com/intelifox/click-deploy/internal/store"
)

// OTPHandler handles OTP-related endpoints
type OTPHandler struct {
	db            *store.DB
	config        *config.Config
	mailtrapClient *email.MailtrapClient
}

// NewOTPHandler creates a new OTP handler
func NewOTPHandler(db *store.DB, cfg *config.Config) *OTPHandler {
	var mailtrapClient *email.MailtrapClient
	if cfg.MailtrapAPIToken != "" {
		mailtrapClient = email.NewMailtrapClient(
			cfg.MailtrapAPIToken,
			cfg.MailtrapSenderEmail,
			cfg.MailtrapSenderName,
		)
	}
	return &OTPHandler{
		db:            db,
		config:        cfg,
		mailtrapClient: mailtrapClient,
	}
}

// SendOTPRequest is the request body for sending OTP
type SendOTPRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"` // "registration", "password_reset"
}

// SendOTPResponse is the response for sending OTP
type SendOTPResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// VerifyOTPRequest is the request body for verifying OTP
type VerifyOTPRequest struct {
	Email   string `json:"email"`
	Code    string `json:"code"`
	Purpose string `json:"purpose"`
}

// VerifyOTPResponse is the response for verifying OTP
type VerifyOTPResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

// CompleteRegistrationRequest is the request body for completing registration after OTP verification
type CompleteRegistrationRequest struct {
	Email           string `json:"email"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// SendOTP sends an OTP code to the user's email
func (h *OTPHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Validate purpose
	purpose := store.OTPPurposeRegistration
	if req.Purpose == "password_reset" {
		purpose = store.OTPPurposePasswordReset
	}

	// For registration, check if user already exists
	if purpose == store.OTPPurposeRegistration {
		existingUser, err := h.db.GetUserByEmail(r.Context(), req.Email)
		if err == nil && existingUser != nil {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
	}

	// Create OTP (expires in 10 minutes)
	otp, err := h.db.CreateOTPCode(r.Context(), req.Email, purpose, 10*time.Minute)
	if err != nil {
		http.Error(w, "Failed to create OTP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send OTP via Mailtrap
	if h.mailtrapClient != nil {
		if err := h.mailtrapClient.SendOTPEmail(req.Email, otp.Code); err != nil {
			log.Printf("Failed to send OTP email via Mailtrap: %v", err)
			// Don't fail the request - in production you might want to retry
			// For now, we log and continue
		}
	} else {
		// Development mode - log OTP to console
		log.Printf("📧 OTP for %s: %s (Mailtrap not configured)", req.Email, otp.Code)
	}

	resp := SendOTPResponse{
		Message:   "Verification code sent to your email",
		ExpiresIn: 600, // 10 minutes in seconds
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// VerifyOTP verifies an OTP code
func (h *OTPHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Code == "" {
		http.Error(w, "Email and code are required", http.StatusBadRequest)
		return
	}

	// Validate purpose
	purpose := store.OTPPurposeRegistration
	if req.Purpose == "password_reset" {
		purpose = store.OTPPurposePasswordReset
	}

	// Verify OTP
	verified, err := h.db.VerifyOTPCode(r.Context(), req.Email, req.Code, purpose)
	if err != nil {
		resp := VerifyOTPResponse{
			Verified: false,
			Message:  err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := VerifyOTPResponse{
		Verified: verified,
		Message:  "Email verified successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CompleteRegistration completes registration after OTP verification
func (h *OTPHandler) CompleteRegistration(w http.ResponseWriter, r *http.Request) {
	var req CompleteRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Name == "" || req.Password == "" {
		http.Error(w, "Email, name, and password are required", http.StatusBadRequest)
		return
	}

	if req.Password != req.ConfirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Check if OTP was verified for this email
	verified, err := h.db.IsOTPVerified(r.Context(), req.Email, store.OTPPurposeRegistration)
	if err != nil || !verified {
		http.Error(w, "Email not verified. Please verify your email first.", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, _ := h.db.GetUserByEmail(r.Context(), req.Email)
	if existingUser != nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Create user with verified email
	user, err := h.db.CreateUserWithVerifiedEmail(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create default organization for user
	orgName := req.Name + "'s Workspace"
	org, err := h.db.CreateOrganization(r.Context(), orgName, user.ID)
	if err != nil {
		http.Error(w, "Failed to create organization: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add user as owner
	_, err = h.db.AddOrgMember(r.Context(), org.ID, user.ID, "owner")
	if err != nil {
		http.Error(w, "Failed to add user to organization: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate tokens
	authHandler := NewCustomAuthHandler(h.db, h.config)
	tokenPair, err := authHandler.jwtService.GenerateTokenPair(
		user.ID,
		user.Email,
		req.Name,
		org.ID,
		org.Slug,
		"owner",
	)
	if err != nil {
		http.Error(w, "Failed to generate tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Clean up OTP codes for this email
	_ = h.db.DeleteOTPCodes(r.Context(), req.Email, store.OTPPurposeRegistration)

	// Return auth response
	resp := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    "Bearer",
		User: &UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Name:          req.Name,
			EmailVerified: true,
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

// RegisterOTPRoutes registers OTP-related routes
func RegisterOTPRoutes(r interface{ Post(pattern string, handlerFn http.HandlerFunc) }, db *store.DB, cfg *config.Config) {
	handler := NewOTPHandler(db, cfg)
	
	r.Post("/auth/otp/send", handler.SendOTP)
	r.Post("/auth/otp/verify", handler.VerifyOTP)
	r.Post("/auth/register/complete", handler.CompleteRegistration)
}

