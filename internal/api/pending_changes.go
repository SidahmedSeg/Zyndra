package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
)

// PendingChangesHandler handles pending changes endpoints
type PendingChangesHandler struct {
	store  *store.DB
	config *config.Config
}

// NewPendingChangesHandler creates a new pending changes handler
func NewPendingChangesHandler(store *store.DB, cfg *config.Config) *PendingChangesHandler {
	return &PendingChangesHandler{
		store:  store,
		config: cfg,
	}
}

// PendingChangesResponse is the response for pending changes
type PendingChangesResponse struct {
	Count   int                    `json:"count"`
	Commits []*store.PendingCommit `json:"commits"`
}

// ListPendingChanges returns pending changes for a service
func (h *PendingChangesHandler) ListPendingChanges(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	commits, err := h.store.ListUnacknowledgedCommits(r.Context(), serviceID)
	if err != nil {
		http.Error(w, "Failed to get pending changes", http.StatusInternalServerError)
		return
	}

	if commits == nil {
		commits = []*store.PendingCommit{}
	}

	response := PendingChangesResponse{
		Count:   len(commits),
		Commits: commits,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AcknowledgeRequest is the request body for acknowledging changes
type AcknowledgeRequest struct {
	CommitIDs []string `json:"commit_ids"` // Optional: if empty, acknowledge all
}

// AcknowledgePendingChanges marks pending changes as acknowledged
func (h *PendingChangesHandler) AcknowledgePendingChanges(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	var req AcknowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body, acknowledge all
		req = AcknowledgeRequest{}
	}

	if len(req.CommitIDs) == 0 {
		// Acknowledge all
		if err := h.store.AcknowledgeAllPendingCommits(r.Context(), serviceID); err != nil {
			http.Error(w, "Failed to acknowledge changes", http.StatusInternalServerError)
			return
		}
	} else {
		// Acknowledge specific commits
		commitIDs := make([]uuid.UUID, 0, len(req.CommitIDs))
		for _, idStr := range req.CommitIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			commitIDs = append(commitIDs, id)
		}
		if err := h.store.AcknowledgePendingCommits(r.Context(), serviceID, commitIDs); err != nil {
			http.Error(w, "Failed to acknowledge changes", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Changes acknowledged"})
}

// DeployPendingRequest is the request for deploying pending changes
type DeployPendingRequest struct {
	UpToCommitSHA string `json:"up_to_commit_sha"` // Deploy all changes up to this commit
}

// DeployPendingChanges triggers a deployment for pending changes
func (h *PendingChangesHandler) DeployPendingChanges(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	var req DeployPendingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the latest pending commit if not specified
	if req.UpToCommitSHA == "" {
		commits, err := h.store.ListUnacknowledgedCommits(r.Context(), serviceID)
		if err != nil || len(commits) == 0 {
			http.Error(w, "No pending changes to deploy", http.StatusBadRequest)
			return
		}
		// Get the latest (first in list, sorted by pushed_at DESC in other queries)
		req.UpToCommitSHA = commits[len(commits)-1].CommitSHA // Get last one (oldest unacknowledged)
	}

	// Get service
	service, err := h.store.GetService(r.Context(), serviceID)
	if err != nil || service == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Create deployment
	deployment := &store.Deployment{
		ServiceID:    serviceID,
		Status:       "queued",
		TriggeredBy:  "pending_changes",
	}
	
	// Set commit info
	deployment.CommitSHA.Valid = true
	deployment.CommitSHA.String = req.UpToCommitSHA

	if err := h.store.CreateDeployment(r.Context(), deployment); err != nil {
		http.Error(w, "Failed to create deployment", http.StatusInternalServerError)
		return
	}

	// Mark commits as deployed
	if err := h.store.MarkCommitsAsDeployed(r.Context(), serviceID, req.UpToCommitSHA); err != nil {
		// Log but don't fail - deployment was created
	}

	// Create job
	job := &store.Job{
		Type:    "build",
		Payload: map[string]interface{}{"deployment_id": deployment.ID.String()},
		Status:  "pending",
	}
	if err := h.store.CreateJob(r.Context(), job); err != nil {
		http.Error(w, "Failed to queue deployment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

// GetPendingChangesCount returns just the count
func (h *PendingChangesHandler) GetPendingChangesCount(w http.ResponseWriter, r *http.Request) {
	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	count, err := h.store.GetPendingChangesCount(r.Context(), serviceID)
	if err != nil {
		http.Error(w, "Failed to get pending count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// RegisterPendingChangesRoutes registers pending changes routes
func RegisterPendingChangesRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	handler := NewPendingChangesHandler(db, cfg)

	r.Get("/services/{id}/pending-changes", handler.ListPendingChanges)
	r.Get("/services/{id}/pending-changes/count", handler.GetPendingChangesCount)
	r.Post("/services/{id}/pending-changes/acknowledge", handler.AcknowledgePendingChanges)
	r.Post("/services/{id}/pending-changes/deploy", handler.DeployPendingChanges)
}

