package k8s

import (
	"context"
	"fmt"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressSpec defines the specification for a Kubernetes Ingress
type IngressSpec struct {
	ServiceID     string
	ServiceName   string
	ProjectID     string
	Environment   string // e.g., "prod", "staging"
	Port          int32
	CustomDomains []string // Custom domains to add
}

// CreateIngress creates a Kubernetes Ingress for a service
func (c *Client) CreateIngress(ctx context.Context, spec IngressSpec) (*networkingv1.Ingress, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	ingressName := c.ingressName(spec.ServiceID)
	serviceName := c.serviceName(spec.ServiceID)

	// Generate default hostname
	defaultHost := c.generateDefaultHost(spec.ServiceName, spec.Environment)

	// Build hosts list (default + custom)
	hosts := []string{defaultHost}
	hosts = append(hosts, spec.CustomDomains...)

	// Build ingress rules
	rules := make([]networkingv1.IngressRule, 0, len(hosts))
	pathType := networkingv1.PathTypePrefix

	for _, host := range hosts {
		rules = append(rules, networkingv1.IngressRule{
			Host: host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: serviceName,
									Port: networkingv1.ServiceBackendPort{
										Number: spec.Port,
									},
								},
							},
						},
					},
				},
			},
		})
	}

	// Build TLS configuration for all hosts
	tls := []networkingv1.IngressTLS{
		{
			Hosts:      hosts,
			SecretName: ingressName + "-tls",
		},
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: namespace,
			Labels:    c.buildLabels(spec.ServiceID, spec.ServiceName, spec.ProjectID),
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":              c.config.IngressClass,
				"cert-manager.io/cluster-issuer":           c.config.CertIssuer,
				"traefik.ingress.kubernetes.io/router.tls": "true",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &c.config.IngressClass,
			Rules:            rules,
			TLS:              tls,
		},
	}

	result, err := c.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create ingress: %w", err)
	}

	return result, nil
}

// UpdateIngress updates an existing Ingress (e.g., to add custom domains)
func (c *Client) UpdateIngress(ctx context.Context, spec IngressSpec) (*networkingv1.Ingress, error) {
	namespace := c.ProjectNamespace(spec.ProjectID)
	ingressName := c.ingressName(spec.ServiceID)
	serviceName := c.serviceName(spec.ServiceID)

	existing, err := c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	// Generate default hostname
	defaultHost := c.generateDefaultHost(spec.ServiceName, spec.Environment)

	// Build hosts list
	hosts := []string{defaultHost}
	hosts = append(hosts, spec.CustomDomains...)

	// Rebuild rules
	rules := make([]networkingv1.IngressRule, 0, len(hosts))
	pathType := networkingv1.PathTypePrefix

	for _, host := range hosts {
		rules = append(rules, networkingv1.IngressRule{
			Host: host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: serviceName,
									Port: networkingv1.ServiceBackendPort{
										Number: spec.Port,
									},
								},
							},
						},
					},
				},
			},
		})
	}

	// Update TLS
	existing.Spec.Rules = rules
	existing.Spec.TLS = []networkingv1.IngressTLS{
		{
			Hosts:      hosts,
			SecretName: ingressName + "-tls",
		},
	}

	result, err := c.clientset.NetworkingV1().Ingresses(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update ingress: %w", err)
	}

	return result, nil
}

// AddCustomDomain adds a custom domain to an existing ingress
func (c *Client) AddCustomDomain(ctx context.Context, projectID, serviceID, domain string) error {
	namespace := c.ProjectNamespace(projectID)
	ingressName := c.ingressName(serviceID)

	existing, err := c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %w", err)
	}

	// Check if domain already exists
	for _, rule := range existing.Spec.Rules {
		if rule.Host == domain {
			return nil // Domain already exists
		}
	}

	// Add new rule for the custom domain
	serviceName := c.serviceName(serviceID)
	pathType := networkingv1.PathTypePrefix
	port := existing.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number

	existing.Spec.Rules = append(existing.Spec.Rules, networkingv1.IngressRule{
		Host: domain,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{
					{
						Path:     "/",
						PathType: &pathType,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: serviceName,
								Port: networkingv1.ServiceBackendPort{
									Number: port,
								},
							},
						},
					},
				},
			},
		},
	})

	// Update TLS to include new domain
	if len(existing.Spec.TLS) > 0 {
		existing.Spec.TLS[0].Hosts = append(existing.Spec.TLS[0].Hosts, domain)
	}

	_, err = c.clientset.NetworkingV1().Ingresses(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress: %w", err)
	}

	return nil
}

// RemoveCustomDomain removes a custom domain from an ingress
func (c *Client) RemoveCustomDomain(ctx context.Context, projectID, serviceID, domain string) error {
	namespace := c.ProjectNamespace(projectID)
	ingressName := c.ingressName(serviceID)

	existing, err := c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %w", err)
	}

	// Remove rule for the domain
	newRules := make([]networkingv1.IngressRule, 0)
	for _, rule := range existing.Spec.Rules {
		if rule.Host != domain {
			newRules = append(newRules, rule)
		}
	}
	existing.Spec.Rules = newRules

	// Remove from TLS hosts
	if len(existing.Spec.TLS) > 0 {
		newHosts := make([]string, 0)
		for _, host := range existing.Spec.TLS[0].Hosts {
			if host != domain {
				newHosts = append(newHosts, host)
			}
		}
		existing.Spec.TLS[0].Hosts = newHosts
	}

	_, err = c.clientset.NetworkingV1().Ingresses(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress: %w", err)
	}

	return nil
}

// DeleteIngress deletes a Kubernetes Ingress
func (c *Client) DeleteIngress(ctx context.Context, projectID, serviceID string) error {
	namespace := c.ProjectNamespace(projectID)
	ingressName := c.ingressName(serviceID)

	err := c.clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, ingressName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete ingress: %w", err)
	}

	return nil
}

// GetIngress retrieves a Kubernetes Ingress
func (c *Client) GetIngress(ctx context.Context, projectID, serviceID string) (*networkingv1.Ingress, error) {
	namespace := c.ProjectNamespace(projectID)
	ingressName := c.ingressName(serviceID)

	return c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
}

// GetServiceURL returns the default URL for a service
func (c *Client) GetServiceURL(serviceName, environment string) string {
	return "https://" + c.generateDefaultHost(serviceName, environment)
}

// GetCNAMETarget returns the CNAME target for custom domains
func (c *Client) GetCNAMETarget() string {
	// Return the ingress controller's domain
	return "ingress." + c.config.BaseDomain
}

func (c *Client) ingressName(serviceID string) string {
	return "ing-" + serviceID[:8]
}

func (c *Client) generateDefaultHost(serviceName, environment string) string {
	// Format: servicename-environment.up.zyndra.app
	name := strings.ToLower(serviceName)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	
	if environment == "" {
		environment = "prod"
	}
	
	return fmt.Sprintf("%s-%s.%s", name, environment, c.config.BaseDomain)
}

