package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretSpec defines the specification for a Kubernetes Secret
type SecretSpec struct {
	ServiceID   string
	ServiceName string
	ProjectID   string
	EnvVars     map[string]string // key-value pairs for environment variables
}

// CreateSecret creates a Kubernetes Secret for environment variables
func (c *Client) CreateSecret(ctx context.Context, spec SecretSpec) (*corev1.Secret, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	secretName := c.secretName(spec.ServiceID)

	// Convert string map to byte map
	data := make(map[string][]byte)
	for key, value := range spec.EnvVars {
		data[key] = []byte(value)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels:    c.buildLabels(spec.ServiceID, spec.ServiceName, spec.ProjectID),
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	result, err := c.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	return result, nil
}

// UpdateSecret updates an existing Kubernetes Secret
func (c *Client) UpdateSecret(ctx context.Context, spec SecretSpec) (*corev1.Secret, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	secretName := c.secretName(spec.ServiceID)

	existing, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		// If secret doesn't exist, create it
		if errors.IsNotFound(err) {
			return c.CreateSecret(ctx, spec)
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Convert string map to byte map
	data := make(map[string][]byte)
	for key, value := range spec.EnvVars {
		data[key] = []byte(value)
	}

	existing.Data = data

	result, err := c.clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	return result, nil
}

// SetEnvVar sets a single environment variable in a secret
func (c *Client) SetEnvVar(ctx context.Context, projectID, serviceID, key, value string) error {
	namespace := c.ProjectNamespace(projectID)
	secretName := c.secretName(serviceID)

	existing, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create secret with single env var
			_, err := c.CreateSecret(ctx, SecretSpec{
				ServiceID: serviceID,
				ProjectID: projectID,
				EnvVars:   map[string]string{key: value},
			})
			return err
		}
		return fmt.Errorf("failed to get secret: %w", err)
	}

	if existing.Data == nil {
		existing.Data = make(map[string][]byte)
	}
	existing.Data[key] = []byte(value)

	_, err = c.clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// DeleteEnvVar deletes a single environment variable from a secret
func (c *Client) DeleteEnvVar(ctx context.Context, projectID, serviceID, key string) error {
	namespace := c.ProjectNamespace(projectID)
	secretName := c.secretName(serviceID)

	existing, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil // Nothing to delete
		}
		return fmt.Errorf("failed to get secret: %w", err)
	}

	delete(existing.Data, key)

	_, err = c.clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// GetEnvVars retrieves all environment variables from a secret
func (c *Client) GetEnvVars(ctx context.Context, projectID, serviceID string) (map[string]string, error) {
	namespace := c.ProjectNamespace(projectID)
	secretName := c.secretName(serviceID)

	secret, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Convert byte map to string map
	result := make(map[string]string)
	for key, value := range secret.Data {
		result[key] = string(value)
	}

	return result, nil
}

// DeleteSecret deletes a Kubernetes Secret
func (c *Client) DeleteSecret(ctx context.Context, projectID, serviceID string) error {
	namespace := c.ProjectNamespace(projectID)
	secretName := c.secretName(serviceID)

	err := c.clientset.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// GetSecret retrieves a Kubernetes Secret
func (c *Client) GetSecret(ctx context.Context, projectID, serviceID string) (*corev1.Secret, error) {
	namespace := c.ProjectNamespace(projectID)
	secretName := c.secretName(serviceID)

	return c.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
}

// SecretName returns the secret name for a service
func (c *Client) SecretName(serviceID string) string {
	return c.secretName(serviceID)
}

func (c *Client) secretName(serviceID string) string {
	return "env-" + serviceID[:8]
}

