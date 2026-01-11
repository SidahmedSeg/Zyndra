package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/k8s"
	"github.com/intelifox/click-deploy/internal/store"
)

// K8sVolumeWorker handles volume management on k8s
type K8sVolumeWorker struct {
	store     *store.DB
	k8sClient *k8s.Client
}

// NewK8sVolumeWorker creates a new k8s volume worker
func NewK8sVolumeWorker(store *store.DB, k8sClient *k8s.Client) *K8sVolumeWorker {
	return &K8sVolumeWorker{
		store:     store,
		k8sClient: k8sClient,
	}
}

// CreateVolume creates a PVC on k8s
func (w *K8sVolumeWorker) CreateVolume(ctx context.Context, volumeID uuid.UUID) error {
	// Get volume
	vol, err := w.store.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	if vol == nil {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	// Get project
	project, err := w.store.GetProject(ctx, vol.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Ensure namespace exists
	if err := w.k8sClient.CreateNamespace(ctx, project.ID.String(), project.Name); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Create PVC
	spec := k8s.PVCSpec{
		VolumeID:   volumeID.String(),
		VolumeName: vol.Name,
		ProjectID:  project.ID.String(),
		SizeMB:     int64(vol.SizeMB),
	}

	_, err = w.k8sClient.CreatePVC(ctx, spec)
	if err != nil {
		w.store.UpdateVolumeStatus(ctx, volumeID, "failed")
		return fmt.Errorf("failed to create PVC: %w", err)
	}

	// Wait for PVC to be bound
	readyCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if err := w.waitForPVCBound(readyCtx, project.ID.String(), volumeID.String()); err != nil {
		w.store.UpdateVolumeStatus(ctx, volumeID, "failed")
		return fmt.Errorf("PVC failed to bind: %w", err)
	}

	// Update volume status
	w.store.UpdateVolumeStatus(ctx, volumeID, "available")

	return nil
}

// waitForPVCBound polls the PVC status until it's bound
func (w *K8sVolumeWorker) waitForPVCBound(ctx context.Context, projectID, volumeID string) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := w.k8sClient.GetPVCStatus(ctx, projectID, volumeID)
			if err != nil {
				return fmt.Errorf("failed to get PVC status: %w", err)
			}

			if status.Bound {
				return nil
			}
		}
	}
}

// ResizeVolume resizes a PVC on k8s
func (w *K8sVolumeWorker) ResizeVolume(ctx context.Context, volumeID uuid.UUID, newSizeMB int) error {
	// Get volume
	vol, err := w.store.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	if vol == nil {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	// Get project
	project, err := w.store.GetProject(ctx, vol.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Resize PVC
	if err := w.k8sClient.ResizePVC(ctx, project.ID.String(), volumeID.String(), int64(newSizeMB)); err != nil {
		return fmt.Errorf("failed to resize PVC: %w", err)
	}

	// Update volume in database
	w.store.UpdateVolumeSize(ctx, volumeID, newSizeMB)

	return nil
}

// DeleteVolume deletes a PVC from k8s
func (w *K8sVolumeWorker) DeleteVolume(ctx context.Context, volumeID uuid.UUID) error {
	// Get volume
	vol, err := w.store.GetVolume(ctx, volumeID)
	if err != nil {
		return fmt.Errorf("failed to get volume: %w", err)
	}
	if vol == nil {
		return nil // Already deleted
	}

	// Get project
	project, err := w.store.GetProject(ctx, vol.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Delete PVC
	if err := w.k8sClient.DeletePVC(ctx, project.ID.String(), volumeID.String()); err != nil {
		return fmt.Errorf("failed to delete PVC: %w", err)
	}

	// Update status
	w.store.UpdateVolumeStatus(ctx, volumeID, "deleted")

	return nil
}

// AttachVolumeToService updates a deployment to mount a volume
func (w *K8sVolumeWorker) AttachVolumeToService(ctx context.Context, volumeID, serviceID uuid.UUID, mountPath string) error {
	// This would require updating the deployment spec to include the volume mount
	// For now, this is handled by the deployment worker when it creates/updates deployments
	
	// Update the volume record in the database
	w.store.AttachVolume(ctx, volumeID, &serviceID, nil) // Attach to service
	
	return nil
}

// DetachVolumeFromService removes a volume mount from a deployment
func (w *K8sVolumeWorker) DetachVolumeFromService(ctx context.Context, volumeID uuid.UUID) error {
	// Detach in database
	w.store.DetachVolume(ctx, volumeID)
	
	return nil
}

