package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
)

type VolumeHandler struct {
	store  *store.DB
	config *config.Config
}

func NewVolumeHandler(store *store.DB, cfg *config.Config) *VolumeHandler {
	return &VolumeHandler{
		store:  store,
		config: cfg,
	}
}

// RegisterVolumeRoutes registers volume-related routes
func RegisterVolumeRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewVolumeHandler(db, cfg)

	r.Get("/projects/{id}/volumes", h.ListVolumes)
	r.Post("/projects/{id}/volumes", h.CreateVolume)
	r.Get("/volumes/{id}", h.GetVolume)
	r.Patch("/volumes/{id}/attach", h.AttachVolume)
	r.Patch("/volumes/{id}/detach", h.DetachVolume)
	r.Delete("/volumes/{id}", h.DeleteVolume)
}

// CreateVolumeRequest represents a request to create a volume
type CreateVolumeRequest struct {
	Name      string `json:"name"`
	SizeMB    int    `json:"size_mb"`
	MountPath string `json:"mount_path,omitempty"` // Optional: e.g., /var/lib/postgresql/data
}

// CreateVolume creates a new volume
func (h *VolumeHandler) CreateVolume(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Verify project belongs to user's organization
	project, err := h.store.GetProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Parse request
	var req CreateVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.SizeMB <= 0 {
		http.Error(w, "Size must be greater than 0", http.StatusBadRequest)
		return
	}

	// Create volume
	volume := &store.Volume{
		ProjectID:  projectID,
		Name:       req.Name,
		SizeMB:     req.SizeMB,
		Status:     "pending",
		VolumeType: "user",
	}

	if req.MountPath != "" {
		volume.MountPath = sql.NullString{String: req.MountPath, Valid: true}
	}

	if err := h.store.CreateVolume(r.Context(), volume); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Queue create_volume job

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(volume)
}

// ListVolumes lists volumes for a project
func (h *VolumeHandler) ListVolumes(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Verify project belongs to user's organization
	project, err := h.store.GetProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	volumes, err := h.store.ListVolumesByProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(volumes)
}

// GetVolume retrieves a volume by ID
func (h *VolumeHandler) GetVolume(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	volumeIDStr := chi.URLParam(r, "id")
	volumeID, err := uuid.Parse(volumeIDStr)
	if err != nil {
		http.Error(w, "Invalid volume ID", http.StatusBadRequest)
		return
	}

	volume, err := h.store.GetVolume(r.Context(), volumeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if volume == nil {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	// Verify volume belongs to user's organization
	project, err := h.store.GetProject(r.Context(), volume.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(volume)
}

// AttachVolumeRequest represents a request to attach a volume
type AttachVolumeRequest struct {
	ServiceID uuid.UUID `json:"service_id"`
	MountPath string    `json:"mount_path"` // e.g., /var/lib/postgresql/data
}

// AttachVolume attaches a volume to a service
func (h *VolumeHandler) AttachVolume(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	volumeIDStr := chi.URLParam(r, "id")
	volumeID, err := uuid.Parse(volumeIDStr)
	if err != nil {
		http.Error(w, "Invalid volume ID", http.StatusBadRequest)
		return
	}

	// Verify volume belongs to user's organization
	volume, err := h.store.GetVolume(r.Context(), volumeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if volume == nil {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	project, err := h.store.GetProject(r.Context(), volume.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	// Parse request
	var req AttachVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify service belongs to the same project
	service, err := h.store.GetService(r.Context(), req.ServiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if service == nil || service.ProjectID != volume.ProjectID {
		http.Error(w, "Service not found or doesn't belong to project", http.StatusBadRequest)
		return
	}

	// TODO: Queue attach_volume job

	if err := h.store.AttachVolumeToService(r.Context(), volumeID, req.ServiceID, req.MountPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DetachVolume detaches a volume from a service
func (h *VolumeHandler) DetachVolume(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	volumeIDStr := chi.URLParam(r, "id")
	volumeID, err := uuid.Parse(volumeIDStr)
	if err != nil {
		http.Error(w, "Invalid volume ID", http.StatusBadRequest)
		return
	}

	// Verify volume belongs to user's organization
	volume, err := h.store.GetVolume(r.Context(), volumeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if volume == nil {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	project, err := h.store.GetProject(r.Context(), volume.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	// TODO: Queue detach_volume job

	if err := h.store.DetachVolumeFromService(r.Context(), volumeID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteVolume deletes a volume
func (h *VolumeHandler) DeleteVolume(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	volumeIDStr := chi.URLParam(r, "id")
	volumeID, err := uuid.Parse(volumeIDStr)
	if err != nil {
		http.Error(w, "Invalid volume ID", http.StatusBadRequest)
		return
	}

	// Verify volume belongs to user's organization
	volume, err := h.store.GetVolume(r.Context(), volumeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if volume == nil {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	project, err := h.store.GetProject(r.Context(), volume.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Volume not found", http.StatusNotFound)
		return
	}

	// Check if volume is attached
	if volume.Status == "attached" {
		http.Error(w, "Cannot delete attached volume. Detach it first.", http.StatusBadRequest)
		return
	}

	// TODO: Queue destroy_volume job

	if err := h.store.DeleteVolume(r.Context(), volumeID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Volume not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

