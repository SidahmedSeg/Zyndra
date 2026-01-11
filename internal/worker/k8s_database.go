package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/k8s"
	"github.com/intelifox/click-deploy/internal/store"
)

// K8sDatabaseWorker handles database provisioning on k8s
type K8sDatabaseWorker struct {
	store     *store.DB
	k8sClient *k8s.Client
}

// NewK8sDatabaseWorker creates a new k8s database worker
func NewK8sDatabaseWorker(store *store.DB, k8sClient *k8s.Client) *K8sDatabaseWorker {
	return &K8sDatabaseWorker{
		store:     store,
		k8sClient: k8sClient,
	}
}

// ProvisionDatabase creates a managed database on k8s
func (w *K8sDatabaseWorker) ProvisionDatabase(ctx context.Context, databaseID uuid.UUID) error {
	// Get database
	db, err := w.store.GetDatabase(ctx, databaseID)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	if db == nil {
		return fmt.Errorf("database not found: %s", databaseID)
	}

	// Get service (to find project)
	if !db.ServiceID.Valid {
		return fmt.Errorf("database has no linked service")
	}
	serviceID, err := uuid.Parse(db.ServiceID.String)
	if err != nil {
		return fmt.Errorf("invalid service ID: %w", err)
	}
	service, err := w.store.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Get project
	project, err := w.store.GetProject(ctx, service.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Update database status
	w.store.UpdateDatabaseStatus(ctx, databaseID, "provisioning")

	// Ensure namespace exists
	if err := w.k8sClient.CreateNamespace(ctx, project.ID.String(), project.Name); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Create database on k8s
	spec := k8s.DatabaseSpec{
		DatabaseID:   databaseID.String(),
		DatabaseName: db.DatabaseName.String,
		ProjectID:    project.ID.String(),
		Engine:       db.Engine,
		Version:      db.Version.String,
		SizeMB:       int64(db.VolumeSizeMB),
	}

	creds, err := w.k8sClient.CreateDatabase(ctx, spec)
	if err != nil {
		w.store.UpdateDatabaseStatus(ctx, databaseID, "failed")
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Wait for database to be ready
	readyCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := w.waitForDatabaseReady(readyCtx, project.ID.String(), databaseID.String()); err != nil {
		w.store.UpdateDatabaseStatus(ctx, databaseID, "failed")
		return fmt.Errorf("database failed to become ready: %w", err)
	}

	// Update database with credentials
	updateData := map[string]interface{}{
		"status":            "active",
		"internal_hostname": creds.Host,
		"port":              creds.Port,
		"username":          creds.Username,
		"password":          creds.Password,
		"database_name":     creds.Database,
		"connection_url":    creds.ConnectionURL,
	}

	if err := w.store.UpdateDatabaseFields(ctx, databaseID, updateData); err != nil {
		return fmt.Errorf("failed to update database fields: %w", err)
	}

	return nil
}

// waitForDatabaseReady polls the database status until it's ready
func (w *K8sDatabaseWorker) waitForDatabaseReady(ctx context.Context, projectID, databaseID string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := w.k8sClient.GetDatabaseStatus(ctx, projectID, databaseID)
			if err != nil {
				return fmt.Errorf("failed to get database status: %w", err)
			}

			if status.Ready {
				return nil
			}
		}
	}
}

// DeleteDatabase removes a database from k8s
func (w *K8sDatabaseWorker) DeleteDatabase(ctx context.Context, databaseID uuid.UUID) error {
	// Get database
	db, err := w.store.GetDatabase(ctx, databaseID)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	if db == nil {
		return nil // Already deleted
	}

	// Get service (to find project)
	if !db.ServiceID.Valid {
		return fmt.Errorf("database has no linked service")
	}
	serviceID, err := uuid.Parse(db.ServiceID.String)
	if err != nil {
		return fmt.Errorf("invalid service ID: %w", err)
	}
	service, err := w.store.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Delete from k8s
	if err := w.k8sClient.DeleteDatabase(ctx, service.ProjectID.String(), databaseID.String()); err != nil {
		return fmt.Errorf("failed to delete database from k8s: %w", err)
	}

	// Update status
	w.store.UpdateDatabaseStatus(ctx, databaseID, "deleted")

	return nil
}
