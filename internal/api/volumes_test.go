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

func TestVolumeHandler_CreateVolume(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewVolumeHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-vol-001"
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
		requestBody    CreateVolumeRequest
		expectedStatus int
	}{
		{
			name: "valid volume",
			requestBody: CreateVolumeRequest{
				Name:      "Test Volume",
				SizeMB:    1000,
				MountPath: "/data",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "volume without mount path",
			requestBody: CreateVolumeRequest{
				Name:   "Test Volume 2",
				SizeMB: 500,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			requestBody: CreateVolumeRequest{
				SizeMB: 1000,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid size",
			requestBody: CreateVolumeRequest{
				Name:   "Test Volume",
				SizeMB: 0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative size",
			requestBody: CreateVolumeRequest{
				Name:   "Test Volume",
				SizeMB: -100,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "POST", "/v1/click-deploy/projects/"+project.ID.String()+"/volumes",
				map[string]string{"id": project.ID.String()}, bytes.NewReader(body), "test-user-123", orgID)
			w := testutil.MockResponseRecorder()

			handler.CreateVolume(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestVolumeHandler_ListVolumes(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewVolumeHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-vol-002"
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

	// Create test volumes
	vol1 := &store.Volume{
		ProjectID:  project.ID,
		Name:       "Volume 1",
		SizeMB:     1000,
		Status:     "active",
		VolumeType: "user",
	}
	vol2 := &store.Volume{
		ProjectID:  project.ID,
		Name:       "Volume 2",
		SizeMB:     500,
		Status:     "active",
		VolumeType: "user",
	}

	if err := dbStore.CreateVolume(ctx, vol1); err != nil {
		t.Fatalf("Failed to create test volume 1: %v", err)
	}
	if err := dbStore.CreateVolume(ctx, vol2); err != nil {
		t.Fatalf("Failed to create test volume 2: %v", err)
	}

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/projects/"+project.ID.String()+"/volumes",
		map[string]string{"id": project.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.ListVolumes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var volumes []*store.Volume
	if err := json.NewDecoder(w.Body).Decode(&volumes); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(volumes) != 2 {
		t.Errorf("Expected 2 volumes, got %d", len(volumes))
	}
}

func TestVolumeHandler_GetVolume(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewVolumeHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-vol-003"
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
		{
			name:           "volume from different org",
			volumeID:       volume.ID.String(),
			orgID:          "different-org",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := testutil.MockRequestWithURLParamAndAuth(t, "GET", "/v1/click-deploy/volumes/"+tt.volumeID,
				map[string]string{"id": tt.volumeID}, nil, "test-user-123", tt.orgID)
			w := testutil.MockResponseRecorder()

			handler.GetVolume(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestVolumeHandler_DeleteVolume(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &store.DB{DB: db}
	handler := NewVolumeHandler(dbStore, &config.Config{})

	// Create a test project
	orgID := "test-org-vol-004"
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

	req, _ := testutil.MockRequestWithURLParamAndAuth(t, "DELETE", "/v1/click-deploy/volumes/"+volume.ID.String(),
		map[string]string{"id": volume.ID.String()}, nil, "test-user-123", orgID)
	w := testutil.MockResponseRecorder()

	handler.DeleteVolume(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusNoContent, w.Code, w.Body.String())
	}

	// Verify the volume was deleted
	deletedVolume, err := dbStore.GetVolume(ctx, volume.ID)
	if err != nil {
		t.Fatalf("Failed to check deleted volume: %v", err)
	}
	if deletedVolume != nil {
		t.Error("Volume should have been deleted")
	}
}

