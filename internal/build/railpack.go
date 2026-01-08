package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RailpackClient handles Railpack builds
// Railpack is a zero-config build tool that auto-detects runtimes
type RailpackClient struct {
	buildkitAddress string
}

// NewRailpackClient creates a new Railpack client
func NewRailpackClient(buildkitAddress string) *RailpackClient {
	return &RailpackClient{
		buildkitAddress: buildkitAddress,
	}
}

// BuildOptions specifies options for a Railpack build
type RailpackBuildOptions struct {
	ContextPath    string            // Path to repository
	ImageTag       string            // Full image tag (registry/image:tag)
	BuildCommand   string            // Optional: override build command
	StartCommand   string            // Optional: override start command
	InstallCommand string            // Optional: override install command
	BuildArgs      map[string]string // Build arguments
	EnvVars        map[string]string // Environment variables for build
}

// DetectRuntime detects the runtime from the repository
func (r *RailpackClient) DetectRuntime(ctx context.Context, contextPath string) (string, error) {
	// Check for runtime indicators
	if _, err := os.Stat(filepath.Join(contextPath, "package.json")); err == nil {
		return "nodejs", nil
	}
	if _, err := os.Stat(filepath.Join(contextPath, "go.mod")); err == nil {
		return "go", nil
	}
	if _, err := os.Stat(filepath.Join(contextPath, "requirements.txt")); err == nil {
		return "python", nil
	}
	if _, err := os.Stat(filepath.Join(contextPath, "composer.json")); err == nil {
		return "php", nil
	}
	if _, err := os.Stat(filepath.Join(contextPath, "Gemfile")); err == nil {
		return "ruby", nil
	}
	if _, err := os.Stat(filepath.Join(contextPath, "index.html")); err == nil {
		return "static", nil
	}
	if _, err := os.Stat(filepath.Join(contextPath, "Staticfile")); err == nil {
		return "static", nil
	}

	return "", fmt.Errorf("unable to detect runtime")
}

// Build builds a container image using Railpack
// Railpack uses BuildKit under the hood, so we'll generate a Dockerfile
// and then use BuildKit to build it
func (r *RailpackClient) Build(ctx context.Context, opts RailpackBuildOptions) error {
	// Detect runtime
	runtime, err := r.DetectRuntime(ctx, opts.ContextPath)
	if err != nil {
		return fmt.Errorf("failed to detect runtime: %w", err)
	}

	// Generate Dockerfile based on runtime
	dockerfile, err := r.generateDockerfile(runtime, opts)
	if err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Write Dockerfile to context
	dockerfilePath := filepath.Join(opts.ContextPath, "Dockerfile.railpack")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}
	defer os.Remove(dockerfilePath) // Clean up after build

	// Use BuildKit to build the image
	buildkit, err := NewBuildKitClient(ctx, r.buildkitAddress)
	if err != nil {
		return fmt.Errorf("failed to create BuildKit client: %w", err)
	}
	defer buildkit.Close()

	buildOpts := BuildOptions{
		ContextPath:    opts.ContextPath,
		DockerfilePath: "Dockerfile.railpack",
		ImageTag:       opts.ImageTag,
		BuildArgs:      opts.BuildArgs,
	}

	return buildkit.BuildImage(ctx, buildOpts)
}

// generateDockerfile generates a Dockerfile based on runtime
func (r *RailpackClient) generateDockerfile(runtime string, opts RailpackBuildOptions) (string, error) {
	var dockerfile strings.Builder

	switch runtime {
	case "nodejs":
		dockerfile.WriteString("FROM node:20-alpine AS builder\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY package*.json ./\n")
		
		installCmd := opts.InstallCommand
		if installCmd == "" {
			installCmd = "npm ci"
		}
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", installCmd))
		
		dockerfile.WriteString("COPY . .\n")
		
		buildCmd := opts.BuildCommand
		if buildCmd == "" {
			dockerfile.WriteString("RUN npm run build || true\n")
		} else {
			dockerfile.WriteString(fmt.Sprintf("RUN %s\n", buildCmd))
		}
		
		dockerfile.WriteString("FROM node:20-alpine\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY --from=builder /app/node_modules ./node_modules\n")
		dockerfile.WriteString("COPY --from=builder /app .\n")
		
		startCmd := opts.StartCommand
		if startCmd == "" {
			startCmd = "npm start"
		}
		dockerfile.WriteString(fmt.Sprintf("CMD [\"sh\", \"-c\", \"%s\"]\n", startCmd))

	case "go":
		dockerfile.WriteString("FROM golang:1.22-alpine AS builder\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY go.mod go.sum ./\n")
		dockerfile.WriteString("RUN go mod download\n")
		dockerfile.WriteString("COPY . .\n")
		
		buildCmd := opts.BuildCommand
		if buildCmd == "" {
			dockerfile.WriteString("RUN go build -o /app/server ./...\n")
		} else {
			dockerfile.WriteString(fmt.Sprintf("RUN %s\n", buildCmd))
		}
		
		dockerfile.WriteString("FROM alpine:latest\n")
		dockerfile.WriteString("RUN apk --no-cache add ca-certificates\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY --from=builder /app/server .\n")
		
		startCmd := opts.StartCommand
		if startCmd == "" {
			startCmd = "./server"
		}
		dockerfile.WriteString(fmt.Sprintf("CMD [\"%s\"]\n", startCmd))

	case "python":
		dockerfile.WriteString("FROM python:3.11-slim AS builder\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY requirements.txt ./\n")
		
		installCmd := opts.InstallCommand
		if installCmd == "" {
			installCmd = "pip install --no-cache-dir -r requirements.txt"
		}
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", installCmd))
		
		dockerfile.WriteString("COPY . .\n")
		
		buildCmd := opts.BuildCommand
		if buildCmd != "" {
			dockerfile.WriteString(fmt.Sprintf("RUN %s\n", buildCmd))
		}
		
		dockerfile.WriteString("FROM python:3.11-slim\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages\n")
		dockerfile.WriteString("COPY --from=builder /app .\n")
		
		startCmd := opts.StartCommand
		if startCmd == "" {
			startCmd = "python app.py"
		}
		dockerfile.WriteString(fmt.Sprintf("CMD [\"sh\", \"-c\", \"%s\"]\n", startCmd))

	case "php":
		dockerfile.WriteString("FROM php:8.2-fpm-alpine\n")
		dockerfile.WriteString("WORKDIR /var/www/html\n")
		dockerfile.WriteString("COPY composer.json composer.lock ./\n")
		
		installCmd := opts.InstallCommand
		if installCmd == "" {
			installCmd = "composer install --no-dev --optimize-autoloader"
		}
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", installCmd))
		
		dockerfile.WriteString("COPY . .\n")
		
		startCmd := opts.StartCommand
		if startCmd == "" {
			startCmd = "php-fpm"
		}
		dockerfile.WriteString(fmt.Sprintf("CMD [\"%s\"]\n", startCmd))

	case "ruby":
		dockerfile.WriteString("FROM ruby:3.2-alpine\n")
		dockerfile.WriteString("WORKDIR /app\n")
		dockerfile.WriteString("COPY Gemfile Gemfile.lock ./\n")
		
		installCmd := opts.InstallCommand
		if installCmd == "" {
			installCmd = "bundle install"
		}
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", installCmd))
		
		dockerfile.WriteString("COPY . .\n")
		
		startCmd := opts.StartCommand
		if startCmd == "" {
			startCmd = "bundle exec rails server"
		}
		dockerfile.WriteString(fmt.Sprintf("CMD [\"sh\", \"-c\", \"%s\"]\n", startCmd))

	case "static":
		dockerfile.WriteString("FROM caddy:2-alpine\n")
		dockerfile.WriteString("WORKDIR /usr/share/caddy\n")
		dockerfile.WriteString("COPY . .\n")
		dockerfile.WriteString("EXPOSE 80\n")
		dockerfile.WriteString("CMD [\"caddy\", \"file-server\", \"--listen\", \":80\"]\n")

	default:
		return "", fmt.Errorf("unsupported runtime: %s", runtime)
	}

	return dockerfile.String(), nil
}

// BuildWithRailpackCLI builds using Railpack CLI if available
// This is a fallback if we want to use the actual Railpack binary
func (r *RailpackClient) BuildWithRailpackCLI(ctx context.Context, opts RailpackBuildOptions) error {
	// Check if railpack CLI is available
	railpackPath, err := exec.LookPath("railpack")
	if err != nil {
		return fmt.Errorf("railpack CLI not found: %w", err)
	}

	// Build command
	cmd := exec.CommandContext(ctx, railpackPath, "build",
		"--context", opts.ContextPath,
		"--tag", opts.ImageTag,
	)

	// Add optional flags
	if opts.BuildCommand != "" {
		cmd.Args = append(cmd.Args, "--build-command", opts.BuildCommand)
	}
	if opts.StartCommand != "" {
		cmd.Args = append(cmd.Args, "--start-command", opts.StartCommand)
	}

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range opts.EnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Run build
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("railpack build failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

