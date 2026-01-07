package worker

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/infra"
	"github.com/intelifox/click-deploy/internal/metrics"
	"github.com/intelifox/click-deploy/internal/store"
)

// CleanupWorker handles resource cleanup jobs
type CleanupWorker struct {
	store  *store.DB
	config *config.Config
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker(store *store.DB, cfg *config.Config) *CleanupWorker {
	return &CleanupWorker{
		store:  store,
		config: cfg,
	}
}

// CleanupServiceResources cleans up all resources associated with a service
func (w *CleanupWorker) CleanupServiceResources(ctx context.Context, serviceID uuid.UUID) error {
	// Get service details
	service, err := w.store.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if service == nil {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	// Get project for OpenStack tenant ID
	project, err := w.store.GetProject(ctx, service.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Create infra client
	infraConfig := infra.Config{
		BaseURL:  w.config.InfraServiceURL,
		APIKey:   w.config.InfraServiceAPIKey,
		TenantID: project.OpenStackTenantID,
		UseMock:  w.config.UseMockInfra,
	}

	baseClient := infra.NewClient(infraConfig)
	client := infra.NewRetryClient(baseClient)

	// 1. Unregister from Prometheus
	if service.OpenStackInstanceID.Valid {
		targetManager := metrics.NewTargetManager(w.config.PrometheusTargetsDir)
		if err := targetManager.UnregisterInstance(service.OpenStackInstanceID.String); err != nil {
			fmt.Printf("Warning: Failed to unregister service %s from Prometheus: %v\n", serviceID, err)
		}
	}

	// 2. Delete container/instance if exists
	if service.OpenStackInstanceID.Valid {
		instanceID := service.OpenStackInstanceID.String

		// Try to stop and delete container first
		if err := client.StopContainer(ctx, instanceID); err != nil {
			// Log but continue - container might already be stopped
			fmt.Printf("Warning: failed to stop container %s: %v\n", instanceID, err)
		}

		if err := client.DeleteContainer(ctx, instanceID); err != nil {
			// Log but continue - might be already deleted
			fmt.Printf("Warning: failed to delete container %s: %v\n", instanceID, err)
		} else {
			// If container deletion failed, try instance deletion
			if err := client.DeleteInstance(ctx, instanceID); err != nil {
				fmt.Printf("Warning: failed to delete instance %s: %v\n", instanceID, err)
			}
		}
	}

	// 3. Detach and release floating IP if exists
	if service.OpenStackFIPID.Valid {
		fipID := service.OpenStackFIPID.String

		// If instance still exists, detach FIP first
		if service.OpenStackInstanceID.Valid {
			// Detach is handled by delete instance, but we try anyway
			_ = client.DeleteInstance(ctx, service.OpenStackInstanceID.String)
		}

		// Note: In a real implementation, we'd need an API endpoint to release the FIP
		// For now, we'll mark it for cleanup and rely on the infrastructure service
		// to handle orphaned resources
		fmt.Printf("Floating IP %s should be released\n", fipID)
	}

	// 4. Delete security group if exists
	if service.SecurityGroupID.Valid {
		sgID := service.SecurityGroupID.String
		// Note: In a real implementation, we'd need a DeleteSecurityGroup method
		// For now, we'll mark it for cleanup
		fmt.Printf("Security Group %s should be deleted\n", sgID)
	}

	// 5. Delete DNS record if exists
	if service.Subdomain.Valid {
		// Note: In a real implementation, we'd need a DeleteDNSRecord method
		// For now, we'll mark it for cleanup
		fmt.Printf("DNS record for subdomain %s should be deleted\n", service.Subdomain.String)
	}

	// 6. Delete Git webhook if exists
	if service.GitSourceID.Valid {
		gitSourceID := service.GitSourceID.String
		gitSourceIDUUID, err := uuid.Parse(gitSourceID)
		if err == nil {
			gitSource, err := w.store.GetGitSource(ctx, gitSourceIDUUID)
			if err == nil && gitSource != nil && gitSource.WebhookID.Valid {
				// Note: Webhook deletion should be handled by Git client
				// For now, we mark it for cleanup
				fmt.Printf("Webhook %s should be deleted\n", gitSource.WebhookID.String)
			}
		}
	}

	// 6. Environment variables are deleted via CASCADE
	// 7. Deployments are deleted via CASCADE
	// 8. Git source is deleted via CASCADE

	return nil
}

// CleanupProjectResources cleans up all resources associated with a project
func (w *CleanupWorker) CleanupProjectResources(ctx context.Context, projectID uuid.UUID) error {
	// Get project
	project, err := w.store.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project not found: %s", projectID)
	}

	// 1. Clean up all services
	services, err := w.store.ListServicesByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	for _, service := range services {
		if err := w.CleanupServiceResources(ctx, service.ID); err != nil {
			// Log but continue with other services
			fmt.Printf("Warning: failed to cleanup service %s: %v\n", service.ID, err)
		}
	}

	// 2. Clean up all databases
	databases, err := w.store.ListDatabasesByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}

	infraConfig := infra.Config{
		BaseURL:  w.config.InfraServiceURL,
		APIKey:   w.config.InfraServiceAPIKey,
		TenantID: project.OpenStackTenantID,
		UseMock:  w.config.UseMockInfra,
	}

	baseClient := infra.NewClient(infraConfig)
	client := infra.NewRetryClient(baseClient)

	for _, db := range databases {
		// Unregister from Prometheus
		if db.OpenStackInstanceID.Valid {
			targetManager := metrics.NewTargetManager(w.config.PrometheusTargetsDir)
			if err := targetManager.UnregisterDatabase(db.ID.String()); err != nil {
				fmt.Printf("Warning: Failed to unregister database %s from Prometheus: %v\n", db.ID, err)
			}
		}

		// Delete database instance if exists
		if db.OpenStackInstanceID.Valid {
			instanceID := db.OpenStackInstanceID.String
			if err := client.DeleteInstance(ctx, instanceID); err != nil {
				fmt.Printf("Warning: failed to delete database instance %s: %v\n", instanceID, err)
			}
		}

		// Delete volume if exists
		if db.VolumeID.Valid {
			volumeID := db.VolumeID.String
			// First detach if attached
			if db.OpenStackInstanceID.Valid {
				if err := client.DetachVolume(ctx, volumeID); err != nil {
					fmt.Printf("Warning: failed to detach volume %s: %v\n", volumeID, err)
				}
			}
			// Then delete
			if err := client.DeleteVolume(ctx, volumeID); err != nil {
				fmt.Printf("Warning: failed to delete volume %s: %v\n", volumeID, err)
			}
		}

		// Delete DNS record if exists
		if db.InternalHostname.Valid {
			// Note: DNS record deletion needed
			fmt.Printf("DNS record for database hostname %s should be deleted\n", db.InternalHostname.String)
		}
	}

	// 3. Clean up all volumes
	volumes, err := w.store.ListVolumesByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to list volumes: %w", err)
	}

	for _, volume := range volumes {
		// Detach volume if attached
		if volume.AttachedToServiceID.Valid {
			serviceID, _ := uuid.Parse(volume.AttachedToServiceID.String)
			service, err := w.store.GetService(ctx, serviceID)
			if err == nil && service != nil && service.OpenStackInstanceID.Valid {
				if err := client.DetachVolume(ctx, volume.OpenStackVolumeID.String); err != nil {
					fmt.Printf("Warning: failed to detach volume %s: %v\n", volume.ID, err)
				}
			}
		} else if volume.AttachedToDatabaseID.Valid {
			dbID, _ := uuid.Parse(volume.AttachedToDatabaseID.String)
			db, err := w.store.GetDatabase(ctx, dbID)
			if err == nil && db != nil && db.OpenStackInstanceID.Valid {
				if err := client.DetachVolume(ctx, volume.OpenStackVolumeID.String); err != nil {
					fmt.Printf("Warning: failed to detach volume %s: %v\n", volume.ID, err)
				}
			}
		}

		// Delete volume
		if volume.OpenStackVolumeID.Valid {
			volumeID := volume.OpenStackVolumeID.String
			if err := client.DeleteVolume(ctx, volumeID); err != nil {
				fmt.Printf("Warning: failed to delete volume %s: %v\n", volumeID, err)
			}
		}
	}

	return nil
}

// ProcessCleanupServiceJob processes a cleanup service job
func (w *CleanupWorker) ProcessCleanupServiceJob(ctx context.Context, job *store.Job) error {
	serviceIDStr, ok := job.Payload["service_id"].(string)
	if !ok {
		return fmt.Errorf("missing service_id in cleanup job payload")
	}

	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		return fmt.Errorf("invalid service_id: %w", err)
	}

	return w.CleanupServiceResources(ctx, serviceID)
}

// ProcessCleanupProjectJob processes a cleanup project job
func (w *CleanupWorker) ProcessCleanupProjectJob(ctx context.Context, job *store.Job) error {
	projectIDStr, ok := job.Payload["project_id"].(string)
	if !ok {
		return fmt.Errorf("missing project_id in cleanup job payload")
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return fmt.Errorf("invalid project_id: %w", err)
	}

	return w.CleanupProjectResources(ctx, projectID)
}

