package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/domain"
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/worker"
)

type ProjectHandler struct {
	Store  *store.DB
	config *config.Config
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(store *store.DB, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{
		Store:  store,
		config: cfg,
	}
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Slug              string  `json:"slug"`
	Description       *string `json:"description,omitempty"`
	CasdoorOrgID      string  `json:"casdoor_org_id"`
	OpenStackTenantID *string `json:"openstack_tenant_id,omitempty"`
	OpenStackNetworkID *string `json:"openstack_network_id,omitempty"`
	DefaultRegion     *string `json:"default_region,omitempty"`
	AutoDeploy        bool    `json:"auto_deploy"`
	CreatedBy         *string `json:"created_by,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// toProjectResponse converts a store.Project to ProjectResponse
func toProjectResponse(p *store.Project) ProjectResponse {
	resp := ProjectResponse{
		ID:           p.ID.String(),
		Name:         p.Name,
		Slug:         p.Slug,
		CasdoorOrgID: p.CasdoorOrgID,
		AutoDeploy:   p.AutoDeploy,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.Description.Valid {
		resp.Description = &p.Description.String
	}
	if p.OpenStackTenantID != "" {
		resp.OpenStackTenantID = &p.OpenStackTenantID
	}
	if p.OpenStackNetworkID.Valid {
		resp.OpenStackNetworkID = &p.OpenStackNetworkID.String
	}
	if p.DefaultRegion.Valid {
		resp.DefaultRegion = &p.DefaultRegion.String
	}
	if p.CreatedBy.Valid {
		resp.CreatedBy = &p.CreatedBy.String
	}

	return resp
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	// Get org_id from context (set by auth middleware)
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	projects, err := h.Store.ListProjectsByOrg(r.Context(), orgID)
	if err != nil {
		// Log the error for debugging
		log.Printf("Error listing projects for org %s: %v", orgID, err)
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Convert store.Project to ProjectResponse
	response := make([]ProjectResponse, 0)
	if projects != nil {
		for _, p := range projects {
			if p != nil {
				response = append(response, toProjectResponse(p))
			}
		}
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
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

	project, err := h.Store.GetProject(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	if project == nil {
		WriteError(w, domain.NewNotFoundError("Project"))
		return
	}

	// Verify project belongs to organization
	if project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Project"))
		return
	}

	WriteJSON(w, http.StatusOK, toProjectResponse(project))
}

// CreateProject handles POST /projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	// Get org_id and user_id from context
	orgID := auth.GetOrgID(r.Context())
	userID := auth.GetUserID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid request body: "+err.Error()))
		return
	}

	// Sanitize and validate request
	req.Name = SanitizeString(req.Name)
	if validationErrs := ValidateCreateProjectRequest(&req); validationErrs.HasErrors() {
		WriteError(w, validationErrs.ToAppError())
		return
	}

	// Generate slug from name
	slug := store.GenerateSlug(req.Name)

	// Check if slug already exists for this org
	existingProjects, err := h.Store.ListProjectsByOrg(r.Context(), orgID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Ensure slug is unique
	baseSlug := slug
	counter := 1
	for {
		unique := true
		for _, p := range existingProjects {
			if p.Slug == slug {
				unique = false
				break
			}
		}
		if unique {
			break
		}
		slug = baseSlug + "-" + strconv.Itoa(counter)
		counter++
	}

	// Create project
	project := &store.Project{
		CasdoorOrgID:      orgID,
		Name:              req.Name,
		Slug:              slug,
		OpenStackTenantID: req.OpenStackTenantID,
		AutoDeploy:        true,
	}

	if req.Description != nil {
		project.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	if req.DefaultRegion != nil {
		project.DefaultRegion = sql.NullString{String: *req.DefaultRegion, Valid: true}
	}

	if req.AutoDeploy != nil {
		project.AutoDeploy = *req.AutoDeploy
	}

	if userID != "" {
		project.CreatedBy = sql.NullString{String: userID, Valid: true}
	}

	if err := h.Store.CreateProject(r.Context(), project); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Fetch created project to return full details
	createdProject, err := h.Store.GetProject(r.Context(), project.ID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteCreated(w, toProjectResponse(createdProject))
}

// UpdateProject handles PATCH /projects/:id
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Organization ID not found in token", http.StatusUnauthorized)
		return
	}

	// Verify project exists and belongs to organization
	exists, err := h.Store.ProjectExists(r.Context(), id, orgID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if !exists {
		WriteError(w, domain.NewNotFoundError("Project"))
		return
	}

	// Get existing project
	project, err := h.Store.GetProject(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid request body: "+err.Error()))
		return
	}

	// Validate request
	if validationErrs := ValidateUpdateProjectRequest(&req); validationErrs.HasErrors() {
		WriteError(w, validationErrs.ToAppError())
		return
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
		// Regenerate slug if name changed
		project.Slug = store.GenerateSlug(*req.Name)
	}

	if req.Description != nil {
		project.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	if req.DefaultRegion != nil {
		project.DefaultRegion = sql.NullString{String: *req.DefaultRegion, Valid: true}
	}

	if req.AutoDeploy != nil {
		project.AutoDeploy = *req.AutoDeploy
	}

	// Update project
	if err := h.Store.UpdateProject(r.Context(), id, project); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	// Fetch updated project
	updatedProject, err := h.Store.GetProject(r.Context(), id)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteJSON(w, http.StatusOK, toProjectResponse(updatedProject))
}

// DeleteProject handles DELETE /projects/:id
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Get org_id from context
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Organization ID not found in token", http.StatusUnauthorized)
		return
	}

	// Verify project exists and belongs to organization
	exists, err := h.Store.ProjectExists(r.Context(), id, orgID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if !exists {
		WriteError(w, domain.NewNotFoundError("Project"))
		return
	}

	// Create cleanup job for project resources
	cleanupWorker := worker.NewCleanupWorker(h.Store, h.config)
	if err := cleanupWorker.CleanupProjectResources(r.Context(), id); err != nil {
		// Log error but continue with deletion
		// The database will be cleaned up via cascade
		fmt.Printf("Warning: failed to cleanup project resources: %v\n", err)
	}

	// Delete project (cascade will delete related resources in DB)
	if err := h.Store.DeleteProject(r.Context(), id, orgID); err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}

	WriteNoContent(w)
}


