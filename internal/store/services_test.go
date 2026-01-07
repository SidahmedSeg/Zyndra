package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestDB_CreateService(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Create a project first (required for foreign key)
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

	// Create a service
	service := &Service{
		ProjectID:    projectID,
		Name:         "Test Service",
		Type:         "app",
		Status:       "pending",
		InstanceSize: "medium",
		Port:         8080,
		CanvasX:      100,
		CanvasY:      200,
	}

	if isSQLite {
		service.ID = uuid.New()
		query := `INSERT INTO services (id, project_id, name, type, status, instance_size, port, canvas_x, canvas_y) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		_, err = db.ExecContext(ctx, query,
			service.ID.String(), service.ProjectID.String(), service.Name, service.Type,
			service.Status, service.InstanceSize, service.Port, service.CanvasX, service.CanvasY)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
		db.QueryRowContext(ctx, "SELECT created_at, updated_at FROM services WHERE id = $1", service.ID.String()).
			Scan(&service.CreatedAt, &service.UpdatedAt)
	} else {
		err = dbStore.CreateService(ctx, service)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
	}

	if service.ID == uuid.Nil {
		t.Error("Service ID should be set after creation")
	}

	if service.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestDB_GetService(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup: Create project and service
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	projectID := uuid.New()
	serviceID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO services (id, project_id, name, type, status, instance_size, port) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			serviceID.String(), projectID.String(), "Test Service", "app", "pending", "medium", 8080)
		if err != nil {
			t.Fatalf("Failed to create test service: %v", err)
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
		service := &Service{
			ID:          serviceID,
			ProjectID:   projectID,
			Name:        "Test Service",
			Type:        "app",
			Status:      "pending",
			InstanceSize: "medium",
			Port:        8080,
		}
		if err := dbStore.CreateService(ctx, service); err != nil {
			t.Fatalf("Failed to create test service: %v", err)
		}
	}

	// Retrieve the service
	retrieved, err := dbStore.GetService(ctx, serviceID)
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Service should exist")
	}

	if retrieved.ID != serviceID {
		t.Errorf("Expected service ID %s, got %s", serviceID, retrieved.ID)
	}

	if retrieved.Name != "Test Service" {
		t.Errorf("Expected service name 'Test Service', got '%s'", retrieved.Name)
	}

	// Test non-existent service
	nonExistentID := uuid.New()
	retrieved, err = dbStore.GetService(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("GetService should return nil for non-existent service, not error: %v", err)
	}
	if retrieved != nil {
		t.Error("GetService should return nil for non-existent service")
	}
}

func TestDB_ListServicesByProject(t *testing.T) {
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

		service1ID := uuid.New()
		service2ID := uuid.New()
		otherServiceID := uuid.New()

		_, err = db.ExecContext(ctx, `INSERT INTO services (id, project_id, name, type, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			service1ID.String(), projectID.String(), "Service 1", "app", "pending")
		if err != nil {
			t.Fatalf("Failed to create service 1: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO services (id, project_id, name, type, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			service2ID.String(), projectID.String(), "Service 2", "app", "pending")
		if err != nil {
			t.Fatalf("Failed to create service 2: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO services (id, project_id, name, type, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			otherServiceID.String(), otherProjectID.String(), "Other Service", "app", "pending")
		if err != nil {
			t.Fatalf("Failed to create other service: %v", err)
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

		service1 := &Service{
			ProjectID:    projectID,
			Name:         "Service 1",
			Type:         "app",
			Status:       "pending",
			InstanceSize: "medium",
			Port:         8080,
		}
		service2 := &Service{
			ProjectID:    projectID,
			Name:         "Service 2",
			Type:         "app",
			Status:       "pending",
			InstanceSize: "medium",
			Port:         8081,
		}
		if err := dbStore.CreateService(ctx, service1); err != nil {
			t.Fatalf("Failed to create service 1: %v", err)
		}
		if err := dbStore.CreateService(ctx, service2); err != nil {
			t.Fatalf("Failed to create service 2: %v", err)
		}
	}

	// List services for the project
	services, err := dbStore.ListServicesByProject(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to list services: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}

	// Verify services belong to the correct project
	for _, s := range services {
		if s.ProjectID != projectID {
			t.Errorf("Service %s belongs to wrong project: %s", s.ID, s.ProjectID)
		}
	}
}

func TestDB_UpdateService(t *testing.T) {
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
	serviceID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO services (id, project_id, name, type, status, instance_size, port) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			serviceID.String(), projectID.String(), "Original Name", "app", "pending", "medium", 8080)
		if err != nil {
			t.Fatalf("Failed to create test service: %v", err)
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
		service := &Service{
			ID:           serviceID,
			ProjectID:    projectID,
			Name:         "Original Name",
			Type:         "app",
			Status:       "pending",
			InstanceSize: "medium",
			Port:         8080,
		}
		if err := dbStore.CreateService(ctx, service); err != nil {
			t.Fatalf("Failed to create test service: %v", err)
		}
	}

	// Update the service
	updates := &Service{
		Name:         "Updated Name",
		Status:       "active",
		InstanceSize: "large",
		Port:         9090,
	}

	if isSQLite {
		query := `UPDATE services SET name = $1, status = $2, instance_size = $3, port = $4, 
			updated_at = datetime('now') WHERE id = $5`
		_, err = db.ExecContext(ctx, query, updates.Name, updates.Status, updates.InstanceSize, 
			updates.Port, serviceID.String())
		if err != nil {
			t.Fatalf("Failed to update service: %v", err)
		}
	} else {
		err = dbStore.UpdateService(ctx, serviceID, updates)
		if err != nil {
			t.Fatalf("Failed to update service: %v", err)
		}
	}

	// Verify updates
	updated, err := dbStore.GetService(ctx, serviceID)
	if err != nil {
		t.Fatalf("Failed to get updated service: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}

	if updated.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", updated.Status)
	}

	if updated.InstanceSize != "large" {
		t.Errorf("Expected instance size 'large', got '%s'", updated.InstanceSize)
	}

	if updated.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", updated.Port)
	}
}

func TestDB_DeleteService(t *testing.T) {
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
	serviceID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO projects (id, casdoor_org_id, name, slug, openstack_tenant_id) 
			VALUES ($1, $2, $3, $4, $5)`, 
			projectID.String(), "test-org", "Test Project", "test-project", "test-tenant")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		_, err = db.ExecContext(ctx, `INSERT INTO services (id, project_id, name, type, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			serviceID.String(), projectID.String(), "Test Service", "app", "pending")
		if err != nil {
			t.Fatalf("Failed to create test service: %v", err)
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
		service := &Service{
			ID:           serviceID,
			ProjectID:    projectID,
			Name:         "Test Service",
			Type:         "app",
			Status:       "pending",
			InstanceSize: "medium",
			Port:         8080,
		}
		if err := dbStore.CreateService(ctx, service); err != nil {
			t.Fatalf("Failed to create test service: %v", err)
		}
	}

	// Delete the service
	err = dbStore.DeleteService(ctx, serviceID)
	if err != nil {
		t.Fatalf("Failed to delete service: %v", err)
	}

	// Verify service is deleted
	deleted, err := dbStore.GetService(ctx, serviceID)
	if err != nil {
		t.Fatalf("Failed to check deleted service: %v", err)
	}

	if deleted != nil {
		t.Error("Service should be deleted")
	}
}

