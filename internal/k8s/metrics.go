package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// MetricsClient handles metrics from the k8s Metrics Server
type MetricsClient struct {
	client *metricsv.Clientset
	config Config
}

// PodMetrics represents CPU/Memory metrics for a pod
type PodMetrics struct {
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	CPUCores  float64 `json:"cpu_cores"`   // CPU in cores (e.g., 0.5 = 500m)
	MemoryMB  float64 `json:"memory_mb"`   // Memory in MB
	CPUUsage  string  `json:"cpu_usage"`   // Human-readable (e.g., "500m")
	MemoryUsage string `json:"memory_usage"` // Human-readable (e.g., "256Mi")
}

// ServiceMetrics represents aggregated metrics for a service
type ServiceMetrics struct {
	ServiceID    string       `json:"service_id"`
	ServiceName  string       `json:"service_name"`
	TotalCPU     float64      `json:"total_cpu_cores"`
	TotalMemory  float64      `json:"total_memory_mb"`
	PodCount     int          `json:"pod_count"`
	Pods         []PodMetrics `json:"pods"`
}

// NewMetricsClient creates a new metrics client
func NewMetricsClient(cfg Config) (*MetricsClient, error) {
	var restConfig *rest.Config
	var err error

	if cfg.InCluster {
		restConfig, err = rest.InClusterConfig()
	} else {
		kubeconfigPath := cfg.KubeconfigPath
		if kubeconfigPath == "" {
			// Try default paths
			kubeconfigPath = "/etc/rancher/k3s/k3s.yaml"
		}
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build metrics config: %w", err)
	}

	client, err := metricsv.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &MetricsClient{
		client: client,
		config: cfg,
	}, nil
}

// GetPodMetrics gets metrics for a specific pod
func (m *MetricsClient) GetPodMetrics(ctx context.Context, namespace, podName string) (*PodMetrics, error) {
	metrics, err := m.client.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	return m.convertPodMetrics(metrics), nil
}

// GetServiceMetrics gets aggregated metrics for all pods of a service
func (m *MetricsClient) GetServiceMetrics(ctx context.Context, projectID, serviceID, serviceName string) (*ServiceMetrics, error) {
	namespace := m.config.NamespacePrefix + projectID

	// List pods with the service label
	podMetricsList, err := m.client.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("zyndra.io/service-id=%s", serviceID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}

	result := &ServiceMetrics{
		ServiceID:   serviceID,
		ServiceName: serviceName,
		Pods:        make([]PodMetrics, 0),
	}

	for _, pm := range podMetricsList.Items {
		podMetric := m.convertPodMetrics(&pm)
		result.Pods = append(result.Pods, *podMetric)
		result.TotalCPU += podMetric.CPUCores
		result.TotalMemory += podMetric.MemoryMB
	}

	result.PodCount = len(result.Pods)

	return result, nil
}

// GetNamespaceMetrics gets metrics for all pods in a project namespace
func (m *MetricsClient) GetNamespaceMetrics(ctx context.Context, projectID string) ([]PodMetrics, error) {
	namespace := m.config.NamespacePrefix + projectID

	podMetricsList, err := m.client.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}

	result := make([]PodMetrics, 0, len(podMetricsList.Items))
	for _, pm := range podMetricsList.Items {
		result = append(result, *m.convertPodMetrics(&pm))
	}

	return result, nil
}

// GetNodeMetrics gets metrics for all nodes
func (m *MetricsClient) GetNodeMetrics(ctx context.Context) ([]NodeMetrics, error) {
	nodeMetricsList, err := m.client.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list node metrics: %w", err)
	}

	result := make([]NodeMetrics, 0, len(nodeMetricsList.Items))
	for _, nm := range nodeMetricsList.Items {
		cpuQuantity := nm.Usage.Cpu()
		memQuantity := nm.Usage.Memory()

		result = append(result, NodeMetrics{
			Name:        nm.Name,
			CPUCores:    float64(cpuQuantity.MilliValue()) / 1000,
			MemoryMB:    float64(memQuantity.Value()) / (1024 * 1024),
			CPUUsage:    cpuQuantity.String(),
			MemoryUsage: memQuantity.String(),
		})
	}

	return result, nil
}

// NodeMetrics represents metrics for a node
type NodeMetrics struct {
	Name        string  `json:"name"`
	CPUCores    float64 `json:"cpu_cores"`
	MemoryMB    float64 `json:"memory_mb"`
	CPUUsage    string  `json:"cpu_usage"`
	MemoryUsage string  `json:"memory_usage"`
}

// convertPodMetrics converts k8s PodMetrics to our PodMetrics
func (m *MetricsClient) convertPodMetrics(pm *metricsv1beta1.PodMetrics) *PodMetrics {
	var totalCPU, totalMem int64

	for _, container := range pm.Containers {
		cpuQuantity := container.Usage.Cpu()
		memQuantity := container.Usage.Memory()
		totalCPU += cpuQuantity.MilliValue()
		totalMem += memQuantity.Value()
	}

	return &PodMetrics{
		Name:        pm.Name,
		Namespace:   pm.Namespace,
		CPUCores:    float64(totalCPU) / 1000,
		MemoryMB:    float64(totalMem) / (1024 * 1024),
		CPUUsage:    fmt.Sprintf("%dm", totalCPU),
		MemoryUsage: fmt.Sprintf("%dMi", totalMem/(1024*1024)),
	}
}

// IsMetricsServerAvailable checks if the metrics server is available
func (m *MetricsClient) IsMetricsServerAvailable(ctx context.Context) bool {
	_, err := m.client.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	return err == nil
}

