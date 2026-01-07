package worker

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/infra"
	"github.com/intelifox/click-deploy/internal/store"
)

// VolumeWorker processes volume management jobs
type VolumeWorker struct {
	store  *store.DB
	config *config.Config
	client infra.Client
}

// NewVolumeWorker creates a new volume worker
func NewVolumeWorker(store *store.DB, cfg *config.Config, client infra.Client) *VolumeWorker {
	return &VolumeWorker{
		store:  store,
		config: cfg,
		client: client,
	}
}

// ProcessCreateVolumeJob processes a volume creation job
func (w *VolumeWorker) ProcessCreateVolumeJob(ctx context.Context, volumeID uuid.UUID) error {
	// Get volume
	volume, err := w.store.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	if volume == nil {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	// Get project for OpenStack tenant ID
	project, err := w.store.GetProject(ctx, volume.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Create infra client config
	infraConfig := infra.Config{
		BaseURL:  w.config.InfraServiceURL,
		APIKey:   w.config.InfraServiceAPIKey,
		TenantID: project.OpenStackTenantID,
		UseMock:  w.config.UseMockInfra,
	}

	baseClient := infra.NewClient(infraConfig)
	client := infra.NewRetryClient(baseClient)

	// Create volume in OpenStack
	volumeSizeGB := (volume.SizeMB + 1023) / 1024 // Round up to GB
	volumeReq := infra.CreateVolumeRequest{
		Name:       volume.Name,
		SizeGB:     volumeSizeGB,
		VolumeType: volume.VolumeType,
	}

	openstackVolume, err := client.CreateVolume(ctx, volumeReq)
	if err != nil {
		w.store.UpdateVolume(ctx, volumeID, &store.Volume{
			Status: "error",
		})
		return fmt.Errorf("failed to create volume: %w", err)
	}

	// Update volume with OpenStack volume ID
	w.store.UpdateVolume(ctx, volumeID, &store.Volume{
		OpenStackVolumeID: sql.NullString{String: openstackVolume.ID, Valid: true},
		Status:            "available",
	})

	return nil
}

// ProcessAttachVolumeJob processes a volume attachment job
func (w *VolumeWorker) ProcessAttachVolumeJob(ctx context.Context, volumeID uuid.UUID, instanceID string, device string) error {
	// Get volume
	volume, err := w.store.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	if volume == nil {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	if !volume.OpenStackVolumeID.Valid {
		return fmt.Errorf("volume not yet created in OpenStack")
	}

	// Get project for OpenStack tenant ID
	project, err := w.store.GetProject(ctx, volume.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Create infra client config
	infraConfig := infra.Config{
		BaseURL:  w.config.InfraServiceURL,
		APIKey:   w.config.InfraServiceAPIKey,
		TenantID: project.OpenStackTenantID,
		UseMock:  w.config.UseMockInfra,
	}

	baseClient := infra.NewClient(infraConfig)
	client := infra.NewRetryClient(baseClient)

	// Attach volume
	if err := client.AttachVolume(ctx, volume.OpenStackVolumeID.String, instanceID, device); err != nil {
		return fmt.Errorf("failed to attach volume: %w", err)
	}

	// Update volume status
	w.store.UpdateVolume(ctx, volumeID, &store.Volume{
		Status: "attached",
	})

	return nil
}

// ProcessDetachVolumeJob processes a volume detachment job
func (w *VolumeWorker) ProcessDetachVolumeJob(ctx context.Context, volumeID uuid.UUID) error {
	// Get volume
	volume, err := w.store.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	if volume == nil {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	if !volume.OpenStackVolumeID.Valid {
		return fmt.Errorf("volume not yet created in OpenStack")
	}

	// Get project for OpenStack tenant ID
	project, err := w.store.GetProject(ctx, volume.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Create infra client config
	infraConfig := infra.Config{
		BaseURL:  w.config.InfraServiceURL,
		APIKey:   w.config.InfraServiceAPIKey,
		TenantID: project.OpenStackTenantID,
		UseMock:  w.config.UseMockInfra,
	}

	baseClient := infra.NewClient(infraConfig)
	client := infra.NewRetryClient(baseClient)

	// Detach volume (OpenStack API might need instance ID, but we'll try without)
	// In real implementation, we'd need to track which instance it's attached to
	if err := client.DetachVolume(ctx, volume.OpenStackVolumeID.String); err != nil {
		return fmt.Errorf("failed to detach volume: %w", err)
	}

	// Update volume status
	w.store.UpdateVolume(ctx, volumeID, &store.Volume{
		Status: "available",
	})

	return nil
}

// ProcessDeleteVolumeJob processes a volume deletion job
func (w *VolumeWorker) ProcessDeleteVolumeJob(ctx context.Context, volumeID uuid.UUID) error {
	// Get volume
	volume, err := w.store.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	if volume == nil {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	if !volume.OpenStackVolumeID.Valid {
		// Volume not created in OpenStack, just delete from DB
		return w.store.DeleteVolume(ctx, volumeID)
	}

	// Get project for OpenStack tenant ID
	project, err := w.store.GetProject(ctx, volume.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Create infra client config
	infraConfig := infra.Config{
		BaseURL:  w.config.InfraServiceURL,
		APIKey:   w.config.InfraServiceAPIKey,
		TenantID: project.OpenStackTenantID,
		UseMock:  w.config.UseMockInfra,
	}

	baseClient := infra.NewClient(infraConfig)
	client := infra.NewRetryClient(baseClient)

	// Delete volume from OpenStack
	if err := client.DeleteVolume(ctx, volume.OpenStackVolumeID.String); err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	// Delete from database
	return w.store.DeleteVolume(ctx, volumeID)
}

