package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/k8s"
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/worker"
)

type DeploymentHandler struct {
	store         *store.DB
	config        *config.Config
	buildWorker   *worker.BuildWorker
	k8sWorker     *worker.K8sDeployWorker
}

func NewDeploymentHandler(store *store.DB, cfg *config.Config, buildWorker *worker.BuildWorker, k8sClient *k8s.Client) *DeploymentHandler {
	var k8sWorker *worker.K8sDeployWorker
	if k8sClient != nil {
		k8sWorker = worker.NewK8sDeployWorker(store, k8sClient)
	}
	
	return &DeploymentHandler{
		store:       store,
		config:      cfg,
		buildWorker: buildWorker,
		k8sWorker:   k8sWorker,
	}
}

// RegisterDeploymentRoutes registers deployment-related routes
func RegisterDeploymentRoutes(r chi.Router, db *store.DB, cfg *config.Config, buildWorker *worker.BuildWorker, k8sClient *k8s.Client) {
	h := NewDeploymentHandler(db, cfg, buildWorker, k8sClient)

	r.Post("/services/{id}/deploy", h.TriggerDeployment)
	r.Get("/deployments/{id}", h.GetDeployment)
	r.Get("/deployments/{id}/logs", h.GetDeploymentLogs)
	r.Post("/deployments/{id}/cancel", h.CancelDeployment)
	r.Get("/services/{id}/deployments", h.ListServiceDeployments)
}

// TriggerDeploymentRequest represents a request to trigger a deployment
type TriggerDeploymentRequest struct {
	CommitSHA string `json:"commit_sha,omitempty"` // Optional: deploy specific commit
	Branch    string `json:"branch,omitempty"`     // Optional: deploy specific branch
}

// TriggerDeployment triggers a new deployment for a service
func (h *DeploymentHandler) TriggerDeployment(w http.ResponseWriter, r *http.Request) {
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

	// Verify project belongs to organization
	project, err := h.store.GetProject(r.Context(), service.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if project == nil || !project.BelongsToOrg(orgID) {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Parse request
	var req TriggerDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get git source
	gitSource, err := h.store.GetGitSourceByService(r.Context(), serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if gitSource == nil {
		http.Error(w, "Git source not found for service", http.StatusBadRequest)
		return
	}

	// Create deployment
	deployment := &store.Deployment{
		ServiceID:   serviceID,
		Status:      "queued",
		TriggeredBy: "manual",
	}

	if req.CommitSHA != "" {
		deployment.CommitSHA = sql.NullString{String: req.CommitSHA, Valid: true}
	}

	// TODO: Get commit message and author from git if commit_sha provided

	if err := h.store.CreateDeployment(r.Context(), deployment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Queue build job asynchronously
	if h.buildWorker != nil && h.k8sWorker != nil {
		go func() {
			ctx := context.Background()
			
			// Run build
			if err := h.buildWorker.ProcessBuildJob(ctx, deployment.ID); err != nil {
				h.store.UpdateDeploymentStatus(ctx, deployment.ID, "failed")
				return
			}
			
			// Deploy to k8s after successful build
			if err := h.k8sWorker.DeployToK8s(ctx, deployment.ID); err != nil {
				h.store.UpdateDeploymentStatus(ctx, deployment.ID, "failed")
				return
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

// GetDeployment retrieves a deployment by ID
func (h *DeploymentHandler) GetDeployment(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentIDStr := chi.URLParam(r, "id")
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	deployment, err := h.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Verify service belongs to user's organization
	service, err := h.store.GetService(r.Context(), deployment.ServiceID)
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
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

// GetDeploymentLogs retrieves logs for a deployment
func (h *DeploymentHandler) GetDeploymentLogs(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentIDStr := chi.URLParam(r, "id")
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	// Verify deployment belongs to user's organization
	deployment, err := h.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	service, err := h.store.GetService(r.Context(), deployment.ServiceID)
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
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Get limit from query parameter
	limit := 1000 // Default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := h.store.GetDeploymentLogs(r.Context(), deploymentID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// CancelDeployment cancels a running deployment
func (h *DeploymentHandler) CancelDeployment(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentIDStr := chi.URLParam(r, "id")
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	// Verify deployment belongs to user's organization
	deployment, err := h.store.GetDeployment(r.Context(), deploymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	service, err := h.store.GetService(r.Context(), deployment.ServiceID)
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
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Check if deployment can be cancelled
	if deployment.Status != "queued" && deployment.Status != "building" && deployment.Status != "pushing" {
		http.Error(w, "Deployment cannot be cancelled", http.StatusBadRequest)
		return
	}

	// Update status
	if err := h.store.UpdateDeploymentStatus(r.Context(), deploymentID, "cancelled"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add log entry
	h.store.AddDeploymentLog(r.Context(), deploymentID, "deploy", "info", "Deployment cancelled by user", nil)

	// TODO: Actually cancel the build process (context cancellation)

	w.WriteHeader(http.StatusNoContent)
}

// ListServiceDeployments lists deployments for a service
func (h *DeploymentHandler) ListServiceDeployments(w http.ResponseWriter, r *http.Request) {
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

	// Get limit from query parameter
	limit := 50 // Default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	deployments, err := h.store.ListDeploymentsByService(r.Context(), serviceID, limit, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

