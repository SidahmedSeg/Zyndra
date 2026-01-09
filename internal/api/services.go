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

// ServiceResponse represents a service in API responses
type ServiceResponse struct {
	ID                  string  `json:"id"`
	ProjectID           string  `json:"project_id"`
	GitSourceID         *string `json:"git_source_id,omitempty"`
	Name                string  `json:"name"`
	Type                string  `json:"type"`
	Status              string  `json:"status"`
	InstanceSize        string  `json:"instance_size"`
	Port                int     `json:"port"`
	OpenStackInstanceID *string `json:"openstack_instance_id,omitempty"`
	OpenStackFIPID      *string `json:"openstack_fip_id,omitempty"`
	OpenStackFIPAddress *string `json:"openstack_fip_address,omitempty"`
	SecurityGroupID     *string `json:"security_group_id,omitempty"`
	Subdomain           *string `json:"subdomain,omitempty"`
	GeneratedURL        *string `json:"generated_url,omitempty"`
	CurrentImageTag     *string `json:"current_image_tag,omitempty"`
	CanvasX             int     `json:"canvas_x"`
	CanvasY             int     `json:"canvas_y"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

// toServiceResponse converts a store.Service to ServiceResponse
func toServiceResponse(s *store.Service) ServiceResponse {
	resp := ServiceResponse{
		ID:           s.ID.String(),
		ProjectID:    s.ProjectID.String(),
		Name:         s.Name,
		Type:         s.Type,
		Status:       s.Status,
		InstanceSize: s.InstanceSize,
		Port:         s.Port,
		CanvasX:      s.CanvasX,
		CanvasY:      s.CanvasY,
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if s.GitSourceID.Valid {
		resp.GitSourceID = &s.GitSourceID.String
	}
	if s.OpenStackInstanceID.Valid {
		resp.OpenStackInstanceID = &s.OpenStackInstanceID.String
	}
	if s.OpenStackFIPID.Valid {
		resp.OpenStackFIPID = &s.OpenStackFIPID.String
	}
	if s.OpenStackFIPAddress.Valid {
		resp.OpenStackFIPAddress = &s.OpenStackFIPAddress.String
	}
	if s.SecurityGroupID.Valid {
		resp.SecurityGroupID = &s.SecurityGroupID.String
	}
	if s.Subdomain.Valid {
		resp.Subdomain = &s.Subdomain.String
	}
	if s.GeneratedURL.Valid {
		resp.GeneratedURL = &s.GeneratedURL.String
	}
	if s.CurrentImageTag.Valid {
		resp.CurrentImageTag = &s.CurrentImageTag.String
	}

	return resp
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

	// Convert store.Service to ServiceResponse
	response := make([]ServiceResponse, 0)
	if services != nil {
		for _, s := range services {
			if s != nil {
				response = append(response, toServiceResponse(s))
			}
		}
	}

	WriteJSON(w, http.StatusOK, response)
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


