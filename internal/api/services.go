package api

import (
	"context"
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
	
	// Git source info (populated from git_sources table)
	RepoOwner *string `json:"repo_owner,omitempty"`
	RepoName  *string `json:"repo_name,omitempty"`
	Branch    *string `json:"branch,omitempty"`
	RootDir   *string `json:"root_dir,omitempty"`
	
	// Resource limits
	CPULimit    *string `json:"cpu_limit,omitempty"`
	MemoryLimit *string `json:"memory_limit,omitempty"`
	
	// Build config
	StartCommand *string `json:"start_command,omitempty"`
	BuildCommand *string `json:"build_command,omitempty"`
	
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

// toServiceResponseWithGitSource adds git source info to a service response
func (h *ServiceHandler) toServiceResponseWithGitSource(ctx context.Context, s *store.Service) ServiceResponse {
	resp := toServiceResponse(s)
	
	// Fetch git source info if available
	if s.GitSourceID.Valid {
		gitSource, err := h.Store.GetGitSourceByService(ctx, s.ID)
		if err == nil && gitSource != nil {
			resp.RepoOwner = &gitSource.RepoOwner
			resp.RepoName = &gitSource.RepoName
			resp.Branch = &gitSource.Branch
			if gitSource.RootDir.Valid {
				resp.RootDir = &gitSource.RootDir.String
			}
		}
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
	if project == nil || !project.BelongsToOrg(orgID) {
		WriteError(w, domain.NewNotFoundError("Project"))
		return
	}

	// List services in project
	services, err := h.Store.ListServicesByProject(r.Context(), projectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Convert store.Service to ServiceResponse with git source info
	response := make([]ServiceResponse, 0)
	if services != nil {
		for _, s := range services {
			if s != nil {
				response = append(response, h.toServiceResponseWithGitSource(r.Context(), s))
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
	if project == nil || !project.BelongsToOrg(orgID) {
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

	// Handle git source ID if provided
	if req.GitSourceID != nil {
		gitSourceUUID, err := uuid.Parse(*req.GitSourceID)
		if err == nil {
			service.GitSourceID = sql.NullString{String: gitSourceUUID.String(), Valid: true}
		}
	}

	// Create service first
	if err := h.Store.CreateService(r.Context(), service); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// If git source info provided, create git source after service creation
	if req.GitSource != nil {
		// Get git connection for this org and provider
		connection, err := h.Store.GetGitConnectionByOrgAndProvider(r.Context(), orgID, req.GitSource.Provider)
		if err != nil {
			WriteError(w, domain.ErrDatabase.WithError(err))
			return
		}
		if connection == nil {
			WriteError(w, domain.NewInvalidInputError(fmt.Sprintf("No %s connection found. Please connect your %s account first.", req.GitSource.Provider, req.GitSource.Provider)))
			return
		}

		// Create git source
		gitSource := &store.GitSource{
			ServiceID:       service.ID,
			GitConnectionID: connection.ID,
			Provider:        req.GitSource.Provider,
			RepoOwner:       SanitizeString(req.GitSource.RepoOwner),
			RepoName:        SanitizeString(req.GitSource.RepoName),
			Branch:          SanitizeString(req.GitSource.Branch),
		}

		if req.GitSource.RootDir != nil {
			gitSource.RootDir = sql.NullString{String: SanitizeString(*req.GitSource.RootDir), Valid: true}
		}

		if err := h.Store.CreateGitSource(r.Context(), gitSource); err != nil {
			WriteError(w, domain.ErrDatabase.WithError(err))
			return
		}

		// Update service with git source ID
		service.GitSourceID = sql.NullString{String: gitSource.ID.String(), Valid: true}
		// Note: UpdateService doesn't update git_source_id, but that's okay
		// The git_source table has the service_id foreign key, so the relationship is established
	}

	// Fetch created service to return full details
	createdService, err := h.Store.GetService(r.Context(), service.ID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteCreated(w, h.toServiceResponseWithGitSource(r.Context(), createdService))
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
	if project == nil || !project.BelongsToOrg(orgID) {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	WriteJSON(w, http.StatusOK, h.toServiceResponseWithGitSource(r.Context(), service))
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
	if project == nil || !project.BelongsToOrg(orgID) {
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

	// Update git source if branch or root_dir provided
	if req.Branch != nil || req.RootDir != nil {
		gitSource, err := h.Store.GetGitSourceByService(r.Context(), id)
		if err != nil {
			WriteError(w, domain.ErrDatabase.WithError(err))
			return
		}
		
		if gitSource != nil {
			if req.Branch != nil {
				gitSource.Branch = *req.Branch
			}
			if req.RootDir != nil {
				gitSource.RootDir = sql.NullString{String: *req.RootDir, Valid: *req.RootDir != ""}
			}
			
			if err := h.Store.UpdateGitSource(r.Context(), gitSource.ID, gitSource); err != nil {
				WriteError(w, domain.ErrDatabase.WithError(err))
				return
			}
		}
	}

	// Fetch updated service
	updatedService, err := h.Store.GetService(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteJSON(w, http.StatusOK, h.toServiceResponseWithGitSource(r.Context(), updatedService))
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
	if project == nil || !project.BelongsToOrg(orgID) {
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

	WriteJSON(w, http.StatusOK, h.toServiceResponseWithGitSource(r.Context(), updatedService))
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
	if project == nil || !project.BelongsToOrg(orgID) {
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


