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
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/worker"
)

type ServiceHandler struct {
	Store  *store.DB
	config *config.Config
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(store *store.DB, cfg *config.Config) *ServiceHandler {
	return &ServiceHandler{
		Store:  store,
		config: cfg,
	}
}

// ListServices handles GET /projects/:id/services
func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid project ID"))
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	// Verify project belongs to organization
	project, err := h.Store.GetProject(r.Context(), projectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Project"))
		return
	}

	// List services in project
	services, err := h.Store.ListServicesByProject(r.Context(), projectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteJSON(w, http.StatusOK, services)
}

// CreateService handles POST /projects/:id/services
func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid project ID"))
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	// Verify project belongs to organization
	project, err := h.Store.GetProject(r.Context(), projectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Project"))
		return
	}

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid request body: "+err.Error()))
		return
	}

	// Sanitize input
	req.Name = SanitizeString(req.Name)

	// Validate request
	if validationErrs := ValidateCreateServiceRequest(&req); validationErrs.HasErrors() {
		WriteError(w, validationErrs.ToAppError())
		return
	}

	// Create service
	service := &store.Service{
		ProjectID:    projectID,
		Name:         req.Name,
		Type:         req.Type,
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
		CanvasX:      0,
		CanvasY:      0,
	}

	if req.InstanceSize != "" {
		service.InstanceSize = req.InstanceSize
	}

	if req.Port != nil {
		service.Port = *req.Port
	}

	if req.CanvasX != nil {
		service.CanvasX = *req.CanvasX
	}

	if req.CanvasY != nil {
		service.CanvasY = *req.CanvasY
	}

	if req.GitSourceID != nil {
		gitSourceUUID, err := uuid.Parse(*req.GitSourceID)
		if err == nil {
			service.GitSourceID = sql.NullString{String: gitSourceUUID.String(), Valid: true}
		}
	}

	if err := h.Store.CreateService(r.Context(), service); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteCreated(w, service)
}

// GetService handles GET /services/:id
func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid service ID"))
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	// Get service
	service, err := h.Store.GetService(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	if service == nil {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Verify service belongs to organization via project
	project, err := h.Store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	WriteJSON(w, http.StatusOK, service)
}

// UpdateService handles PATCH /services/:id
func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid service ID"))
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	// Get service
	service, err := h.Store.GetService(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	if service == nil {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Verify service belongs to organization
	project, err := h.Store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	var req UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid request body: "+err.Error()))
		return
	}

	// Validate request
	if validationErrs := ValidateUpdateServiceRequest(&req); validationErrs.HasErrors() {
		WriteError(w, validationErrs.ToAppError())
		return
	}

	// Update fields if provided
	if req.Name != nil {
		service.Name = *req.Name
	}

	if req.Type != nil {
		service.Type = *req.Type
	}

	if req.InstanceSize != nil {
		service.InstanceSize = *req.InstanceSize
	}

	if req.Port != nil {
		service.Port = *req.Port
	}

	if req.Status != nil {
		service.Status = *req.Status
	}

	// Update service
	if err := h.Store.UpdateService(r.Context(), id, service); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Fetch updated service
	updatedService, err := h.Store.GetService(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteJSON(w, http.StatusOK, updatedService)
}

// UpdateServicePosition handles PATCH /services/:id/position
func (h *ServiceHandler) UpdateServicePosition(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid service ID"))
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	// Get service
	service, err := h.Store.GetService(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	if service == nil {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Verify service belongs to organization
	project, err := h.Store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	var req UpdateServicePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid request body: "+err.Error()))
		return
	}

	// Validate position
	if validationErrs := ValidateUpdateServicePositionRequest(&req); validationErrs.HasErrors() {
		WriteError(w, validationErrs.ToAppError())
		return
	}

	// Update position
	if err := h.Store.UpdateServicePosition(r.Context(), id, req.X, req.Y); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Fetch updated service
	updatedService, err := h.Store.GetService(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteJSON(w, http.StatusOK, updatedService)
}

// DeleteService handles DELETE /services/:id
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid service ID"))
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	// Get service
	service, err := h.Store.GetService(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	if service == nil {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Verify service belongs to organization
	project, err := h.Store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Create cleanup job for service resources
	cleanupWorker := worker.NewCleanupWorker(h.Store, h.config)
	if err := cleanupWorker.CleanupServiceResources(r.Context(), id); err != nil {
		// Log error but continue with deletion
		// The database will be cleaned up via cascade
		fmt.Printf("Warning: failed to cleanup service resources: %v\n", err)
	}

	// Delete service (cascade will delete related resources in DB)
	if err := h.Store.DeleteService(r.Context(), id); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteNoContent(w)
}


