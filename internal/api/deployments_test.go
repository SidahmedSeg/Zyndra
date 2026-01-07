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

func TestDeploymentHandler_TriggerDeployment(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewDeploymentHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-dep-001"
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
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create a git connection first
	gitConn := &store.GitConnection{
		CasdoorOrgID: orgID,
		Provider:     "github",
		AccessToken:   "test-token",
	}
	if err := dbStore.CreateGitConnection(ctx, gitConn); err != nil {
		t.Fatalf("Failed to create test git connection: %v", err)
	}

	// Create a git source for the service
	gitSource := &store.GitSource{
		ServiceID:      service.ID,
		GitConnectionID: gitConn.ID,
		Provider:       "github",
		RepoOwner:      "test-owner",
		RepoName:       "test-repo",
		Branch:         "main",
	}
	if err := dbStore.CreateGitSource(ctx, gitSource); err != nil {
		t.Fatalf("Failed to create test git source: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    TriggerDeploymentRequest
		expectedStatus int
	}{
		{
			name:           "valid deployment",
			requestBody:    TriggerDeploymentRequest{},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "deployment with commit SHA",
			requestBody: TriggerDeploymentRequest{
				CommitSHA: "abc123def456",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "deployment with branch",
			requestBody: TriggerDeploymentRequest{
				Branch: "develop",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "POST", "/v1/click-deploy/services/"+service.ID.String()+"/deploy",
				map[string]string{"id": service.ID.String()}, bytes.NewReader(body), "test-user-123", orgID)
			w := testutil.MockResponseRecorder()

			handler.TriggerDeployment(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestDeploymentHandler_GetDeployment(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewDeploymentHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-dep-002"
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
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create a test deployment
	deployment := &store.Deployment{
		ServiceID:   service.ID,
		Status:      "queued",
		TriggeredBy: "manual",
	}
	if err := dbStore.CreateDeployment(ctx, deployment); err != nil {
		t.Fatalf("Failed to create test deployment: %v", err)
	}

	tests := []struct {
		name           string
		deploymentID   string
		orgID          string
		expectedStatus int
	}{
		{
			name:           "valid deployment",
			deploymentID:   deployment.ID.String(),
			orgID:          orgID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent deployment",
			deploymentID:   uuid.New().String(),
			orgID:          orgID,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "deployment from different org",
			deploymentID:   deployment.ID.String(),
			orgID:          "different-org",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/deployments/"+tt.deploymentID,
				map[string]string{"id": tt.deploymentID}, nil, "test-user-123", tt.orgID)
			w := testutil.MockResponseRecorder()

			handler.GetDeployment(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestDeploymentHandler_ListServiceDeployments(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewDeploymentHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-dep-003"
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
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
	}
	if err := dbStore.CreateService(ctx, service); err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Create test deployments
	dep1 := &store.Deployment{
		ServiceID:   service.ID,
		Status:      "queued",
		TriggeredBy: "manual",
	}
	dep2 := &store.Deployment{
		ServiceID:   service.ID,
		Status:      "success",
		TriggeredBy: "webhook",
	}
	if err := dbStore.CreateDeployment(ctx, dep1); err != nil {
		t.Fatalf("Failed to create test deployment 1: %v", err)
	}
	if err := dbStore.CreateDeployment(ctx, dep2); err != nil {
		t.Fatalf("Failed to create test deployment 2: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/services/"+service.ID.String()+"/deployments",
		map[string]string{"id": service.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.ListServiceDeployments(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var deployments []*store.Deployment
	if err := json.NewDecoder(w.Body).Decode(&deployments); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(deployments) != 2 {
		t.Errorf("Expected 2 deployments, got %d", len(deployments))
	}
}

