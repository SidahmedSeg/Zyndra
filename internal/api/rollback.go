package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
)

type RollbackHandler struct {
	store  *store.DB
	config *config.Config
}

func NewRollbackHandler(store *store.DB, cfg *config.Config) *RollbackHandler {
	return &RollbackHandler{
		store:  store,
		config: cfg,
	}
}

// RegisterRollbackRoutes registers rollback-related routes
func RegisterRollbackRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewRollbackHandler(db, cfg)

	r.Post("/services/{id}/rollback/{deploymentId}", h.RollbackDeployment)
	r.Get("/services/{id}/rollback-candidates", h.GetRollbackCandidates)
}

// RollbackDeployment rolls back a service to a previous deployment
func (h *RollbackHandler) RollbackDeployment(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	deploymentIDStr := chi.URLParam(r, "deploymentId")
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if service == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Get the deployment to rollback to
	targetDeployment, err := h.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if targetDeployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Verify deployment belongs to the service
	if targetDeployment.ServiceID != serviceID {
		http.Error(w, "Deployment does not belong to this service", http.StatusBadRequest)
		return
	}

	// Verify deployment was successful and has an image tag
	if targetDeployment.Status != "success" {
		http.Error(w, "Can only rollback to successful deployments", http.StatusBadRequest)
		return
	}

	if !targetDeployment.ImageTag.Valid {
		http.Error(w, "Deployment does not have an image tag", http.StatusBadRequest)
		return
	}

	// Create a new deployment record for the rollback
	rollbackDeployment := &store.Deployment{
		ServiceID:     serviceID,
		CommitSHA:     targetDeployment.CommitSHA,
		CommitMessage: sql.NullString{String: "Rollback to " + targetDeployment.ID.String()[:8], Valid: true},
		CommitAuthor:  sql.NullString{String: "System", Valid: true},
		Status:        "queued",
		ImageTag:      targetDeployment.ImageTag,
		TriggeredBy:   "rollback",
		StartedAt:     sql.NullTime{Time: time.Now(), Valid: true},
	}

	if err := h.store.CreateDeployment(r.Context(), rollbackDeployment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a rollback job
	job := &store.Job{
		Type:    "rollback",
		Payload: map[string]interface{}{
			"deployment_id":            rollbackDeployment.ID.String(),
			"target_image_tag":         targetDeployment.ImageTag.String,
			"rollback_to_deployment_id": targetDeployment.ID.String(),
		},
		Status:     "queued",
		Attempts:   0,
		MaxAttempts: 3,
	}

	if err := h.store.CreateJob(r.Context(), job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rollbackDeployment)
}

// GetRollbackCandidates returns successful deployments that can be rolled back to
func (h *RollbackHandler) GetRollbackCandidates(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if service == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Get successful deployments (rollback candidates)
	candidates, err := h.store.GetSuccessfulDeploymentsByService(r.Context(), serviceID, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(candidates)
}

