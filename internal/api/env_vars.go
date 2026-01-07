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

type EnvVarHandler struct {
	store  *store.DB
	config *config.Config
}

func NewEnvVarHandler(store *store.DB, cfg *config.Config) *EnvVarHandler {
	return &EnvVarHandler{
		store:  store,
		config: cfg,
	}
}

// RegisterEnvVarRoutes registers environment variable routes
func RegisterEnvVarRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewEnvVarHandler(db, cfg)

	r.Get("/services/{id}/env", h.ListEnvVars)
	r.Post("/services/{id}/env", h.CreateEnvVar)
	r.Patch("/services/{id}/env/{key}", h.UpdateEnvVar)
	r.Delete("/services/{id}/env/{key}", h.DeleteEnvVar)
}

// CreateEnvVarRequest represents a request to create an environment variable
type CreateEnvVarRequest struct {
	Key              string    `json:"key"`
	Value            string    `json:"value,omitempty"`            // Optional if linked to database
	IsSecret         bool      `json:"is_secret,omitempty"`
	LinkedDatabaseID uuid.UUID `json:"linked_database_id,omitempty"` // Optional
	LinkType         string    `json:"link_type,omitempty"`          // connection_url, host, port, username, password, database
}

// CreateEnvVar creates a new environment variable
func (h *EnvVarHandler) CreateEnvVar(w http.ResponseWriter, r *http.Request) {
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
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Parse request
	var req CreateEnvVarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validation
	if req.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	// If linked to database, verify database exists and belongs to same project
	var linkedDatabaseID sql.NullString
	var linkType sql.NullString
	if req.LinkedDatabaseID != uuid.Nil {
		database, err := h.store.GetDatabase(r.Context(), req.LinkedDatabaseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if database == nil {
			http.Error(w, "Database not found", http.StatusBadRequest)
			return
		}

		// Verify database belongs to same service or project
		if database.ServiceID.Valid {
			dbServiceID, _ := uuid.Parse(database.ServiceID.String)
			if dbServiceID != serviceID {
				http.Error(w, "Database does not belong to this service", http.StatusBadRequest)
				return
			}
		}

		if req.LinkType == "" {
			http.Error(w, "Link type is required when linking to database", http.StatusBadRequest)
			return
		}

		validLinkTypes := map[string]bool{
			"connection_url": true,
			"host":           true,
			"port":           true,
			"username":       true,
			"password":       true,
			"database":       true,
		}
		if !validLinkTypes[req.LinkType] {
			http.Error(w, "Invalid link type", http.StatusBadRequest)
			return
		}

		linkedDatabaseID = sql.NullString{String: req.LinkedDatabaseID.String(), Valid: true}
		linkType = sql.NullString{String: req.LinkType, Valid: true}
	} else if req.Value == "" {
		http.Error(w, "Value is required if not linking to database", http.StatusBadRequest)
		return
	}

	// Create environment variable
	envVar := &store.EnvVar{
		ServiceID:       serviceID,
		Key:             req.Key,
		IsSecret:        req.IsSecret,
		LinkedDatabaseID: linkedDatabaseID,
		LinkType:        linkType,
	}

	if req.Value != "" {
		envVar.Value = sql.NullString{String: req.Value, Valid: true}
	}

	if err := h.store.CreateEnvVar(r.Context(), envVar); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(envVar)
}

// ListEnvVars lists environment variables for a service
func (h *EnvVarHandler) ListEnvVars(w http.ResponseWriter, r *http.Request) {
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
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	envVars, err := h.store.ListEnvVarsByService(r.Context(), serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Don't expose secret values
	for _, ev := range envVars {
		if ev.IsSecret && ev.Value.Valid {
			ev.Value = sql.NullString{String: "***", Valid: true}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVars)
}

// UpdateEnvVar updates an environment variable
func (h *EnvVarHandler) UpdateEnvVar(w http.ResponseWriter, r *http.Request) {
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

	key := chi.URLParam(r, "key")

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
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Get existing env var
	envVars, err := h.store.ListEnvVarsByService(r.Context(), serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var envVar *store.EnvVar
	for _, ev := range envVars {
		if ev.Key == key {
			envVar = ev
			break
		}
	}

	if envVar == nil {
		http.Error(w, "Environment variable not found", http.StatusNotFound)
		return
	}

	// Parse request
	var req CreateEnvVarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update values
	if req.Value != "" {
		envVar.Value = sql.NullString{String: req.Value, Valid: true}
	}
	if req.IsSecret {
		envVar.IsSecret = req.IsSecret
	}
	if req.LinkedDatabaseID != uuid.Nil {
		envVar.LinkedDatabaseID = sql.NullString{String: req.LinkedDatabaseID.String(), Valid: true}
		if req.LinkType != "" {
			envVar.LinkType = sql.NullString{String: req.LinkType, Valid: true}
		}
	}

	if err := h.store.UpdateEnvVar(r.Context(), envVar.ID, envVar); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVar)
}

// DeleteEnvVar deletes an environment variable
func (h *EnvVarHandler) DeleteEnvVar(w http.ResponseWriter, r *http.Request) {
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

	key := chi.URLParam(r, "key")

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
	if project == nil || project.CasdoorOrgID != orgID {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Get existing env var
	envVars, err := h.store.ListEnvVarsByService(r.Context(), serviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var envVar *store.EnvVar
	for _, ev := range envVars {
		if ev.Key == key {
			envVar = ev
			break
		}
	}

	if envVar == nil {
		http.Error(w, "Environment variable not found", http.StatusNotFound)
		return
	}

	if err := h.store.DeleteEnvVar(r.Context(), envVar.ID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Environment variable not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

