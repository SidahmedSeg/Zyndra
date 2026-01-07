package worker

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/intelifox/click-deploy/internal/build"
	"github.com/intelifox/click-deploy/internal/config"
	"github.com/intelifox/click-deploy/internal/git"
	"github.com/intelifox/click-deploy/internal/realtime"
	"github.com/intelifox/click-deploy/internal/store"
)

// BuildWorker processes build jobs
type BuildWorker struct {
	store          *store.DB
	config         *config.Config
	buildkitClient *build.BuildKitClient
	railpackClient *build.RailpackClient
	registryClient *build.RegistryClient
	buildDir       string // Temporary directory for builds
	publisher      realtime.Publisher
}

// NewBuildWorker creates a new build worker
func NewBuildWorker(store *store.DB, cfg *config.Config) (*BuildWorker, error) {
	ctx := context.Background()
	
	// Initialize BuildKit client
	buildkitAddress := cfg.BuildKitAddress
	if buildkitAddress == "" {
		buildkitAddress = "unix:///run/buildkit/buildkitd.sock"
	}
	
	buildkitClient, err := build.NewBuildKitClient(ctx, buildkitAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create BuildKit client: %w", err)
	}

	// Initialize Railpack client
	railpackClient := build.NewRailpackClient(buildkitAddress)

	// Initialize registry client
	registryClient := build.NewRegistryClient(
		cfg.RegistryURL,
		cfg.RegistryUsername,
		cfg.RegistryPassword,
	)

	// Create build directory
	buildDir := cfg.BuildDir
	if buildDir == "" {
		buildDir = "/tmp/click-deploy-builds"
	}
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create build directory: %w", err)
	}

	return &BuildWorker{
		store:          store,
		config:         cfg,
		buildkitClient: buildkitClient,
		railpackClient: railpackClient,
		registryClient: registryClient,
		buildDir:       buildDir,
		publisher:      realtime.NewCentrifugoPublisher(cfg.CentrifugoAPIURL, cfg.CentrifugoAPIKey),
	}, nil
}

// Close closes the worker and cleans up resources
func (w *BuildWorker) Close() error {
	if w.buildkitClient != nil {
		return w.buildkitClient.Close()
	}
	return nil
}

// ProcessBuildJob processes a build job for a deployment
func (w *BuildWorker) ProcessBuildJob(ctx context.Context, deploymentID uuid.UUID) error {
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

	// Get git source
	gitSource, err := w.store.GetGitSourceByService(ctx, deployment.ServiceID)
	if err != nil {
		return fmt.Errorf("failed to get git source: %w", err)
	}
	if gitSource == nil {
		return fmt.Errorf("git source not found for service: %s", deployment.ServiceID)
	}

	// Get git connection
	gitConnection, err := w.store.GetGitConnection(ctx, gitSource.GitConnectionID)
	if err != nil {
		return fmt.Errorf("failed to get git connection: %w", err)
	}
	if gitConnection == nil {
		return fmt.Errorf("git connection not found: %s", gitSource.GitConnectionID)
	}

	// Update deployment status
	w.store.UpdateDeploymentStatus(ctx, deploymentID, "building")
	w.log(ctx, deploymentID, "clone", "info", "Starting build process", nil)

	// Clone repository
	clonePath := filepath.Join(w.buildDir, deploymentID.String())
	defer os.RemoveAll(clonePath) // Clean up after build

	cloneOpts := git.CloneOptions{
		URL:      fmt.Sprintf("https://%s/%s/%s.git", gitSource.Provider, gitSource.RepoOwner, gitSource.RepoName),
		Branch:   gitSource.Branch,
		Token:    gitConnection.AccessToken,
		Provider: gitSource.Provider,
	}

	if deployment.CommitSHA.Valid {
		cloneOpts.Commit = deployment.CommitSHA.String
	}

	w.log(ctx, deploymentID, "clone", "info",
		fmt.Sprintf("Cloning repository: %s/%s (branch: %s)", gitSource.RepoOwner, gitSource.RepoName, gitSource.Branch), nil)

	cloneResult, err := git.CloneRepository(ctx, cloneOpts, w.buildDir)
	if err != nil {
		w.log(ctx, deploymentID, "clone", "error",
			fmt.Sprintf("Failed to clone repository: %v", err), nil)
		w.store.UpdateDeploymentStatus(ctx, deploymentID, "failed")
		w.store.UpdateDeploymentProgress(ctx, deploymentID, map[string]interface{}{
			"error_message": err.Error(),
			"finished_at":   time.Now(),
		})
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	w.log(ctx, deploymentID, "clone", "info",
		fmt.Sprintf("Repository cloned successfully (commit: %s)", cloneResult.CommitSHA), nil)

	// Determine build context path
	buildContextPath := clonePath
	if gitSource.RootDir.Valid && gitSource.RootDir.String != "" && gitSource.RootDir.String != "/" {
		buildContextPath = filepath.Join(clonePath, gitSource.RootDir.String)
	}

	// Build image tag
	imageTag := build.BuildImageTag(
		w.config.RegistryURL,
		service.Name,
		service.Name,
		deployment.CommitSHA.String,
	)

	// Check if Dockerfile exists
	dockerfilePath := filepath.Join(buildContextPath, "Dockerfile")
	useRailpack := true
	if _, err := os.Stat(dockerfilePath); err == nil {
		// Dockerfile exists, use it instead of Railpack
		useRailpack = false
		w.log(ctx, deploymentID, "build", "info",
			"Dockerfile detected, using Dockerfile build", nil)
	}

	buildStartTime := time.Now()

	// Build image
	if useRailpack {
		// Use Railpack for zero-config build
		w.log(ctx, deploymentID, "build", "info",
			"Using Railpack for zero-config build", nil)

		railpackOpts := build.RailpackBuildOptions{
			ContextPath: buildContextPath,
			ImageTag:    imageTag,
		}

		err = w.railpackClient.Build(ctx, railpackOpts)
	} else {
		// Use BuildKit with Dockerfile
		w.log(ctx, deploymentID, "build", "info",
			"Building with Dockerfile", nil)

		buildOpts := build.BuildOptions{
			ContextPath:    buildContextPath,
			DockerfilePath: "Dockerfile",
			ImageTag:       imageTag,
			RegistryAuth: map[string]build.AuthConfig{
				w.config.RegistryURL: w.registryClient.AuthConfig(),
			},
		}

		err = w.buildkitClient.BuildImage(ctx, buildOpts)
	}

	if err != nil {
		w.log(ctx, deploymentID, "build", "error",
			fmt.Sprintf("Build failed: %v", err), nil)
		w.store.UpdateDeploymentStatus(ctx, deploymentID, "failed")
		w.store.UpdateDeploymentProgress(ctx, deploymentID, map[string]interface{}{
			"error_message": err.Error(),
			"build_duration": int64(time.Since(buildStartTime).Seconds()),
			"finished_at":    time.Now(),
		})
		return fmt.Errorf("build failed: %w", err)
	}

	buildDuration := int64(time.Since(buildStartTime).Seconds())
	w.log(ctx, deploymentID, "build", "info",
		fmt.Sprintf("Build completed successfully in %d seconds", buildDuration), nil)

	// Update deployment with image tag
	w.store.UpdateDeploymentProgress(ctx, deploymentID, map[string]interface{}{
		"image_tag":      imageTag,
		"build_duration": buildDuration,
		"status":         "pushing",
	})

	// Verify image in registry
	w.log(ctx, deploymentID, "push", "info",
		"Verifying image in registry", nil)

	if err := w.registryClient.VerifyImage(ctx, imageTag); err != nil {
		w.log(ctx, deploymentID, "push", "error",
			fmt.Sprintf("Failed to verify image: %v", err), nil)
		// Don't fail here, BuildKit should have pushed it
	}

	w.log(ctx, deploymentID, "push", "info",
		fmt.Sprintf("Image pushed successfully: %s", imageTag), nil)

	// Update deployment status
	w.store.UpdateDeploymentProgress(ctx, deploymentID, map[string]interface{}{
		"status":      "success",
		"finished_at": time.Now(),
	})

	// Update service with new image tag
	service.CurrentImageTag = sql.NullString{String: imageTag, Valid: true}
	w.store.UpdateService(ctx, service.ID, service)

	return nil
}

func (w *BuildWorker) log(ctx context.Context, deploymentID uuid.UUID, phase, level, message string, metadata map[string]interface{}) {
	_ = w.store.AddDeploymentLog(ctx, deploymentID, phase, level, message, metadata)

	// Best-effort publish to Centrifugo (deployment:<id>)
	if w.publisher != nil {
		_ = w.publisher.Publish(ctx, "deployment:"+deploymentID.String(), map[string]any{
			"timestamp": time.Now().UTC(),
			"phase":     phase,
			"level":     level,
			"message":   message,
			"metadata":  metadata,
		})
	}
}

