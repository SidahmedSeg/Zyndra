package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVCSpec defines the specification for a PersistentVolumeClaim
type PVCSpec struct {
	VolumeID     string
	VolumeName   string
	ProjectID    string
	SizeMB       int64  // Size in megabytes
	StorageClass string // e.g., "longhorn" for Longhorn volumes
	AccessMode   string // ReadWriteOnce, ReadWriteMany, ReadOnlyMany
}

// CreatePVC creates a PersistentVolumeClaim
func (c *Client) CreatePVC(ctx context.Context, spec PVCSpec) (*corev1.PersistentVolumeClaim, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	pvcName := c.pvcName(spec.VolumeID)

	// Set defaults
	storageClass := spec.StorageClass
	if storageClass == "" {
		storageClass = "longhorn" // Default to Longhorn
	}

	accessMode := corev1.ReadWriteOnce
	if spec.AccessMode == "ReadWriteMany" {
		accessMode = corev1.ReadWriteMany
	} else if spec.AccessMode == "ReadOnlyMany" {
		accessMode = corev1.ReadOnlyMany
	}

	// Convert MB to resource.Quantity
	sizeStr := fmt.Sprintf("%dMi", spec.SizeMB)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "zyndra",
				"zyndra.io/volume-id":           spec.VolumeID,
				"zyndra.io/volume-name":         spec.VolumeName,
				"zyndra.io/project-id":          spec.ProjectID,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{accessMode},
			StorageClassName: &storageClass,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(sizeStr),
				},
			},
		},
	}

	result, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC: %w", err)
	}

	return result, nil
}

// ResizePVC resizes a PersistentVolumeClaim (requires storage class support)
func (c *Client) ResizePVC(ctx context.Context, projectID, volumeID string, newSizeMB int64) error {
	namespace := c.ProjectNamespace(projectID)
	pvcName := c.pvcName(volumeID)

	pvc, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get PVC: %w", err)
	}

	sizeStr := fmt.Sprintf("%dMi", newSizeMB)
	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(sizeStr)

	_, err = c.clientset.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to resize PVC: %w", err)
	}

	return nil
}

// DeletePVC deletes a PersistentVolumeClaim
func (c *Client) DeletePVC(ctx context.Context, projectID, volumeID string) error {
	namespace := c.ProjectNamespace(projectID)
	pvcName := c.pvcName(volumeID)

	err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete PVC: %w", err)
	}

	return nil
}

// GetPVC retrieves a PersistentVolumeClaim
func (c *Client) GetPVC(ctx context.Context, projectID, volumeID string) (*corev1.PersistentVolumeClaim, error) {
	namespace := c.ProjectNamespace(projectID)
	pvcName := c.pvcName(volumeID)

	return c.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
}

// GetPVCStatus returns the status of a PVC
func (c *Client) GetPVCStatus(ctx context.Context, projectID, volumeID string) (*PVCStatus, error) {
	pvc, err := c.GetPVC(ctx, projectID, volumeID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &PVCStatus{Exists: false}, nil
		}
		return nil, err
	}

	var capacityMB int64
	if cap, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
		capacityMB = cap.Value() / (1024 * 1024)
	}

	return &PVCStatus{
		Exists:     true,
		Phase:      string(pvc.Status.Phase),
		Bound:      pvc.Status.Phase == corev1.ClaimBound,
		CapacityMB: capacityMB,
	}, nil
}

// PVCStatus represents the status of a PVC
type PVCStatus struct {
	Exists     bool
	Phase      string // Pending, Bound, Lost
	Bound      bool
	CapacityMB int64
}

// ListPVCs lists all PVCs in a project's namespace
func (c *Client) ListPVCs(ctx context.Context, projectID string) (*corev1.PersistentVolumeClaimList, error) {
	namespace := c.ProjectNamespace(projectID)
	return c.clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/managed-by=zyndra",
	})
}

// PVCName returns the PVC name for a volume
func (c *Client) PVCName(volumeID string) string {
	return c.pvcName(volumeID)
}

func (c *Client) pvcName(volumeID string) string {
	return "vol-" + volumeID[:8]
}

