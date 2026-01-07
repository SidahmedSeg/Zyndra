package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestEnvVarHandler_CreateEnvVar(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewEnvVarHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-env-001"
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
		requestBody    CreateEnvVarRequest
		expectedStatus int
	}{
		{
			name: "valid env var",
			requestBody: CreateEnvVarRequest{
				Key:   "DATABASE_URL",
				Value: "postgres://user:pass@localhost:5432/db",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "secret env var",
			requestBody: CreateEnvVarRequest{
				Key:      "API_KEY",
				Value:    "secret-key-123",
				IsSecret: true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing key",
			requestBody: CreateEnvVarRequest{
				Value: "some-value",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "POST", "/v1/click-deploy/services/"+service.ID.String()+"/env",
				map[string]string{"id": service.ID.String()}, bytes.NewReader(body), "test-user-123", orgID)
			w := testutil.MockResponseRecorder()

			handler.CreateEnvVar(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestEnvVarHandler_ListEnvVars(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewEnvVarHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-env-002"
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

	// Create test environment variables
	envVar1 := &store.EnvVar{
		ServiceID: service.ID,
		Key:       "DATABASE_URL",
		Value:     sql.NullString{String: "postgres://localhost:5432/db", Valid: true},
		IsSecret:  false,
	}
	envVar2 := &store.EnvVar{
		ServiceID: service.ID,
		Key:       "API_KEY",
		Value:     sql.NullString{String: "secret-key", Valid: true},
		IsSecret:  true,
	}
	if err := dbStore.CreateEnvVar(ctx, envVar1); err != nil {
		t.Fatalf("Failed to create test env var 1: %v", err)
	}
	if err := dbStore.CreateEnvVar(ctx, envVar2); err != nil {
		t.Fatalf("Failed to create test env var 2: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/services/"+service.ID.String()+"/env",
		map[string]string{"id": service.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.ListEnvVars(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var envVars []*store.EnvVar
	if err := json.NewDecoder(w.Body).Decode(&envVars); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(envVars) != 2 {
		t.Errorf("Expected 2 env vars, got %d", len(envVars))
	}
}

func TestEnvVarHandler_DeleteEnvVar(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewEnvVarHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-env-003"
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

	// Create a test environment variable
	envVar := &store.EnvVar{
		ServiceID: service.ID,
		Key:       "DATABASE_URL",
		Value:     sql.NullString{String: "postgres://localhost:5432/db", Valid: true},
		IsSecret:  false,
	}
	if err := dbStore.CreateEnvVar(ctx, envVar); err != nil {
		t.Fatalf("Failed to create test env var: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "DELETE", "/v1/click-deploy/services/"+service.ID.String()+"/env/DATABASE_URL",
		map[string]string{"id": service.ID.String(), "key": "DATABASE_URL"}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.DeleteEnvVar(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusNoContent, w.Code, w.Body.String())
	}

	// Verify the env var was deleted by listing env vars
	envVars, err := dbStore.ListEnvVarsByService(ctx, service.ID)
	if err != nil {
		t.Fatalf("Failed to list env vars: %v", err)
	}
	if len(envVars) != 0 {
		t.Errorf("Expected 0 env vars after deletion, got %d", len(envVars))
	}
}

