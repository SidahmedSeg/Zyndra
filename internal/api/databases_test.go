package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestDatabaseHandler_CreateDatabase(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewDatabaseHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-db-001"
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

	tests := []struct {
		name           string
		requestBody    CreateDatabaseRequest
		expectedStatus int
	}{
		{
			name: "valid postgresql database",
			requestBody: CreateDatabaseRequest{
				Engine:       "postgresql",
				Version:      "14",
				Size:         "small",
				VolumeSizeMB: 500,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "valid mysql database",
			requestBody: CreateDatabaseRequest{
				Engine:       "mysql",
				Version:      "8.0",
				Size:         "medium",
				VolumeSizeMB: 1000,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "valid redis database",
			requestBody: CreateDatabaseRequest{
				Engine:       "redis",
				Size:         "small",
				VolumeSizeMB: 500,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid engine",
			requestBody: CreateDatabaseRequest{
				Engine:       "invalid",
				Size:         "small",
				VolumeSizeMB: 500,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid size",
			requestBody: CreateDatabaseRequest{
				Engine:       "postgresql",
				Size:         "invalid",
				VolumeSizeMB: 500,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "default size and volume",
			requestBody: CreateDatabaseRequest{
				Engine: "postgresql",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "POST", "/v1/click-deploy/projects/"+project.ID.String()+"/databases",
				map[string]string{"id": project.ID.String()}, bytes.NewReader(body), "test-user-123", orgID)
			w := testutil.MockResponseRecorder()

			handler.CreateDatabase(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestDatabaseHandler_ListDatabases(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewDatabaseHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-db-002"
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

	// Create a test service to link databases to
	service := &store.Service{
		ProjectID:    project.ID,
		Name:         "Test Service",
		Type:         "app",
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create test databases linked to the service
	db1 := &store.Database{
		ServiceID:    sql.NullString{String: service.ID.String(), Valid: true},
		Engine:       "postgresql",
		Version:      sql.NullString{String: "14", Valid: true},
		Size:         "small",
		VolumeSizeMB: 500,
		Status:       "active",
	}
	db2 := &store.Database{
		ServiceID:    sql.NullString{String: service.ID.String(), Valid: true},
		Engine:       "mysql",
		Version:      sql.NullString{String: "8.0", Valid: true},
		Size:         "medium",
		VolumeSizeMB: 1000,
		Status:       "active",
	}

	if err := dbStore.CreateDatabase(ctx, db1); err != nil {
		t.Fatalf("Failed to create test database 1: %v", err)
	}
	if err := dbStore.CreateDatabase(ctx, db2); err != nil {
		t.Fatalf("Failed to create test database 2: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/projects/"+project.ID.String()+"/databases",
		map[string]string{"id": project.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.ListDatabases(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var databases []*store.Database
	if err := json.NewDecoder(w.Body).Decode(&databases); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(databases) != 2 {
		t.Errorf("Expected 2 databases, got %d", len(databases))
	}
}

func TestDatabaseHandler_GetDatabase(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewDatabaseHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-db-003"
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

	// Create a test service to link database to
	service := &store.Service{
		ProjectID:    project.ID,
		Name:         "Test Service",
		Type:         "app",
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create a test database linked to the service
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
		{
			name:           "database from different org",
			databaseID:     database.ID.String(),
			orgID:          "different-org",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/databases/"+tt.databaseID,
				map[string]string{"id": tt.databaseID}, nil, "test-user-123", tt.orgID)
			w := testutil.MockResponseRecorder()

			handler.GetDatabase(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestDatabaseHandler_DeleteDatabase(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewDatabaseHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-db-004"
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

	// Create a test service to link database to
	service := &store.Service{
		ProjectID:    project.ID,
		Name:         "Test Service",
		Type:         "app",
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create a test database linked to the service
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

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "DELETE", "/v1/click-deploy/databases/"+database.ID.String(),
		map[string]string{"id": database.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.DeleteDatabase(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusNoContent, w.Code, w.Body.String())
	}

	// Verify the database was deleted
	deletedDB, err := dbStore.GetDatabase(ctx, database.ID)
	if err != nil {
		t.Fatalf("Failed to check deleted database: %v", err)
	}
	if deletedDB != nil {
		t.Error("Database should have been deleted")
	}
}

