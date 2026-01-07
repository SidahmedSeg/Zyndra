package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/intelifox/click-deploy/internal/store"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestProjectHandler_CreateProject(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewProjectHandler(dbStore, nil)

	tests := []struct {
		name           string
		requestBody    CreateProjectRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid project creation",
			requestBody: CreateProjectRequest{
				Name:              "Test Project",
				Description:       stringPtr("Test Description"),
				OpenStackTenantID: "test-tenant-123",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "missing name",
			requestBody: CreateProjectRequest{
				Name:              "",
				OpenStackTenantID: "test-tenant-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "name too long",
			requestBody: CreateProjectRequest{
				Name:              string(make([]byte, 300)), // 300 characters
				OpenStackTenantID: "test-tenant-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestJSON(t, "POST", "/v1/click-deploy/projects", tt.requestBody)
			w := testutil.MockResponseRecorder()

			handler.CreateProject(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusCreated {
				// Verify response is valid JSON (basic check)
				var result map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Verify name field exists
				if name, ok := result["name"].(string); ok && name != tt.requestBody.Name {
					t.Errorf("Expected project name %s, got %s", tt.requestBody.Name, name)
				}
			}
		})
	}
}

func TestProjectHandler_GetProject(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewProjectHandler(dbStore, nil)

	// Create a test project
	orgID := "test-org-456"
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

	tests := []struct {
		name           string
		projectID      string
		orgID          string
		expectedStatus int
	}{
		{
			name:           "valid project",
			projectID:      project.ID.String(),
			orgID:          orgID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent project",
			projectID:      uuid.New().String(),
			orgID:          orgID,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "project from different org",
			projectID:      project.ID.String(),
			orgID:          "different-org",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestWithURLParam(t, "GET", "/v1/click-deploy/projects/"+tt.projectID, 
				map[string]string{"id": tt.projectID}, nil)
			req = req.WithContext(testutil.MockAuthContext(req.Context(), "test-user-123", tt.orgID))
			w := testutil.MockResponseRecorder()

			handler.GetProject(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProjectHandler_ListProjects(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewProjectHandler(dbStore, nil)

	orgID := "test-org-456"
	ctx := testutil.MockAuthContext(context.Background(), "test-user-123", orgID)

	// Create test projects
	for i := 0; i < 3; i++ {
		project := &store.Project{
			ID:                uuid.New(),
			Name:              "Test Project " + string(rune('A'+i)),
			Slug:              "test-project-" + string(rune('a'+i)),
			CasdoorOrgID:      orgID,
			OpenStackTenantID: "test-tenant-123",
		}
		if err := dbStore.CreateProject(ctx, project); err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
	}

	req, _ := testutil.MockRequest(t, "GET", "/v1/click-deploy/projects", nil)
	w := testutil.MockResponseRecorder()

	handler.ListProjects(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var projects []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &projects); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}
}

func stringPtr(s string) *string {
	return &s
}

