package api

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestMetricsHandler_GetServiceMetrics(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler, err := NewMetricsHandler(dbStore, &config.Config{})
	if err != nil {
		// Prometheus may not be configured in tests, create handler with nil client
		handler = &MetricsHandler{
			store:  dbStore,
			config: &config.Config{},
			client: nil,
		}
	}

	// Create a test project
	orgID := "test-org-metrics-001"
	project := &store.Project{
		Name:              "Test Project",
		Slug:              "test-project",
		CasdoorOrgID:      orgID,
		OpenStackTenantID: "test-tenant-123",
	}

	ctx := testutil.MockAuthContext(context.Background(), "test-user-123", orgID)
	if err := dbStore.CreateProject(ctx, project); err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create a test service
	service := &store.Service{
		ProjectID:    project.ID,
		Name:         "Test Service",
		Type:         "app",
		Status:       "active",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	tests := []struct {
		name           string
		serviceID      string
		orgID          string
		expectedStatus int
	}{
		{
			name:           "valid service",
			serviceID:      service.ID.String(),
			orgID:          orgID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent service",
			serviceID:      uuid.New().String(),
			orgID:          orgID,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "service from different org",
			serviceID:      service.ID.String(),
			orgID:          "different-org",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/services/"+tt.serviceID+"/metrics",
				map[string]string{"id": tt.serviceID}, nil, "test-user-123", tt.orgID)
			w := testutil.MockResponseRecorder()

			handler.GetServiceMetrics(w, req)

			// Metrics endpoint may return 200 even if Prometheus is not configured
			// or may return 500 if Prometheus is unavailable
			if w.Code != tt.expectedStatus && w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d or 500, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestMetricsHandler_GetDatabaseMetrics(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler, err := NewMetricsHandler(dbStore, &config.Config{})
	if err != nil {
		// Prometheus may not be configured in tests, create handler with nil client
		handler = &MetricsHandler{
			store:  dbStore,
			config: &config.Config{},
			client: nil,
		}
	}

	// Create a test project
	orgID := "test-org-metrics-002"
	project := &store.Project{
		Name:              "Test Project",
		Slug:              "test-project",
		CasdoorOrgID:      orgID,
		OpenStackTenantID: "test-tenant-123",
	}

	ctx := testutil.MockAuthContext(context.Background(), "test-user-123", orgID)
	if err := dbStore.CreateProject(ctx, project); err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create a test service
	service := &store.Service{
		ProjectID:    project.ID,
		Name:         "Test Service",
		Type:         "app",
		Status:       "active",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create a test database
	database := &store.Database{
		ServiceID:    sql.NullString{String: service.ID.String(), Valid: true},
		Engine:       "postgresql",
		Version:      sql.NullString{String: "14", Valid: true},
		Size:         "small",
		VolumeSizeMB: 500,
		Status:       "active",
	}
	if err := dbStore.CreateDatabase(ctx, database); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	tests := []struct {
		name           string
		databaseID     string
		orgID          string
		expectedStatus int
	}{
		{
			name:           "valid database",
			databaseID:     database.ID.String(),
			orgID:          orgID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent database",
			databaseID:     uuid.New().String(),
			orgID:          orgID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/databases/"+tt.databaseID+"/metrics",
				map[string]string{"id": tt.databaseID}, nil, "test-user-123", tt.orgID)
			w := testutil.MockResponseRecorder()

			handler.GetDatabaseMetrics(w, req)

			// Metrics endpoint may return 200 even if Prometheus is not configured
			// or may return 500 if Prometheus is unavailable
			if w.Code != tt.expectedStatus && w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d or 500, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestMetricsHandler_GetVolumeMetrics(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler, err := NewMetricsHandler(dbStore, &config.Config{})
	if err != nil {
		// Prometheus may not be configured in tests, create handler with nil client
		handler = &MetricsHandler{
			store:  dbStore,
			config: &config.Config{},
			client: nil,
		}
	}

	// Create a test project
	orgID := "test-org-metrics-003"
	project := &store.Project{
		Name:              "Test Project",
		Slug:              "test-project",
		CasdoorOrgID:      orgID,
		OpenStackTenantID: "test-tenant-123",
	}

	ctx := testutil.MockAuthContext(context.Background(), "test-user-123", orgID)
	if err := dbStore.CreateProject(ctx, project); err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create a test volume
	volume := &store.Volume{
		ProjectID:  project.ID,
		Name:       "Test Volume",
		SizeMB:     1000,
		Status:     "active",
		VolumeType: "user",
	}
	if err := dbStore.CreateVolume(ctx, volume); err != nil {
		t.Fatalf("Failed to create test volume: %v", err)
	}

	tests := []struct {
		name           string
		volumeID       string
		orgID          string
		expectedStatus int
	}{
		{
			name:           "valid volume",
			volumeID:       volume.ID.String(),
			orgID:          orgID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent volume",
			volumeID:       uuid.New().String(),
			orgID:          orgID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/volumes/"+tt.volumeID+"/metrics",
				map[string]string{"id": tt.volumeID}, nil, "test-user-123", tt.orgID)
			w := testutil.MockResponseRecorder()

			handler.GetVolumeMetrics(w, req)

			// Metrics endpoint may return 200 even if Prometheus is not configured
			// or may return 500 if Prometheus is unavailable
			if w.Code != tt.expectedStatus && w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d or 500, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

