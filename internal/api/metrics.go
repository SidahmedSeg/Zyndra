package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/k8s"
	"github.com/intelifox/click-deploy/internal/store"
)

// MetricsHandler handles metrics endpoints
type MetricsHandler struct {
	store         *store.DB
	config        *config.Config
	metricsClient *k8s.MetricsClient
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(store *store.DB, cfg *config.Config, metricsClient *k8s.MetricsClient) *MetricsHandler {
	return &MetricsHandler{
		store:         store,
		config:        cfg,
		metricsClient: metricsClient,
	}
}

// GetServiceMetrics returns live metrics for a service
func (h *MetricsHandler) GetServiceMetrics(w http.ResponseWriter, r *http.Request) {
	if h.metricsClient == nil {
		// Return mock metrics if not using k8s
		h.returnMockMetrics(w)
		return
	}

	serviceIDStr := chi.URLParam(r, "id")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	// Get service and project
	service, err := h.store.GetService(r.Context(), serviceID)
	if err != nil || service == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	metrics, err := h.metricsClient.GetServiceMetrics(
		r.Context(),
		service.ProjectID.String(),
		serviceID.String(),
		service.Name,
	)
	if err != nil {
		// If metrics server is not ready, return empty data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available": false,
			"error":     err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": true,
		"metrics":   metrics,
	})
}

// GetProjectMetrics returns metrics for all services in a project
func (h *MetricsHandler) GetProjectMetrics(w http.ResponseWriter, r *http.Request) {
	if h.metricsClient == nil {
		h.returnMockMetrics(w)
		return
	}

	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	metrics, err := h.metricsClient.GetNamespaceMetrics(r.Context(), projectID.String())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available": false,
			"error":     err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": true,
		"pods":      metrics,
	})
}

// GetClusterMetrics returns cluster-wide metrics (admin only)
func (h *MetricsHandler) GetClusterMetrics(w http.ResponseWriter, r *http.Request) {
	if h.metricsClient == nil {
		h.returnMockMetrics(w)
		return
	}

	nodes, err := h.metricsClient.GetNodeMetrics(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available": false,
			"error":     err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": true,
		"nodes":     nodes,
	})
}

// CheckMetricsAvailability checks if metrics server is available
func (h *MetricsHandler) CheckMetricsAvailability(w http.ResponseWriter, r *http.Request) {
	if h.metricsClient == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"available": false})
		return
	}

	available := h.metricsClient.IsMetricsServerAvailable(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"available": available})
}

// returnMockMetrics returns mock metrics for development
func (h *MetricsHandler) returnMockMetrics(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": true,
		"is_mock":   true,
		"metrics": k8s.ServiceMetrics{
			ServiceID:   "mock-service",
			ServiceName: "mock",
			TotalCPU:    0.25,
			TotalMemory: 256,
			PodCount:    1,
			Pods: []k8s.PodMetrics{
				{
					Name:        "mock-pod-1",
					Namespace:   "mock-namespace",
					CPUCores:    0.25,
					MemoryMB:    256,
					CPUUsage:    "250m",
					MemoryUsage: "256Mi",
				},
			},
		},
	})
}

// RegisterMetricsRoutes registers metrics routes
func RegisterMetricsRoutes(r chi.Router, db *store.DB, cfg *config.Config, metricsClient *k8s.MetricsClient) {
	handler := NewMetricsHandler(db, cfg, metricsClient)

	r.Get("/services/{id}/metrics", handler.GetServiceMetrics)
	r.Get("/projects/{id}/metrics", handler.GetProjectMetrics)
	r.Get("/cluster/metrics", handler.GetClusterMetrics)
	r.Get("/metrics/available", handler.CheckMetricsAvailability)
}
