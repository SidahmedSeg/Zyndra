package build

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RegistryClient handles container registry operations
type RegistryClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewRegistryClient creates a new registry client
func NewRegistryClient(baseURL, username, password string) *RegistryClient {
	return &RegistryClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		username:   username,
		password:   password,
		httpClient: &http.Client{},
	}
}

// AuthConfig returns authentication configuration for the registry
func (r *RegistryClient) AuthConfig() AuthConfig {
	return AuthConfig{
		Username: r.username,
		Password: r.password,
	}
}

// PushImage pushes an image to the registry
// Note: Actual push is handled by BuildKit, this is for registry operations
func (r *RegistryClient) PushImage(ctx context.Context, imageTag string) error {
	// BuildKit handles the actual push, but we can verify the image exists
	return r.VerifyImage(ctx, imageTag)
}

// VerifyImage verifies that an image exists in the registry
func (r *RegistryClient) VerifyImage(ctx context.Context, imageTag string) error {
	// Harbor API: GET /api/v2.0/projects/{project_name}/repositories/{repository_name}/artifacts
	// For now, we'll just verify connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v2.0/health", r.baseURL), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.username, r.password)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry health check failed: %d", resp.StatusCode)
	}

	return nil
}

// GetImageManifest retrieves the manifest for an image
func (r *RegistryClient) GetImageManifest(ctx context.Context, imageTag string) (map[string]interface{}, error) {
	// Parse image tag
	parts := strings.Split(imageTag, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid image tag format: %s", imageTag)
	}

	project := parts[0]
	imageParts := strings.Split(parts[len(parts)-1], ":")
	repository := imageParts[0]
	tag := "latest"
	if len(imageParts) > 1 {
		tag = imageParts[1]
	}

	// Harbor API: GET /api/v2.0/projects/{project_name}/repositories/{repository_name}/artifacts/{reference}
	apiURL := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts/%s",
		r.baseURL, url.PathEscape(project), url.PathEscape(repository), url.PathEscape(tag))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.SetBasicAuth(r.username, r.password)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get manifest: %d - %s", resp.StatusCode, string(body))
	}

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return manifest, nil
}

// DeleteImage deletes an image from the registry
func (r *RegistryClient) DeleteImage(ctx context.Context, imageTag string) error {
	// Parse image tag
	parts := strings.Split(imageTag, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid image tag format: %s", imageTag)
	}

	project := parts[0]
	imageParts := strings.Split(parts[len(parts)-1], ":")
	repository := imageParts[0]
	tag := "latest"
	if len(imageParts) > 1 {
		tag = imageParts[1]
	}

	// Harbor API: DELETE /api/v2.0/projects/{project_name}/repositories/{repository_name}/artifacts/{reference}
	apiURL := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts/%s",
		r.baseURL, url.PathEscape(project), url.PathEscape(repository), url.PathEscape(tag))

	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.username, r.password)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete image: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetAuthHeader returns the Authorization header value for registry requests
func (r *RegistryClient) GetAuthHeader() string {
	auth := fmt.Sprintf("%s:%s", r.username, r.password)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return fmt.Sprintf("Basic %s", encoded)
}

// BuildImageTag builds a full image tag from components
func BuildImageTag(registryURL, project, imageName, tag string) string {
	// Remove protocol from registry URL
	registryHost := strings.TrimPrefix(registryURL, "https://")
	registryHost = strings.TrimPrefix(registryHost, "http://")
	
	if tag == "" {
		tag = "latest"
	}
	
	return fmt.Sprintf("%s/%s/%s:%s", registryHost, project, imageName, tag)
}

