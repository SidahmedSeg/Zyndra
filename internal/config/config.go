package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Server
	Port string `envconfig:"PORT" default:"8080"`

	// Database
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// Casdoor (optional if DISABLE_AUTH=true)
	CasdoorEndpoint     string `envconfig:"CASDOOR_ENDPOINT"`
	CasdoorClientID     string `envconfig:"CASDOOR_CLIENT_ID"`
	CasdoorClientSecret string `envconfig:"CASDOOR_CLIENT_SECRET"`

	// OpenStack
	InfraServiceURL    string `envconfig:"INFRA_SERVICE_URL"`
	InfraServiceAPIKey string `envconfig:"INFRA_SERVICE_API_KEY"`
	UseMockInfra       bool   `envconfig:"USE_MOCK_INFRA" default:"true"` // Use mock OpenStack client

	// Registry
	RegistryURL      string `envconfig:"REGISTRY_URL" required:"true"`
	RegistryUsername string `envconfig:"REGISTRY_USERNAME" required:"true"`
	RegistryPassword string `envconfig:"REGISTRY_PASSWORD" required:"true"`

	// GitHub OAuth
	GitHubClientID     string `envconfig:"GITHUB_CLIENT_ID"`
	GitHubClientSecret string `envconfig:"GITHUB_CLIENT_SECRET"`
	GitHubRedirectURL  string `envconfig:"GITHUB_REDIRECT_URL" default:"http://localhost:8080/api/git/callback/github"`

	// GitLab OAuth
	GitLabClientID     string `envconfig:"GITLAB_CLIENT_ID"`
	GitLabClientSecret string `envconfig:"GITLAB_CLIENT_SECRET"`
	GitLabRedirectURL  string `envconfig:"GITLAB_REDIRECT_URL" default:"http://localhost:8080/api/git/callback/gitlab"`
	GitLabBaseURL      string `envconfig:"GITLAB_BASE_URL"` // Optional, for self-hosted GitLab

	// Webhook
	WebhookSecret string `envconfig:"WEBHOOK_SECRET" required:"true"`
	BaseURL       string `envconfig:"BASE_URL" default:"http://localhost:8080"`

	// BuildKit
	BuildKitAddress string `envconfig:"BUILDKIT_ADDRESS" default:"unix:///run/buildkit/buildkitd.sock"`
	BuildDir        string `envconfig:"BUILD_DIR" default:"/tmp/click-deploy-builds"`

	// DNS (for database internal hostnames)
	DNSZoneID string `envconfig:"DNS_ZONE_ID"` // OpenStack Designate zone ID

	// Caddy
	CaddyAdminURL string `envconfig:"CADDY_ADMIN_URL" default:"http://localhost:2019"`

	// Prometheus
	PrometheusURL        string `envconfig:"PROMETHEUS_URL" default:"http://localhost:9090"`
	PrometheusTargetsDir string `envconfig:"PROMETHEUS_TARGETS_DIR" default:"/tmp/prometheus-targets"`

	// Performance
	DBMaxOpenConns    int `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	DBMaxIdleConns    int `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	DBConnMaxLifetime int `envconfig:"DB_CONN_MAX_LIFETIME" default:"300"` // seconds

	// Security
	RateLimitRequests int `envconfig:"RATE_LIMIT_REQUESTS" default:"100"` // requests per window
	RateLimitWindow   int `envconfig:"RATE_LIMIT_WINDOW" default:"60"`     // window in seconds

	// Centrifugo
	CentrifugoWSURL            string `envconfig:"CENTRIFUGO_WS_URL"`              // e.g. wss://centrifugo.example.com/connection/websocket
	CentrifugoAPIURL           string `envconfig:"CENTRIFUGO_API_URL"`             // e.g. http://centrifugo:8000/api
	CentrifugoAPIKey           string `envconfig:"CENTRIFUGO_API_KEY"`             // HTTP API key
	CentrifugoTokenHMACSecret  string `envconfig:"CENTRIFUGO_TOKEN_HMAC_SECRET"`   // JWT HMAC secret

	// CORS
	CORSOrigins string `envconfig:"CORS_ORIGINS" default:"*"` // Comma-separated list of allowed origins

	// Development
	DisableAuth bool `envconfig:"DISABLE_AUTH" default:"true"` // Use mock auth for development (set to false for Casdoor)
}

func Load() (*Config, error) {
	// Load .env file (optional, for local development)
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

