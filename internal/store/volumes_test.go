package store

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestDB_CreateVolume(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup: Create a project first
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	projectID := uuid.New()
	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
	} else {
		project := &Project{
			ID:                projectID,
			CasdoorOrgID:      "test-org",
			Name:              "Test Project",
			Slug:              "test-project",
			OpenStackTenantID: "test-tenant",
			AutoDeploy:        true,
		}
		if err := dbStore.CreateProject(ctx, project); err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
	}

	// Create a volume
	volume := &Volume{
		ProjectID:  projectID,
		Name:       "Test Volume",
		SizeMB:     10240,
		VolumeType: "user",
		Status:     "pending",
	}

	if isSQLite {
		volume.ID = uuid.New()
		query := `INSERT INTO volumes (id, project_id, name, size_mb, volume_type, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = db.ExecContext(ctx, query,
			volume.ID.String(), volume.ProjectID.String(), volume.Name,
			volume.SizeMB, volume.VolumeType, volume.Status)
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		db.QueryRowContext(ctx, "SELECT created_at FROM volumes WHERE id = $1", volume.ID.String()).
			Scan(&volume.CreatedAt)
	} else {
		err = dbStore.CreateVolume(ctx, volume)
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
	}

	if volume.ID == uuid.Nil {
		t.Error("Volume ID should be set after creation")
	}

	if volume.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestDB_GetVolume(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	projectID := uuid.New()
	volumeID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO volumes (id, project_id, name, size_mb, volume_type, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			volumeID.String(), projectID.String(), "Test Volume", 10240, "user", "pending")
		if err != nil {
			t.Fatalf("Failed to create test volume: %v", err)
		}
	} else {
		project := &Project{
			ID:                projectID,
			CasdoorOrgID:      "test-org",
			Name:              "Test Project",
			Slug:              "test-project",
			OpenStackTenantID: "test-tenant",
			AutoDeploy:        true,
		}
		if err := dbStore.CreateProject(ctx, project); err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		volume := &Volume{
			ID:         volumeID,
			ProjectID:  projectID,
			Name:       "Test Volume",
			SizeMB:     10240,
			VolumeType: "user",
			Status:     "pending",
		}
		if err := dbStore.CreateVolume(ctx, volume); err != nil {
			t.Fatalf("Failed to create test volume: %v", err)
		}
	}

	// Retrieve the volume
	retrieved, err := dbStore.GetVolume(ctx, volumeID)
	if err != nil {
		t.Fatalf("Failed to get volume: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Volume should exist")
	}

	if retrieved.ID != volumeID {
		t.Errorf("Expected volume ID %s, got %s", volumeID, retrieved.ID)
	}

	if retrieved.Name != "Test Volume" {
		t.Errorf("Expected volume name 'Test Volume', got '%s'", retrieved.Name)
	}

	// Test non-existent volume
	nonExistentID := uuid.New()
	retrieved, err = dbStore.GetVolume(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("GetVolume should return nil for non-existent volume, not error: %v", err)
	}
	if retrieved != nil {
		t.Error("GetVolume should return nil for non-existent volume")
	}
}

func TestDB_ListVolumesByProject(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	projectID := uuid.New()
	otherProjectID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			otherProjectID.String(), "test-org", "Other Project", "other-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create other project: %v", err)
		}

		volume1ID := uuid.New()
		volume2ID := uuid.New()
		otherVolumeID := uuid.New()

		_, err = db.ExecContext(ctx, `INSERT INTO volumes (id, project_id, name, size_mb, volume_type, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			volume1ID.String(), projectID.String(), "Volume 1", 10240, "user", "pending")
		if err != nil {
			t.Fatalf("Failed to create volume 1: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO volumes (id, project_id, name, size_mb, volume_type, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			volume2ID.String(), projectID.String(), "Volume 2", 20480, "user", "available")
		if err != nil {
			t.Fatalf("Failed to create volume 2: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO volumes (id, project_id, name, size_mb, volume_type, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			otherVolumeID.String(), otherProjectID.String(), "Other Volume", 10240, "user", "pending")
		if err != nil {
			t.Fatalf("Failed to create other volume: %v", err)
		}
	} else {
		project := &Project{
			ID:                projectID,
			CasdoorOrgID:      "test-org",
			Name:              "Test Project",
			Slug:              "test-project",
			OpenStackTenantID: "test-tenant",
			AutoDeploy:        true,
		}
		if err := dbStore.CreateProject(ctx, project); err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}

		volume1 := &Volume{
			ProjectID:  projectID,
			Name:       "Volume 1",
			SizeMB:     10240,
			VolumeType: "user",
			Status:     "pending",
		}
		volume2 := &Volume{
			ProjectID:  projectID,
			Name:       "Volume 2",
			SizeMB:     20480,
			VolumeType: "user",
			Status:     "available",
		}
		if err := dbStore.CreateVolume(ctx, volume1); err != nil {
			t.Fatalf("Failed to create volume 1: %v", err)
		}
		if err := dbStore.CreateVolume(ctx, volume2); err != nil {
			t.Fatalf("Failed to create volume 2: %v", err)
		}
	}

	// List volumes for the project
	volumes, err := dbStore.ListVolumesByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to list volumes: %v", err)
	}

	if len(volumes) != 2 {
		t.Errorf("Expected 2 volumes, got %d", len(volumes))
	}

	// Verify volumes belong to the correct project
	for _, v := range volumes {
		if v.ProjectID != projectID {
			t.Errorf("Volume %s belongs to wrong project: %s", v.ID, v.ProjectID)
		}
	}
}

func TestDB_UpdateVolume(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	projectID := uuid.New()
	volumeID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO volumes (id, project_id, name, size_mb, volume_type, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			volumeID.String(), projectID.String(), "Original Name", 10240, "user", "pending")
		if err != nil {
			t.Fatalf("Failed to create test volume: %v", err)
		}
	} else {
		project := &Project{
			ID:                projectID,
			CasdoorOrgID:      "test-org",
			Name:              "Test Project",
			Slug:              "test-project",
			OpenStackTenantID: "test-tenant",
			AutoDeploy:        true,
		}
		if err := dbStore.CreateProject(ctx, project); err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		volume := &Volume{
			ID:         volumeID,
			ProjectID:  projectID,
			Name:       "Original Name",
			SizeMB:     10240,
			VolumeType: "user",
			Status:     "pending",
		}
		if err := dbStore.CreateVolume(ctx, volume); err != nil {
			t.Fatalf("Failed to create test volume: %v", err)
		}
	}

	// Update the volume
	updates := &Volume{
		Name:       "Updated Name",
		Status:     "available",
		MountPath:  sql.NullString{String: "/mnt/data", Valid: true},
	}

	if isSQLite {
		query := `UPDATE volumes SET name = $1, status = $2, mount_path = $3 WHERE id = $4`
		_, err = db.ExecContext(ctx, query, updates.Name, updates.Status, updates.MountPath.String, volumeID.String())
		if err != nil {
			t.Fatalf("Failed to update volume: %v", err)
		}
	} else {
		err = dbStore.UpdateVolume(ctx, volumeID, updates)
		if err != nil {
			t.Fatalf("Failed to update volume: %v", err)
		}
	}

	// Verify updates
	updated, err := dbStore.GetVolume(ctx, volumeID)
	if err != nil {
		t.Fatalf("Failed to get updated volume: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}

	if updated.Status != "available" {
		t.Errorf("Expected status 'available', got '%s'", updated.Status)
	}

	if !updated.MountPath.Valid || updated.MountPath.String != "/mnt/data" {
		t.Errorf("Expected mount path '/mnt/data', got valid=%v, string=%s",
			updated.MountPath.Valid, updated.MountPath.String)
	}
}

func TestDB_DeleteVolume(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	projectID := uuid.New()
	volumeID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO volumes (id, project_id, name, size_mb, volume_type, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			volumeID.String(), projectID.String(), "Test Volume", 10240, "user", "pending")
		if err != nil {
			t.Fatalf("Failed to create test volume: %v", err)
		}
	} else {
		project := &Project{
			ID:                projectID,
			CasdoorOrgID:      "test-org",
			Name:              "Test Project",
			Slug:              "test-project",
			OpenStackTenantID: "test-tenant",
			AutoDeploy:        true,
		}
		if err := dbStore.CreateProject(ctx, project); err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		volume := &Volume{
			ID:         volumeID,
			ProjectID:  projectID,
			Name:       "Test Volume",
			SizeMB:     10240,
			VolumeType: "user",
			Status:     "pending",
		}
		if err := dbStore.CreateVolume(ctx, volume); err != nil {
			t.Fatalf("Failed to create test volume: %v", err)
		}
	}

	// Delete the volume
	err = dbStore.DeleteVolume(ctx, volumeID)
	if err != nil {
		t.Fatalf("Failed to delete volume: %v", err)
	}

	// Verify volume is deleted
	deleted, err := dbStore.GetVolume(ctx, volumeID)
	if err != nil {
		t.Fatalf("Failed to check deleted volume: %v", err)
	}

	if deleted != nil {
		t.Error("Volume should be deleted")
	}
}

