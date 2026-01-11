package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/k8s"
	"github.com/intelifox/click-deploy/internal/store"
)

// K8sDeployWorker handles k8s deployments after builds complete
type K8sDeployWorker struct {
	store     *store.DB
	k8sClient *k8s.Client
}

// NewK8sDeployWorker creates a new k8s deployment worker
func NewK8sDeployWorker(store *store.DB, k8sClient *k8s.Client) *K8sDeployWorker {
	return &K8sDeployWorker{
		store:     store,
		k8sClient: k8sClient,
	}
}

// DeployToK8s deploys a service to Kubernetes after a successful build
func (w *K8sDeployWorker) DeployToK8s(ctx context.Context, deploymentID uuid.UUID) error {
	// Get deployment
	deployment, err := w.store.GetDeployment(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if deployment == nil {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	// Get service
	service, err := w.store.GetService(ctx, deployment.ServiceID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if service == nil {
		return fmt.Errorf("service not found: %s", deployment.ServiceID)
	}

	// Get project
	project, err := w.store.GetProject(ctx, service.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project not found: %s", service.ProjectID)
	}

	// Update deployment status to deploying
	w.store.UpdateDeploymentStatus(ctx, deploymentID, "deploying")
	w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "info", "Starting Kubernetes deployment", nil)

	// Ensure namespace exists
	if err := w.k8sClient.CreateNamespace(ctx, project.ID.String(), project.Name); err != nil {
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "error", fmt.Sprintf("Failed to create namespace: %v", err), nil)
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	projectID := project.ID.String()
	serviceID := service.ID.String()

	// Get environment variables for the service
	envVars, err := w.store.ListEnvVarsByService(ctx, service.ID)
	if err != nil {
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "warn", fmt.Sprintf("Failed to get env vars: %v", err), nil)
		envVars = []*store.EnvVar{} // Continue with empty env vars
	}

	// Create/update secret with environment variables
	envMap := make(map[string]string)
	for _, ev := range envVars {
		if ev.Value.Valid {
			envMap[ev.Key] = ev.Value.String
		}
	}

	if len(envMap) > 0 {
		_, err = w.k8sClient.UpdateSecret(ctx, k8s.SecretSpec{
			ServiceID:   serviceID,
			ServiceName: service.Name,
			ProjectID:   projectID,
			EnvVars:     envMap,
		})
		if err != nil {
			w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "error", fmt.Sprintf("Failed to create secret: %v", err), nil)
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	// Check if deployment exists
	deployStatus, err := w.k8sClient.GetDeploymentStatus(ctx, projectID, serviceID)
	if err != nil {
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "error", fmt.Sprintf("Failed to check deployment status: %v", err), nil)
		return fmt.Errorf("failed to check deployment status: %w", err)
	}

	imageTag := ""
	if service.CurrentImageTag.Valid {
		imageTag = service.CurrentImageTag.String
	} else {
		return fmt.Errorf("no image tag available for service")
	}

	deploySpec := k8s.DeploymentSpec{
		ServiceID:     serviceID,
		ServiceName:   service.Name,
		ProjectID:     projectID,
		Image:         imageTag,
		Port:          int32(service.Port),
		Replicas:      1,
		EnvSecretName: w.k8sClient.SecretName(serviceID),
		HealthCheckPath: "/health", // Default health check path
	}

	if deployStatus.Exists {
		// Update existing deployment
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "info", "Updating existing deployment", nil)
		_, err = w.k8sClient.UpdateDeployment(ctx, deploySpec)
	} else {
		// Create new deployment
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "info", "Creating new deployment", nil)
		_, err = w.k8sClient.CreateDeployment(ctx, deploySpec)
	}

	if err != nil {
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "error", fmt.Sprintf("Failed to deploy: %v", err), nil)
		w.store.UpdateDeploymentStatus(ctx, deploymentID, "failed")
		return fmt.Errorf("failed to deploy: %w", err)
	}

	// Create/update Service
	svcSpec := k8s.ServiceSpec{
		ServiceID:   serviceID,
		ServiceName: service.Name,
		ProjectID:   projectID,
		Port:        int32(service.Port),
		TargetPort:  int32(service.Port),
	}

	_, err = w.k8sClient.GetService(ctx, projectID, serviceID)
	if err != nil {
		// Service doesn't exist, create it
		_, err = w.k8sClient.CreateService(ctx, svcSpec)
		if err != nil {
			w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "error", fmt.Sprintf("Failed to create service: %v", err), nil)
			return fmt.Errorf("failed to create k8s service: %w", err)
		}
	}

	// Create/update Ingress
	environment := "prod" // Could be dynamic based on project environment
	ingressSpec := k8s.IngressSpec{
		ServiceID:   serviceID,
		ServiceName: service.Name,
		ProjectID:   projectID,
		Environment: environment,
		Port:        int32(service.Port),
	}

	// Get custom domains for this service
	customDomains, err := w.store.ListCustomDomainsByService(ctx, service.ID)
	if err == nil && len(customDomains) > 0 {
		for _, cd := range customDomains {
			if cd.Status == "active" {
				ingressSpec.CustomDomains = append(ingressSpec.CustomDomains, cd.Domain)
			}
		}
	}

	_, err = w.k8sClient.GetIngress(ctx, projectID, serviceID)
	if err != nil {
		// Ingress doesn't exist, create it
		_, err = w.k8sClient.CreateIngress(ctx, ingressSpec)
	} else {
		// Update existing ingress
		_, err = w.k8sClient.UpdateIngress(ctx, ingressSpec)
	}

	if err != nil {
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "warn", fmt.Sprintf("Failed to configure ingress: %v", err), nil)
		// Don't fail deployment for ingress issues
	}

	// Wait for deployment to be ready
	w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "info", "Waiting for deployment to be ready", nil)
	
	readyCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := w.waitForDeploymentReady(readyCtx, projectID, serviceID, deploymentID); err != nil {
		w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "error", fmt.Sprintf("Deployment failed to become ready: %v", err), nil)
		w.store.UpdateDeploymentStatus(ctx, deploymentID, "failed")
		return fmt.Errorf("deployment failed to become ready: %w", err)
	}

	// Update service status and URL
	generatedURL := w.k8sClient.GetServiceURL(service.Name, environment)
	if service.GeneratedURL.Valid {
		service.GeneratedURL.String = generatedURL
	}
	service.Status = "running"
	w.store.UpdateService(ctx, service.ID, service)

	// Update deployment status
	w.store.UpdateDeploymentStatus(ctx, deploymentID, "success")
	w.store.UpdateDeploymentProgress(ctx, deploymentID, map[string]interface{}{
		"finished_at": time.Now(),
	})
	w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "info", 
		fmt.Sprintf("Deployment successful! Service available at %s", generatedURL), nil)

	return nil
}

// waitForDeploymentReady polls the deployment status until it's ready
func (w *K8sDeployWorker) waitForDeploymentReady(ctx context.Context, projectID, serviceID string, deploymentID uuid.UUID) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := w.k8sClient.GetDeploymentStatus(ctx, projectID, serviceID)
			if err != nil {
				return fmt.Errorf("failed to get deployment status: %w", err)
			}

			if status.Available {
				return nil
			}

			w.store.AddDeploymentLog(ctx, deploymentID, "deploy", "info",
				fmt.Sprintf("Waiting for pods... (%d/%d ready)", status.ReadyReplicas, status.Replicas), nil)
		}
	}
}

// CleanupK8sResources removes all k8s resources for a service
func (w *K8sDeployWorker) CleanupK8sResources(ctx context.Context, projectID, serviceID string) error {
	var errs []error

	// Delete Ingress
	if err := w.k8sClient.DeleteIngress(ctx, projectID, serviceID); err != nil {
		errs = append(errs, fmt.Errorf("ingress: %w", err))
	}

	// Delete Service
	if err := w.k8sClient.DeleteService(ctx, projectID, serviceID); err != nil {
		errs = append(errs, fmt.Errorf("service: %w", err))
	}

	// Delete Deployment
	if err := w.k8sClient.DeleteDeployment(ctx, projectID, serviceID); err != nil {
		errs = append(errs, fmt.Errorf("deployment: %w", err))
	}

	// Delete Secret
	if err := w.k8sClient.DeleteSecret(ctx, projectID, serviceID); err != nil {
		errs = append(errs, fmt.Errorf("secret: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// CleanupK8sProject removes the entire namespace for a project
func (w *K8sDeployWorker) CleanupK8sProject(ctx context.Context, projectID string) error {
	return w.k8sClient.DeleteNamespace(ctx, projectID)
}

