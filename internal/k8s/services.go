package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ServiceSpec defines the specification for a Kubernetes Service
type ServiceSpec struct {
	ServiceID   string
	ServiceName string
	ProjectID   string
	Port        int32
	TargetPort  int32
}

// CreateService creates a Kubernetes Service for a deployment
func (c *Client) CreateService(ctx context.Context, spec ServiceSpec) (*corev1.Service, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	serviceName := c.serviceName(spec.ServiceID)

	targetPort := spec.TargetPort
	if targetPort == 0 {
		targetPort = spec.Port
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels:    c.buildLabels(spec.ServiceID, spec.ServiceName, spec.ProjectID),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"zyndra.io/service-id": spec.ServiceID,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       spec.Port,
					TargetPort: intstr.FromInt32(targetPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	result, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return result, nil
}

// UpdateService updates an existing Kubernetes Service
func (c *Client) UpdateService(ctx context.Context, spec ServiceSpec) (*corev1.Service, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	serviceName := c.serviceName(spec.ServiceID)

	existing, err := c.clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	targetPort := spec.TargetPort
	if targetPort == 0 {
		targetPort = spec.Port
	}

	existing.Spec.Ports[0].Port = spec.Port
	existing.Spec.Ports[0].TargetPort = intstr.FromInt32(targetPort)

	result, err := c.clientset.CoreV1().Services(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	return result, nil
}

// DeleteService deletes a Kubernetes Service
func (c *Client) DeleteService(ctx context.Context, projectID, serviceID string) error {
	namespace := c.ProjectNamespace(projectID)
	serviceName := c.serviceName(serviceID)

	err := c.clientset.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// GetService retrieves a Kubernetes Service
func (c *Client) GetService(ctx context.Context, projectID, serviceID string) (*corev1.Service, error) {
	namespace := c.ProjectNamespace(projectID)
	serviceName := c.serviceName(serviceID)

	return c.clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
}

func (c *Client) serviceName(serviceID string) string {
	return "svc-" + serviceID[:8]
}

