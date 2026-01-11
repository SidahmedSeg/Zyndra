package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds the k8s client configuration
type Config struct {
	KubeconfigPath    string // Path to kubeconfig file (for local dev)
	InCluster         bool   // Use in-cluster config
	NamespacePrefix   string // Prefix for project namespaces (e.g., "zyndra-")
	DefaultRegistry   string // Default container registry
	BaseDomain        string // Base domain for generated URLs (e.g., "up.zyndra.app")
	IngressClass      string // Ingress class (e.g., "traefik")
	CertIssuer        string // cert-manager ClusterIssuer name
}

// Client wraps the Kubernetes clientset
type Client struct {
	clientset *kubernetes.Clientset
	config    Config
}

// NewClient creates a new Kubernetes client
func NewClient(cfg Config) (*Client, error) {
	var config *rest.Config
	var err error

	if cfg.InCluster {
		// Use in-cluster config (when running inside k8s)
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig file
		kubeconfigPath := cfg.KubeconfigPath
		if kubeconfigPath == "" {
			// Default to ~/.kube/config or /etc/rancher/k3s/k3s.yaml
			homeDir, _ := os.UserHomeDir()
			defaultPath := filepath.Join(homeDir, ".kube", "config")
			k3sPath := "/etc/rancher/k3s/k3s.yaml"
			
			if _, err := os.Stat(k3sPath); err == nil {
				kubeconfigPath = k3sPath
			} else if _, err := os.Stat(defaultPath); err == nil {
				kubeconfigPath = defaultPath
			} else {
				return nil, fmt.Errorf("no kubeconfig found at %s or %s", k3sPath, defaultPath)
			}
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Set defaults
	if cfg.NamespacePrefix == "" {
		cfg.NamespacePrefix = "zyndra-"
	}
	if cfg.IngressClass == "" {
		cfg.IngressClass = "traefik"
	}
	if cfg.CertIssuer == "" {
		cfg.CertIssuer = "letsencrypt-prod"
	}

	return &Client{
		clientset: clientset,
		config:    cfg,
	}, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() Config {
	return c.config
}

// ProjectNamespace returns the namespace name for a project
func (c *Client) ProjectNamespace(projectID string) string {
	return c.config.NamespacePrefix + projectID
}

// CreateNamespace creates a namespace for a project
func (c *Client) CreateNamespace(ctx context.Context, projectID, projectName string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.ProjectNamespace(projectID),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "zyndra",
				"zyndra.io/project-id":          projectID,
				"zyndra.io/project-name":        projectName,
			},
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	return nil
}

// DeleteNamespace deletes a project's namespace and all resources in it
func (c *Client) DeleteNamespace(ctx context.Context, projectID string) error {
	err := c.clientset.CoreV1().Namespaces().Delete(ctx, c.ProjectNamespace(projectID), metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}
	return nil
}

// GetNamespace retrieves a project's namespace
func (c *Client) GetNamespace(ctx context.Context, projectID string) (*corev1.Namespace, error) {
	ns, err := c.clientset.CoreV1().Namespaces().Get(ctx, c.ProjectNamespace(projectID), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}
	return ns, nil
}

// NamespaceExists checks if a project's namespace exists
func (c *Client) NamespaceExists(ctx context.Context, projectID string) (bool, error) {
	_, err := c.GetNamespace(ctx, projectID)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListDeployments lists all deployments in a project's namespace
func (c *Client) ListDeployments(ctx context.Context, projectID string) (*appsv1.DeploymentList, error) {
	return c.clientset.AppsV1().Deployments(c.ProjectNamespace(projectID)).List(ctx, metav1.ListOptions{})
}

// ListServices lists all services in a project's namespace
func (c *Client) ListServices(ctx context.Context, projectID string) (*corev1.ServiceList, error) {
	return c.clientset.CoreV1().Services(c.ProjectNamespace(projectID)).List(ctx, metav1.ListOptions{})
}

// ListIngresses lists all ingresses in a project's namespace
func (c *Client) ListIngresses(ctx context.Context, projectID string) (*networkingv1.IngressList, error) {
	return c.clientset.NetworkingV1().Ingresses(c.ProjectNamespace(projectID)).List(ctx, metav1.ListOptions{})
}

// Ping checks if the Kubernetes API is reachable
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes: %w", err)
	}
	return nil
}

