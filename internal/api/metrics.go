package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/intelifox/click-deploy/internal/auth"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/domain"
	"github.com/intelifox/click-deploy/internal/store"
)

type MetricsHandler struct {
	store  *store.DB
	config *config.Config
	client v1.API
}

func NewMetricsHandler(store *store.DB, cfg *config.Config) (*MetricsHandler, error) {
	// Create Prometheus API client
	promClient, err := api.NewClient(api.Config{
		Address: cfg.PrometheusURL,
	})
	if err != nil {
		return nil, err
	}

	v1API := v1.NewAPI(promClient)

	return &MetricsHandler{
		store:  store,
		config: cfg,
		client: v1API,
	}, nil
}

// RegisterMetricsRoutes registers metrics-related routes
func RegisterMetricsRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h, err := NewMetricsHandler(db, cfg)
	if err != nil {
		// If Prometheus is not configured, metrics endpoints will return errors
		// This is acceptable for development
		return
	}

	r.Get("/services/{id}/metrics", h.GetServiceMetrics)
	r.Get("/databases/{id}/metrics", h.GetDatabaseMetrics)
	r.Get("/volumes/{id}/metrics", h.GetVolumeMetrics)
}

// MetricsResponse represents a metrics response
type MetricsResponse struct {
	CPU          []DataPoint `json:"cpu"`
	Memory       []DataPoint `json:"memory"`
	NetworkIn    []DataPoint `json:"network_in"`
	NetworkOut   []DataPoint `json:"network_out"`
	RequestCount []DataPoint `json:"request_count"`
	ResponseTime []DataPoint `json:"response_time"`
	ErrorRate    []DataPoint `json:"error_rate"`
}

// DataPoint represents a single metric data point
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// GetServiceMetrics handles GET /services/:id/metrics
func (h *MetricsHandler) GetServiceMetrics(w http.ResponseWriter, r *http.Request) {
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
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Service"))
		return
	}

	// Get time range from query params (default: last 1 hour)
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	step := r.URL.Query().Get("step")
	if step == "" {
		step = "30s" // Default step: 30 seconds
	}

	var startTime, endTime time.Time
	if start != "" && end != "" {
		startTime, _ = time.Parse(time.RFC3339, start)
		endTime, _ = time.Parse(time.RFC3339, end)
	} else {
		// Default: last 1 hour
		endTime = time.Now()
		startTime = endTime.Add(-1 * time.Hour)
	}

	ctx := r.Context()
	timeRange := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  parseDuration(step),
	}

	// Query Prometheus for service metrics
	metrics := MetricsResponse{}

	// CPU Usage
	cpuQuery := `click_deploy_service_cpu_usage_percent{service_id="` + serviceID.String() + `"}`
	cpuResult, _, err := h.client.QueryRange(ctx, cpuQuery, timeRange)
	if err == nil {
		metrics.CPU = parsePrometheusResult(cpuResult)
	}

	// Memory Usage
	memoryQuery := `click_deploy_service_memory_usage_bytes{service_id="` + serviceID.String() + `"}`
	memoryResult, _, err := h.client.QueryRange(ctx, memoryQuery, timeRange)
	if err == nil {
		metrics.Memory = parsePrometheusResult(memoryResult)
	}

	// Network Traffic In
	networkInQuery := `rate(click_deploy_service_network_traffic_in_bytes_total{service_id="` + serviceID.String() + `"}[5m])`
	networkInResult, _, err := h.client.QueryRange(ctx, networkInQuery, timeRange)
	if err == nil {
		metrics.NetworkIn = parsePrometheusResult(networkInResult)
	}

	// Network Traffic Out
	networkOutQuery := `rate(click_deploy_service_network_traffic_out_bytes_total{service_id="` + serviceID.String() + `"}[5m])`
	networkOutResult, _, err := h.client.QueryRange(ctx, networkOutQuery, timeRange)
	if err == nil {
		metrics.NetworkOut = parsePrometheusResult(networkOutResult)
	}

	// Request Count
	requestCountQuery := `rate(click_deploy_service_requests_total{service_id="` + serviceID.String() + `"}[5m])`
	requestCountResult, _, err := h.client.QueryRange(ctx, requestCountQuery, timeRange)
	if err == nil {
		metrics.RequestCount = parsePrometheusResult(requestCountResult)
	}

	// Response Time
	responseTimeQuery := `click_deploy_service_request_duration_seconds{service_id="` + serviceID.String() + `"}`
	responseTimeResult, _, err := h.client.QueryRange(ctx, responseTimeQuery, timeRange)
	if err == nil {
		metrics.ResponseTime = parsePrometheusResult(responseTimeResult)
	}

	// Error Rate
	errorRateQuery := `click_deploy_service_error_rate{service_id="` + serviceID.String() + `"}`
	errorRateResult, _, err := h.client.QueryRange(ctx, errorRateQuery, timeRange)
	if err == nil {
		metrics.ErrorRate = parsePrometheusResult(errorRateResult)
	}

	WriteJSON(w, http.StatusOK, metrics)
}

// GetDatabaseMetrics handles GET /databases/:id/metrics
func (h *MetricsHandler) GetDatabaseMetrics(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	databaseIDStr := chi.URLParam(r, "id")
	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid database ID"))
		return
	}

	// Verify database belongs to user's organization
	database, err := h.store.GetDatabase(r.Context(), databaseID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if database == nil {
		WriteError(w, domain.NewNotFoundError("Database"))
		return
	}

	// Get service to verify project
	var serviceID uuid.UUID
	if database.ServiceID.Valid {
		serviceID, _ = uuid.Parse(database.ServiceID.String)
		service, err := h.store.GetService(r.Context(), serviceID)
		if err == nil && service != nil {
			project, err := h.store.GetProject(r.Context(), service.ProjectID)
			if err == nil && project != nil && project.CasdoorOrgID != orgID {
				WriteError(w, domain.NewNotFoundError("Database"))
				return
			}
		}
	}

	// Get time range
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	step := r.URL.Query().Get("step")
	if step == "" {
		step = "30s"
	}

	var startTime, endTime time.Time
	if start != "" && end != "" {
		startTime, _ = time.Parse(time.RFC3339, start)
		endTime, _ = time.Parse(time.RFC3339, end)
	} else {
		endTime = time.Now()
		startTime = endTime.Add(-1 * time.Hour)
	}

	ctx := r.Context()
	timeRange := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  parseDuration(step),
	}

	metrics := MetricsResponse{}

	// Query Prometheus for database metrics
	cpuQuery := `click_deploy_database_cpu_usage_percent{database_id="` + databaseID.String() + `"}`
	cpuResult, _, err := h.client.QueryRange(ctx, cpuQuery, timeRange)
	if err == nil {
		metrics.CPU = parsePrometheusResult(cpuResult)
	}

	memoryQuery := `click_deploy_database_memory_usage_bytes{database_id="` + databaseID.String() + `"}`
	memoryResult, _, err := h.client.QueryRange(ctx, memoryQuery, timeRange)
	if err == nil {
		metrics.Memory = parsePrometheusResult(memoryResult)
	}

	networkInQuery := `rate(click_deploy_database_network_traffic_in_bytes_total{database_id="` + databaseID.String() + `"}[5m])`
	networkInResult, _, err := h.client.QueryRange(ctx, networkInQuery, timeRange)
	if err == nil {
		metrics.NetworkIn = parsePrometheusResult(networkInResult)
	}

	networkOutQuery := `rate(click_deploy_database_network_traffic_out_bytes_total{database_id="` + databaseID.String() + `"}[5m])`
	networkOutResult, _, err := h.client.QueryRange(ctx, networkOutQuery, timeRange)
	if err == nil {
		metrics.NetworkOut = parsePrometheusResult(networkOutResult)
	}

	WriteJSON(w, http.StatusOK, metrics)
}

// GetVolumeMetrics handles GET /volumes/:id/metrics
func (h *MetricsHandler) GetVolumeMetrics(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		WriteError(w, domain.ErrUnauthorized.WithDetails("Organization ID not found in token"))
		return
	}

	volumeIDStr := chi.URLParam(r, "id")
	volumeID, err := uuid.Parse(volumeIDStr)
	if err != nil {
		WriteError(w, domain.NewInvalidInputError("Invalid volume ID"))
		return
	}

	// Verify volume belongs to user's organization
	volume, err := h.store.GetVolume(r.Context(), volumeID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if volume == nil {
		WriteError(w, domain.NewNotFoundError("Volume"))
		return
	}

	project, err := h.store.GetProject(r.Context(), volume.ProjectID)
	if err != nil {
		WriteError(w, domain.ErrDatabase.WithError(err))
		return
	}
	if project == nil || project.CasdoorOrgID != orgID {
		WriteError(w, domain.NewNotFoundError("Volume"))
		return
	}

	// Get time range
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	step := r.URL.Query().Get("step")
	if step == "" {
		step = "30s"
	}

	var startTime, endTime time.Time
	if start != "" && end != "" {
		startTime, _ = time.Parse(time.RFC3339, start)
		endTime, _ = time.Parse(time.RFC3339, end)
	} else {
		endTime = time.Now()
		startTime = endTime.Add(-1 * time.Hour)
	}

	ctx := r.Context()
	timeRange := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  parseDuration(step),
	}

	metrics := MetricsResponse{}

	// Query Prometheus for volume metrics
	usageQuery := `click_deploy_volume_usage_bytes{volume_id="` + volumeID.String() + `"}`
	usageResult, _, err := h.client.QueryRange(ctx, usageQuery, timeRange)
	if err == nil {
		metrics.Memory = parsePrometheusResult(usageResult) // Using Memory field for volume usage
	}

	readQuery := `rate(click_deploy_volume_io_read_bytes_total{volume_id="` + volumeID.String() + `"}[5m])`
	readResult, _, err := h.client.QueryRange(ctx, readQuery, timeRange)
	if err == nil {
		metrics.NetworkIn = parsePrometheusResult(readResult) // Using NetworkIn for read
	}

	writeQuery := `rate(click_deploy_volume_io_write_bytes_total{volume_id="` + volumeID.String() + `"}[5m])`
	writeResult, _, err := h.client.QueryRange(ctx, writeQuery, timeRange)
	if err == nil {
		metrics.NetworkOut = parsePrometheusResult(writeResult) // Using NetworkOut for write
	}

	WriteJSON(w, http.StatusOK, metrics)
}

// parsePrometheusResult converts Prometheus query result to DataPoint array
func parsePrometheusResult(result model.Value) []DataPoint {
	var points []DataPoint

	switch result.Type() {
	case model.ValMatrix:
		matrix := result.(model.Matrix)
		for _, series := range matrix {
			for _, sample := range series.Values {
				points = append(points, DataPoint{
					Timestamp: sample.Timestamp.Time(),
					Value:     float64(sample.Value),
				})
			}
		}
	case model.ValVector:
		vector := result.(model.Vector)
		for _, sample := range vector {
			points = append(points, DataPoint{
				Timestamp: sample.Timestamp.Time(),
				Value:     float64(sample.Value),
			})
		}
	}

	return points
}

// parseDuration parses a duration string (e.g., "30s", "1m", "5m")
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second // Default
	}
	return d
}

