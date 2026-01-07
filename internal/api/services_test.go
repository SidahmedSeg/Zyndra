package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestServiceHandler_CreateService(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewServiceHandler(dbStore, &config.Config{})

	// Create a test project first
	orgID := "test-org-789"
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

	// Verify project was created and ID was set
	if project.ID == uuid.Nil {
		t.Fatalf("Project ID was not set after creation")
	}

	verifyProject, err := dbStore.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("Failed to verify project creation: %v", err)
	}
	if verifyProject == nil {
		t.Fatalf("Project was not created - GetProject returned nil")
	}
	if verifyProject.CasdoorOrgID != orgID {
		t.Fatalf("Project orgID mismatch: got %s, want %s", verifyProject.CasdoorOrgID, orgID)
	}

	tests := []struct {
		name           string
		requestBody    CreateServiceRequest
		expectedStatus int
	}{
		{
			name: "valid service",
			requestBody: CreateServiceRequest{
				Name:         "Test Service",
				Type:         "app",
				InstanceSize: "medium",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			requestBody: CreateServiceRequest{
				Type:         "app",
				InstanceSize: "medium",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name too long",
			requestBody: CreateServiceRequest{
				Name:         string(make([]byte, 256)), // 256 characters
				Type:         "app",
				InstanceSize: "medium",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			// Create request with correct orgID and URL params from the start
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "POST", "/v1/click-deploy/projects/"+project.ID.String()+"/services",
				map[string]string{"id": project.ID.String()}, bytes.NewReader(body), "test-user-123", orgID)
			w := testutil.MockResponseRecorder()

			handler.CreateService(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestServiceHandler_ListServices(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewServiceHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-101"
	project := &store.Project{
		ID:                uuid.New(),
		Name:              "Test Project",
		Slug:              "test-project",
		CasdoorOrgID:      orgID,
		OpenStackTenantID: "test-tenant-123",
	}

	ctx := testutil.MockAuthContext(context.Background(), "test-user-123", orgID)
	if err := dbStore.CreateProject(ctx, project); err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create test services
	service1 := &store.Service{
		ProjectID:    project.ID,
		Name:         "Service 1",
		Type:         "app",
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}
	service2 := &store.Service{
		ProjectID:    project.ID,
		Name:         "Service 2",
		Type:         "app",
		Status:       "pending",
		InstanceSize: "small",
		Port:         8081,
	}

	if err := dbStore.CreateService(ctx, service1); err != nil {
		t.Fatalf("Failed to create test service 1: %v", err)
	}
	if err := dbStore.CreateService(ctx, service2); err != nil {
		t.Fatalf("Failed to create test service 2: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/projects/"+project.ID.String()+"/services",
		map[string]string{"id": project.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.ListServices(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var services []*store.Service
	if err := json.NewDecoder(w.Body).Decode(&services); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}
}

func TestServiceHandler_GetService(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewServiceHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-202"
	project := &store.Project{
		ID:                uuid.New(),
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
		Status:       "pending",
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
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/services/"+tt.serviceID,
				map[string]string{"id": tt.serviceID}, nil, "test-user-123", tt.orgID)
			w := testutil.MockResponseRecorder()

			handler.GetService(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServiceHandler_UpdateService(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewServiceHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-303"
	project := &store.Project{
		ID:                uuid.New(),
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
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}

	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	newName := "Updated Service Name"
	requestBody := UpdateServiceRequest{
		Name: &newName,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "PATCH", "/v1/click-deploy/services/"+service.ID.String(),
		map[string]string{"id": service.ID.String()}, bytes.NewReader(body), "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.UpdateService(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify the service was updated
	updatedService, err := dbStore.GetService(ctx, service.ID)
	if err != nil {
		t.Fatalf("Failed to get updated service: %v", err)
	}
	if updatedService == nil {
		t.Fatal("Service should exist after update")
	}
	if updatedService.Name != newName {
		t.Errorf("Expected service name %s, got %s", newName, updatedService.Name)
	}
}

func TestServiceHandler_DeleteService(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewServiceHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-404"
	project := &store.Project{
		ID:                uuid.New(),
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
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}

	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "DELETE", "/v1/click-deploy/services/"+service.ID.String(),
		map[string]string{"id": service.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.DeleteService(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify the service was deleted
	deletedService, err := dbStore.GetService(ctx, service.ID)
	if err != nil {
		t.Fatalf("Failed to check deleted service: %v", err)
	}
	if deletedService != nil {
		t.Error("Service should have been deleted")
	}
}

