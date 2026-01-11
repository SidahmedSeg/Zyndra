package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeploymentSpec defines the specification for a service deployment
type DeploymentSpec struct {
	// Identifiers
	ServiceID   string
	ServiceName string
	ProjectID   string
	
	// Container
	Image       string
	Port        int32
	Replicas    int32
	
	// Resources
	CPURequest    string // e.g., "100m"
	CPULimit      string // e.g., "500m"
	MemoryRequest string // e.g., "128Mi"
	MemoryLimit   string // e.g., "512Mi"
	
	// Environment variables (from Secret)
	EnvSecretName string
	
	// Volume mounts
	VolumeMounts []VolumeMount
	
	// Health checks
	HealthCheckPath string
	HealthCheckPort int32
}

// VolumeMount defines a volume to mount in the container
type VolumeMount struct {
	Name      string
	MountPath string
	PVCName   string
}

// CreateDeployment creates a Kubernetes Deployment for a service
func (c *Client) CreateDeployment(ctx context.Context, spec DeploymentSpec) (*appsv1.Deployment, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	deploymentName := c.deploymentName(spec.ServiceID)

	// Build container spec
	container := corev1.Container{
		Name:  spec.ServiceName,
		Image: spec.Image,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: spec.Port,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Resources: c.buildResourceRequirements(spec),
	}

	// Add environment variables from secret
	if spec.EnvSecretName != "" {
		container.EnvFrom = []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: spec.EnvSecretName,
					},
				},
			},
		}
	}

	// Add volume mounts
	if len(spec.VolumeMounts) > 0 {
		for _, vm := range spec.VolumeMounts {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      vm.Name,
				MountPath: vm.MountPath,
			})
		}
	}

	// Add health checks if path specified
	if spec.HealthCheckPath != "" {
		port := spec.HealthCheckPort
		if port == 0 {
			port = spec.Port
		}
		container.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: spec.HealthCheckPath,
					Port: intstr.FromInt32(port),
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
			FailureThreshold:    3,
		}
		container.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: spec.HealthCheckPath,
					Port: intstr.FromInt32(port),
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
			FailureThreshold:    3,
		}
	}

	// Build pod spec
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container},
	}

	// Add volumes for PVCs
	if len(spec.VolumeMounts) > 0 {
		for _, vm := range spec.VolumeMounts {
			podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
				Name: vm.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: vm.PVCName,
					},
				},
			})
		}
	}

	// Build deployment
	replicas := spec.Replicas
	if replicas == 0 {
		replicas = 1
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels:    c.buildLabels(spec.ServiceID, spec.ServiceName, spec.ProjectID),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"zyndra.io/service-id": spec.ServiceID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: c.buildLabels(spec.ServiceID, spec.ServiceName, spec.ProjectID),
				},
				Spec: podSpec,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 0},
					MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
				},
			},
		},
	}

	result, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	return result, nil
}

// UpdateDeployment updates an existing deployment
func (c *Client) UpdateDeployment(ctx context.Context, spec DeploymentSpec) (*appsv1.Deployment, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	deploymentName := c.deploymentName(spec.ServiceID)

	// Get existing deployment
	existing, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Update image
	existing.Spec.Template.Spec.Containers[0].Image = spec.Image

	// Update resources if specified
	if spec.CPURequest != "" || spec.MemoryRequest != "" {
		existing.Spec.Template.Spec.Containers[0].Resources = c.buildResourceRequirements(spec)
	}

	// Update replicas if specified
	if spec.Replicas > 0 {
		existing.Spec.Replicas = &spec.Replicas
	}

	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment: %w", err)
	}

	return result, nil
}

// DeleteDeployment deletes a deployment
func (c *Client) DeleteDeployment(ctx context.Context, projectID, serviceID string) error {
	namespace := c.ProjectNamespace(projectID)
	deploymentName := c.deploymentName(serviceID)

	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	return nil
}

// GetDeployment retrieves a deployment
func (c *Client) GetDeployment(ctx context.Context, projectID, serviceID string) (*appsv1.Deployment, error) {
	namespace := c.ProjectNamespace(projectID)
	deploymentName := c.deploymentName(serviceID)

	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
}

// GetDeploymentStatus returns the deployment status (ready replicas, etc.)
func (c *Client) GetDeploymentStatus(ctx context.Context, projectID, serviceID string) (*DeploymentStatus, error) {
	deployment, err := c.GetDeployment(ctx, projectID, serviceID)
	if err != nil {
		if errors.IsNotFound(err) {
			return &DeploymentStatus{Exists: false}, nil
		}
		return nil, err
	}

	return &DeploymentStatus{
		Exists:          true,
		Replicas:        deployment.Status.Replicas,
		ReadyReplicas:   deployment.Status.ReadyReplicas,
		UpdatedReplicas: deployment.Status.UpdatedReplicas,
		Available:       deployment.Status.ReadyReplicas > 0,
	}, nil
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus struct {
	Exists          bool
	Replicas        int32
	ReadyReplicas   int32
	UpdatedReplicas int32
	Available       bool
}

// ScaleDeployment scales a deployment to the specified number of replicas
func (c *Client) ScaleDeployment(ctx context.Context, projectID, serviceID string, replicas int32) error {
	namespace := c.ProjectNamespace(projectID)
	deploymentName := c.deploymentName(serviceID)

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = &replicas

	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

// RestartDeployment triggers a rolling restart of a deployment
func (c *Client) RestartDeployment(ctx context.Context, projectID, serviceID string) error {
	namespace := c.ProjectNamespace(projectID)
	deploymentName := c.deploymentName(serviceID)

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	// Add/update restart annotation to trigger rolling restart
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().Format("2006-01-02T15:04:05Z")

	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart deployment: %w", err)
	}

	return nil
}

// Helper functions

func (c *Client) deploymentName(serviceID string) string {
	return "svc-" + serviceID[:8] // Use first 8 chars of UUID for shorter names
}

func (c *Client) buildLabels(serviceID, serviceName, projectID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       serviceName,
		"app.kubernetes.io/managed-by": "zyndra",
		"zyndra.io/service-id":          serviceID,
		"zyndra.io/project-id":          projectID,
	}
}

func (c *Client) buildResourceRequirements(spec DeploymentSpec) corev1.ResourceRequirements {
	requirements := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}

	// Set defaults if not specified
	cpuRequest := spec.CPURequest
	if cpuRequest == "" {
		cpuRequest = "100m"
	}
	cpuLimit := spec.CPULimit
	if cpuLimit == "" {
		cpuLimit = "500m"
	}
	memoryRequest := spec.MemoryRequest
	if memoryRequest == "" {
		memoryRequest = "128Mi"
	}
	memoryLimit := spec.MemoryLimit
	if memoryLimit == "" {
		memoryLimit = "512Mi"
	}

	requirements.Requests[corev1.ResourceCPU] = resource.MustParse(cpuRequest)
	requirements.Requests[corev1.ResourceMemory] = resource.MustParse(memoryRequest)
	requirements.Limits[corev1.ResourceCPU] = resource.MustParse(cpuLimit)
	requirements.Limits[corev1.ResourceMemory] = resource.MustParse(memoryLimit)

	return requirements
}

