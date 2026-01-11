package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/caddy"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/domain"
	"github.com/intelifox/click-deploy/internal/store"
)

type CustomDomainHandler struct {
	store  *store.DB
	config *config.Config
	caddy  *caddy.Client
}

func NewCustomDomainHandler(store *store.DB, cfg *config.Config) *CustomDomainHandler {
	return &CustomDomainHandler{
		store:  store,
		config: cfg,
		caddy:  caddy.NewClient(cfg.CaddyAdminURL),
	}
}

// RegisterCustomDomainRoutes registers custom domain routes
func RegisterCustomDomainRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewCustomDomainHandler(db, cfg)

	r.Get("/services/{id}/domains", h.ListCustomDomains)
	r.Post("/services/{id}/domains", h.AddCustomDomain)
	r.Get("/domains/{id}", h.GetCustomDomain)
	r.Post("/domains/{id}/verify", h.VerifyCustomDomain)
	r.Delete("/domains/{id}", h.DeleteCustomDomain)
}

// AddCustomDomainRequest represents a request to add a custom domain
type AddCustomDomainRequest struct {
	Domain string `json:"domain" validate:"required,hostname"`
}

// AddCustomDomain handles POST /services/:id/domains
func (h *CustomDomainHandler) AddCustomDomain(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid service ID"))
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), serviceID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if service == nil {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Parse request
	var req AddCustomDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid request body"))
		return
	}

	// Sanitize input
	req.Domain = SanitizeDomain(req.Domain)

	// Validate domain
	if validationErr := ValidateAddCustomDomainRequest(&req); validationErr != nil {
		WriteError(w, validationErr)
		return
	}

	// Check if domain already exists for this service
	existingDomains, err := h.store.ListCustomDomainsByService(r.Context(), serviceID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	for _, existingDomain := range existingDomains {
		if existingDomain.Domain == req.Domain {
			WriteError(w, domain.NewAppError(domain.ErrCodeAlreadyExists, "Domain already exists for this service", http.StatusConflict))
			return
		}
	}

	// Get service's floating IP for CNAME target
	if !service.OpenStackFIPAddress.Valid {
		WriteError(w, domain.NewAppError(domain.ErrCodeValidation, "Service does not have a floating IP address", http.StatusBadRequest))
		return
	}

	targetIP := service.OpenStackFIPAddress.String

	// Create custom domain record
	customDomain := &store.CustomDomain{
		ServiceID:     serviceID,
		Domain:        req.Domain,
		Status:        "pending",
		CNAMETarget:   store.StringToNullString(targetIP),
		SSLEnabled:    true, // Enable SSL by default
		ValidationToken: store.StringToNullString(uuid.New().String()),
	}

	if err := h.store.CreateCustomDomain(r.Context(), customDomain); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Add route to Caddy (even if not verified yet, Caddy will handle it)
	if err := h.caddy.AddRoute(r.Context(), req.Domain, targetIP, service.Port, true); err != nil {
		// Log error but don't fail - route can be added later
		// Update status to error
		customDomain.Status = "error"
		h.store.UpdateCustomDomain(r.Context(), customDomain.ID, customDomain)
		WriteError(w, domain.NewAppError(domain.ErrCodeExternalAPI, "Failed to add route to Caddy: "+err.Error(), http.StatusInternalServerError))
		return
	}

	// Update status to active
	customDomain.Status = "active"
	if err := h.store.UpdateCustomDomain(r.Context(), customDomain.ID, customDomain); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteJSON(w, http.StatusCreated, customDomain)
}

// ListCustomDomains handles GET /services/:id/domains
func (h *CustomDomainHandler) ListCustomDomains(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid service ID"))
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), serviceID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if service == nil {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Get custom domains
	domains, err := h.store.ListCustomDomainsByService(r.Context(), serviceID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteJSON(w, http.StatusOK, domains)
}

// GetCustomDomain handles GET /domains/:id
func (h *CustomDomainHandler) GetCustomDomain(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid domain ID"))
		return
	}

	// Get custom domain
	customDomain, err := h.store.GetCustomDomain(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if customDomain == nil {
		WriteError(w, domain.NewNotFoundError("Custom Domain"))
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), customDomain.ServiceID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		WriteError(w, domain.NewNotFoundError("Custom Domain"))
		return
	}

	WriteJSON(w, http.StatusOK, customDomain)
}

// VerifyCustomDomain handles POST /domains/:id/verify
func (h *CustomDomainHandler) VerifyCustomDomain(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid domain ID"))
		return
	}

	// Get custom domain
	customDomain, err := h.store.GetCustomDomain(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if customDomain == nil {
		WriteError(w, domain.NewNotFoundError("Custom Domain"))
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), customDomain.ServiceID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		WriteError(w, domain.NewNotFoundError("Custom Domain"))
		return
	}

	// Verify CNAME record exists
	verified, err := h.caddy.ValidateDomain(r.Context(), customDomain.Domain, customDomain.CNAMETarget.String)
	if err != nil {
		WriteError(w, domain.NewAppError(domain.ErrCodeExternalAPI, "Failed to verify domain: "+err.Error(), http.StatusInternalServerError))
		return
	}

	if verified {
		customDomain.Status = "verified"
		if err := h.store.UpdateCustomDomain(r.Context(), id, customDomain); err != nil {
			WriteError(w, domain.ErrDatabase.WithError(err))
			return
		}
	}

	WriteJSON(w, http.StatusOK, customDomain)
}

// DeleteCustomDomain handles DELETE /domains/:id
func (h *CustomDomainHandler) DeleteCustomDomain(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid domain ID"))
		return
	}

	// Get custom domain
	customDomain, err := h.store.GetCustomDomain(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if customDomain == nil {
		WriteError(w, domain.NewNotFoundError("Custom Domain"))
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), customDomain.ServiceID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		WriteError(w, domain.NewNotFoundError("Custom Domain"))
		return
	}

	// Remove route from Caddy
	if err := h.caddy.RemoveRoute(r.Context(), customDomain.Domain); err != nil {
		// Log error but continue with deletion
		// Route can be manually removed later
	}

	// Delete custom domain
	if err := h.store.DeleteCustomDomain(r.Context(), id); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteNoContent(w)
}

