package caddy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client handles Caddy Admin API interactions
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Caddy client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Route represents a Caddy route configuration
type Route struct {
	Match []MatchRule `json:"match"`
	Handle []Handle   `json:"handle"`
	Terminal bool     `json:"terminal,omitempty"`
}

// MatchRule represents a route match rule
type MatchRule struct {
	Host []string `json:"host"`
}

// Handle represents a route handler
type Handle struct {
	Handler     string                 `json:"handler"`
	Upstreams   []Upstream             `json:"upstreams,omitempty"`
	Transport   *Transport             `json:"transport,omitempty"`
	Routes      []Route                `json:"routes,omitempty"`
	Headers     map[string]interface{} `json:"headers,omitempty"`
}

// Upstream represents an upstream server
type Upstream struct {
	Dial string `json:"dial"`
}

// Transport represents transport configuration
type Transport struct {
	Protocol     string `json:"protocol"`
	TLSSkipVerify bool  `json:"tls_skip_verify,omitempty"`
}

// AddRoute adds a route to Caddy for a custom domain
func (c *Client) AddRoute(ctx context.Context, domain string, targetHost string, targetPort int, enableSSL bool) error {
	// Construct route configuration
	route := Route{
		Match: []MatchRule{
			{
				Host: []string{domain},
			},
		},
		Handle: []Handle{
			{
				Handler: "reverse_proxy",
				Upstreams: []Upstream{
					{
						Dial: fmt.Sprintf("%s:%d", targetHost, targetPort),
					},
				},
				Transport: &Transport{
					Protocol: "http",
				},
			},
		},
		Terminal: true,
	}

	// If SSL is enabled, add SSL handler (Caddy will auto-provision certificates)
	if enableSSL {
		// Caddy handles SSL automatically via automatic HTTPS
		// The route will be served over HTTPS by default
	}

	// POST to Caddy Admin API to add route
	// URL is constructed in setRoutes
	
	// Get existing routes first
	existingRoutes, err := c.getRoutes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get existing routes: %w", err)
	}

	// Add new route
	allRoutes := append(existingRoutes, route)

	// Update routes
	return c.setRoutes(ctx, allRoutes)
}

// RemoveRoute removes a route from Caddy
func (c *Client) RemoveRoute(ctx context.Context, domain string) error {
	// Get existing routes
	routes, err := c.getRoutes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get existing routes: %w", err)
	}

	// Filter out route matching the domain
	filteredRoutes := make([]Route, 0)
	for _, route := range routes {
		shouldKeep := true
		for _, match := range route.Match {
			for _, host := range match.Host {
				if host == domain {
					shouldKeep = false
					break
				}
			}
			if !shouldKeep {
				break
			}
		}
		if shouldKeep {
			filteredRoutes = append(filteredRoutes, route)
		}
	}

	// Update routes
	return c.setRoutes(ctx, filteredRoutes)
}

// UpdateRoute updates an existing route
func (c *Client) UpdateRoute(ctx context.Context, domain string, targetHost string, targetPort int) error {
	// Remove old route
	if err := c.RemoveRoute(ctx, domain); err != nil {
		return fmt.Errorf("failed to remove old route: %w", err)
	}

	// Add new route
	return c.AddRoute(ctx, domain, targetHost, targetPort, true)
}

// getRoutes gets all routes from Caddy
func (c *Client) getRoutes(ctx context.Context) ([]Route, error) {
	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("caddy API returned status %d", resp.StatusCode)
	}

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		// If no routes exist, return empty array
		if resp.StatusCode == http.StatusNotFound {
			return []Route{}, nil
		}
		return nil, err
	}

	return routes, nil
}

// setRoutes sets routes in Caddy
func (c *Client) setRoutes(ctx context.Context, routes []Route) error {
	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", c.baseURL)

	body, err := json.Marshal(routes)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("caddy API returned status %d", resp.StatusCode)
	}

	return nil
}

// ValidateDomain validates if a CNAME record exists for a domain
func (c *Client) ValidateDomain(ctx context.Context, domain string, expectedTarget string) (bool, error) {
	// This would typically use DNS lookup to check CNAME
	// For now, we'll return true (assume valid if user adds it)
	// In production, implement actual DNS lookup
	return true, nil
}

