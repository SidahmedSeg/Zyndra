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

type DatabaseHandler struct {
	store  *store.DB
	config *config.Config
}

func NewDatabaseHandler(store *store.DB, cfg *config.Config) *DatabaseHandler {
	return &DatabaseHandler{
		store:  store,
		config: cfg,
	}
}

// RegisterDatabaseRoutes registers database-related routes
func RegisterDatabaseRoutes(r chi.Router, db *store.DB, cfg *config.Config) {
	h := NewDatabaseHandler(db, cfg)

	r.Get("/projects/{id}/databases", h.ListDatabases)
	r.Post("/projects/{id}/databases", h.CreateDatabase)
	r.Get("/databases/{id}", h.GetDatabase)
	r.Get("/databases/{id}/credentials", h.GetDatabaseCredentials)
	r.Delete("/databases/{id}", h.DeleteDatabase)
}

// CreateDatabaseRequest represents a request to create a database
type CreateDatabaseRequest struct {
	ServiceID uuid.UUID `json:"service_id,omitempty"` // Optional: link to service
	Engine    string    `json:"engine"`                // postgresql, mysql, redis
	Version   string    `json:"version,omitempty"`    // Optional: e.g., "14", "8.0"
	Size      string    `json:"size,omitempty"`        // small, medium, large (default: small)
	VolumeSizeMB int    `json:"volume_size_mb,omitempty"` // Default: 500
}

// CreateDatabase creates a new database
func (h *DatabaseHandler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
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
	if project == nil || !project.BelongsToOrg(orgID) {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Parse request
	var req CreateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate engine
	if req.Engine != "postgresql" && req.Engine != "mysql" && req.Engine != "redis" {
		http.Error(w, "Invalid engine. Must be postgresql, mysql, or redis", http.StatusBadRequest)
		return
	}

	// Validate size
	if req.Size == "" {
		req.Size = "small"
	}
	if req.Size != "small" && req.Size != "medium" && req.Size != "large" {
		http.Error(w, "Invalid size. Must be small, medium, or large", http.StatusBadRequest)
		return
	}

	// Set default volume size
	if req.VolumeSizeMB == 0 {
		req.VolumeSizeMB = 500
	}

	// If service_id provided, verify it belongs to the project
	var serviceID sql.NullString
	if req.ServiceID != uuid.Nil {
		service, err := h.store.GetService(r.Context(), req.ServiceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if service == nil || service.ProjectID != projectID {
			http.Error(w, "Service not found or doesn't belong to project", http.StatusBadRequest)
			return
		}
		serviceID = sql.NullString{String: req.ServiceID.String(), Valid: true}
	}

	// Create database
	database := &store.Database{
		ServiceID:    serviceID,
		Engine:       req.Engine,
		Size:         req.Size,
		VolumeSizeMB: req.VolumeSizeMB,
		Status:       "pending",
	}

	if req.Version != "" {
		database.Version = sql.NullString{String: req.Version, Valid: true}
	}

	if err := h.store.CreateDatabase(r.Context(), database); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Queue provision_db job

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(database)
}

// ListDatabases lists databases for a project
func (h *DatabaseHandler) ListDatabases(w http.ResponseWriter, r *http.Request) {
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
	if project == nil || !project.BelongsToOrg(orgID) {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	databases, err := h.store.ListDatabasesByProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Don't expose passwords
	for _, db := range databases {
		if db.Password.Valid {
			db.Password = sql.NullString{}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(databases)
}

// GetDatabase retrieves a database by ID
func (h *DatabaseHandler) GetDatabase(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	databaseIDStr := chi.URLParam(r, "id")
	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		http.Error(w, "Invalid database ID", http.StatusBadRequest)
		return
	}

	database, err := h.store.GetDatabase(r.Context(), databaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if database == nil {
		http.Error(w, "Database not found", http.StatusNotFound)
		return
	}

	// Verify database belongs to user's organization (via service -> project)
	if database.ServiceID.Valid {
		serviceID, _ := uuid.Parse(database.ServiceID.String)
		service, err := h.store.GetService(r.Context(), serviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if service != nil {
			project, err := h.store.GetProject(r.Context(), service.ProjectID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if project == nil || !project.BelongsToOrg(orgID) {
				http.Error(w, "Database not found", http.StatusNotFound)
				return
			}
		}
	}

	// Don't expose password
	if database.Password.Valid {
		database.Password = sql.NullString{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(database)
}

// GetDatabaseCredentials retrieves database credentials
func (h *DatabaseHandler) GetDatabaseCredentials(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	databaseIDStr := chi.URLParam(r, "id")
	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		http.Error(w, "Invalid database ID", http.StatusBadRequest)
		return
	}

	// Verify database belongs to user's organization
	database, err := h.store.GetDatabase(r.Context(), databaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if database == nil {
		http.Error(w, "Database not found", http.StatusNotFound)
		return
	}

	if database.ServiceID.Valid {
		serviceID, _ := uuid.Parse(database.ServiceID.String)
		service, err := h.store.GetService(r.Context(), serviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if service != nil {
			project, err := h.store.GetProject(r.Context(), service.ProjectID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if project == nil || !project.BelongsToOrg(orgID) {
				http.Error(w, "Database not found", http.StatusNotFound)
				return
			}
		}
	}

	creds, err := h.store.GetDatabaseCredentials(r.Context(), databaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if creds == nil {
		http.Error(w, "Database not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
}

// DeleteDatabase deletes a database
func (h *DatabaseHandler) DeleteDatabase(w http.ResponseWriter, r *http.Request) {
	orgID := auth.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	databaseIDStr := chi.URLParam(r, "id")
	databaseID, err := uuid.Parse(databaseIDStr)
	if err != nil {
		http.Error(w, "Invalid database ID", http.StatusBadRequest)
		return
	}

	// Verify database belongs to user's organization
	database, err := h.store.GetDatabase(r.Context(), databaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if database == nil {
		http.Error(w, "Database not found", http.StatusNotFound)
		return
	}

	if database.ServiceID.Valid {
		serviceID, _ := uuid.Parse(database.ServiceID.String)
		service, err := h.store.GetService(r.Context(), serviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if service != nil {
			project, err := h.store.GetProject(r.Context(), service.ProjectID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if project == nil || !project.BelongsToOrg(orgID) {
				http.Error(w, "Database not found", http.StatusNotFound)
				return
			}
		}
	}

	// TODO: Queue destroy job (delete OpenStack resources first)

	if err := h.store.DeleteDatabase(r.Context(), databaseID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Database not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

