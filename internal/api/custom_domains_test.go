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

func TestCustomDomainHandler_AddCustomDomain(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewCustomDomainHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-cd-001"
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

	// Update service to have a floating IP address (required for custom domains)
	service.OpenStackFIPAddress = sql.NullString{String: "192.168.1.100", Valid: true}
	if err := dbStore.UpdateService(ctx, service.ID, service); err != nil {
		t.Fatalf("Failed to update service with FIP: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    AddCustomDomainRequest
		expectedStatus int
	}{
		{
			name: "valid domain",
			requestBody: AddCustomDomainRequest{
				Domain: "example.com",
			},
			// Will fail on Caddy call, but that's expected in tests
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "subdomain",
			requestBody: AddCustomDomainRequest{
				Domain: "api.example.com",
			},
			// Will fail on Caddy call, but that's expected in tests
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "missing domain",
			requestBody: AddCustomDomainRequest{
				Domain: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "POST", "/v1/click-deploy/services/"+service.ID.String()+"/custom-domains",
				map[string]string{"id": service.ID.String()}, bytes.NewReader(body), "test-user-123", orgID)
			w := testutil.MockResponseRecorder()

			handler.AddCustomDomain(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestCustomDomainHandler_ListCustomDomains(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewCustomDomainHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-cd-002"
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

	// Update service to have a floating IP address (required for custom domains)
	service.OpenStackFIPAddress = sql.NullString{String: "192.168.1.100", Valid: true}
	if err := dbStore.UpdateService(ctx, service.ID, service); err != nil {
		t.Fatalf("Failed to update service with FIP: %v", err)
	}

	// Create test custom domains
	domain1 := &store.CustomDomain{
		ServiceID: service.ID,
		Domain:    "example.com",
		Status:    "active",
	}
	domain2 := &store.CustomDomain{
		ServiceID: service.ID,
		Domain:    "api.example.com",
		Status:    "pending",
	}
	if err := dbStore.CreateCustomDomain(ctx, domain1); err != nil {
		t.Fatalf("Failed to create test domain 1: %v", err)
	}
	if err := dbStore.CreateCustomDomain(ctx, domain2); err != nil {
		t.Fatalf("Failed to create test domain 2: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/services/"+service.ID.String()+"/custom-domains",
		map[string]string{"id": service.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.ListCustomDomains(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var domains []*store.CustomDomain
	if err := json.NewDecoder(w.Body).Decode(&domains); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(domains))
	}
}

func TestCustomDomainHandler_DeleteCustomDomain(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewCustomDomainHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-cd-003"
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

	// Update service to have a floating IP address (required for custom domains)
	service.OpenStackFIPAddress = sql.NullString{String: "192.168.1.100", Valid: true}
	if err := dbStore.UpdateService(ctx, service.ID, service); err != nil {
		t.Fatalf("Failed to update service with FIP: %v", err)
	}

	// Create a test custom domain
	domain := &store.CustomDomain{
		ServiceID: service.ID,
		Domain:    "example.com",
		Status:    "active",
	}
	if err := dbStore.CreateCustomDomain(ctx, domain); err != nil {
		t.Fatalf("Failed to create test domain: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "DELETE", "/v1/click-deploy/custom-domains/"+domain.ID.String(),
		map[string]string{"id": domain.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.DeleteCustomDomain(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusNoContent, w.Code, w.Body.String())
	}

	// Verify the domain was deleted
	deletedDomain, err := dbStore.GetCustomDomain(ctx, domain.ID)
	if err != nil {
		t.Fatalf("Failed to check deleted domain: %v", err)
	}
	if deletedDomain != nil {
		t.Error("Domain should have been deleted")
	}
}

